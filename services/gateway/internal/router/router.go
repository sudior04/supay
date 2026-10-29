package router

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"supay.local/platform/internal/authn"
	"supay.local/platform/internal/platform"
	"time"
)

func Register(r *gin.Engine) error {
	secret := []byte(os.Getenv("JWT_SECRET"))
	if len(secret) < 32 {
		return fmt.Errorf("JWT_SECRET must contain at least 32 bytes")
	}
	groups := []struct {
		env, fallback string
		prefixes      []string
	}{
		{"USER_SERVICE_URL", "http://localhost:8081", []string{"auth", "users", "merchants"}},
		{"COMMERCE_SERVICE_URL", "http://localhost:8082", []string{"products", "categories", "orders", "inventory"}},
		{"WALLET_SERVICE_URL", "http://localhost:8083", []string{"wallets", "payments", "transfers", "transactions", "refunds"}},
	}
	for _, group := range groups {
		target, err := url.Parse(platform.Env(group.env, group.fallback))
		if err != nil || target.Host == "" || (target.Scheme != "http" && target.Scheme != "https") {
			return fmt.Errorf("invalid upstream %s", group.env)
		}
		proxy := httputil.NewSingleHostReverseProxy(target)
		original := proxy.Director
		proxy.Director = func(req *http.Request) {
			original(req)
			for _, header := range []string{"X-User-ID", "X-User-Role", "X-Merchant-ID"} {
				req.Header.Del(header)
			}
			if identity, ok := authn.FromContext(req.Context()); ok {
				req.Header.Set("X-User-ID", identity.ID)
				req.Header.Set("X-User-Role", identity.Role)
			}
		}
		proxy.Transport = &http.Transport{Proxy: http.ProxyFromEnvironment, MaxIdleConns: 100, MaxIdleConnsPerHost: 10, IdleConnTimeout: 90 * time.Second, ResponseHeaderTimeout: 8 * time.Second}
		proxy.ErrorHandler = func(w http.ResponseWriter, req *http.Request, err error) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadGateway)
			_, _ = fmt.Fprintf(w, "{\"error\":{\"code\":\"UPSTREAM_UNAVAILABLE\",\"message\":\"Upstream is unavailable\"},\"request_id\":%q}", req.Header.Get("X-Request-ID"))
		}
		handler := func(c *gin.Context) {
			path := c.Request.URL.Path
			public := strings.HasPrefix(path, "/api/v1/auth/") || path == "/api/v1/users/status" || path == "/api/v1/products/status" || path == "/api/v1/wallets/status"
			if !public {
				parts := strings.Fields(c.GetHeader("Authorization"))
				if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
					platform.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "A valid access token is required")
					return
				}
				claims, err := authn.Verify(parts[1], secret)
				if err != nil {
					platform.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "A valid access token is required")
					return
				}
				c.Request = c.Request.WithContext(authn.WithIdentity(c.Request.Context(), claims.Subject, claims.Role))
			}
			ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
			defer cancel()
			proxy.ServeHTTP(c.Writer, c.Request.WithContext(ctx))
		}
		for _, prefix := range group.prefixes {
			r.Any("/api/v1/"+prefix, handler)
			r.Any("/api/v1/"+prefix+"/*path", handler)
		}
	}
	return nil
}
