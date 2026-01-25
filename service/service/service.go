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
	"github.com/dkrizic/feature/service/service/persistence"
	"github.com/dkrizic/feature/service/service/persistence/factory"
	"github.com/dkrizic/feature/service/telemetry"
	"github.com/urfave/cli/v3"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/stats"

	metaversion "github.com/dkrizic/feature/service/meta"

	"github.com/dkrizic/feature/service/service/feature"
	"github.com/dkrizic/feature/service/service/feature/v1"
	"github.com/dkrizic/feature/service/service/meta"
	"github.com/dkrizic/feature/service/service/meta/v1"
	"github.com/dkrizic/feature/service/service/workload"
	workloadv1 "github.com/dkrizic/feature/service/service/workload/v1"
)

var otelShutdown func(ctx context.Context) error = nil

func Before(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	slog.Info("Starting service", "version", metaversion.Version)

	otelEnabled := cmd.Bool(constant.OpenTelemetryEnabled)
	otelEndpoint := cmd.String(constant.OpenTelemetryEndpoint)

	if otelEnabled {
		slog.InfoContext(ctx, "OpenTelemetry enabled", "endpoint", otelEndpoint)
		if otelEndpoint == "" {
			slog.Error("OTLP endpoint is required when OpenTelemetry is enabled")
			return ctx, fmt.Errorf("otlp endpoint is required when OpenTelemetry is enabled")
		}
		shutdown, err := telemetry.OpenTelemetryConfig{
			ServiceName:    metaversion.Service,
			ServiceVersion: metaversion.Version,
			OTLPEndpoint:   otelEndpoint,
		}.InitOpenTelemetry(ctx)
		if err != nil {
			slog.Error("Failed to initialize OpenTelemetry", "error", err)
			return ctx, fmt.Errorf("failed to initialize OpenTelemetry: %w", err)
		}
		otelShutdown = shutdown
	} else {
		slog.InfoContext(ctx, "OpenTelemetry disabled")
	}

	return ctx, nil
}

func After(ctx context.Context, cmd *cli.Command) error {
	if otelShutdown != nil {
		slog.InfoContext(ctx, "Shutting down OpenTelemetry")
		err := otelShutdown(ctx)
		if err != nil {
			slog.Error("Failed to shut down OpenTelemetry", "error", err)
			return fmt.Errorf("failed to shut down OpenTelemetry: %w", err)
		}
	}
	slog.Info("Shutting down service", "version", metaversion.Version)
	return nil
}

func Service(ctx context.Context, cmd *cli.Command) error {
	// get the port
	port := cmd.Int("port")
	slog.InfoContext(ctx, "Configuration", "port", port)

	// configure persistence based on storage type
	pers, err := factory.NewPersistence(ctx, cmd)

	// check if there is a preset
	preset := cmd.StringSlice(constant.PreSet)
	for _, kv := range preset {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) != 2 {
			slog.WarnContext(ctx, "Invalid preset format, expected key=value", "preset", kv)
			continue
		}
		key := parts[0]
		value := parts[1]
		slog.InfoContext(ctx, "Pre-setting key-value", "key", key, "value", value)
		err := pers.PreSet(ctx, persistence.KeyValue{
			Key:   key,
			Value: value,
		})
		if err != nil {
			slog.ErrorContext(ctx, "Failed to pre-set key-value", "key", key, "value", value, "error", err)
			return fmt.Errorf("failed to pre-set key-value: %w", err)
		}
	}

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

	// Get editable fields configuration
	editableFields := cmd.String(constant.Editable)

	// feature
	featureService, err := feature.NewFeatureService(pers, editableFields)
	if err != nil {
		slog.Error("Failed to create feature service", "error", err)
		return fmt.Errorf("failed to create feature service: %w", err)
	}
	featurev1.RegisterFeatureServer(grpcServer, featureService)

	// workload
	// Get the namespace from the environment, default to "default"
	namespace := os.Getenv("POD_NAMESPACE")
	if namespace == "" {
		namespace = "default"
	}

	// Get restart configuration from flags
	restartEnabled := cmd.Bool(constant.RestartEnabled)
	restartTypeStr := cmd.String(constant.RestartType)
	restartName := cmd.String(constant.RestartName)

	// Convert restart type string to protobuf enum
	var restartType workloadv1.WorkloadType
	switch restartTypeStr {
	case "deployment":
		restartType = workloadv1.WorkloadType_WORKLOAD_TYPE_DEPLOYMENT
	case "statefulset":
		restartType = workloadv1.WorkloadType_WORKLOAD_TYPE_STATEFULSET
	case "daemonset":
		restartType = workloadv1.WorkloadType_WORKLOAD_TYPE_DAEMONSET
	default:
		restartType = workloadv1.WorkloadType_WORKLOAD_TYPE_DEPLOYMENT
	}

	workloadService, err := workload.NewWorkloadService(namespace, restartEnabled, restartType, restartName)
	if err != nil {
		slog.WarnContext(ctx, "Failed to create workload service (workload restart feature will be disabled)", "error", err)
	} else {
		workloadv1.RegisterWorkloadServer(grpcServer, workloadService)
		slog.InfoContext(ctx, "Workload service enabled", "namespace", namespace, "restartEnabled", restartEnabled, "restartType", restartTypeStr, "restartName", restartName)
	}

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
