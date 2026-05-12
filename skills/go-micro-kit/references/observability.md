# Observability

## Telemetry Setup

```go
otelShutdown, err := telemetry.SetupTelemetry(ctx, cfg.Telemetry)
if err != nil {
    panic(err)
}
defer func() {
    err = errors.Join(err, otelShutdown(context.Background()))
}()

log := telemetry.CreateLoggerWithTelemetry()
tracer := telemetry.CreateTracer()
meter := telemetry.CreateMeter()
```

`SetupTelemetry` configures:

- W3C trace context and baggage propagation.
- trace provider with OTLP gRPC or HTTP exporter.
- meter provider with periodic OTLP export.
- log provider with OTLP gRPC or HTTP exporter.

## Protocol

`telemetry.Config` supports:

- `grpc`
- `http/json`

With the recommended `env:",prefix=OTEL_"` config field, set:

```env
OTEL_EXPORTER_OTLP_PROTOCOL=grpc
OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317
OTEL_SERVICE_NAME=item-service
```

For HTTP:

```env
OTEL_EXPORTER_OTLP_PROTOCOL=http/json
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318
OTEL_SERVICE_NAME=item-service
```

## Logger Choices

Use `telemetry.CreateLoggerWithTelemetry()` for services that already call `SetupTelemetry`. It fans out to:

- JSON stdout.
- OTel slog bridge.
- Sentry handler for error level.

Use `logger.CreateLogger()` only when the application wants JSON stdout plus Sentry without the OTel slog bridge.

## Handler Spans

```go
ctx, span := h.tracer.Start(ctx, "handler.ItemService.CreateItem")
defer span.End()

if err != nil {
    span.RecordError(err)
}
```

Use the returned `ctx` for repository and service calls.

## Built-in Instrumentation

- gRPC server: `otelgrpc.NewServerHandler`.
- gRPC client inside HTTP gateway: `otelgrpc.NewClientHandler`.
- HTTP gateway: `otelhttp.NewHandler`.
- GORM: `otelgorm.NewPlugin`.
- Cron tasks: span named `cron.<task-name>` and histogram `cron.<task-name>.duration`.

