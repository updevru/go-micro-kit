package database

import (
	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	driver "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(dsn string) (*gorm.DB, error) {
	cfg := &gorm.Config{
		Logger:      logger.Default.LogMode(logger.Silent),
		QueryFields: false,
	}

	db, err := gorm.Open(driver.Open(dsn), cfg)
	if err == nil {
		err = db.Use(otelgorm.NewPlugin())
	}

	return db, err
}
