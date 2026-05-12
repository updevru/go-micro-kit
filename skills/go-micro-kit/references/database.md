# Database

## Connect

The `database.Connect` helper opens PostgreSQL through GORM and installs the `otelgorm` plugin:

```go
db, err := database.Connect(cfg.Database.DSN)
if err != nil {
    return err
}
```

It uses silent GORM logging by default.

## Model Registration

Register models before calling `database.Migrate`:

```go
package model

import (
    "github.com/updevru/go-micro-kit/database"
    "gorm.io/gorm"
)

func init() {
    database.AddModel(&Item{})
}

type Item struct {
    gorm.Model
    Name string `gorm:"not null"`
}
```

Using `init` is optional, but it prevents forgetting registration in `main.go`. If using `init`, ensure the model package is imported by the application.

## Migrate

```go
if err := database.Migrate(ctx, tracer, db); err != nil {
    log.ErrorContext(ctx, "failed to migrate database", "error", err)
    return err
}
```

`Migrate` returns nil when no models were registered. It uses GORM `AutoMigrate`.

## Repository Rules

- Accept `context.Context` in repository methods.
- Use `db.WithContext(ctx)` for all queries.
- Return domain/model errors upward and map them to gRPC status errors in handlers.
- Keep database transactions inside repositories or service-layer units of work, not in protobuf handlers unless the project already follows that style.

## Wiring Example

```go
db, err := database.Connect(cfg.Database.DSN)
if err != nil {
    panic(err)
}

if err := database.Migrate(ctx, tracer, db); err != nil {
    panic(err)
}

itemRepository := itemrepo.New(db)
itemHandler := itemhandler.New(log, tracer, itemRepository)
```

