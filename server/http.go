package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/updevru/go-micro-kit/config"
	"github.com/updevru/go-micro-kit/server/middleware"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/protobuf/encoding/protojson"
	"log/slog"
	"net/http"
)

type HttpHandler func(context.Context, *runtime.ServeMux, *grpc.ClientConn) error

func customHeaderMatcher(key string) (string, bool) {
	switch key {
	case "Authorization":
		return key, true
	default:
		return runtime.DefaultHeaderMatcher(key)
	}
}

func (s *Server) Http(cfg *config.Http, cfgRpc *config.Grpc, opts []runtime.ServeMuxOption, handlers ...HttpHandler) {
	s.httpServer = func() error {
		con, _ := grpc.NewClient(
			fmt.Sprintf(":%s", cfgRpc.Port),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		)

		options := []runtime.ServeMuxOption{
			runtime.WithHealthzEndpoint(grpc_health_v1.NewHealthClient(con)),
			runtime.WithIncomingHeaderMatcher(customHeaderMatcher),
			runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
				MarshalOptions: protojson.MarshalOptions{
					EmitUnpopulated:   false,
					EmitDefaultValues: false,
					UseEnumNumbers:    false,
				},
			}),
		}

		for _, opt := range opts {
			options = append(options, opt)
		}

		srv := runtime.NewServeMux(options...)

		for _, handler := range handlers {
			if err := handler(s.ctx, srv, con); err != nil {
				s.logger.ErrorContext(s.ctx, "Failed to register handler: %v", slog.String("error", err.Error()))
				return err
			}
		}

		corsMiddleware := middleware.NewCorsMiddleware(cfg.AllowedOrigins, cfg.AllowedHeaders)

		address := fmt.Sprintf(":%s", cfg.Port)
		httpServer := &http.Server{
			Addr:    address,
			Handler: corsMiddleware(srv),
		}

		go func() {
			<-s.ctx.Done()
			s.logger.Info("rest server stopping")
			if err := httpServer.Shutdown(context.Background()); err != nil {
				s.logger.ErrorContext(s.ctx, "Failed to shutdown rest gateway server: %v", slog.String("error", err.Error()))
			}
		}()

		var err error
		s.logger.Info("rest server listening at", slog.String("address", address))
		if err = httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("failed to serve: %v", slog.String("error", err.Error()))
			return err
		}

		return nil
	}
}
