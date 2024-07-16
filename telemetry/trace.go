package telemetry

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

func CreateTracer() trace.Tracer {
	return otel.Tracer("main")
}
