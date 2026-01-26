package delete

import (
	"context"
	"log/slog"

	"github.com/dkrizic/feature/cli/command"
	feature "github.com/dkrizic/feature/cli/repository/feature/v1"
	"github.com/urfave/cli/v3"
	"go.opentelemetry.io/otel"
)

func Delete(ctx context.Context, cmd *cli.Command) error {
	ctx, span := otel.Tracer("cli/command/getall").Start(ctx, "Delete")
	defer span.End()

	fc, err := command.FeatureClient(cmd)
	if err != nil {
		return err
	}

	key := cmd.StringArg("key")
	app := command.GetApplicationName(cmd)

	slog.Info("Deleting feature", "key", key, "application", app)
	_, err = fc.Delete(ctx, &feature.Key{
		Name:        key,
		Application: app,
	})
	return err
}
