package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/gellen89/mam-update/internal/app"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	err := app.Run(ctx, os.Args[1:])
	if err != nil {
		panic(err)
	}
}
