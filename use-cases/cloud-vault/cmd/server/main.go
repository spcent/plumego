package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"cloud-vault/internal/app"
	"cloud-vault/internal/backup"
	"cloud-vault/internal/config"
)

var version = "dev"

func main() {
	// Check for CLI subcommands
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "restore":
			if err := runRestore(); err != nil {
				log.Printf("restore failed: %v", err)
				os.Exit(1)
			}
			return
		case "help", "-h", "--help":
			printUsage()
			return
		}
	}

	if err := run(); err != nil {
		log.Printf("server stopped: %v", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Printf(`Cloud Vault %s

Usage:
  cloud-vault                          Start the server
  cloud-vault restore --file <zip>     Restore from backup
  cloud-vault help                     Show this help

`, version)
}

func runRestore() error {
	var file, dataDir string
	for i := 2; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--file", "-f":
			if i+1 < len(os.Args) {
				file = os.Args[i+1]
				i++
			}
		case "--data-dir", "-d":
			if i+1 < len(os.Args) {
				dataDir = os.Args[i+1]
				i++
			}
		}
	}

	if file == "" {
		return fmt.Errorf("required: --file <backup.zip>")
	}
	if dataDir == "" {
		dataDir = "./data"
	}

	return backup.RestoreCLI(file, dataDir)
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

	return a.Start(ctx)
}
