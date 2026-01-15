package factory

import (
	"errors"

	"github.com/dkrizic/feature/service/constant"
	"github.com/dkrizic/feature/service/notifier"
	"github.com/dkrizic/feature/service/notifier/log"
	"github.com/dkrizic/feature/service/notifier/none"
	"github.com/urfave/cli/v3"

	"context"
	"log/slog"
)

func NewNotifier(ctx context.Context, cmd *cli.Command) (notifier.Notifier, error) {
	enabled := cmd.Bool(constant.NotificationEnabled)
	ntype := cmd.String(constant.NotificationType)

	if !enabled {
		slog.InfoContext(ctx, "Notifications disabled")
		return none.NewNoneNotifier(), nil
	}

	switch ntype {
	case constant.NotificationTypeLog:
		slog.InfoContext(ctx, "Log notifier selected")
		return log.NewLogNotifier(), nil
	default:
		slog.ErrorContext(ctx, "Invalid notifier type", "type", ntype)
		return nil, errors.New("Invalid notifier type")
	}
}
