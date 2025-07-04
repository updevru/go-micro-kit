package config

import (
	"context"
	"github.com/joho/godotenv"
	"github.com/sethvargo/go-envconfig"
)

type App struct {
	AppName string `env:"APP_NAME"`
}

type Http struct {
	Host string `env:"HOST, default=localhost"`
	Port string `env:"PORT, default=8080"`

	// Настройки CORS
	AllowedOrigins []string `env:"ALLOW_ORIGINS, default=*"`
	AllowedHeaders []string `env:"ALLOW_HEADERS, default=*"`
}

type Grpc struct {
	Host string `env:"HOST, default=localhost"`
	Port string `env:"PORT, default=8081"`
}

func CreateConfig(ctx context.Context, obj any) error {
	_ = godotenv.Load()

	return envconfig.Process(ctx, obj)
}
