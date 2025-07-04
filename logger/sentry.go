package logger

import (
	"github.com/getsentry/sentry-go"
)

func InitSentry() error {
	return sentry.Init(sentry.ClientOptions{
		EnableTracing: false,
	})
}
