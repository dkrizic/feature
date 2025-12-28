package service

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/urfave/cli/v3"
)

func Service(ctx context.Context, cmd *cli.Command) error {
	port := cmd.Int("port")
	slog.Info("Starting the feature service", "port", port)
	
	cancelChan := make(chan os.Signal, 1)
	// catch SIGETRM or SIGINTERRUPT
	signal.Notify(cancelChan, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		for {
			slog.Info("Service running...")
			time.Sleep(1 * time.Second)
		}
	}()
	sig := <-cancelChan
	slog.Info("Shutting down feature service", "signal", sig)
	return nil
}
