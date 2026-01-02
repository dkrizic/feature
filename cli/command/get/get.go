package get

import (
	"context"
	"log/slog"

	"github.com/dkrizic/feature/cli/command"
	feature "github.com/dkrizic/feature/cli/repository/feature/v1"
	"github.com/urfave/cli/v3"
	"go.opentelemetry.io/otel"
)

func Get(ctx context.Context, cmd *cli.Command) error {
	ctx, span := otel.Tracer("cli/command/get").Start(ctx, "Get")
	defer span.End()

	fc, err := command.FeatureClient(cmd)
	if err != nil {
		return err
	}

	key := cmd.StringArg("key")

	slog.InfoContext(ctx, "Getting feature", "key", key)
	result, err := fc.Get(ctx, &feature.Key{
		Name: key,
	})
	if err == nil {
		cmd.Writer.Write([]byte(result.Name + "\n"))
	}
	return err
}
