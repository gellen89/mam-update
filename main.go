package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/gellen89/mam-update/internal/app"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Initialize the application
	application, err := app.New(os.Args[1:])
	if err != nil {
		slog.Error(fmt.Sprintf("failed to initialize application: %v", err))
		os.Exit(1)
	}

	// Run the application
	if err := application.Run(ctx); err != nil {
		slog.Error(fmt.Sprintf("failed to run application: %v", err))
		os.Exit(1)
	}
}
