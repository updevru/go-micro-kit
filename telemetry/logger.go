package telemetry

import (
	slogmulti "github.com/samber/slog-multi"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/log/global"
	"log/slog"
	"os"
)

func CreateLogger() *slog.Logger {
	return slog.New(
		slogmulti.Fanout(
			slog.NewJSONHandler(os.Stdout, nil),
			otelslog.NewHandler("main", otelslog.WithLoggerProvider(global.GetLoggerProvider())),
		),
	)
}
