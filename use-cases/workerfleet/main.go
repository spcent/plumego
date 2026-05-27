package main

import (
	"context"
	"fmt"
	"log"
	"os"

	workerapp "workerfleet/internal/app"
	"workerfleet/internal/handler"
)

func main() {
	if err := run(context.Background(), os.LookupEnv); err != nil {
		log.Fatalf("workerfleet stopped: %v", err)
	}
}

func run(ctx context.Context, lookup func(string) (string, bool)) error {
	appCfg, err := workerapp.LoadConfig(lookup)
	if err != nil {
		return fmt.Errorf("load app config: %w", err)
	}
	serverCfg, err := workerapp.LoadServerConfig(lookup)
	if err != nil {
		return fmt.Errorf("load server config: %w", err)
	}

	app, err := workerapp.New(ctx, appCfg, serverCfg, handler.RegisterServiceRoutes)
	if err != nil {
		return fmt.Errorf("construct workerfleet app: %w", err)
	}
	return app.Run(ctx)
}
