package platform

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Logger() *slog.Logger { return slog.New(slog.NewJSONHandler(os.Stdout, nil)) }
func Error(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{"error": gin.H{"code": code, "message": message}, "request_id": c.GetString("request_id")})
}
func validRequestID(id string) bool {
	if len(id) == 0 || len(id) > 128 {
		return false
	}
	for _, r := range id {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_') {
			return false
		}
	}
	return true
}
func Middleware(log *slog.Logger, service string) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if !validRequestID(id) {
			var bytes [16]byte
			if _, err := rand.Read(bytes[:]); err != nil {
				panic(err)
			}
			id = hex.EncodeToString(bytes[:])
		}
		c.Set("request_id", id)
		c.Request.Header.Set("X-Request-ID", id)
		c.Header("X-Request-ID", id)
		start := time.Now()
		c.Next()
		log.Info("http_request", "service", service, "request_id", id, "method", c.Request.Method, "path", c.Request.URL.Path, "status", c.Writer.Status(), "duration_ms", time.Since(start).Milliseconds())
	}
}

// Run owns infrastructure only; domain code belongs to its service.
func Run(name string, database bool, register func(*gin.Engine) error) error {
	return RunWithDatabase(name, database, func(r *gin.Engine, _ *pgxpool.Pool) error {
		if register == nil {
			return nil
		}
		return register(r)
	})
}

// RunWithDatabase exposes the service's pool to domain handlers.
func RunWithDatabase(name string, database bool, register func(*gin.Engine, *pgxpool.Pool) error) error {
	log := Logger()
	cfg, err := Load(name, database)
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	var pool *pgxpool.Pool
	if database {
		pc, err := pgxpool.ParseConfig(cfg.DatabaseURL)
		if err != nil {
			return err
		}
		pc.MaxConns = cfg.MaxConnections
		pc.MinConns = 1
		pc.MaxConnLifetime = 30 * time.Minute
		pool, err = pgxpool.NewWithConfig(ctx, pc)
		if err != nil {
			return err
		}
		defer pool.Close()
		check, cancel := context.WithTimeout(ctx, 5*time.Second)
		err = pool.Ping(check)
		cancel()
		if err != nil {
			return err
		}
	}
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(Middleware(log, name))
	r.Use(gin.CustomRecovery(func(c *gin.Context, recovered any) {
		log.Error("request_panic", "service", name, "request_id", c.GetString("request_id"))
		Error(c, 500, "INTERNAL_ERROR", "Internal server error")
		c.Abort()
	}))
	r.GET("/healthz", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok", "service": name}) })
	r.GET("/readyz", func(c *gin.Context) {
		if ctx.Err() != nil {
			Error(c, 503, "NOT_READY", "Service is shutting down")
			return
		}
		if pool != nil {
			check, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
			defer cancel()
			if err := pool.Ping(check); err != nil {
				Error(c, 503, "NOT_READY", "Database is unavailable")
				return
			}
		}
		c.JSON(200, gin.H{"status": "ready", "service": name})
	})
	r.NoRoute(func(c *gin.Context) { Error(c, 404, "NOT_FOUND", "Endpoint is not implemented") })
	if register != nil {
		if err := register(r, pool); err != nil {
			return err
		}
	}
	server := &http.Server{Addr: cfg.Address, Handler: r, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 20 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 1 << 20}
	done := make(chan error, 1)
	go func() { done <- server.ListenAndServe() }()
	log.Info("service_started", "service", name, "address", cfg.Address)
	select {
	case err := <-done:
		if err != http.ErrServerClosed {
			return err
		}
	case <-ctx.Done():
		log.Info("service_stopping", "service", name)
		shutdown, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		if err := server.Shutdown(shutdown); err != nil {
			_ = server.Close()
			return err
		}
	}
	return nil
}
