package configmap

import (
	"context"
	"testing"

	"github.com/dkrizic/feature/service/service/persistence"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// fakeConfigMapClient implements configMapClient interface for testing
type fakeConfigMapClient struct {
	configMaps map[string]*v1.ConfigMap
}

func (f *fakeConfigMapClient) Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.ConfigMap, error) {
	cm, exists := f.configMaps[name]
	if !exists {
		return nil, errors.NewNotFound(schema.GroupResource{Group: "", Resource: "configmaps"}, name)
	}
	// Return a copy to avoid test interference
	return cm.DeepCopy(), nil
}

func (f *fakeConfigMapClient) Create(ctx context.Context, configMap *v1.ConfigMap, opts metav1.CreateOptions) (*v1.ConfigMap, error) {
	if _, exists := f.configMaps[configMap.Name]; exists {
		return nil, errors.NewAlreadyExists(schema.GroupResource{Group: "", Resource: "configmaps"}, configMap.Name)
	}
	// Store a copy
	cmCopy := configMap.DeepCopy()
	f.configMaps[configMap.Name] = cmCopy
	return cmCopy, nil
}

func (f *fakeConfigMapClient) Update(ctx context.Context, configMap *v1.ConfigMap, opts metav1.UpdateOptions) (*v1.ConfigMap, error) {
	if _, exists := f.configMaps[configMap.Name]; !exists {
		return nil, errors.NewNotFound(schema.GroupResource{Group: "", Resource: "configmaps"}, configMap.Name)
	}
	// Store a copy
	cmCopy := configMap.DeepCopy()
	f.configMaps[configMap.Name] = cmCopy
	return cmCopy, nil
}

// setupFakeK8s sets up fake k8s client for tests and returns the fake client for seeding data
func setupFakeK8s(namespace string) *fakeConfigMapClient {
	fakeClient := &fakeConfigMapClient{
		configMaps: make(map[string]*v1.ConfigMap),
	}

	// Save original functions to restore later
	originalK8sClientFn := k8sClientFn
	originalOwnNamespaceFn := ownNamespaceFn

	// Override the injectable functions
	k8sClientFn = func(ctx context.Context, configMapName string) (configMapClient, *string, error) {
		ns := namespace
		return fakeClient, &ns, nil
	}

	ownNamespaceFn = func(ctx context.Context) (*string, error) {
		ns := namespace
		return &ns, nil
	}

	// Store original functions to potentially restore (though in unit tests this usually doesn't matter)
	_ = originalK8sClientFn
	_ = originalOwnNamespaceFn

	return fakeClient
}

func TestNewConfigMapPersistence(t *testing.T) {
	configMapName := "test-configmap"
	p := NewConfigMapPersistence(configMapName)

	assert.NotNil(t, p)
	assert.Equal(t, configMapName, p.configMapName)
}

func TestConfigMapPersistence_GetAll(t *testing.T) {
	fakeClient := setupFakeK8s("test-namespace")
	ctx := context.Background()

	// Seed data
	fakeClient.configMaps["test-configmap"] = &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-configmap",
		},
		Data: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}

	p := NewConfigMapPersistence("test-configmap")
	result, err := p.GetAll(ctx)

	assert.NoError(t, err)
	assert.Len(t, result, 2)

	// Convert to map for easier assertion
	resultMap := make(map[string]string)
	for _, kv := range result {
		resultMap[kv.Key] = kv.Value
	}
	assert.Equal(t, "value1", resultMap["key1"])
	assert.Equal(t, "value2", resultMap["key2"])
}

func TestConfigMapPersistence_GetAll_Empty(t *testing.T) {
	setupFakeK8s("test-namespace")
	ctx := context.Background()

	p := NewConfigMapPersistence("test-configmap")
	result, err := p.GetAll(ctx)

	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestConfigMapPersistence_PreSet(t *testing.T) {
	setupFakeK8s("test-namespace")
	ctx := context.Background()
	p := NewConfigMapPersistence("test-configmap")

	// PreSet should set the value when key doesn't exist
	err := p.PreSet(ctx, persistence.KeyValue{Key: "key1", Value: "value1"})
	assert.NoError(t, err)

	// Verify the value was set
	kv, err := p.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Equal(t, "key1", kv.Key)
	assert.Equal(t, "value1", kv.Value)

	// PreSet should not change the value when key already exists with different value
	err = p.PreSet(ctx, persistence.KeyValue{Key: "key1", Value: "newvalue"})
	assert.NoError(t, err)

	// Verify the value was NOT changed
	kv, err = p.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", kv.Value)
}

func TestConfigMapPersistence_Set(t *testing.T) {
	setupFakeK8s("test-namespace")
	ctx := context.Background()
	p := NewConfigMapPersistence("test-configmap")

	// Set should set the value
	err := p.Set(ctx, persistence.KeyValue{Key: "key1", Value: "value1"})
	assert.NoError(t, err)

	// Verify the value was set
	kv, err := p.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Equal(t, "key1", kv.Key)
	assert.Equal(t, "value1", kv.Value)

	// Set should update the value when key already exists
	err = p.Set(ctx, persistence.KeyValue{Key: "key1", Value: "newvalue"})
	assert.NoError(t, err)

	// Verify the value was changed
	kv, err = p.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Equal(t, "newvalue", kv.Value)
}

func TestConfigMapPersistence_Get(t *testing.T) {
	fakeClient := setupFakeK8s("test-namespace")
	ctx := context.Background()

	// Seed data
	fakeClient.configMaps["test-configmap"] = &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-configmap",
		},
		Data: map[string]string{
			"key1": "value1",
		},
	}

	p := NewConfigMapPersistence("test-configmap")
	result, err := p.Get(ctx, "key1")

	assert.NoError(t, err)
	assert.Equal(t, "key1", result.Key)
	assert.Equal(t, "value1", result.Value)
}

func TestConfigMapPersistence_Get_NotFound(t *testing.T) {
	fakeClient := setupFakeK8s("test-namespace")
	ctx := context.Background()

	// Seed data with empty ConfigMap
	fakeClient.configMaps["test-configmap"] = &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-configmap",
		},
		Data: map[string]string{},
	}

	p := NewConfigMapPersistence("test-configmap")
	result, err := p.Get(ctx, "nonexistent")

	assert.Error(t, err)
	assert.Equal(t, persistence.ErrKeyNotFound, err)
	assert.Equal(t, "", result.Key)
	assert.Equal(t, "", result.Value)
}

func TestConfigMapPersistence_Delete(t *testing.T) {
	fakeClient := setupFakeK8s("test-namespace")
	ctx := context.Background()

	// Seed data
	fakeClient.configMaps["test-configmap"] = &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-configmap",
		},
		Data: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}

	p := NewConfigMapPersistence("test-configmap")

	// Delete a key
	err := p.Delete(ctx, "key1")
	assert.NoError(t, err)

	// Verify key was deleted
	_, err = p.Get(ctx, "key1")
	assert.Error(t, err)
	assert.Equal(t, persistence.ErrKeyNotFound, err)

	// Verify other key still exists
	kv, err := p.Get(ctx, "key2")
	assert.NoError(t, err)
	assert.Equal(t, "value2", kv.Value)
}

func TestConfigMapPersistence_Count(t *testing.T) {
	fakeClient := setupFakeK8s("test-namespace")
	ctx := context.Background()

	// Seed data
	fakeClient.configMaps["test-configmap"] = &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-configmap",
		},
		Data: map[string]string{
			"key1": "value1",
			"key2": "value2",
			"key3": "value3",
		},
	}

	p := NewConfigMapPersistence("test-configmap")
	count, err := p.Count(ctx)

	assert.NoError(t, err)
	assert.Equal(t, 3, count)
}
