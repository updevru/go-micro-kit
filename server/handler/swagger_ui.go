package handler

import (
	"context"
	"errors"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/swaggest/swgui/v5emb"
	"github.com/updevru/go-micro-kit/server"
	"google.golang.org/grpc"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var (
	ErrSwaggerUIBadServiceUrl = errors.New("bad service url")
)

type SwaggerOptions struct {
	// Название сервиса, будет видно в UI
	ServiceName string

	// Публичный адрес сервиса для выполнения запросов из UI
	ServiceUrl string

	// Путь к OpenAPI файлу в формате JSON
	OpenAPIFileJson string
}

// NewSwaggerUIHandler Обработчик для отображения Swagger UI по адресу /swagger/
func NewSwaggerUIHandler(opt SwaggerOptions) server.HttpHandler {
	return func(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
		var address *url.URL
		var err error

		if opt.ServiceUrl != "" {
			address, err = url.Parse(opt.ServiceUrl)
			if err != nil {
				return ErrSwaggerUIBadServiceUrl
			}
		}

		err = mux.HandlePath(http.MethodGet, "/docs/api.swagger.json", func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
			content, err := readOpenAPIContent(opt.OpenAPIFileJson, address)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Установка заголовков для правильной обработки JSON в ответе
			w.Header().Set("Content-Type", "application/json")

			// Отправка модифицированного содержимого
			_, err = w.Write(content)
			if err != nil {
				http.Error(w, "Unable to write response", http.StatusInternalServerError)
			}
		})

		if err != nil {
			return err
		}

		handler := v5emb.New(
			opt.ServiceName,
			"/docs/api.swagger.json",
			"/swagger/",
		)
		return mux.HandlePath(http.MethodGet, "/swagger/**", func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
			handler.ServeHTTP(w, r)
		})
	}
}

func readOpenAPIContent(file string, address *url.URL) ([]byte, error) {
	// Чтение содержимого файла
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	content := string(data)

	// Замена адреса сервера и протокола, чтобы проходили запросы из UI
	if address != nil {
		content = strings.ReplaceAll(content, "localhost:8080", address.Host)
		content = strings.ReplaceAll(content, "\"http\"", fmt.Sprintf("\"%s\"", address.Scheme))
	}

	return []byte(content), nil
}
