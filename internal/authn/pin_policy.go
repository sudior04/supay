package authn

import (
	"net/http"
	"strings"
)

// PIN confirms each catalog/order write, never login, reads, categories or wallets.
func RequiresActionPIN(method, path string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
	default:
		return false
	}
	for _, prefix := range []string{"/api/v1/products", "/api/v1/orders"} {
		if path == prefix || strings.HasPrefix(path, prefix+"/") {
			return true
		}
	}
	return false
}
