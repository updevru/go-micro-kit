# Configuration and Environment

## Framework Types

```go
type App struct {
    AppName string `env:"APP_NAME"`
}

type Http struct {
    Host string `env:"HOST, default=localhost"`
    Port string `env:"PORT, default=8080"`
    AllowedOrigins []string `env:"ALLOW_ORIGINS, default=*"`
    AllowedHeaders []string `env:"ALLOW_HEADERS, default=*"`
}

type Grpc struct {
    Host string `env:"HOST, default=localhost"`
    Port string `env:"PORT, default=8081"`
}

func CreateConfig(ctx context.Context, obj any) error
```

`CreateConfig` calls `godotenv.Load()` and then `envconfig.Process(ctx, obj)`.

## Recommended Config Struct

```go
package config

import (
    kitconfig "github.com/updevru/go-micro-kit/config"
    "github.com/updevru/go-micro-kit/telemetry"
)

type Config struct {
    kitconfig.App
    Http      kitconfig.Http  `env:",prefix=HTTP_"`
    Grpc      kitconfig.Grpc  `env:",prefix=GRPC_"`
    Telemetry telemetry.Config `env:",prefix=OTEL_"`
    Database  Database        `env:",prefix=DB_"`
}

type Database struct {
    DSN string `env:"DSN, required"`
}
```

This maps to:

- `APP_NAME`
- `HTTP_HOST`, `HTTP_PORT`, `HTTP_ALLOW_ORIGINS`, `HTTP_ALLOW_HEADERS`
- `GRPC_HOST`, `GRPC_PORT`
- `OTEL_EXPORTER_OTLP_PROTOCOL`
- `DB_DSN`

## Prefix Pitfalls

- Embedding `kitconfig.App` maps `AppName` to `APP_NAME`.
- Using a named field like `App kitconfig.App` with the tag `env:",prefix=APP_"` maps `AppName` to `APP_APP_NAME`; avoid that unless intended.
- `telemetry.Config.Protocol` has tag `EXPORTER_OTLP_PROTOCOL`. With prefix `OTEL_`, the env var is `OTEL_EXPORTER_OTLP_PROTOCOL`, which matches common OpenTelemetry env naming.
- If you do not prefix telemetry config, the env var is `EXPORTER_OTLP_PROTOCOL`.

## Example .env

```env
APP_NAME=item-service

HTTP_HOST=localhost
HTTP_PORT=8080
HTTP_ALLOW_ORIGINS=*
HTTP_ALLOW_HEADERS=*

GRPC_HOST=localhost
GRPC_PORT=8081

DB_DSN=host=localhost user=postgres password=postgres dbname=items port=5432 sslmode=disable

OTEL_EXPORTER_OTLP_PROTOCOL=grpc
OTEL_SERVICE_NAME=item-service
OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317

SENTRY_DSN=
CONSUL_HTTP_ADDR=127.0.0.1:8500
```

## Public URL

If Swagger UI needs an externally reachable URL, add application config:

```go
type AppConfig struct {
    PublicURL string `env:"PUBLIC_URL"`
}
```

Then pass it to `handler.SwaggerOptions.ServiceUrl`.
