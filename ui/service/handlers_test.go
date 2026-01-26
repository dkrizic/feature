package service

import (
	"context"
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	featurev1 "github.com/dkrizic/feature/ui/repository/feature/v1"
	metav1 "github.com/dkrizic/feature/ui/repository/meta/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

// MockFeatureClient is a mock for FeatureClient
type MockFeatureClient struct {
	mock.Mock
}

func (m *MockFeatureClient) GetAll(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (grpc.ServerStreamingClient[featurev1.KeyValue], error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(grpc.ServerStreamingClient[featurev1.KeyValue]), args.Error(1)
}

func (m *MockFeatureClient) Get(ctx context.Context, in *featurev1.Key, opts ...grpc.CallOption) (*featurev1.Value, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*featurev1.Value), args.Error(1)
}

func (m *MockFeatureClient) Set(ctx context.Context, in *featurev1.KeyValue, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (m *MockFeatureClient) PreSet(ctx context.Context, in *featurev1.KeyValue, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (m *MockFeatureClient) Delete(ctx context.Context, in *featurev1.Key, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

// MockMetaClient is a mock for MetaClient
type MockMetaClient struct {
	mock.Mock
}

func (m *MockMetaClient) Meta(ctx context.Context, in *metav1.MetaRequest, opts ...grpc.CallOption) (*metav1.MetaResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*metav1.MetaResponse), args.Error(1)
}

// MockStreamClient is a mock for grpc.ServerStreamingClient
type MockStreamClient struct {
	mock.Mock
	items []*featurev1.KeyValue
	index int
}

func (m *MockStreamClient) Recv() (*featurev1.KeyValue, error) {
	if m.index >= len(m.items) {
		return nil, io.EOF
	}
	item := m.items[m.index]
	m.index++
	return item, nil
}

func (m *MockStreamClient) Header() (metadata.MD, error) {
	return nil, nil
}

func (m *MockStreamClient) Trailer() metadata.MD {
	return nil
}

func (m *MockStreamClient) CloseSend() error {
	return nil
}

func (m *MockStreamClient) Context() context.Context {
	return context.Background()
}

func (m *MockStreamClient) SendMsg(msg interface{}) error {
	return nil
}

func (m *MockStreamClient) RecvMsg(msg interface{}) error {
	return nil
}

func TestHandleHealth(t *testing.T) {
	server := &Server{}

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()

	server.handleHealth(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, "OK", string(body))
}

func TestHandleIndex(t *testing.T) {
	// Create a minimal template for testing
	tmpl := template.Must(template.New("index.gohtml").Parse(`UI: {{.UIVersion}}, Backend: {{.BackendVersion}}`))

	server := &Server{
		templates:      tmpl,
		uiVersion:      "1.0.0",
		backendVersion: "2.0.0",
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	server.handleIndex(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Contains(t, string(body), "UI: 1.0.0")
	assert.Contains(t, string(body), "Backend: 2.0.0")
}

func TestHandleIndex_TemplateError(t *testing.T) {
	// Create an invalid template that will error during execution
	tmpl := template.Must(template.New("index.gohtml").Parse(`{{.InvalidField}}`))

	server := &Server{
		templates:      tmpl,
		uiVersion:      "1.0.0",
		backendVersion: "2.0.0",
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	server.handleIndex(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestHandleFeaturesList(t *testing.T) {
	tests := []struct {
		name               string
		mockItems          []*featurev1.KeyValue
		mockError          error
		expectedStatus     int
		expectedBodyChecks []string
	}{
		{
			name: "successful with features",
			mockItems: []*featurev1.KeyValue{
				{Key: "feature1", Value: "value1"},
				{Key: "feature2", Value: "value2"},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "empty features",
			mockItems:      []*featurev1.KeyValue{},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "error from backend",
			mockItems:      nil,
			mockError:      io.EOF,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFeatureClient := new(MockFeatureClient)
			
			if tt.mockError == nil {
				mockStream := &MockStreamClient{
					items: tt.mockItems,
					index: 0,
				}
				mockFeatureClient.On("GetAll", mock.Anything, mock.Anything).Return(mockStream, nil)
			} else {
				mockFeatureClient.On("GetAll", mock.Anything, mock.Anything).Return(nil, tt.mockError)
			}

			// Create a minimal template for testing
			tmpl := template.Must(template.New("features_list.gohtml").Parse(`Features: {{len .Features}}`))

			server := &Server{
				templates:     tmpl,
				featureClient: mockFeatureClient,
			}

			req := httptest.NewRequest(http.MethodGet, "/features/list", nil)
			w := httptest.NewRecorder()

			server.handleFeaturesList(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedStatus == http.StatusOK {
				body, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)
				assert.Contains(t, string(body), "Features:")
			}

			mockFeatureClient.AssertExpectations(t)
		})
	}
}

func TestHandleFeaturesList_AlphabeticalSorting(t *testing.T) {
	// Test that features are sorted alphabetically regardless of backend order
	mockFeatureClient := new(MockFeatureClient)
	
	// Provide features in non-alphabetical order
	mockStream := &MockStreamClient{
		items: []*featurev1.KeyValue{
			{Key: "zebra", Value: "value1"},
			{Key: "apple", Value: "value2"},
			{Key: "mango", Value: "value3"},
			{Key: "banana", Value: "value4"},
		},
		index: 0,
	}
	mockFeatureClient.On("GetAll", mock.Anything, mock.Anything).Return(mockStream, nil)

	// Create a template that outputs the keys in order
	tmpl := template.Must(template.New("features_list.gohtml").Parse(`{{range .Features}}{{.Key}},{{end}}`))

	server := &Server{
		templates:     tmpl,
		featureClient: mockFeatureClient,
	}

	req := httptest.NewRequest(http.MethodGet, "/features/list", nil)
	w := httptest.NewRecorder()

	server.handleFeaturesList(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	
	// Verify features are returned in alphabetical order
	assert.Equal(t, "apple,banana,mango,zebra,", string(body))

	mockFeatureClient.AssertExpectations(t)
}

func TestHandleFeatureDelete(t *testing.T) {
	tests := []struct {
		name           string
		formKey        string
		mockError      error
		expectedStatus int
		mockGetAll     bool
	}{
		{
			name:           "successful delete",
			formKey:        "feature1",
			mockError:      nil,
			expectedStatus: http.StatusOK,
			mockGetAll:     true,
		},
		{
			name:           "missing key",
			formKey:        "",
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			mockGetAll:     false,
		},
		{
			name:           "delete error",
			formKey:        "feature1",
			mockError:      io.EOF,
			expectedStatus: http.StatusInternalServerError,
			mockGetAll:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFeatureClient := new(MockFeatureClient)

			if tt.formKey != "" && tt.mockError == nil {
				mockFeatureClient.On("Delete", mock.Anything, &featurev1.Key{Name: tt.formKey}).Return(&emptypb.Empty{}, nil)
			} else if tt.formKey != "" && tt.mockError != nil {
				mockFeatureClient.On("Delete", mock.Anything, &featurev1.Key{Name: tt.formKey}).Return(nil, tt.mockError)
			}

			if tt.mockGetAll {
				mockStream := &MockStreamClient{
					items: []*featurev1.KeyValue{},
					index: 0,
				}
				mockFeatureClient.On("GetAll", mock.Anything, mock.Anything).Return(mockStream, nil)
			}

			// Create a minimal template for testing
			tmpl := template.Must(template.New("features_list.gohtml").Parse(`Features: {{len .Features}}`))

			server := &Server{
				templates:     tmpl,
				featureClient: mockFeatureClient,
			}

			form := url.Values{}
			if tt.formKey != "" {
				form.Add("key", tt.formKey)
			}

			req := httptest.NewRequest(http.MethodPost, "/features/delete", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()

			server.handleFeatureDelete(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			mockFeatureClient.AssertExpectations(t)
		})
	}
}

func TestHandleVersion(t *testing.T) {
	tests := []struct {
		name           string
		mockVersion    string
		mockError      error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "successful version fetch",
			mockVersion:    "v1.2.3",
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody:   "v1.2.3",
		},
		{
			name:           "error fetching version",
			mockVersion:    "",
			mockError:      io.EOF,
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMetaClient := new(MockMetaClient)

			if tt.mockError == nil {
				mockMetaClient.On("Meta", mock.Anything, mock.Anything).Return(&metav1.MetaResponse{Version: tt.mockVersion}, nil)
			} else {
				mockMetaClient.On("Meta", mock.Anything, mock.Anything).Return(nil, tt.mockError)
			}

			server := &Server{
				metaClient: mockMetaClient,
			}

			req := httptest.NewRequest(http.MethodGet, "/version", nil)
			w := httptest.NewRecorder()

			server.handleVersion(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBody, string(body))

			if tt.expectedBody != "" {
				assert.Equal(t, "text/plain", resp.Header.Get("Content-Type"))
			}

			mockMetaClient.AssertExpectations(t)
		})
	}
}

func TestRegisterHandlers(t *testing.T) {
	// Create minimal templates for testing
	tmpl := template.Must(template.New("index.gohtml").Parse(`Test`))
	template.Must(tmpl.New("features_list.gohtml").Parse(`Test`))

	mockFeatureClient := new(MockFeatureClient)
	mockMetaClient := new(MockMetaClient)

	server := &Server{
		templates:     tmpl,
		featureClient: mockFeatureClient,
		metaClient:    mockMetaClient,
	}
	mux := http.NewServeMux()

	server.registerHandlers(mux)

	// Test that routes are registered by making test requests
	testRoutes := []struct {
		method         string
		path           string
		skipBadRequest bool
	}{
		{http.MethodGet, "/", false},
		{http.MethodGet, "/healthz", false},
		// Skip these as they require mock setup
		// {http.MethodGet, "/features/list", true},
		// {http.MethodPost, "/features/delete", true},
	}

	for _, route := range testRoutes {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			req := httptest.NewRequest(route.method, route.path, nil)
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)

			// Should not return 404 if route is registered
			assert.NotEqual(t, http.StatusNotFound, w.Code, "Route should be registered")
		})
	}
}

// TestSubpathSupport verifies that routes work correctly with a subpath prefix.
func TestSubpathSupport(t *testing.T) {
	tests := []struct {
		name     string
		subpath  string
		testPath string
		expected int
	}{
		{
			name:     "no subpath - root",
			subpath:  "",
			testPath: "/",
			expected: http.StatusOK,
		},
		{
			name:     "no subpath - health",
			subpath:  "",
			testPath: "/health",
			expected: http.StatusOK,
		},
		{
			name:     "with subpath /feature - root",
			subpath:  "/feature",
			testPath: "/feature/",
			expected: http.StatusOK,
		},
		{
			name:     "with subpath /feature - health",
			subpath:  "/feature",
			testPath: "/feature/health",
			expected: http.StatusOK,
		},
		{
			name:     "with subpath /app/v1 - root",
			subpath:  "/app/v1",
			testPath: "/app/v1/",
			expected: http.StatusOK,
		},
		{
			name:     "with subpath /app/v1 - health",
			subpath:  "/app/v1",
			testPath: "/app/v1/health",
			expected: http.StatusOK,
		},
		{
			name:     "with subpath - wrong path should 404",
			subpath:  "/feature",
			testPath: "/",
			expected: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFeatureClient := new(MockFeatureClient)
			mockMetaClient := new(MockMetaClient)

			tmpl := template.Must(template.New("index.gohtml").Parse(`<html><body>Subpath: {{.Subpath}}</body></html>`))

			server := &Server{
				subpath:       tt.subpath,
				templates:     tmpl,
				featureClient: mockFeatureClient,
				metaClient:    mockMetaClient,
				uiVersion:     "test",
			}

			mux := http.NewServeMux()
			server.registerHandlers(mux)

			req := httptest.NewRequest(http.MethodGet, tt.testPath, nil)
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)

			assert.Equal(t, tt.expected, w.Code, "Expected status code %d, got %d", tt.expected, w.Code)

			// If we expect OK and it's the index page, verify subpath is in the response
			if tt.expected == http.StatusOK && (tt.testPath == "/" || tt.testPath == "/feature/" || tt.testPath == "/app/v1/") {
				body, _ := io.ReadAll(w.Body)
				assert.Contains(t, string(body), "Subpath: "+tt.subpath)
			}
		})
	}
}
