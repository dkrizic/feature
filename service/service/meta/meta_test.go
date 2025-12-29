package meta

import (
	"context"
	"testing"

	metav1 "github.com/dkrizic/feature/service/service/meta/v1"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	ms := New()
	assert.NotNil(t, ms)
}

func TestMetaService_Meta(t *testing.T) {
	ms := New()
	req := &metav1.MetaRequest{}

	resp, err := ms.Meta(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.ServiceName)
	assert.NotEmpty(t, resp.Version)
}
