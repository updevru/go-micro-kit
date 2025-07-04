package database

import (
	"context"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

var modelsForMigration []interface{}

func AddModel(model interface{}) {
	modelsForMigration = append(modelsForMigration, model)
}

func Migrate(ctx context.Context, tracer trace.Tracer, db *gorm.DB) error {
	if modelsForMigration == nil {
		return nil
	}

	ctxSpan, span := tracer.Start(ctx, "Update database schema")
	span.End()
	err := db.WithContext(ctxSpan).AutoMigrate(modelsForMigration...)

	if err != nil {
		span.RecordError(err)
	}

	return err
}
