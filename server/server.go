package server

import (
	"context"
	"github.com/updevru/go-micro-kit/discovery"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"
	"log/slog"
)

type Server struct {
	ctx    context.Context
	logger *slog.Logger
	tracer trace.Tracer
	meter  metric.Meter
	events []Event

	grpcServer func() error
	httpServer func() error
	cronServer func() error
}

func NewServer(logger *slog.Logger, tracer trace.Tracer, meter metric.Meter) *Server {
	return &Server{
		logger: logger,
		tracer: tracer,
		meter:  meter,
		events: make([]Event, 0),
	}
}

func (s *Server) AddDiscovery(discovery discovery.Discovery) {
	s.events = append(s.events, Event{
		ServiceStart: discovery.RegisterService,
		ServiceStop:  discovery.DeregisterService,
		WorkerStart:  discovery.RegisterWorker,
		WorkerStop:   discovery.DeregisterWorker,
	})
}

func (s *Server) Run(ctx context.Context) error {
	group, groupCtx := errgroup.WithContext(ctx)
	s.ctx = groupCtx

	if s.grpcServer != nil {
		group.Go(s.grpcServer)
	}

	if s.httpServer != nil {
		group.Go(s.httpServer)
	}

	if s.cronServer != nil {
		group.Go(s.cronServer)
	}

	if err := s.runEventServiceStart(); err != nil {
		return err
	}
	defer func(s *Server) {
		err := s.runEventServiceStop()
		if err != nil {
			s.logger.ErrorContext(s.ctx, "RunEventServiceStop error", slog.String("error", err.Error()))
		}
	}(s)

	if err := group.Wait(); err != nil {
		s.logger.ErrorContext(groupCtx, "exit reason: %s", slog.String("error", err.Error()))
		return err
	}

	return nil
}

func (s *Server) runEventServiceStart() error {
	for _, event := range s.events {
		if err := event.ServiceStart(); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) runEventServiceStop() error {
	for _, event := range s.events {
		if err := event.ServiceStop(); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) runEventWorkerStart() error {
	for _, event := range s.events {
		if err := event.WorkerStart(); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) runEventWorkerStop() error {
	for _, event := range s.events {
		if err := event.WorkerStop(); err != nil {
			return err
		}
	}
	return nil
}
