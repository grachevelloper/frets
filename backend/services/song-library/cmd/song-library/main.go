package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/grachevelloper/frets/backend/services/song-library/internal/app"
)

func main() {
	if err := run(); err != nil {
		slog.Error("song-library stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	application, err := app.New(ctx)
	if err != nil {
		return err
	}
	defer application.Close()
	return application.Run(ctx)
}
