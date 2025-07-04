package logger

import (
	slogmulti "github.com/samber/slog-multi"
	slogsentry "github.com/samber/slog-sentry/v2"
	"log/slog"
	"os"
)

func CreateLogger() *slog.Logger {
	return slog.New(
		slogmulti.Fanout(
			slog.NewJSONHandler(os.Stdout, nil),
			slogsentry.Option{Level: slog.LevelError}.NewSentryHandler(),
		),
	)
}
