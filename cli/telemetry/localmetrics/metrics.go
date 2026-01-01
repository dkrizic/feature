package localmetrics

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

var (
	accessCount    metric.Int64Counter
	activeGauge    metric.Int64Gauge
	deletedCounter metric.Int64Counter
)

func New() error {
	// create access counter metric
	var err error
	accessCount, err = otel.Meter("telemetry/localmetrics").Int64Counter("yasm.external.access.count",
		metric.WithDescription("Number of accesses to external service"),
		metric.WithUnit("count"))
	if err != nil {
		return err
	}

	activeGauge, err = otel.Meter("telemetry/localmetrics").Int64Gauge("yasm.external.profile.remaining.hours",
		metric.WithDescription("Remaining hours for active profiles"),
		metric.WithUnit("hours"))
	if err != nil {
		return err
	}

	deletedCounter, err = otel.Meter("telemetry/localmetrics").Int64Counter("yasm.external.profile.deleted.count",
		metric.WithDescription("Number of deleted profiles"),
		metric.WithUnit("count"))
	if err != nil {
		return err
	}

	return nil
}

func AccessCount() metric.Int64Counter {
	return accessCount
}

func ActiveGauge() metric.Int64Gauge {
	return activeGauge
}

func DeletedCounter() metric.Int64Counter {
	return deletedCounter
}
