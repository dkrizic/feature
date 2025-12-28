package inmemory

// Test inmemory persistence implementation

import (
	"context"
	"testing"

	"github.com/dkrizic/feature/service/service/persistence"
	"github.com/stretchr/testify/assert"
)

func TestInMemoryPersistence(t *testing.T) {
	ctx := context.Background()
	p := NewPersistence()

	// Test PreSet
	err := p.PreSet(ctx, persistence.KeyValue{Key: "key1", Value: "value1"})
	assert.NoError(t, err)

	// Test Get
	kv, err := p.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", kv.Value)

	// Test PreSet on existing key
	err = p.PreSet(ctx, persistence.KeyValue{Key: "key1", Value: "newvalue"})
	assert.NoError(t, err)

	kv, err = p.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", kv.Value) // value should not change

	// Test Set
	err = p.Set(ctx, persistence.KeyValue{Key: "key1", Value: "newvalue"})
	assert.NoError(t, err)

	kv, err = p.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Equal(t, "newvalue", kv.Value)
	// Test GetAll
	err = p.PreSet(ctx, persistence.KeyValue{Key: "key2", Value: "value2"})
	assert.NoError(t, err)

	allKVs, err := p.GetAll(ctx)
	assert.NoError(t, err)
	assert.Len(t, allKVs, 2)
	expected := map[string]string{
		"key1": "newvalue",
		"key2": "value2",
	}
	for _, kv := range allKVs {
		assert.Equal(t, expected[kv.Key], kv.Value)
	}
}
