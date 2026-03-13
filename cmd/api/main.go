package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/call-notes-ai-service/internal/boot"
	"github.com/call-notes-ai-service/internal/logger"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app, err := boot.Initialize(ctx)
	if err != nil {
		panic("failed to initialize application: " + err.Error())
	}
	defer logger.Sync()

	app.Start()
	waitForShutdown(ctx, app, cancel)
}

func waitForShutdown(ctx context.Context, app *boot.App, cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	<-sigChan

	app.Shutdown(ctx)
	cancel()
}
