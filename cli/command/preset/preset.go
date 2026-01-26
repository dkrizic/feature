package preset

import (
	"context"
	"log/slog"

	"github.com/dkrizic/feature/cli/command"
	feature "github.com/dkrizic/feature/cli/repository/feature/v1"
	"github.com/urfave/cli/v3"
	"go.opentelemetry.io/otel"
)

func PreSet(ctx context.Context, cmd *cli.Command) error {
	ctx, span := otel.Tracer("cli/command/preset").Start(ctx, "PreSet")
	defer span.End()

	fc, err := command.FeatureClient(cmd)
	if err != nil {
		return err
	}

	key := cmd.StringArg("key")
	value := cmd.StringArg("value")
	app := command.GetApplicationName(cmd)

	slog.Info("PreSetting feature", "key", key, "value", value, "application", app)
	_, err = fc.PreSet(ctx, &feature.KeyValue{
		Key:         key,
		Value:       value,
		Application: app,
	})
	return err
}
