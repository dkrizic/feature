package getall

import (
	"context"
	"log/slog"

	"github.com/dkrizic/feature/cli/command"
	"github.com/urfave/cli/v3"
	"go.opentelemetry.io/otel"
	"google.golang.org/protobuf/types/known/emptypb"
)

func GetAll(ctx context.Context, cmd *cli.Command) error {
	ctx, span := otel.Tracer("cli/command/getall").Start(ctx, "GetAll")
	defer span.End()

	fc, err := command.FeatureClient(cmd)
	if err != nil {
		return err
	}

	slog.InfoContext(ctx, "Getting all features")
	all, err := fc.GetAll(ctx, &emptypb.Empty{})
	if err != nil {
		return err
	}
	for {
		kv, err := all.Recv()
		if err != nil {
			break
		}
		slog.InfoContext(ctx, "Feature", "key", kv.Key, "value", kv.Value)
	}
	return nil
}
