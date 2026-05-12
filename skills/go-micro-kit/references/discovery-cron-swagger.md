# Discovery, Cron, Swagger, and HTTP Details

## Consul Discovery

```go
consul, err := discovery.NewConsul(&cfg.App, &cfg.Http, &cfg.Grpc)
if err != nil {
    return err
}
app.AddDiscovery(consul)
```

The discovery interface is:

```go
type Discovery interface {
    RegisterService() error
    DeregisterService() error
    RegisterWorker() error
    DeregisterWorker() error
}
```

`NewConsul` uses HashiCorp Consul's default config, so Consul env such as `CONSUL_HTTP_ADDR` applies.

Registration behavior:

- HTTP service address and health check are added when `configHttp` is not nil.
- gRPC service address and health check are added when `configGrpc` is not nil.
- Worker registration is used by `Cron` lifecycle events.

## Cron

```go
app.Cron([]server.CronTask{
    {
        Name: "cleanup",
        Cron: "0 3 * * *",
        Fn: func(ctx context.Context) error {
            return cleanup.Run(ctx)
        },
    },
})
```

Rules:

- Cron syntax is parsed by `go-co-op/gocron/v2`.
- Each task function receives a context with an active OTel span.
- Return errors from the task function; the framework records them on the span.
- Keep long-running task logic in a service, not inline in `main.go`.

## Swagger UI

```go
swagger := handler.NewSwaggerUIHandler(handler.SwaggerOptions{
    ServiceName:     cfg.AppName,
    ServiceUrl:      cfg.PublicURL,
    OpenAPIFileJson: "./docs/api.swagger.json",
})
```

Add `swagger` to the `server.HttpHandler` list. It serves:

- `/swagger/`: Swagger UI.
- `/docs/api.swagger.json`: OpenAPI JSON.

If `ServiceUrl` is set, the handler rewrites `localhost:8080` and the `"http"` scheme inside the OpenAPI file so requests from the UI target the public service URL.

## HTTP Gateway Details

`server.Http` automatically:

- Adds `/healthz` backed by gRPC health.
- Wraps the gateway with CORS middleware.
- Wraps the gateway with OTel HTTP middleware.
- Uses protobuf JSON marshaling.
- Forwards `Authorization`, `Traceparent`, `Tracestate`, and `Baggage` headers to gRPC metadata.

