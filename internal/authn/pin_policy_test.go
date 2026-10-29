package authn

import "testing"

func TestPINOnlyProtectsProductAndOrderWrites(t *testing.T) {
	for _, path := range []string{"/api/v1/products", "/api/v1/products/123/variants/456/inventory", "/api/v1/orders", "/api/v1/orders/123"} {
		for _, method := range []string{"POST", "PUT", "PATCH", "DELETE"} {
			if !RequiresActionPIN(method, path) {
				t.Errorf("missing PIN policy for %s %s", method, path)
			}
		}
		for _, method := range []string{"GET", "HEAD", "OPTIONS"} {
			if RequiresActionPIN(method, path) {
				t.Errorf("read requires PIN: %s %s", method, path)
			}
		}
	}
	for _, path := range []string{"/api/v1/auth/login", "/api/v1/auth/register", "/api/v1/auth/refresh", "/api/v1/categories", "/api/v1/categories/123", "/api/v1/wallets", "/api/v1/users/me/password", "/api/v1/products-archive", "/api/v1/orders-report"} {
		if RequiresActionPIN("POST", path) {
			t.Errorf("unexpected PIN policy for %s", path)
		}
	}
}
