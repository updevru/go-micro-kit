package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/updevru/go-micro-kit/config"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
	"log/slog"
	"net/http"
)

type HttpHandler func(context.Context, *runtime.ServeMux, *grpc.ClientConn) error

func (s *Server) Http(cfg *config.Http, cfgRpc *config.Grpc, handlers ...HttpHandler) {
	s.group.Go(func() error {

		con, _ := grpc.NewClient(
			fmt.Sprintf(":%s", cfgRpc.Port),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		)
		srv := runtime.NewServeMux(
			runtime.WithHealthzEndpoint(grpc_health_v1.NewHealthClient(con)),
		)

		for _, handler := range handlers {
			if err := handler(s.ctx, srv, con); err != nil {
				s.logger.ErrorContext(s.ctx, "Failed to register handler: %v", err)
				return err
			}
		}

		address := fmt.Sprintf(":%s", cfg.Port)
		httpServer := &http.Server{
			Addr:    address,
			Handler: srv,
		}

		go func() {
			<-s.ctx.Done()
			s.logger.Info("rest server stopping")
			if err := httpServer.Shutdown(context.Background()); err != nil {
				s.logger.ErrorContext(s.ctx, "Failed to shutdown rest gateway server: %v", err)
			}
		}()

		var err error
		s.logger.Info("rest server listening at", slog.String("address", address))
		if err = httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("failed to serve: %v", err)
			return err
		}

		return nil
	})
}
