package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/dkrizic/feature/cli/command/delete"
	"github.com/dkrizic/feature/cli/command/get"
	"github.com/dkrizic/feature/cli/command/getall"
	"github.com/dkrizic/feature/cli/command/preset"
	"github.com/dkrizic/feature/cli/command/set"
	"github.com/dkrizic/feature/cli/constant"
	"github.com/dkrizic/feature/cli/meta"
	"github.com/urfave/cli/v3" // imports as package "cli"
)

func main() {
	cmd := &cli.Command{
		Name:  "feature",
		Usage: "Feature CLI",
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
				Name:   "getall",
				Usage:  "Get all features",
				Action: getall.GetAll,
			},
			&cli.Command{
				Name:   "get",
				Usage:  "Get a feature by key",
				Action: get.Get,
				Arguments: []cli.Argument{
					&cli.StringArg{
						Name: "key",
					},
				},
			},
			&cli.Command{
				Name:   "set",
				Usage:  "Set a feature key-value",
				Action: set.Set,
				Arguments: []cli.Argument{
					&cli.StringArg{
						Name: "key",
					},
					&cli.StringArg{
						Name: "value",
					},
				},
			},
			&cli.Command{
				Name:   "delete",
				Usage:  "Delete a feature by key",
				Action: delete.Delete,
				Arguments: []cli.Argument{
					&cli.StringArg{
						Name: "key",
					},
				},
			},
			&cli.Command{
				Name:   "preset",
				Usage:  "Pre-set a feature key-value",
				Action: preset.PreSet,
				Arguments: []cli.Argument{
					&cli.StringArg{
						Name: "key",
					},
					&cli.StringArg{
						Name: "value",
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

	logger := slog.New(handler)
	slog.SetDefault(logger)

	return ctx, nil
}
