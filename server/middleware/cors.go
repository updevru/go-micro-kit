package middleware

import (
	"github.com/rs/cors"
	"net/http"
)

// NewCorsMiddleware Добавляет в ответ CORS-заголовки
func NewCorsMiddleware(allowedOrigins, allowedHeaders []string) func(http.Handler) http.Handler {
	corsHandler := cors.New(cors.Options{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{
			http.MethodHead,
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
		},
		AllowedHeaders: allowedHeaders,
	})

	return corsHandler.Handler
}
