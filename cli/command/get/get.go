package get

import (
	"context"
	"log/slog"

	"github.com/dkrizic/feature/cli/command"
	feature "github.com/dkrizic/feature/cli/repository/feature/v1"
	"github.com/urfave/cli/v3"
)

func Get(ctx context.Context, cmd *cli.Command) error {
	fc, err := command.FeatureClient(cmd)
	if err != nil {
		return err
	}

	key := cmd.StringArg("key")

	slog.Info("Getting feature", "key", key)
	result, err := fc.Get(ctx, &feature.Key{
		Name: key,
	})
	if err == nil {
		cmd.Writer.Write([]byte(result.Name + "\n"))
	}
	return err
}
