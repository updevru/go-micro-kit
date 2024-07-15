# Micro Kit

Набор преднастроенных бибилотек для быстрого создания микро-сервисов на go.

Набор инструментов создан для Schema-First подхода, когда описываются сервисы в формает [protocol buffers](https://developers.google.com/protocol-buffers)
формате и далее генерируется код этих сервисов.

## Генерация кода

Установка утилит:

```
$ go install \
    github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
    github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
    google.golang.org/protobuf/cmd/protoc-gen-go \
    google.golang.org/grpc/cmd/protoc-gen-go-grpc
```

Установленные утилиты должны находиться в папке, на которую указывает переменная окружения `$GOBIN`.

Пример генерации кода:

```bash
protoc -I proto .\proto\store\store.proto \ 
  --go_out=./gen/ --go_opt=paths=source_relative --go-grpc_out=./gen/ --go-grpc_opt=paths=source_relative \ 
  --grpc-gateway_out ./gen --grpc-gateway_opt paths=source_relative --grpc-gateway_opt generate_unbound_methods=true \ 
  --openapiv2_out ./docs --openapiv2_opt allow_merge=true,merge_file_name=api
```

Команда генерирует весь необходимый go код, gRPC-Gateway и openapi спецификацию.

## Возможности

### Bootstrap manager

Удобный запуск нескольких сервисов в одном приложении (например gRPC и HTTP API на разных портах).

### API Gateway

Генерация gRPC API из protobuf и автоматическое создание HTTP API с помощью [gRPC-Gateway](https://github.com/grpc-ecosystem/grpc-gateway).

### Configuration

Конфигурация на основе env переменных или .env файлов, это позволяет удобно запускать сервисы в виде контейнеров.

### Cron

Простой функционал для запуска фоновых задач по расписанию.

### Discovery

Автоматическая регистрация сервиса в системах service discovery ([Consul](https://developer.hashicorp.com/consul)).

### Observability

Трассировка, метрики и логи в формет [OpenTelemetry](https://opentelemetry.io/docs/languages/go/).

## Примеры использования

### main.go

```go
func main() {
	// Единый context, который закрывается при завершении программы
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Инициализируем конфигурацию и наполняем структуру переменными окружения
	var cfg config.Config
	if err := configPkg.CreateConfig(ctx, &cfg); err != nil {
		panic(err)
	}

	// Настройка OpenTelemetry
	otelShutdown, err := telemetry.SetupOTelSDK(ctx)
	if err != nil {
		panic(err)
	}
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	logger := telemetry.CreateLogger()
	tracer := telemetry.CreateTracer()
	meter := telemetry.CreateMeter()

	// Создаем Bootstrap manager
	app := server.NewServer(ctx, logger, tracer, meter)
    
	// Инифализируем и добавляем gRPC сервисы
	app.Grpc(&cfg.Grpc, func(g *grpc.Server) {
        pbStore.RegisterStoreServer(g)
    })

	//Инициализируем и добавляем gRPC-Gateway, так же можем добавить дополнительные роуты
    app.Http(&cfg.Http, &cfg.Grpc, func(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
        return []server.HttpHandler{
            pbStore.RegisterStoreHandler,
        }
    })

    // Добавляем фоновые задачи по расписанию
    app.Cron([]server.CronTask{
        {
            Name: "clock",
            Cron: "*/1 * * * *",
            Fn: func(ctx context.Context) error {
                // TODO: implement cron job
                return nil
            },
        },
    })

	// Добавляем автоматический service discovery
	consul, err := discovery.NewConsul(&cfg.App, &cfg.Http, &cfg.Grpc)
	app.AddDiscovery(consul)

	// Запуск приложения
	if err := app.Run(); err != nil {
		logger.ErrorContext(ctx, "Failed to run server: %v", err)
		panic(err)
	}
}
```

### Конфигурация приложения

Используются библиотеки [go-envconfig](https://github.com/sethvargo/go-envconfig) и [godotenv](https://github.com/joho/godotenv).

```go 
package config
import "github.com/updevru/go-micro-kit/config"

type Config struct {
	config.App
	Http config.Http `env:",prefix=HTTP_"`
	Grpc config.Grpc `env:",prefix=GRPC_"`
	Option string `env:"OPTION, default=value"`
}
```