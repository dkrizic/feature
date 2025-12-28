package service

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/dkrizic/feature/service/constant"
	persitence "github.com/dkrizic/feature/service/service/persistence"
	"github.com/dkrizic/feature/service/service/persistence/configmap"
	"github.com/dkrizic/feature/service/service/persistence/inmemory"
	"github.com/urfave/cli/v3"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/stats"

	"github.com/dkrizic/feature/service/service/meta"
	"github.com/dkrizic/feature/service/service/meta/metav1"
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

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// create gRPC server
	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler(otelgrpc.WithFilter(
			func(stats *stats.RPCTagInfo) bool {
				// slog.Info("Called filter", "method", stats.FullMethodName)
				if strings.Contains(stats.FullMethodName, "grpc.health") {
					return false
				}
				return true
			},
		))),
	)

	// meta
	metav1.RegisterMetaServer(grpcServer, meta.New())

	// reflection
	reflection.Register(grpcServer)

	// health
	healthServer := health.NewServer()
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)

	cancelChan := make(chan os.Signal, 1)
	// catch SIGETRM or SIGINTERRUPT
	signal.Notify(cancelChan, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			slog.Error("gRPC server stopped with error", "error", err)
			cancelChan <- syscall.SIGTERM
		}
	}()
	sig := <-cancelChan
	slog.Info("Shutting down feature service", "signal", sig)
	return nil

}
