# Handlers, Factories, and Repositories

## Handler Pattern

```go
package itemhandler

import (
    "log/slog"

    pb "example/gen/item/v1"
    "example/internal/repository/itemrepo"
    "go.opentelemetry.io/otel/trace"
)

type Handler struct {
    pb.UnimplementedItemServiceServer
    log    *slog.Logger
    tracer trace.Tracer
    repo   itemrepo.Repository
}

func New(log *slog.Logger, tracer trace.Tracer, repo itemrepo.Repository) *Handler {
    return &Handler{log: log, tracer: tracer, repo: repo}
}
```

## RPC Method Pattern

```go
package itemhandler

import (
    "context"

    pb "example/gen/item/v1"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

func (h *Handler) CreateItem(ctx context.Context, in *pb.CreateItemRequest) (*pb.CreateItemResponse, error) {
    ctx, span := h.tracer.Start(ctx, "handler.ItemService.CreateItem")
    defer span.End()

    item, err := h.repo.Create(ctx, in.GetName())
    if err != nil {
        span.RecordError(err)
        h.log.ErrorContext(ctx, "failed to create item", "error", err)
        return nil, status.Error(codes.Internal, "failed to create item")
    }

    return &pb.CreateItemResponse{Id: uint64(item.ID), Name: item.Name}, nil
}
```

Rules:

- Start a span at the top of each non-trivial RPC.
- Use the returned `ctx` for downstream calls.
- Record errors on the span before returning.
- Return `status.Error` or `status.Errorf` for public gRPC errors.
- Keep mapping between GORM models and protobuf DTOs explicit.

## gRPC Factory

Use a factory when the project separates bootstrap wiring from `main.go`:

```go
package grpcserver

import (
    pb "example/gen/item/v1"
    "github.com/updevru/go-micro-kit/server"
    "google.golang.org/grpc"
)

func New(handler pb.ItemServiceServer) ([]grpc.ServerOption, server.GrpcHandler) {
    return []grpc.ServerOption{},
        func(s *grpc.Server) {
            pb.RegisterItemServiceServer(s, handler)
        }
}
```

## REST Factory

```go
package restserver

import (
    pb "example/gen/item/v1"
    "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
    "github.com/updevru/go-micro-kit/server"
    "github.com/updevru/go-micro-kit/server/handler"
)

func New(serviceName, publicURL string) ([]runtime.ServeMuxOption, []server.HttpHandler) {
    swagger := handler.NewSwaggerUIHandler(handler.SwaggerOptions{
        ServiceName:     serviceName,
        ServiceUrl:      publicURL,
        OpenAPIFileJson: "./docs/api.swagger.json",
    })

    return []runtime.ServeMuxOption{}, []server.HttpHandler{
        pb.RegisterItemServiceHandler,
        swagger,
    }
}
```

## Repository Pattern

```go
package itemrepo

import (
    "context"

    "example/internal/model"
    "gorm.io/gorm"
)

type Repository interface {
    Create(ctx context.Context, name string) (*model.Item, error)
}

type repository struct {
    db *gorm.DB
}

func New(db *gorm.DB) Repository {
    return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, name string) (*model.Item, error) {
    item := &model.Item{Name: name}
    if err := r.db.WithContext(ctx).Create(item).Error; err != nil {
        return nil, err
    }
    return item, nil
}
```

Always use `db.WithContext(ctx)` in repository methods.

