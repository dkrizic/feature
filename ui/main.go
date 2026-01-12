package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/dkrizic/feature/ui/constant"
	"github.com/dkrizic/feature/ui/meta"
	"github.com/dkrizic/feature/ui/service"
	"github.com/dkrizic/feature/ui/telemetry/injectctx"
	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:  "feature-ui",
		Usage: "Feature UI",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     constant.LogFormat,
				Value:    constant.LogFormatText,
				Category: "logging",
				Usage:    "Log format: text or json",
				Sources:  cli.EnvVars("LOG_FORMAT"),
				Action: func(ctx context.Context, command *cli.Command, s string) error {
					if s != constant.LogFormatText && s != constant.LogFormatJSON {
						return fmt.Errorf("invalid log format: %s", s)
					}
					return nil
				},
			},
			&cli.StringFlag{
				Name:     constant.LogLevel,
				Value:    constant.LogLevelInfo,
				Category: "logging",
				Usage:    "Log level: debug, info, warn, error",
				Sources:  cli.EnvVars("LOG_LEVEL"),
				Action: func(ctx context.Context, command *cli.Command, s string) error {
					if s != constant.LogLevelDebug && s != constant.LogLevelInfo && s != constant.LogLevelWarn && s != constant.LogLevelError {
						return fmt.Errorf("invalid log level: %s", s)
					}
					return nil
				},
			},
			&cli.StringFlag{
				Name:     constant.Endpoint,
				Value:    "localhost:8000",
				Category: "connection",
				Usage:    "Feature service endpoint",
				Required: true,
				Sources:  cli.EnvVars("ENDPOINT"),
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
				Before: service.Before,
				Action: service.Service,
				After:  service.After,
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:     constant.Port,
						Value:    8080,
						Category: "service",
						Usage:    "Port to run the service on",
						Sources:  cli.EnvVars("PORT"),
					},
					&cli.BoolFlag{
						Name:     constant.EnableOpenTelemetry,
						Value:    false,
						Category: "observability",
						Usage:    "Enable OpenTelemetry tracing",
						Sources:  cli.EnvVars("ENABLE_OPENTELEMETRY"),
					},
					&cli.StringFlag{
						Name:     constant.OTLPEndpoint,
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
	logFormat := cmd.String(constant.LogFormat)
	logLevel := cmd.String(constant.LogLevel)

	level := slog.LevelInfo
	switch logLevel {
	case constant.LogLevelDebug:
		level = slog.LevelDebug
	case constant.LogLevelInfo:
		level = slog.LevelInfo
	case constant.LogLevelWarn:
		level = slog.LevelWarn
	case constant.LogLevelError:
		level = slog.LevelError
	default:
		return ctx, fmt.Errorf("invalid log level: %s", logLevel)
	}

	var handler slog.Handler
	if logFormat == constant.LogFormatJSON {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	}

	otelhttp := injectctx.NewHandler(handler)

	logger := slog.New(otelhttp)
	slog.SetDefault(logger)

	return ctx, nil
}
