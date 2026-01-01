package localmetrics

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func TestNew(t *testing.T) {
	// Set up a test meter provider
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	otel.SetMeterProvider(provider)

	err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Verify that metrics are created
	if AccessCount() == nil {
		t.Error("AccessCount() returned nil")
	}

	if ActiveGauge() == nil {
		t.Error("ActiveGauge() returned nil")
	}

	if DeletedCounter() == nil {
		t.Error("DeletedCounter() returned nil")
	}
}

func TestAccessCount(t *testing.T) {
	// Set up a test meter provider
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	otel.SetMeterProvider(provider)

	err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	counter := AccessCount()
	if counter == nil {
		t.Fatal("AccessCount() returned nil")
	}

	// Test adding to counter
	ctx := context.Background()
	counter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("path", "/test"),
		attribute.String("remote_ip", "127.0.0.1"),
	))

	// Collect metrics
	var rm metricdata.ResourceMetrics
	err = reader.Collect(ctx, &rm)
	if err != nil {
		t.Fatalf("Failed to collect metrics: %v", err)
	}

	// Verify metric was recorded
	found := false
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name == "yasm.external.access.count" {
				found = true
				break
			}
		}
	}

	if !found {
		t.Error("Access count metric was not recorded")
	}
}

func TestActiveGauge(t *testing.T) {
	// Set up a test meter provider
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	otel.SetMeterProvider(provider)

	err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	gauge := ActiveGauge()
	if gauge == nil {
		t.Fatal("ActiveGauge() returned nil")
	}

	// Test recording to gauge
	ctx := context.Background()
	gauge.Record(ctx, 42, metric.WithAttributes(
		attribute.String("directory", "test-dir"),
	))

	// Collect metrics
	var rm metricdata.ResourceMetrics
	err = reader.Collect(ctx, &rm)
	if err != nil {
		t.Fatalf("Failed to collect metrics: %v", err)
	}

	// Verify metric was recorded
	found := false
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name == "yasm.external.profile.remaining.hours" {
				found = true
				break
			}
		}
	}

	if !found {
		t.Error("Active gauge metric was not recorded")
	}
}

func TestDeletedCounter(t *testing.T) {
	// Set up a test meter provider
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	otel.SetMeterProvider(provider)

	err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	counter := DeletedCounter()
	if counter == nil {
		t.Fatal("DeletedCounter() returned nil")
	}

	// Test adding to counter
	ctx := context.Background()
	counter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("directory", "test-dir"),
	))

	// Collect metrics
	var rm metricdata.ResourceMetrics
	err = reader.Collect(ctx, &rm)
	if err != nil {
		t.Fatalf("Failed to collect metrics: %v", err)
	}

	// Verify metric was recorded
	found := false
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name == "yasm.external.profile.deleted.count" {
				found = true
				break
			}
		}
	}

	if !found {
		t.Error("Deleted counter metric was not recorded")
	}
}

func TestMetricsCanBeCalledMultipleTimes(t *testing.T) {
	// Set up a test meter provider
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	otel.SetMeterProvider(provider)

	err := New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	ctx := context.Background()

	// Call AccessCount multiple times
	for i := 0; i < 5; i++ {
		counter := AccessCount()
		if counter == nil {
			t.Fatalf("AccessCount() returned nil on iteration %d", i)
		}
		counter.Add(ctx, 1, metric.WithAttributes(
			attribute.String("path", "/test"),
		))
	}

	// Call ActiveGauge multiple times
	for i := 0; i < 5; i++ {
		gauge := ActiveGauge()
		if gauge == nil {
			t.Fatalf("ActiveGauge() returned nil on iteration %d", i)
		}
		gauge.Record(ctx, int64(i*10), metric.WithAttributes(
			attribute.String("directory", "test"),
		))
	}

	// Call DeletedCounter multiple times
	for i := 0; i < 5; i++ {
		counter := DeletedCounter()
		if counter == nil {
			t.Fatalf("DeletedCounter() returned nil on iteration %d", i)
		}
		counter.Add(ctx, 1, metric.WithAttributes(
			attribute.String("directory", "test"),
		))
	}
}
