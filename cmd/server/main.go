package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"goph-keeper/internal/app"
)

func main() {
	root, err := app.New()
	if err != nil {
		slog.Default().Error("app init failed", slog.Any("err", err))
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := root.Run(ctx); err != nil {
		root.Logger().Error("app stopped with error", slog.Any("err", err))
		os.Exit(1)
	}
}
