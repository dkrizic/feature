package configmap

import (
	"context"
	"testing"

	"github.com/dkrizic/feature/service/service/persistence"
	"github.com/stretchr/testify/assert"
)

func TestNewPersistence(t *testing.T) {
	configMapName := "test-configmap"
	p := NewPersistence(configMapName)

	assert.NotNil(t, p)
	assert.Equal(t, configMapName, p.configMapName)
}

func TestConfigMapPersistence_GetAll(t *testing.T) {
	ctx := context.Background()
	p := NewPersistence("test-configmap")

	result, err := p.GetAll(ctx)

	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestConfigMapPersistence_PreSet(t *testing.T) {
	ctx := context.Background()
	p := NewPersistence("test-configmap")

	err := p.PreSet(ctx, persistence.KeyValue{Key: "key1", Value: "value1"})

	assert.NoError(t, err)
}

func TestConfigMapPersistence_Set(t *testing.T) {
	ctx := context.Background()
	p := NewPersistence("test-configmap")

	err := p.Set(ctx, persistence.KeyValue{Key: "key1", Value: "value1"})

	assert.NoError(t, err)
}

func TestConfigMapPersistence_Get(t *testing.T) {
	ctx := context.Background()
	p := NewPersistence("test-configmap")

	result, err := p.Get(ctx, "key1")

	assert.NoError(t, err)
	assert.Equal(t, "", result.Key)
	assert.Equal(t, "", result.Value)
}

func TestConfigMapPersistence_Delete(t *testing.T) {
	ctx := context.Background()
	p := NewPersistence("test-configmap")

	err := p.Delete(ctx, "key1")

	assert.NoError(t, err)
}
