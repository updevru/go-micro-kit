package telemetry

import (
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/log/global"
	"log/slog"
)

func CreateLogger() *slog.Logger {
	return otelslog.NewLogger("main", otelslog.WithLoggerProvider(global.GetLoggerProvider()))
}
