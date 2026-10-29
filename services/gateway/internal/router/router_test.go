package router

import (
	"github.com/gin-gonic/gin"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"supay.local/platform/internal/authn"
	"supay.local/platform/internal/platform"
	"testing"
	"time"
)

func TestRoutingPreservesContractAndRemovesSpoofedIdentity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var token string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RequestURI() != "/api/v1/orders/order-1?include=items" {
			t.Errorf("path/query changed: %s", r.URL.RequestURI())
		}
		if r.Header.Get("X-Request-ID") != "payment-123" {
			t.Error("request ID lost")
		}
		if r.Header.Get("Authorization") != "Bearer "+token {
			t.Error("authorization lost")
		}
		if r.Header.Get("X-User-ID") != "user-123" || r.Header.Get("X-User-Role") != "CUSTOMER" || r.Header.Get("X-Merchant-ID") != "" {
			t.Error("trusted identity mismatch or spoofed merchant forwarded")
		}
		w.WriteHeader(202)
	}))
	defer upstream.Close()
	t.Setenv("COMMERCE_SERVICE_URL", upstream.URL)
	t.Setenv("JWT_SECRET", "test-jwt-secret-that-is-at-least-32-bytes-long")
	signed, err := authn.Sign(authn.Claims{Subject: "user-123", Email: "test@example.com", Role: "CUSTOMER", Issued: time.Now().Unix(), Expires: time.Now().Add(time.Minute).Unix()}, []byte("test-jwt-secret-that-is-at-least-32-bytes-long"))
	if err != nil {
		t.Fatal(err)
	}
	token = signed
	r := gin.New()
	r.Use(platform.Middleware(slog.New(slog.NewTextHandler(io.Discard, nil)), "gateway"))
	if err := Register(r); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/orders/order-1?include=items", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Request-ID", "payment-123")
	req.Header.Set("X-User-ID", "attacker")
	req.Header.Set("X-User-Role", "ADMIN")
	req.Header.Set("X-Merchant-ID", "other")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 202 {
		t.Fatalf("unexpected upstream status: %d", w.Code)
	}
}
func TestInvalidUpstreamFailsRegistration(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-jwt-secret-that-is-at-least-32-bytes-long")
	t.Setenv("WALLET_SERVICE_URL", "file:///tmp/wallet")
	if err := Register(gin.New()); err == nil {
		t.Fatal("invalid upstream accepted")
	}
}
