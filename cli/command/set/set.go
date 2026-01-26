package set

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/dkrizic/feature/cli/command"
	feature "github.com/dkrizic/feature/cli/repository/feature/v1"
	"github.com/urfave/cli/v3"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	app := command.GetApplicationName(cmd)

	slog.Info("Setting feature", "key", key, "value", value, "application", app)
	_, err = fc.Set(ctx, &feature.KeyValue{
		Key:         key,
		Value:       value,
		Application: app,
	})
	
	// Check if the error is a PermissionDenied error
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.PermissionDenied {
			slog.Warn("Permission denied", "key", key, "error", st.Message())
			return fmt.Errorf("%s", st.Message())
		}
	}
	
	return err
}
