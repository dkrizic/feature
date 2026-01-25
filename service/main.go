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
						Name:     constant.OpenTelemetryEnabled,
						Value:    false,
						Category: "observability",
						Usage:    "Enable OpenTelemetry tracing",
						Sources:  cli.EnvVars("OPENTELEMETRY_ENABLED"),
					},
					&cli.StringFlag{
						Name:     constant.OpenTelemetryEndpoint,
						Value:    "",
						Category: "observability",
						Usage:    "OTLP endpoint for OpenTelemetry",
						Sources:  cli.EnvVars("OPENTELEMETRY_ENDPOINT"),
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
					&cli.BoolFlag{
						Name:    constant.NotificationEnabled,
						Usage:   "Enable notifications for feature changes",
						Value:   false,
						Sources: cli.EnvVars("NOTIFICATION_ENABLED"),
					},
					&cli.StringFlag{
						Name:    constant.NotificationType,
						Usage:   "Type of notification to use: log, redis_topic",
						Value:   constant.NotificationTypeLog,
						Sources: cli.EnvVars("NOTIFICATION_TYPE"),
						Action: func(ctx context.Context, cmd *cli.Command, s string) error {
							if s != constant.NotificationTypeLog && s != constant.NotificationTypeRedisTopic {
								return fmt.Errorf("invalid notification type: %s", s)
							}
							// if notification type is redis_topic, redis endpoint and redis topic must be set
							if s == constant.NotificationTypeRedisTopic {
								redisEndpoint := cmd.String(constant.RedisEndpoint)
								redisTopic := cmd.String(constant.RedisNotificationTopic)
								if redisEndpoint == "" {
									return fmt.Errorf("redis-endpoint cannot be empty when notification-type is redis_topic")
								}
								if redisTopic == "" {
									return fmt.Errorf("redis-notification-topic cannot be empty when notification-type is redis_topic")
								}
							}
							return nil
						},
					},
					&cli.StringFlag{
						Name:    constant.RedisEndpoint,
						Usage:   "Redis endpoint for redis_topic notifications",
						Sources: cli.EnvVars("REDIS_ENDPOINT"),
					},
					&cli.StringFlag{
						Name:    constant.RedisNotificationTopic,
						Usage:   "Redis topic for notifications",
						Value:   "feature_notifications",
						Sources: cli.EnvVars("REDIS_NOTIFICATION_TOPIC"),
					},
					&cli.BoolFlag{
						Name:     constant.RestartEnabled,
						Usage:    "Enable workload restart feature",
						Value:    false,
						Category: "restart",
						Sources:  cli.EnvVars("RESTART_ENABLED"),
					},
					&cli.StringFlag{
						Name:     constant.RestartType,
						Usage:    "Type of workload to restart: deployment, statefulset, daemonset",
						Value:    "deployment",
						Category: "restart",
						Sources:  cli.EnvVars("RESTART_TYPE"),
						Action: func(ctx context.Context, cmd *cli.Command, s string) error {
							if s != "" && s != "deployment" && s != "statefulset" && s != "daemonset" {
								return fmt.Errorf("invalid restart type: %s (must be deployment, statefulset, or daemonset)", s)
							}
							return nil
						},
					},
					&cli.StringFlag{
						Name:     constant.RestartName,
						Usage:    "Name of the workload to restart",
						Value:    "",
						Category: "restart",
						Sources:  cli.EnvVars("RESTART_NAME"),
					},
					&cli.StringFlag{
						Name:     constant.Editable,
						Usage:    "Comma-separated list of editable field names (empty means all fields are editable)",
						Value:    "",
						Category: "service",
						Sources:  cli.EnvVars("EDITABLE"),
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
