package telemetry

import (
	"context"
	"errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	"time"
)

func SetupTelemetry(ctx context.Context, cfg Config) (shutdown func(context.Context) error, err error) {
	var shutdownFuncs []func(context.Context) error

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	// handleErr calls shutdown for cleanup and makes sure that all errors are returned.
	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	// Set up propagator.
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	// Set up trace provider.
	tracerProvider, err := newTraceProvider(ctx, cfg)
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	// Set up meter provider.
	meterProvider, err := newMeterProvider(ctx, cfg)
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)

	// Set up logger provider.
	loggerProvider, err := newLoggerProvider(ctx, cfg)
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, loggerProvider.Shutdown)
	global.SetLoggerProvider(loggerProvider)

	return
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newTraceProvider(ctx context.Context, cfg Config) (*trace.TracerProvider, error) {
	res, err := resource.New(
		ctx,
		resource.WithFromEnv(),      // Discover and provide attributes from OTEL_RESOURCE_ATTRIBUTES and OTEL_SERVICE_NAME environment variables.
		resource.WithTelemetrySDK(), // Discover and provide information about the OpenTelemetry SDK used.
		resource.WithProcess(),      // Discover and provide process information.
		resource.WithOS(),           // Discover and provide OS information.
		resource.WithContainer(),    // Discover and provide container information.
		resource.WithHost(),         // Discover and provide host information.
	)

	if err != nil {
		return nil, err
	}

	opts := []trace.TracerProviderOption{
		trace.WithResource(res),
	}

	switch {
	case cfg.IsProtocolHttp():
		traceExporter, err := otlptracehttp.New(ctx)
		if err != nil {
			return nil, err
		}

		opts = append(opts, trace.WithBatcher(
			traceExporter,
			// Default is 5s. Set to 1s for demonstrative purposes.
			trace.WithBatchTimeout(time.Second),
		))
	case cfg.IsProtocolGrpc():
		traceExporter, err := otlptracegrpc.New(ctx)
		if err != nil {
			return nil, err
		}

		opts = append(opts, trace.WithBatcher(
			traceExporter,
			// Default is 5s. Set to 1s for demonstrative purposes.
			trace.WithBatchTimeout(time.Second),
		))
	}
	return trace.NewTracerProvider(opts...), nil
}

func newMeterProvider(ctx context.Context, cfg Config) (*metric.MeterProvider, error) {
	var opts []metric.Option

	switch {
	case cfg.IsProtocolHttp():
		metricExporter, err := otlpmetrichttp.New(ctx)
		if err != nil {
			return nil, err
		}

		opts = append(opts, metric.WithReader(
			metric.NewPeriodicReader(
				metricExporter,
				// Default is 1m. Set to 3s for demonstrative purposes.
				metric.WithInterval(3*time.Second),
			),
		))
	case cfg.IsProtocolGrpc():
		metricExporter, err := otlpmetricgrpc.New(ctx)
		if err != nil {
			return nil, err
		}

		opts = append(opts, metric.WithReader(
			metric.NewPeriodicReader(
				metricExporter,
				// Default is 1m. Set to 3s for demonstrative purposes.
				metric.WithInterval(3*time.Second),
			),
		))
	}

	return metric.NewMeterProvider(opts...), nil
}

func newLoggerProvider(ctx context.Context, cfg Config) (*log.LoggerProvider, error) {
	var opts []log.LoggerProviderOption

	switch {
	case cfg.IsProtocolHttp():
		httpExporter, err := otlploghttp.New(ctx)
		if err != nil {
			return nil, err
		}

		opts = append(opts, log.WithProcessor(log.NewSimpleProcessor(httpExporter)))
	case cfg.IsProtocolGrpc():
		grpcExporter, err := otlploggrpc.New(ctx)
		if err != nil {
			return nil, err
		}
		opts = append(opts, log.WithProcessor(log.NewSimpleProcessor(grpcExporter)))
	}

	return log.NewLoggerProvider(opts...), nil
}
