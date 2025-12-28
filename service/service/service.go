package service

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dkrizic/feature/service/constant"
	persitence "github.com/dkrizic/feature/service/service/persistence"
	"github.com/dkrizic/feature/service/service/persistence/configmap"
	"github.com/dkrizic/feature/service/service/persistence/inmemory"
	"github.com/urfave/cli/v3"
)

func Service(ctx context.Context, cmd *cli.Command) error {
	// get the port
	port := cmd.Int("port")
	slog.Info("Starting the feature service", "port", port)

	// configure persistence based on storage type
	var persistence persitence.Persistence
	storageType := cmd.String(constant.StorageType)
	switch storageType {
	case constant.StorageTypeInMemory:
		slog.Info("Using in-memory storage")
		persistence = inmemory.NewPersistence()
	case constant.StorageTypeConfigMap:
		configMapName := cmd.String(constant.ConfigMapName)
		slog.Info("Using ConfigMap storage", "configMapName", configMapName)
		persistence = configmap.NewPersistence(configMapName)
	default:
		slog.Error("Invalid storage type", "storageType", storageType)
		return fmt.Errorf("invalid storage type: %s", storageType)
	}

	_ = persistence

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
