package meta

import (
	"context"
	"testing"

	metav1 "github.com/dkrizic/feature/service/service/meta/v1"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	ms := New(false)
	assert.NotNil(t, ms)
}

func TestMetaService_Meta(t *testing.T) {
	ms := New(false)
	req := &metav1.MetaRequest{}

	resp, err := ms.Meta(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.ServiceName)
	assert.NotEmpty(t, resp.Version)
	assert.False(t, resp.AuthenticationRequired)
}

func TestMetaService_MetaWithAuth(t *testing.T) {
	ms := New(true)
	req := &metav1.MetaRequest{}

	resp, err := ms.Meta(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.ServiceName)
	assert.NotEmpty(t, resp.Version)
	assert.True(t, resp.AuthenticationRequired)
}
