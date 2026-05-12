---
name: go-micro-kit
description: Use when building, modifying, or debugging Go microservices with github.com/updevru/go-micro-kit, including protobuf-first gRPC services, gRPC-Gateway HTTP APIs, env configuration, OpenTelemetry setup, GORM/PostgreSQL migrations, cron tasks, Consul discovery, Swagger UI, and service bootstrap wiring.
---

# go-micro-kit

Use this skill as the project-specific guide for `github.com/updevru/go-micro-kit`. Prefer the framework's packages and conventions when the user asks to scaffold or change a Go microservice that uses this kit.

## First Steps

Inspect the target service before editing:

1. Read `go.mod` to confirm the module path and `github.com/updevru/go-micro-kit` version.
2. Read `main.go` or `cmd/**/main.go` to see how the bootstrap manager is wired.
3. Read `internal/config`, `.env.example`, or deployment env files before changing configuration.
4. Read `proto/`, `gen/`, and `docs/api.swagger.json` before changing APIs.
5. Preserve local architecture if it already has factories, repositories, handlers, or generated-code conventions.

## Framework Map

- `config`: env and `.env` loading through `go-envconfig` and `godotenv`.
- `server`: bootstrap manager for gRPC, HTTP gRPC-Gateway, cron, discovery events, and graceful shutdown.
- `server/handler`: embedded Swagger UI handler for generated OpenAPI specs.
- `database`: PostgreSQL GORM connection with `otelgorm`, model registration, and `AutoMigrate`.
- `telemetry`: OpenTelemetry traces, metrics, logs, providers, and slog bridge.
- `logger`: JSON slog logger with Sentry, without the OTel slog bridge.
- `discovery`: service discovery interface and Consul implementation.
- `server/middleware`: CORS middleware used by HTTP gateway.

## Core Rules

- Use a schema-first flow: edit `.proto`, regenerate code, implement generated server interfaces, and register the gRPC-Gateway handler.
- Register gRPC before HTTP because `server.Http` dials the local gRPC server.
- Keep `main.go` mostly wiring: config, telemetry, database, repositories/services, handlers, server registration, `Run`.
- Always call `server.NewServer(logger, tracer, meter)` before `Grpc`, `Http`, `Cron`, or `AddDiscovery`.
- Call `app.Run(ctx)` last; it blocks until cancellation or a server error.
- Use `signal.NotifyContext` for graceful shutdown and call the telemetry shutdown function on exit.
- Register GORM models with `database.AddModel` before `database.Migrate`.
- Use `db.WithContext(ctx)` in repositories so traces propagate into SQL spans.
- Return gRPC errors with `status.Error(codes.Xxx, "...")`; gRPC-Gateway maps these to HTTP responses.
- Treat authentication as application code. The framework only forwards `Authorization` from HTTP to gRPC metadata.

## Reference Files

Read only the files relevant to the task:

- `references/protobuf.md`: protobuf layout, gRPC-Gateway annotations, generation commands, and OpenAPI output.
- `references/server-bootstrap.md`: canonical `main.go`, server lifecycle, and registration order.
- `references/handlers-and-factories.md`: handler structs, RPC methods, gRPC/REST factories, repositories.
- `references/config-and-env.md`: config structs, prefix pitfalls, and environment variables.
- `references/database.md`: PostgreSQL connection, GORM model registration, migrations, repository rules.
- `references/observability.md`: OpenTelemetry setup, logging, tracing, metrics, OTLP env.
- `references/discovery-cron-swagger.md`: Consul discovery, cron tasks, Swagger UI, CORS/headers.
- `references/devops.md`: Taskfile, Docker, generated artifacts, and local development commands.

