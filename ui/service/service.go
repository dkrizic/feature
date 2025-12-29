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
	// get the port
	port := cmd.Int("port")
	slog.InfoContext(ctx, "Starting the feature UI service", "port", port)

	cancelChan := make(chan os.Signal, 1)
	// catch SIGETRM or SIGINTERRUPT
	signal.Notify(cancelChan, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		slog.Info("Running")
		time.Sleep(5 * time.Second)
	}()
	sig := <-cancelChan
	slog.Info("Shutting down feature UI service", "signal", sig)
	return nil

}
