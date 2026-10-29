package transport

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"supay.local/platform/internal/authn"
)

func TestPINIsVerifiedForEachWriteButNotReadsOrCategories(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := []byte("test-commerce-secret-at-least-32-bytes")
	token, _ := authn.Sign(authn.Claims{Subject: "user-1", Role: "MERCHANT", Expires: time.Now().Add(time.Minute).Unix()}, secret)
	calls := 0
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Header.Get("Authorization") != "Bearer "+token {
			t.Error("caller token not forwarded")
		}
		var body struct {
			PIN string `json:"pin"`
		}
		if json.NewDecoder(r.Body).Decode(&body) != nil {
			t.Error("bad PIN verification request")
		}
		if body.PIN != "012345" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(401)
			w.Write([]byte(`{"error":{"code":"INVALID_PIN","message":"Incorrect PIN"}}`))
			return
		}
		w.WriteHeader(204)
	}))
	defer upstream.Close()
	a := &Catalog{secret: secret, pinURL: upstream.URL, pinClient: upstream.Client()}
	r := gin.New()
	g := r.Group("/api/v1", a.authenticate, a.confirmActionPIN)
	for _, path := range []string{"/products", "/orders", "/categories"} {
		g.Any(path, func(c *gin.Context) { c.Status(204) })
	}
	check := func(method, path, pin string, want int) {
		t.Helper()
		req := httptest.NewRequest(method, path, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		if pin != "" {
			req.Header.Set("X-PIN", pin)
		}
		out := httptest.NewRecorder()
		r.ServeHTTP(out, req)
		if out.Code != want {
			t.Fatalf("%s %s got %d want %d: %s", method, path, out.Code, want, out.Body.String())
		}
	}
	check("GET", "/api/v1/products", "", 204)
	check("POST", "/api/v1/categories", "", 204)
	check("POST", "/api/v1/products", "", 403)
	check("POST", "/api/v1/orders", "", 403)
	if calls != 0 {
		t.Fatal("PIN service called on reads/categories/missing PIN")
	}
	check("POST", "/api/v1/products", "012345", 204)
	check("POST", "/api/v1/products", "", 403)
	check("POST", "/api/v1/orders", "012345", 204)
	check("POST", "/api/v1/products", "999999", 401)
	if calls != 3 {
		t.Fatalf("PIN result was cached or checked more than once: %d calls", calls)
	}
}

func TestPINServiceFailureBlocksMutation(t *testing.T) {
	secret := []byte("test-commerce-secret-at-least-32-bytes")
	token, _ := authn.Sign(authn.Claims{Subject: "user", Role: "MERCHANT", Expires: time.Now().Add(time.Minute).Unix()}, secret)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(500); w.Write([]byte("unavailable")) }))
	defer upstream.Close()
	a := &Catalog{secret: secret, pinURL: upstream.URL, pinClient: upstream.Client()}
	r := gin.New()
	r.POST("/api/v1/products", a.authenticate, a.confirmActionPIN, func(c *gin.Context) { t.Error("mutation executed during verifier outage"); c.Status(204) })
	req := httptest.NewRequest("POST", "/api/v1/products", strings.NewReader(`{}`))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-PIN", "012345")
	out := httptest.NewRecorder()
	r.ServeHTTP(out, req)
	if out.Code != 503 {
		t.Fatalf("got %d want 503", out.Code)
	}
}
