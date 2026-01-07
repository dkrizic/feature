package localmetrics

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

var (
	getAllCounter metric.Int64Counter
	getCounter    metric.Int64Counter
	activeGauge   metric.Int64Gauge
	setCounter    metric.Int64Counter
	presetCounter metric.Int64Counter
	deleteCounter metric.Int64Counter
)

func New() error {
	// create access counter metric
	var err error
	getAllCounter, err = otel.Meter("telemetry/localmetrics").Int64Counter("feature.getall.count",
		metric.WithDescription("Number of GetAll requests"),
		metric.WithUnit("count"))
	if err != nil {
		return err
	}

	getCounter, err = otel.Meter("telemetry/localmetrics").Int64Counter("feature.get.count",
		metric.WithDescription("Number of Get requests"),
		metric.WithUnit("count"))
	if err != nil {
		return err
	}

	setCounter, err = otel.Meter("telemetry/localmetrics").Int64Counter("feature.set.count",
		metric.WithDescription("Number of Set requests"),
		metric.WithUnit("count"))
	if err != nil {
		return err
	}

	presetCounter, err = otel.Meter("telemetry/localmetrics").Int64Counter("feature.preset.count",
		metric.WithDescription("Number of PreSet requests"),
		metric.WithUnit("count"))
	if err != nil {
		return err
	}

	// gauge for active features
	activeGauge, err = otel.Meter("telemetry/localmetrics").Int64Gauge("feature.active.gauge",
		metric.WithDescription("Number of active features"),
		metric.WithUnit("count"))
	if err != nil {
		return err
	}

	// counter for deleted features
	deleteCounter, err = otel.Meter("telemetry/localmetrics").Int64Counter("feature.delete.count",
		metric.WithDescription("Number of deleted features"),
		metric.WithUnit("count"))
	if err != nil {
		return err
	}

	return nil
}

func GetAllCounter() metric.Int64Counter {
	return getAllCounter
}

func GetCounter() metric.Int64Counter {
	return getCounter
}

func ActiveGauge() metric.Int64Gauge {
	return activeGauge
}

func SetCounter() metric.Int64Counter {
	return setCounter
}

func PresetCounter() metric.Int64Counter {
	return presetCounter
}

func DeleteCounter() metric.Int64Counter {
	return deleteCounter
}
