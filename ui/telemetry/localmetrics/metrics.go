package localmetrics

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

var (
	accessCount  metric.Int64Counter
	presetCount  metric.Int64Counter
	setCount     metric.Int64Counter
	activeGauge  metric.Int64UpDownCounter
	deletedCount metric.Int64Counter
)

func New() error {
	// create access counter metric
	var err error
	accessCount, err = otel.Meter("telemetry/localmetrics").Int64Counter("feature.access.count",
		metric.WithDescription("Number of read accesses"),
		metric.WithUnit("count"))
	if err != nil {
		return err
	}

	// counter for preset actions
	presetCount, err = otel.Meter("telemetry/localmetrics").Int64Counter("feature.preset.count",
		metric.WithDescription("Number of preset actions"),
		metric.WithUnit("count"))
	if err != nil {
		return err
	}

	setCount, err = otel.Meter("telemetry/localmetrics").Int64Counter("feature.set.count",
		metric.WithDescription("Number of set actions"),
		metric.WithUnit("count"))
	if err != nil {
		return err
	}

	// gauge for active features
	activeGauge, err = otel.Meter("telemetry/localmetrics").Int64UpDownCounter("feature.active.gauge",
		metric.WithDescription("Number of active features"),
		metric.WithUnit("count"))
	if err != nil {
		return err
	}

	// counter

	return nil
}

func AccessCount() metric.Int64Counter {
	return accessCount
}

func ActiveGauge() metric.Int64UpDownCounter {
	return activeGauge
}

func DeletedCounter() metric.Int64Counter {
	return deletedCount
}

func PreSetCount() metric.Int64Counter {
	return presetCount
}

func SetCount() metric.Int64Counter {
	return setCount
}
