package info

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/dkrizic/feature/cli/command"
	workload "github.com/dkrizic/feature/cli/repository/workload/v1"
	"github.com/urfave/cli/v3"
	"go.opentelemetry.io/otel"
)

func Info(ctx context.Context, cmd *cli.Command) error {
	ctx, span := otel.Tracer("cli/command/info").Start(ctx, "Info")
	defer span.End()

	wc, err := command.WorkloadClient(cmd)
	if err != nil {
		return err
	}

	slog.InfoContext(ctx, "Getting service info")
	result, err := wc.Info(ctx, &workload.InfoRequest{})
	if err != nil {
		return err
	}

	// Format output
	output := fmt.Sprintf("Restart enabled: %t\n", result.Enabled)
	if result.Enabled {
		output += fmt.Sprintf("Restart type: %s\n", result.Type.String())
		output += fmt.Sprintf("Restart name: %s\n", result.Name)
	}

	cmd.Writer.Write([]byte(output))
	return nil
}
