package platform

import (
	"github.com/gin-gonic/gin"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestCorrelation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, id := range []string{"order-123", "", strings.Repeat("x", 129), "invalid value"} {
		t.Run(id, func(t *testing.T) {
			r := gin.New()
			r.Use(Middleware(slog.New(slog.NewTextHandler(io.Discard, nil)), "test"))
			r.GET("/", func(c *gin.Context) {
				if c.Request.Header.Get("X-Request-ID") != c.GetString("request_id") {
					t.Error("downstream correlation differs")
				}
				Error(c, 409, "CONFLICT", "Conflict")
			})
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("X-Request-ID", id)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			got := w.Header().Get("X-Request-ID")
			if !validRequestID(got) {
				t.Fatalf("invalid generated request ID: %q", got)
			}
			if id == "order-123" && got != id {
				t.Error("valid correlation ID was replaced")
			}
			if w.Code != 409 || !strings.Contains(w.Body.String(), got) {
				t.Error("error response lost correlation")
			}
		})
	}
}
func TestInvalidConfigurationFailsEarly(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	if _, err := Load("wallet-service", true); err == nil {
		t.Fatal("missing DB URL accepted")
	}
	for _, n := range []string{"0", "-1", "invalid", "101"} {
		t.Setenv("DB_MAX_CONNECTIONS", n)
		if _, err := Load("gateway", false); err == nil {
			t.Fatalf("invalid pool size accepted: %s", n)
		}
	}
}
