package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"dbadmin/internal/app"
	"dbadmin/internal/config"
	"dbadmin/web"
)

var version = "dev"

func main() {
	if err := run(); err != nil {
		log.Printf("server stopped: %v", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	cfg.App.Version = version

	a, err := app.New(cfg)
	if err != nil {
		return err
	}

	if err := a.RegisterRoutes(); err != nil {
		return err
	}

	printStartupBanner(cfg.Core.Addr, web.Available)

	return a.Start(ctx)
}

// printStartupBanner displays access address and security notice on startup.
func printStartupBanner(addr string, embedded bool) {
	scheme := "http"
	if os.Getenv("APP_TLS_ENABLED") == "true" {
		scheme = "https"
	}

	// Build display URL — normalize :port to 127.0.0.1:port for display.
	displayAddr := addr
	if strings.HasPrefix(addr, ":") {
		displayAddr = "127.0.0.1" + addr
	}
	url := fmt.Sprintf("%s://%s", scheme, displayAddr)

	assetMode := "disk (development)"
	if embedded {
		assetMode = "embedded (production)"
	}

	fmt.Println()
	fmt.Println("┌─────────────────────────────────────────────────────┐")
	fmt.Println("│         Developer Data Workbench                    │")
	fmt.Println("├─────────────────────────────────────────────────────┤")
	fmt.Printf("│  Access:  %-42s│\n", url)
	fmt.Printf("│  Version: %-42s│\n", version)
	fmt.Printf("│  Assets:  %-42s│\n", assetMode)
	fmt.Println("├─────────────────────────────────────────────────────┤")
	fmt.Println("│  ⚠ This is a local developer tool.                  │")
	fmt.Println("│  ⚠ Do not expose directly to the public internet.   │")
	fmt.Println("│  ⚠ Use a reverse proxy with TLS for production.     │")
	fmt.Println("└─────────────────────────────────────────────────────┘")
	fmt.Println()

	// Warn if binding to non-localhost address.
	host := addr
	if idx := strings.LastIndex(addr, ":"); idx > 0 {
		host = addr[:idx]
	}
	if host != "" && host != "127.0.0.1" && host != "localhost" && host != "::1" {
		fmt.Println("⚠ WARNING: Binding to non-localhost address:", addr)
		fmt.Println("  This exposes the service to the network. Ensure firewall rules are in place.")
		fmt.Println("  For local development, use 127.0.0.1 (the default).")
		fmt.Println()
	}
}
