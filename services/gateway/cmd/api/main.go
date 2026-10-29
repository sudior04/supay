package main

import (
	"os"
	"supay.local/platform/internal/platform"
	"supay.local/platform/services/gateway/internal/router"
)

func main() {
	if err := platform.Run("gateway", false, router.Register); err != nil {
		platform.Logger().Error("service_failed", "error", err)
		os.Exit(1)
	}
}
