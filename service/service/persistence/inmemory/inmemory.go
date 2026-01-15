package inmemory

import (
	"context"
	"log/slog"

	"github.com/dkrizic/feature/service/service/persistence"
	"go.opentelemetry.io/otel"
)

type Persistence struct {
	data map[string]string
}

func NewInMemoryPersistence() *Persistence {
	return &Persistence{
		data: make(map[string]string),
	}
}

func (p *Persistence) GetAll(ctx context.Context) ([]persistence.KeyValue, error) {
	ctx, span := otel.Tracer("service/persistence/inmemory").Start(ctx, "GetAll")
	defer span.End()

	var result []persistence.KeyValue
	for k, v := range p.data {
		result = append(result, persistence.KeyValue{Key: k, Value: v})
	}
	return result, nil
}

// set the value only if it does not exist
func (p *Persistence) PreSet(ctx context.Context, kv persistence.KeyValue) error {
	ctx, span := otel.Tracer("service/persistence/inmemory").Start(ctx, "PreSet")
	defer span.End()

	oldvalue, exist := p.data[kv.Key]
	if !exist {
		slog.DebugContext(ctx, "PreSetting", "key", kv.Key, "value", kv.Value)
		p.data[kv.Key] = kv.Value
	} else {
		slog.InfoContext(ctx, "Key already exists, not presetting", "key", kv.Key, "value", kv.Value, "oldvalue", oldvalue)
	}
	return nil

}

func (p *Persistence) Set(ctx context.Context, kv persistence.KeyValue) error {
	ctx, span := otel.Tracer("service/persistence/inmemory").Start(ctx, "Set")
	defer span.End()

	p.data[kv.Key] = kv.Value
	slog.DebugContext(ctx, "Setting", "key", kv.Key, "value", kv.Value)
	return nil
}

func (p *Persistence) Get(ctx context.Context, key string) (persistence.KeyValue, error) {
	ctx, span := otel.Tracer("service/persistence/inmemory").Start(ctx, "Get")
	defer span.End()

	result := persistence.KeyValue{Key: key, Value: p.data[key]}
	slog.DebugContext(ctx, "Getting", "key", key, "value", result.Value)
	return result, nil
}

func (p *Persistence) Delete(ctx context.Context, key string) error {
	ctx, span := otel.Tracer("service/persistence/inmemory").Start(ctx, "Delete")
	defer span.End()

	delete(p.data, key)
	slog.DebugContext(ctx, "Deleting", "key", key)
	return nil
}

func (p *Persistence) Count(ctx context.Context) (int, error) {
	ctx, span := otel.Tracer("service/persistence/inmemory").Start(ctx, "Count")
	defer span.End()

	count := len(p.data)
	slog.DebugContext(ctx, "Counting", "count", count)
	return count, nil
}
