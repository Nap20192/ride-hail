package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"ride-hail/internal/runner"
	"ride-hail/internal/shared/config"
)

func main() {
	ctx := context.Background()

	cfg, err := config.LoadConfig()

	if err != nil {
		fmt.Println("Failed to load config:", err)
		os.Exit(1)
	}

	err = runner.Run(ctx, *cfg)

	if err != nil {
		slog.Error("Application error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
