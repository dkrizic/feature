package service

import (
	"context"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dkrizic/feature/ui/constant"
	"github.com/dkrizic/feature/ui/meta"
	featurev1 "github.com/dkrizic/feature/ui/repository/feature/v1"
	metav1 "github.com/dkrizic/feature/ui/repository/meta/v1"
	"github.com/urfave/cli/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Server holds the HTTP server and gRPC clients.
type Server struct {
	address        string
	templates      *template.Template
	featureClient  featurev1.FeatureClient
	metaClient     metav1.MetaClient
	backendVersion string
	uiVersion      string
	httpServer     *http.Server
}

// Service is the CLI entrypoint for starting the HTTP UI service.
func Service(ctx context.Context, cmd *cli.Command) error {
	port := cmd.Int(constant.Port)
	endpoint := cmd.String(constant.Endpoint)

	slog.Info("Starting the feature UI service", "port", port, "endpoint", endpoint)

	// Parse templates
	templates := ParseTemplates()
	if templates == nil {
		return fmt.Errorf("failed to parse templates")
	}

	// Dial the gRPC backend
	conn, err := grpc.NewClient(endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("Failed to dial gRPC backend", "endpoint", endpoint, "error", err)
		return fmt.Errorf("failed to dial gRPC backend: %w", err)
	}
	defer conn.Close()

	// Initialize gRPC clients
	featureClient := featurev1.NewFeatureClient(conn)
	metaClient := metav1.NewMetaClient(conn)

	// Fetch backend version
	backendVersion := ""
	metaCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	metaResp, err := metaClient.Meta(metaCtx, &metav1.MetaRequest{})
	if err != nil {
		slog.Warn("Failed to fetch backend version", "error", err)
	} else {
		backendVersion = metaResp.Version
		slog.Info("Backend version retrieved", "version", backendVersion)
	}

	// Create server
	server := &Server{
		address:        fmt.Sprintf(":%d", port),
		templates:      templates,
		featureClient:  featureClient,
		metaClient:     metaClient,
		backendVersion: backendVersion,
		uiVersion:      meta.Version,
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
		slog.Info("HTTP server listening", "address", server.address)
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
		slog.Info("Context canceled, shutting down")
	case sig := <-cancelChan:
		slog.Info("Received signal, shutting down", "signal", sig)
	case err := <-errChan:
		slog.Error("HTTP server error", "error", err)
		return err
	}

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	slog.Info("Shutting down HTTP server gracefully")
	if err := server.httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("Failed to shutdown HTTP server gracefully", "error", err)
		return err
	}

	slog.Info("Feature UI service stopped")
	return nil
}
