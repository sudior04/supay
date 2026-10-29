package main

import (
	"os"
	"supay.local/platform/internal/platform"
	"supay.local/platform/services/wallet-service/internal/transport"
)

func main() {
	if err := platform.RunWithDatabase("wallet-service", true, transport.Register); err != nil {
		platform.Logger().Error("service_failed", "error", err)
		os.Exit(1)
	}
}
