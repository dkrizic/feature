package service

import (
	"context"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/dkrizic/feature/ui/constant"
	"github.com/dkrizic/feature/ui/meta"
	featurev1 "github.com/dkrizic/feature/ui/repository/feature/v1"
	metav1 "github.com/dkrizic/feature/ui/repository/meta/v1"
	workloadv1 "github.com/dkrizic/feature/ui/repository/workload/v1"
	"github.com/urfave/cli/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	metaversion "github.com/dkrizic/feature/ui/meta"
	"github.com/dkrizic/feature/ui/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
)

// Server holds the HTTP server and gRPC clients.
type Server struct {
	address         string
	subpath         string
	templates       *template.Template
	featureClient   featurev1.FeatureClient
	metaClient      metav1.MetaClient
	workloadClient  workloadv1.WorkloadClient
	backendVersion  string
	uiVersion       string
	httpServer      *http.Server
	restartEnabled  bool
	restartName     string
	restartType     string
}

var otelShutdown func(ctx context.Context) error = nil

func Before(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	slog.InfoContext(ctx, "Starting UI", "version", metaversion.Version)

	otelEnabled := cmd.Bool(constant.EnableOpenTelemetry)
	otelEndpoint := cmd.String(constant.OTLPEndpoint)

	if otelEnabled {
		slog.InfoContext(ctx, "OpenTelemetry enabled", "endpoint", otelEndpoint)
		if otelEndpoint == "" {
			slog.ErrorContext(ctx, "OTLP endpoint is required when OpenTelemetry is enabled")
			return ctx, fmt.Errorf("otlp endpoint is required when OpenTelemetry is enabled")
		}
		shutdown, err := telemetry.OpenTelemetryConfig{
			ServiceName:    metaversion.Service,
			ServiceVersion: metaversion.Version,
			OTLPEndpoint:   otelEndpoint,
		}.InitOpenTelemetry(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to initialize OpenTelemetry", "error", err)
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
			slog.ErrorContext(ctx, "Failed to shut down OpenTelemetry", "error", err)
			return fmt.Errorf("failed to shut down OpenTelemetry: %w", err)
		}
	}
	slog.InfoContext(ctx, "Shutting down UI", "version", metaversion.Version)
	return nil
}

// Service is the CLI entrypoint for starting the HTTP UI service.
func Service(ctx context.Context, cmd *cli.Command) error {
	port := cmd.Int(constant.Port)
	endpoint := cmd.String(constant.Endpoint)
	subpath := cmd.String(constant.Subpath)

	// Normalize subpath: ensure it starts with / and doesn't end with /
	if subpath != "" {
		if !strings.HasPrefix(subpath, "/") {
			subpath = "/" + subpath
		}
		subpath = strings.TrimSuffix(subpath, "/")
	}

	slog.InfoContext(ctx, "Configuration", "port", port, "endpoint", endpoint, "subpath", subpath)

	// Parse templates
	templates := ParseTemplates(ctx)
	if templates == nil {
		return fmt.Errorf("failed to parse templates")
	}

	// Dial the gRPC backend
	conn, err := grpc.NewClient(endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to dial gRPC backend", "endpoint", endpoint, "error", err)
		return fmt.Errorf("failed to dial gRPC backend: %w", err)
	}
	defer conn.Close()

	// Initialize gRPC clients
	featureClient := featurev1.NewFeatureClient(conn)
	metaClient := metav1.NewMetaClient(conn)
	workloadClient := workloadv1.NewWorkloadClient(conn)

	// Fetch backend version
	const grpcCallTimeout = 5 * time.Second
	backendVersion := ""
	metaCtx, cancel := context.WithTimeout(ctx, grpcCallTimeout)
	defer cancel()
	metaResp, err := metaClient.Meta(metaCtx, &metav1.MetaRequest{})
	if err != nil {
		slog.WarnContext(ctx, "Failed to fetch backend version", "error", err)
	} else {
		backendVersion = metaResp.Version
		slog.InfoContext(ctx, "Backend version retrieved", "version", backendVersion)
	}

	// Fetch service restart info
	restartEnabled := false
	restartName := ""
	restartType := ""
	infoCtx, infoCancel := context.WithTimeout(ctx, grpcCallTimeout)
	defer infoCancel()
	infoResp, err := workloadClient.Info(infoCtx, &workloadv1.InfoRequest{})
	if err != nil {
		slog.WarnContext(ctx, "Failed to fetch service info", "error", err)
	} else {
		restartEnabled = infoResp.Enabled
		restartName = infoResp.Name
		restartType = infoResp.Type.String()
		slog.InfoContext(ctx, "Service info retrieved", "enabled", restartEnabled, "name", restartName, "type", restartType)
	}

	// Create server
	server := &Server{
		address:        fmt.Sprintf(":%d", port),
		subpath:        subpath,
		templates:      templates,
		featureClient:  featureClient,
		metaClient:     metaClient,
		workloadClient: workloadClient,
		backendVersion: backendVersion,
		uiVersion:      meta.Version,
		restartEnabled: restartEnabled,
		restartName:    restartName,
		restartType:    restartType,
	}

	// Setup HTTP routes
	mux := http.NewServeMux()
	server.registerHandlers(mux)

	// Create HTTP server
	server.httpServer = &http.Server{
		Addr:    server.address,
		Handler: mux,
	}

	// Start HTTP server in a goroutine
	errChan := make(chan error, 1)
	go func() {
		slog.InfoContext(ctx, "HTTP server listening", "address", server.address)
		if err := server.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// Setup signal handling for graceful shutdown
	cancelChan := make(chan os.Signal, 1)
	signal.Notify(cancelChan, syscall.SIGTERM, syscall.SIGINT)

	// Wait for shutdown signal or error
	select {
	case <-ctx.Done():
		slog.InfoContext(ctx, "Context canceled, shutting down")
	case sig := <-cancelChan:
		slog.InfoContext(ctx, "Received signal, shutting down", "signal", sig)
	case err := <-errChan:
		slog.ErrorContext(ctx, "HTTP server error", "error", err)
		return err
	}

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	slog.InfoContext(ctx, "Shutting down HTTP server gracefully")
	if err := server.httpServer.Shutdown(shutdownCtx); err != nil {
		slog.ErrorContext(ctx, "Failed to shutdown HTTP server gracefully", "error", err)
		return err
	}

	slog.InfoContext(ctx, "Feature UI service stopped")
	return nil
}
