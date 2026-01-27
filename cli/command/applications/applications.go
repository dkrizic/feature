package applications

import (
	"context"
	"log/slog"

	"github.com/dkrizic/feature/cli/command"
	feature "github.com/dkrizic/feature/cli/repository/feature/v1"
	"github.com/urfave/cli/v3"
	"go.opentelemetry.io/otel"
)

func Applications(ctx context.Context, cmd *cli.Command) error {
	ctx, span := otel.Tracer("cli/command/applications").Start(ctx, "Applications")
	defer span.End()

	fc, err := command.FeatureClient(cmd)
	if err != nil {
		return err
	}

	slog.InfoContext(ctx, "Getting all applications")
	all, err := fc.GetApplications(ctx, &feature.ApplicationsRequest{})
	if err != nil {
		return err
	}
	for {
		app, err := all.Recv()
		if err != nil {
			break
		}
		slog.InfoContext(ctx, "Application", "name", app.Name, "namespace", app.Namespace, "storage_type", app.StorageType)
	}
	return nil
}
