package none

import (
	"context"

	"go.opentelemetry.io/otel"
)
import . "github.com/dkrizic/feature/service/notifier"

// implements the Notifer interace and does nothing

type NoneNotifier struct {
}

func NewNoneNotifier() *NoneNotifier {
	return &NoneNotifier{}
}

func (n *NoneNotifier) Notify(ctx context.Context, notification Notification) error {
	ctx, span := otel.Tracer("notifier/none").Start(ctx, "Notify")
	defer span.End()

	return nil
}
