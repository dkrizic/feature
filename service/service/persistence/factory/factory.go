package factory

import (
	"errors"

	"github.com/dkrizic/feature/service/constant"
	nf "github.com/dkrizic/feature/service/notifier/factory"
	"github.com/dkrizic/feature/service/service/persistence"
	"github.com/dkrizic/feature/service/service/persistence/configmap"
	"github.com/dkrizic/feature/service/service/persistence/inmemory"
	"github.com/dkrizic/feature/service/service/persistence/notifying"
	"github.com/urfave/cli/v3"

	"context"
	"log/slog"
)

func NewPersistence(ctx context.Context, cmd *cli.Command) (persistence.Persistence, error) {
	stype := cmd.String(constant.StorageType)

	notifier, err := nf.NewNotifier(ctx, cmd)
	if err != nil {
		return nil, err
	}

	switch stype {
	case constant.StorageTypeInMemory:
		slog.InfoContext(ctx, "In-memory storage selected")
		return notifying.NewNotifyingPersistence(
			inmemory.NewInMemoryPersistence(), notifier,
		), nil
	case constant.StorageTypeConfigMap:
		slog.InfoContext(ctx, "ConfigMap storage selected")
		cmName := cmd.String(constant.ConfigMapName)
		return notifying.NewNotifyingPersistence(
			configmap.NewConfigMapPersistence(cmName), notifier,
		), nil
	default:
		slog.ErrorContext(ctx, "Invalid storage type", "type", stype)
		return nil, errors.New("Invalid storage type")
	}
}
