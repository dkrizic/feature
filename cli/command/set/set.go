package set

import (
	"context"
	"log/slog"

	"github.com/dkrizic/feature/cli/command"
	feature "github.com/dkrizic/feature/cli/repository/feature/v1"
	"github.com/urfave/cli/v3"
	"go.opentelemetry.io/otel"
)

func Set(ctx context.Context, cmd *cli.Command) error {
	ctx, span := otel.Tracer("cli/command/set").Start(ctx, "Set")
	defer span.End()

	fc, err := command.FeatureClient(cmd)
	if err != nil {
		return err
	}

	key := cmd.StringArg("key")
	value := cmd.StringArg("value")

	slog.Info("Setting feature", "key", key, "value", value)
	_, err = fc.Set(ctx, &feature.KeyValue{
		Key:   key,
		Value: value,
	})
	return err
}
