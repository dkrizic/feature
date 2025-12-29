package feature

import (
	"context"
	"io"
	"testing"

	featurev1 "github.com/dkrizic/feature/service/service/feature/v1"
	"github.com/dkrizic/feature/service/service/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

// MockPersistence is a mock implementation of persistence.Persistence
type MockPersistence struct {
	mock.Mock
}

func (m *MockPersistence) GetAll(ctx context.Context) ([]persistence.KeyValue, error) {
	args := m.Called(ctx)
	return args.Get(0).([]persistence.KeyValue), args.Error(1)
}

func (m *MockPersistence) PreSet(ctx context.Context, kv persistence.KeyValue) error {
	args := m.Called(ctx, kv)
	return args.Error(0)
}

func (m *MockPersistence) Set(ctx context.Context, kv persistence.KeyValue) error {
	args := m.Called(ctx, kv)
	return args.Error(0)
}

func (m *MockPersistence) Get(ctx context.Context, key string) (persistence.KeyValue, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(persistence.KeyValue), args.Error(1)
}

func (m *MockPersistence) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

// MockServerStream is a mock implementation of grpc.ServerStreamingServer
type MockServerStream struct {
	mock.Mock
	ctx context.Context
}

func (m *MockServerStream) Context() context.Context {
	return m.ctx
}

func (m *MockServerStream) Send(kv *featurev1.KeyValue) error {
	args := m.Called(kv)
	return args.Error(0)
}

func (m *MockServerStream) SetHeader(md metadata.MD) error {
	return nil
}

func (m *MockServerStream) SendHeader(md metadata.MD) error {
	return nil
}

func (m *MockServerStream) SetTrailer(md metadata.MD) {}

func (m *MockServerStream) SendMsg(msg interface{}) error {
	return nil
}

func (m *MockServerStream) RecvMsg(msg interface{}) error {
	return nil
}

func TestNewFeatureService(t *testing.T) {
	mockPers := new(MockPersistence)
	fs := NewFeatureService(mockPers)

	assert.NotNil(t, fs)
	assert.Equal(t, mockPers, fs.persistence)
}

func TestFeatureService_Get(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		mockReturn    persistence.KeyValue
		mockError     error
		expectedValue string
		expectError   bool
	}{
		{
			name:          "successful get",
			key:           "testkey",
			mockReturn:    persistence.KeyValue{Key: "testkey", Value: "testvalue"},
			mockError:     nil,
			expectedValue: "testvalue",
			expectError:   false,
		},
		{
			name:        "error from persistence",
			key:         "errorkey",
			mockReturn:  persistence.KeyValue{},
			mockError:   io.EOF,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPers := new(MockPersistence)
			mockPers.On("Get", mock.Anything, tt.key).Return(tt.mockReturn, tt.mockError)

			fs := NewFeatureService(mockPers)
			result, err := fs.Get(context.Background(), &featurev1.Key{Name: tt.key})

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedValue, result.Name)
			}

			mockPers.AssertExpectations(t)
		})
	}
}

func TestFeatureService_Set(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		value       string
		mockError   error
		expectError bool
	}{
		{
			name:        "successful set",
			key:         "key1",
			value:       "value1",
			mockError:   nil,
			expectError: false,
		},
		{
			name:        "error from persistence",
			key:         "key2",
			value:       "value2",
			mockError:   io.EOF,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPers := new(MockPersistence)
			expectedKV := persistence.KeyValue{Key: tt.key, Value: tt.value}
			mockPers.On("Set", mock.Anything, expectedKV).Return(tt.mockError)

			fs := NewFeatureService(mockPers)
			result, err := fs.Set(context.Background(), &featurev1.KeyValue{Key: tt.key, Value: tt.value})

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockPers.AssertExpectations(t)
		})
	}
}

func TestFeatureService_PreSet(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		value       string
		mockError   error
		expectError bool
	}{
		{
			name:        "successful preset",
			key:         "key1",
			value:       "value1",
			mockError:   nil,
			expectError: false,
		},
		{
			name:        "error from persistence",
			key:         "key2",
			value:       "value2",
			mockError:   io.EOF,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPers := new(MockPersistence)
			expectedKV := persistence.KeyValue{Key: tt.key, Value: tt.value}
			mockPers.On("PreSet", mock.Anything, expectedKV).Return(tt.mockError)

			fs := NewFeatureService(mockPers)
			result, err := fs.PreSet(context.Background(), &featurev1.KeyValue{Key: tt.key, Value: tt.value})

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockPers.AssertExpectations(t)
		})
	}
}

func TestFeatureService_Delete(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		mockError   error
		expectError bool
	}{
		{
			name:        "successful delete",
			key:         "key1",
			mockError:   nil,
			expectError: false,
		},
		{
			name:        "error from persistence",
			key:         "key2",
			mockError:   io.EOF,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPers := new(MockPersistence)
			mockPers.On("Delete", mock.Anything, tt.key).Return(tt.mockError)

			fs := NewFeatureService(mockPers)
			result, err := fs.Delete(context.Background(), &featurev1.Key{Name: tt.key})

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockPers.AssertExpectations(t)
		})
	}
}

func TestFeatureService_GetAll(t *testing.T) {
	tests := []struct {
		name           string
		mockReturn     []persistence.KeyValue
		mockError      error
		expectedCount  int
		expectError    bool
		sendError      error
		expectSendFail bool
	}{
		{
			name: "successful getall with multiple items",
			mockReturn: []persistence.KeyValue{
				{Key: "key1", Value: "value1"},
				{Key: "key2", Value: "value2"},
			},
			mockError:     nil,
			expectedCount: 2,
			expectError:   false,
		},
		{
			name:          "empty result",
			mockReturn:    []persistence.KeyValue{},
			mockError:     nil,
			expectedCount: 0,
			expectError:   false,
		},
		{
			name:        "error from persistence",
			mockReturn:  []persistence.KeyValue{},
			mockError:   io.EOF,
			expectError: true,
		},
		{
			name: "send error",
			mockReturn: []persistence.KeyValue{
				{Key: "key1", Value: "value1"},
			},
			mockError:      nil,
			expectedCount:  1,
			sendError:      io.EOF,
			expectSendFail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPers := new(MockPersistence)
			mockPers.On("GetAll", mock.Anything).Return(tt.mockReturn, tt.mockError)

			mockStream := &MockServerStream{
				ctx: context.Background(),
			}

			if !tt.expectError {
				for range tt.mockReturn {
					mockStream.On("Send", mock.Anything).Return(tt.sendError).Once()
				}
			}

			fs := NewFeatureService(mockPers)
			err := fs.GetAll(&emptypb.Empty{}, mockStream)

			if tt.expectError || tt.expectSendFail {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockPers.AssertExpectations(t)
			if !tt.expectError {
				mockStream.AssertExpectations(t)
			}
		})
	}
}
