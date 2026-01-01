package telemetry

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/exemplar"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv/v1.21.0"
)

// OpenTelemetryConfig holds config values
type OpenTelemetryConfig struct {
	ServiceName    string
	ServiceVersion string
	OTLPEndpoint   string
}

var serviceName string

func (config OpenTelemetryConfig) InitOpenTelemetry(ctx context.Context) (shutdown func(ctx context.Context) error, err error) {
	var shutdownFuncs []func(ctx context.Context) error

	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	res, err := newResource(config.ServiceName, config.ServiceVersion)
	if err != nil {
		handleErr(err)
		return
	}

	// --- Propagators ---
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	// --- Tracing ---
	traceProvider, err := newTraceProvider(res, config.OTLPEndpoint)
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, traceProvider.Shutdown)

	// --- Metrics ---
	meterProvider, err := newMeterProvider(res, config.OTLPEndpoint)
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	// ✅ Set global MeterProvider before starting runtime metrics

	return shutdown, nil
}

func newResource(serviceName, serviceVersion string) (*resource.Resource, error) {
	return resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(serviceName),
		semconv.ServiceVersion(serviceVersion),
	), nil
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

// --- Modern TraceProvider ---
func newTraceProvider(res *resource.Resource, url string) (*trace.TracerProvider, error) {
	traceExporter, err := otlptracegrpc.New(
		context.Background(),
		otlptracegrpc.WithEndpoint(url),
		otlptracegrpc.WithInsecure(), // replace with TLS if needed
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create trace exporter: %w", err)
	}

	traceProvider := trace.NewTracerProvider(
		trace.WithSampler(trace.ParentBased(trace.TraceIDRatioBased(0.1))),
		trace.WithBatcher(traceExporter, trace.WithBatchTimeout(5*time.Second)),
		trace.WithResource(res),
	)
	otel.SetTracerProvider(traceProvider)

	return traceProvider, nil
}

// --- Modern MeterProvider with runtime metrics ---
func newMeterProvider(res *resource.Resource, url string) (*sdkmetric.MeterProvider, error) {
	metricExporter, err := otlpmetricgrpc.New(
		context.Background(),
		otlpmetricgrpc.WithEndpoint(url),
		otlpmetricgrpc.WithInsecure(), // replace with TLS if needed
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create metric exporter: %w", err)
	}

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(
			sdkmetric.NewPeriodicReader(metricExporter,
				sdkmetric.WithInterval(60*time.Second), // push interval
			),
		),
		sdkmetric.WithExemplarFilter(exemplar.TraceBasedFilter),
	)

	otel.SetMeterProvider(meterProvider)
	// ✅ Attach runtime metrics to this MeterProvider
	if err := runtime.Start(
		runtime.WithMeterProvider(meterProvider),
		runtime.WithMinimumReadMemStatsInterval(30*time.Second),
	); err != nil {
		return nil, fmt.Errorf("unable to start runtime metrics: %w", err)
	}

	return meterProvider, nil
}
