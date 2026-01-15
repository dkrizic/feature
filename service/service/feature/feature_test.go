package feature

import (
	"context"
	"errors"
	"testing"

	featurev1 "github.com/dkrizic/feature/service/service/feature/v1"
	"github.com/dkrizic/feature/service/service/persistence"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type fakePersistence struct {
	values      []persistence.KeyValue
	getAllErr   error
	preSetErr   error
	setErr      error
	getResult   persistence.KeyValue
	getErr      error
	deleteErr   error
	countResult int
	countErr    error
}

func (f *fakePersistence) GetAll(ctx context.Context) ([]persistence.KeyValue, error) {
	return f.values, f.getAllErr
}

func (f *fakePersistence) PreSet(ctx context.Context, kv persistence.KeyValue) error {
	return f.preSetErr
}

func (f *fakePersistence) Set(ctx context.Context, kv persistence.KeyValue) error {
	return f.setErr
}

func (f *fakePersistence) Get(ctx context.Context, key string) (persistence.KeyValue, error) {
	return f.getResult, f.getErr
}

func (f *fakePersistence) Delete(ctx context.Context, key string) error {
	return f.deleteErr
}

func (f *fakePersistence) Count(ctx context.Context) (int, error) {
	return f.countResult, f.countErr
}

type fakeServerStream struct {
	grpc.ServerStreamingServer[featurev1.KeyValue]
	ctx  context.Context
	sent []*featurev1.KeyValue
}

func (f *fakeServerStream) Context() context.Context { return f.ctx }

func (f *fakeServerStream) Send(kv *featurev1.KeyValue) error {
	f.sent = append(f.sent, kv)
	return nil
}

func TestFeatureService_GetAll_Success(t *testing.T) {
	fp := &fakePersistence{
		values: []persistence.KeyValue{{Key: "k1", Value: "v1"}, {Key: "k2", Value: "v2"}},
	}
	fs, err := NewFeatureService(fp)
	assert.NoError(t, err)

	ctx := context.Background()
	stream := &fakeServerStream{ctx: ctx}

	err = fs.GetAll(&emptypb.Empty{}, stream)
	assert.NoError(t, err)
	assert.Len(t, stream.sent, 2)
	assert.Equal(t, "k1", stream.sent[0].Key)
	assert.Equal(t, "v1", stream.sent[0].Value)
}

func TestFeatureService_GetAll_PersistenceError(t *testing.T) {
	fp := &fakePersistence{getAllErr: errors.New("boom")}
	fs, err := NewFeatureService(fp)
	assert.NoError(t, err)

	ctx := context.Background()
	stream := &fakeServerStream{ctx: ctx}

	err = fs.GetAll(&emptypb.Empty{}, stream)
	assert.Error(t, err)
}

func TestFeatureService_PreSet_Success(t *testing.T) {
	fp := &fakePersistence{countResult: 1}
	fs, err := NewFeatureService(fp)
	assert.NoError(t, err)

	ctx := context.Background()
	_, err = fs.PreSet(ctx, &featurev1.KeyValue{Key: "k1", Value: "v1"})
	assert.NoError(t, err)
}

func TestFeatureService_PreSet_PersistenceError(t *testing.T) {
	fp := &fakePersistence{preSetErr: errors.New("boom")}
	fs, err := NewFeatureService(fp)
	assert.NoError(t, err)

	ctx := context.Background()
	_, err = fs.PreSet(ctx, &featurev1.KeyValue{Key: "k1", Value: "v1"})
	assert.Error(t, err)
}

func TestFeatureService_Set_Success(t *testing.T) {
	fp := &fakePersistence{countResult: 1}
	fs, err := NewFeatureService(fp)
	assert.NoError(t, err)

	ctx := context.Background()
	_, err = fs.Set(ctx, &featurev1.KeyValue{Key: "k1", Value: "v1"})
	assert.NoError(t, err)
}

func TestFeatureService_Set_PersistenceError(t *testing.T) {
	fp := &fakePersistence{setErr: errors.New("boom")}
	fs, err := NewFeatureService(fp)
	assert.NoError(t, err)

	ctx := context.Background()
	_, err = fs.Set(ctx, &featurev1.KeyValue{Key: "k1", Value: "v1"})
	assert.Error(t, err)
}

func TestFeatureService_Get_Found(t *testing.T) {
	fp := &fakePersistence{getResult: persistence.KeyValue{Key: "k1", Value: "v1"}, countResult: 1}
	fs, err := NewFeatureService(fp)
	assert.NoError(t, err)

	ctx := context.Background()
	val, err := fs.Get(ctx, &featurev1.Key{Name: "k1"})
	assert.NoError(t, err)
	assert.Equal(t, "v1", val.Name)
}

func TestFeatureService_Get_Error(t *testing.T) {
	fp := &fakePersistence{getErr: errors.New("boom")}
	fs, err := NewFeatureService(fp)
	assert.NoError(t, err)

	ctx := context.Background()
	_, err = fs.Get(ctx, &featurev1.Key{Name: "k1"})
	assert.Error(t, err)
}

func TestFeatureService_Delete_Success(t *testing.T) {
	fp := &fakePersistence{countResult: 0}
	fs, err := NewFeatureService(fp)
	assert.NoError(t, err)

	ctx := context.Background()
	_, err = fs.Delete(ctx, &featurev1.Key{Name: "k1"})
	assert.NoError(t, err)
}

func TestFeatureService_Delete_Error(t *testing.T) {
	fp := &fakePersistence{deleteErr: errors.New("boom")}
	fs, err := NewFeatureService(fp)
	assert.NoError(t, err)

	ctx := context.Background()
	_, err = fs.Delete(ctx, &featurev1.Key{Name: "k1"})
	assert.Error(t, err)
}
