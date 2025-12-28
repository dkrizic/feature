package configmap

import (
	"context"

	"github.com/dkrizic/feature/service/service/persistence"
)

type Persistence struct {
	configMapName string
}

func NewPersistence(configMapName string) *Persistence {
	return &Persistence{
		configMapName: configMapName,
	}
}

func (p *Persistence) GetAll(ctx context.Context) ([]persistence.KeyValue, error) {
	return nil, nil
}

func (p *Persistence) PreSet(ctx context.Context, kv persistence.KeyValue) error {
	return nil
}

func (p *Persistence) Set(ctx context.Context, kv persistence.KeyValue) error {
	return nil
}

func (p *Persistence) Get(ctx context.Context, key string) (persistence.KeyValue, error) {
	return persistence.KeyValue{}, nil
}
