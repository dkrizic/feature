package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/dkrizic/feature/service/constant"
	"github.com/dkrizic/feature/service/meta"
	"github.com/dkrizic/feature/service/service"
	"github.com/dkrizic/feature/service/telemetry/injectctx"
	"github.com/urfave/cli/v3" // imports as package "cli"
)

func main() {
	cmd := &cli.Command{
		Name:  "feature",
		Usage: "Feature service",
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
						Value:    "",
						Category: "observability",
						Usage:    "OTLP endpoint for OpenTelemetry",
						Sources:  cli.EnvVars("OTLP_ENDPOINT"),
					},
					&cli.StringFlag{
						Name:    constant.StorageType,
						Value:   constant.StorageTypeInMemory,
						Usage:   "Type of storage to use: inmemory, configmap",
						Sources: cli.EnvVars("STORAGE_TYPE"),
						Action: func(ctx context.Context, cmd *cli.Command, s string) error {
							if s != constant.StorageTypeInMemory && s != constant.StorageTypeConfigMap {
								return fmt.Errorf("invalid storage type: %s", s)
							}
							if s == constant.StorageTypeConfigMap {
								configMapName := cmd.String(constant.ConfigMapName)
								if configMapName == "" {
									return fmt.Errorf("configmap-name cannot be empty when storage-type is configmap")
								}
							}
							return nil
						},
					},
					&cli.StringFlag{
						Name:    constant.ConfigMapName,
						Usage:   "Name of the ConfigMap to use for configmap storage",
						Sources: cli.EnvVars("CONFIGMAP_NAME"),
					},
					&cli.StringSliceFlag{
						Name:    constant.PreSet,
						Usage:   "Pre-set key-value pairs in the format key=value before starting the service",
						Sources: cli.EnvVars("PRESET"),
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

	otelhandler := injectctx.NewHandler(handler)

	logger := slog.New(otelhandler)
	slog.SetDefault(logger)

	return ctx, nil
}
