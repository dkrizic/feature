package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/dkrizic/feature/service/meta"
	"github.com/dkrizic/feature/service/service"
	"github.com/urfave/cli/v3" // imports as package "cli"
)

func main() {
	cmd := &cli.Command{
		Name:  "feature",
		Usage: "Feature service",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "log-format",
				Value:    "text",
				Category: "logging",
				Usage:    "Log format: text or json",
				Sources:  cli.EnvVars("LOG_FORMAT"),
			},
			&cli.StringFlag{
				Name:     "log-level",
				Value:    "info",
				Category: "logging",
				Usage:    "Log level: debug, info, warn, error",
				Sources:  cli.EnvVars("LOG_LEVEL"),
			},
		},
		Before: beforeAction,
		Commands: []*cli.Command{
			&cli.Command{
				Name:  "version",
				Usage: "Print the version number of the feature service",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					slog.Info("Feature Service", "name", meta.Service, "version", meta.Version)
					return nil
				},
			},
			&cli.Command{
				Name:   "service",
				Usage:  "Start the feature service",
				Action: service.Service,
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:     "port",
						Value:    8080,
						Category: "service",
						Usage:    "Port to run the service on",
						Sources:  cli.EnvVars("PORT"),
					},
					&cli.BoolFlag{
						Name:     "enable-opentelemetry",
						Value:    false,
						Category: "observability",
						Usage:    "Enable OpenTelemetry tracing",
						Sources:  cli.EnvVars("ENABLE_OPENTELEMETRY"),
					},
					&cli.StringFlag{
						Name:     "otlp-endpoint",
						Value:    "localhost:4317",
						Category: "observability",
						Usage:    "OTLP endpoint for OpenTelemetry",
						Sources:  cli.EnvVars("OTLP_ENDPOINT"),
					},
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func beforeAction(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	logFormat := cmd.String("log-format")
	logLevel := cmd.String("log-level")

	level := slog.LevelInfo
	switch logLevel {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		return ctx, fmt.Errorf("invalid log level: %s", logLevel)
	}

	var handler slog.Handler
	if logFormat == "json" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	return ctx, nil
}
