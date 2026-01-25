package restart

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/dkrizic/feature/cli/command"
	workload "github.com/dkrizic/feature/cli/repository/workload/v1"
	"github.com/urfave/cli/v3"
	"go.opentelemetry.io/otel"
)

func Restart(ctx context.Context, cmd *cli.Command) error {
	ctx, span := otel.Tracer("cli/command/restart").Start(ctx, "Restart")
	defer span.End()

	wc, err := command.WorkloadClient(cmd)
	if err != nil {
		return err
	}

	slog.InfoContext(ctx, "Restarting configured service")
	result, err := wc.Restart(ctx, &workload.SimpleRestartRequest{})
	if err != nil {
		return err
	}

	if result.Success {
		cmd.Writer.Write([]byte(fmt.Sprintf("✓ %s\n", result.Message)))
	} else {
		cmd.Writer.Write([]byte(fmt.Sprintf("✗ %s\n", result.Message)))
		return fmt.Errorf("restart failed: %s", result.Message)
	}

	return nil
}
