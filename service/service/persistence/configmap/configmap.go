package configmap

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/dkrizic/feature/service/service/persistence"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// configMapClient is an interface for ConfigMap operations to allow testing
type configMapClient interface {
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.ConfigMap, error)
	Create(ctx context.Context, configMap *v1.ConfigMap, opts metav1.CreateOptions) (*v1.ConfigMap, error)
	Update(ctx context.Context, configMap *v1.ConfigMap, opts metav1.UpdateOptions) (*v1.ConfigMap, error)
}

type Persistence struct {
	configMapName string
}

// Injectable function variables for testing
var (
	k8sClientFn    func(context.Context, string) (configMapClient, *string, error) = k8sClient
	ownNamespaceFn                                                                 = ownNamespace
)

func NewPersistence(configMapName string) *Persistence {
	return &Persistence{
		configMapName: configMapName,
	}
}

func (p *Persistence) GetAll(ctx context.Context) ([]persistence.KeyValue, error) {
	ctx, span := otel.Tracer("service/persistence/configmap").Start(ctx, "GetAll")
	defer span.End()

	configMap, err := p.createOrLoadConfigMap(ctx)
	if err != nil {
		return nil, err
	}

	var keyValues []persistence.KeyValue
	for key, value := range configMap.Data {
		keyValues = append(keyValues, persistence.KeyValue{
			Key:   key,
			Value: value,
		})
	}
	return keyValues, nil
}

func (p *Persistence) PreSet(ctx context.Context, kv persistence.KeyValue) error {
	ctx, span := otel.Tracer("service/persistence/configmap").Start(ctx, "PreSet")
	defer span.End()

	configMap, err := p.createOrLoadConfigMap(ctx)
	if err != nil {
		return err
	}

	if configMap.Data == nil {
		configMap.Data = make(map[string]string)
	}

	// Only set if key doesn't exist
	if _, exists := configMap.Data[kv.Key]; exists {
		// do not change if there is already a value
		return nil
	}

	configMap.Data[kv.Key] = kv.Value
	return p.saveConfigMap(ctx, *configMap)
}

func (p *Persistence) Set(ctx context.Context, kv persistence.KeyValue) error {
	ctx, span := otel.Tracer("service/persistence/configmap").Start(ctx, "Set")
	defer span.End()

	configMap, err := p.createOrLoadConfigMap(ctx)
	if err != nil {
		return err
	}

	if configMap.Data == nil {
		configMap.Data = make(map[string]string)
	}
	configMap.Data[kv.Key] = kv.Value

	err = p.saveConfigMap(ctx, *configMap)
	return err
}

func (p *Persistence) Get(ctx context.Context, key string) (persistence.KeyValue, error) {
	ctx, span := otel.Tracer("service/persistence/configmap").Start(ctx, "Get")
	defer span.End()

	configMap, err := p.createOrLoadConfigMap(ctx)
	if err != nil {
		return persistence.KeyValue{}, err
	}
	value, exists := configMap.Data[key]
	if !exists {
		return persistence.KeyValue{}, persistence.ErrKeyNotFound
	}
	return persistence.KeyValue{
		Key:   key,
		Value: value,
	}, nil
}

func (p *Persistence) Delete(ctx context.Context, key string) error {
	ctx, span := otel.Tracer("service/persistence/configmap").Start(ctx, "Delete")
	defer span.End()

	configMap, err := p.createOrLoadConfigMap(ctx)
	if err != nil {
		return err
	}
	delete(configMap.Data, key)

	err = p.saveConfigMap(ctx, *configMap)
	return err
}

func (p *Persistence) createOrLoadConfigMap(ctx context.Context) (*v1.ConfigMap, error) {
	ctx, span := otel.Tracer("service/persistence/configmap").Start(ctx, "createOrLoadConfigMap")
	defer span.End()

	configMapClient, namespace, err := k8sClientFn(ctx, p.configMapName)
	if err != nil {
		return nil, err
	}
	slog.DebugContext(ctx, "Running in namespace", "namespace", *namespace)

	configMap, err := configMapClient.Get(ctx, p.configMapName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			// ConfigMap does not exist, create it
			newConfigMap := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: p.configMapName,
				},
				Data: map[string]string{},
			}
			createdConfigMap, err := configMapClient.Create(ctx, newConfigMap, metav1.CreateOptions{})
			if err != nil {
				return nil, err
			}
			return createdConfigMap, nil
		} else {
			return nil, err
		}
	}
	return configMap, nil
}

func k8sClient(ctx context.Context, configMapName string) (client configMapClient, namespace *string, err error) {
	ctx, span := otel.Tracer("service/persistence/configmap").Start(ctx, "k8sClient")
	defer span.End()

	rc, err := rest.InClusterConfig()
	if err != nil {
		return nil, nil, err
	}

	// get own namespace BEFORE using it in WrapTransport
	namespace, err = ownNamespaceFn(ctx)
	if err != nil {
		return nil, nil, err
	}

	// Instrument the transport with otelhttp and set peer.service to "kubernetes"
	rc.WrapTransport = func(rt http.RoundTripper) http.RoundTripper {
		return otelhttp.NewTransport(rt,
			otelhttp.WithSpanOptions(
				trace.WithSpanKind(trace.SpanKindClient),
				trace.WithAttributes(
					attribute.String("peer.service", "kubernetes"),
					attribute.String("namespace", *namespace),
					attribute.String("configmap", configMapName),
				),
			),
		)
	}

	// use Kubernetes API to load ConfigMap
	clientset, err := kubernetes.NewForConfig(rc)
	if err != nil {
		return nil, nil, err
	}

	return clientset.CoreV1().ConfigMaps(*namespace), namespace, nil
}

func (p *Persistence) saveConfigMap(ctx context.Context, configMap v1.ConfigMap) error {
	ctx, span := otel.Tracer("service/persistence/configmap").Start(ctx, "saveConfigMap")
	defer span.End()

	configMapClient, _, err := k8sClientFn(ctx, p.configMapName)
	if err != nil {
		return err
	}

	_, err = configMapClient.Update(ctx, &configMap, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func ownNamespace(ctx context.Context) (namespace *string, err error) {
	ctx, span := otel.Tracer("service/persistence/configmap").Start(ctx, "ownNamespace")
	defer span.End()

	data, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		return nil, err
	}
	ns := string(data)

	span.SetAttributes(attribute.String("namespace", ns))
	return &ns, nil
}

func (p *Persistence) Count(ctx context.Context) (int, error) {
	ctx, span := otel.Tracer("service/persistence/configmap").Start(ctx, "Count")
	defer span.End()

	configMap, err := p.createOrLoadConfigMap(ctx)
	if err != nil {
		return 0, err
	}
	count := len(configMap.Data)
	return count, nil
}
