# Server Bootstrap

## Canonical main.go Shape

```go
package main

import (
    "context"
    "errors"
    "os"
    "os/signal"
    "syscall"

    "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
    configpkg "github.com/updevru/go-micro-kit/config"
    "github.com/updevru/go-micro-kit/server"
    "github.com/updevru/go-micro-kit/telemetry"
    "google.golang.org/grpc"
)

func main() {
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()

    var cfg config.Config
    if err := configpkg.CreateConfig(ctx, &cfg); err != nil {
        panic(err)
    }

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

    app := server.NewServer(log, tracer, meter)

    app.Grpc(&cfg.Grpc, []grpc.ServerOption{}, func(g *grpc.Server) {
        // pb.RegisterItemServiceServer(g, itemHandler)
    })

    app.Http(&cfg.Http, &cfg.Grpc, []runtime.ServeMuxOption{},
        func(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
            // return pb.RegisterItemServiceHandler(ctx, mux, conn)
            return nil
        },
    )

    if err := app.Run(ctx); err != nil {
        log.ErrorContext(ctx, "failed to run server", "error", err)
        panic(err)
    }
}
```

## Registration Order

1. Load config.
2. Setup telemetry and create logger, tracer, meter.
3. Connect database and migrate if the service uses persistence.
4. Construct repositories, services, handlers.
5. Create `server.NewServer`.
6. Register gRPC with `app.Grpc`.
7. Register HTTP gateway with `app.Http`.
8. Register cron tasks with `app.Cron` if needed.
9. Register discovery with `app.AddDiscovery` if needed.
10. Call `app.Run(ctx)`.

## Server Behavior

- `Grpc` adds OpenTelemetry gRPC server stats handler and gRPC health service.
- `Http` creates a gRPC client to `:<cfgGrpc.Port>`, builds a `runtime.ServeMux`, adds health, CORS, JSON marshaling options, and OTel HTTP middleware.
- `Cron` starts a `gocron` scheduler and instruments each task with a span and duration histogram.
- `Run` starts configured servers in an `errgroup`, runs discovery service start events, waits for exit, then runs stop events.

## Common Mistakes

- Do not call `app.Run` before all `Grpc`, `Http`, `Cron`, and `AddDiscovery` calls.
- Do not register only HTTP for generated gateway APIs; the gateway dials the gRPC server.
- Do not ignore `NewConsul` errors before calling `AddDiscovery`.
- Do not put business logic in `main.go`; keep it in handlers, services, repositories, or factories.

