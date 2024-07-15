package telemetry

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

func CreateMeter() metric.Meter {
	return otel.Meter("main")
}
