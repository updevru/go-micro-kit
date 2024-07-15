package server

import (
	"fmt"
	"github.com/updevru/go-micro-kit/pkg/config"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"
	"log/slog"
	"net"
)

type GrpcHandler func(*grpc.Server)

func (s *Server) Grpc(cfg *config.Grpc, handlers ...GrpcHandler) {
	s.group.Go(func() error {
		srv := grpc.NewServer(
			grpc.StatsHandler(otelgrpc.NewServerHandler()),
		)

		healthcheck := health.NewServer()
		healthgrpc.RegisterHealthServer(srv, healthcheck)

		for _, handler := range handlers {
			handler(srv)
		}

		listen, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.Port))
		if err != nil {
			s.logger.Error("failed to listen", err)
			return err
		}

		go func() {
			<-s.ctx.Done()
			s.logger.Info("grpc server stopping")
			srv.GracefulStop()
		}()

		s.logger.Info("grpc server listening at", slog.String("address", listen.Addr().String()))
		if err := srv.Serve(listen); err != nil {
			s.logger.Error("failed to serve: %v", err)
			return err
		}

		return nil
	})
}
