package main

import (
	"os"
	"supay.local/platform/internal/platform"
	"supay.local/platform/services/user-service/internal/transport"
)

func main() {
	if err := platform.RunWithDatabase("user-service", true, transport.Register); err != nil {
		platform.Logger().Error("service_failed", "error", err)
		os.Exit(1)
	}
}
