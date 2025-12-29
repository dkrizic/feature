package set

import (
	"context"
	"log/slog"

	"github.com/dkrizic/feature/cli/command"
	feature "github.com/dkrizic/feature/cli/repository/feature/v1"
	"github.com/urfave/cli/v3"
)

func Set(ctx context.Context, cmd *cli.Command) error {
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
