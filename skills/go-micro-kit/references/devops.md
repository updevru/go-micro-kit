# DevOps and Local Commands

## Taskfile Pattern

```yaml
version: '3'
silent: true

vars:
  PROTO_PATH: './proto/item/v1/*.proto'
  PROTO_GO_OUT: './gen/'
  PROTO_OPENAPI_OUT: './docs'

tasks:
  test:
    cmds:
      - go test ./...

  lint:
    cmds:
      - golangci-lint run

  gen-proto:
    cmds:
      - protoc -I proto {{.PROTO_PATH}} --go_out={{.PROTO_GO_OUT}} --go_opt=paths=source_relative --go-grpc_out={{.PROTO_GO_OUT}} --go-grpc_opt=paths=source_relative --grpc-gateway_out {{.PROTO_GO_OUT}} --grpc-gateway_opt paths=source_relative --grpc-gateway_opt generate_unbound_methods=true --openapiv2_out {{.PROTO_OPENAPI_OUT}} --openapiv2_opt allow_merge=true,merge_file_name=api
```

Use the project's existing task runner if present. This repository uses `taskfile.yaml`.

## Dockerfile Pattern

```dockerfile
FROM golang:1.24 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go test ./...
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM alpine:latest
RUN apk add --no-cache tzdata
WORKDIR /app
COPY --from=builder /app/main /app/main
COPY --from=builder /app/docs/api.swagger.json /app/docs/api.swagger.json
EXPOSE 8080
EXPOSE 8081
ENTRYPOINT ["./main"]
```

Copy `docs/api.swagger.json` when the service uses the bundled Swagger UI.

## Local Observability Stack

Common local services:

- PostgreSQL on `5432`.
- OTel Collector on `4317` for gRPC and `4318` for HTTP.
- Grafana on `3000`.
- Tempo for traces.
- Mimir or Prometheus for metrics.
- Loki for logs.
- Consul on `8500` when using discovery.

## Validation Commands

Use the commands that fit the changed surface:

```bash
go test ./...
go test ./... -race
task gen-proto
task lint
```

After proto changes, verify generated Go files compile and `docs/api.swagger.json` exists if Swagger UI is configured.

