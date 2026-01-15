package factory

import (
	"context"
	"testing"

	"github.com/dkrizic/feature/service/constant"
	"github.com/dkrizic/feature/service/service/persistence"
	"github.com/dkrizic/feature/service/service/persistence/notifying"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v3"
)

func newTestCommand(storageType, configMapName string) *cli.Command {
	cmd := &cli.Command{}
	cmd.Flags = []cli.Flag{
		&cli.StringFlag{Name: constant.StorageType, Value: storageType},
		&cli.StringFlag{Name: constant.ConfigMapName, Value: configMapName},
	}
	return cmd
}

func TestNewPersistence_InMemory(t *testing.T) {
	ctx := context.Background()
	cmd := newTestCommand(constant.StorageTypeInMemory, "")

	p, err := NewPersistence(ctx, cmd)
	assert.NoError(t, err)
	assert.NotNil(t, p)

	// The factory wraps the underlying persistence with notifying.NotifyingPersistence
	_, ok := p.(*notifying.NotifyingPersistence)
	assert.True(t, ok, "expected NotifyingPersistence wrapper for in-memory storage")
}

func TestNewPersistence_ConfigMap(t *testing.T) {
	ctx := context.Background()
	cmd := newTestCommand(constant.StorageTypeConfigMap, "test-configmap")

	p, err := NewPersistence(ctx, cmd)
	assert.NoError(t, err)
	assert.NotNil(t, p)

	// The factory wraps the underlying persistence with notifying.NotifyingPersistence
	_, ok := p.(*notifying.NotifyingPersistence)
	assert.True(t, ok, "expected NotifyingPersistence wrapper for configmap storage")
}

func TestNewPersistence_InvalidType(t *testing.T) {
	ctx := context.Background()
	cmd := newTestCommand("invalid", "")

	p, err := NewPersistence(ctx, cmd)
	assert.Error(t, err)
	assert.Nil(t, p)
}

// Ensure factory returns an implementation that fulfills the Persistence interface
func TestNewPersistence_ImplementsInterface(t *testing.T) {
	ctx := context.Background()
	cmd := newTestCommand(constant.StorageTypeInMemory, "")

	p, err := NewPersistence(ctx, cmd)
	assert.NoError(t, err)

	var _ persistence.Persistence = p
}
