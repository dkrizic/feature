package log

// implements a Notifier that logs notifications using slog.

import (
	"context"
	"github.com/dkrizic/feature/service/notifier"
	"go.opentelemetry.io/otel"
	"log/slog"
)

type LogNotifier struct {
}

func NewLogNotifier() *LogNotifier {
	return &LogNotifier{}
}

func (n *LogNotifier) Notify(ctx context.Context, notification notifier.Notification) error {
	ctx, span := otel.Tracer("notifier/log").Start(ctx, "Notify")
	defer span.End()

	slog.Info("Notification", "action_type", notification.Action.Type, "key", notification.Action.Key, "value", notification.Action.Value)
	return nil
}
