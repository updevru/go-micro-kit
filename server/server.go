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
	group  *errgroup.Group
	logger *slog.Logger
	tracer trace.Tracer
	meter  metric.Meter
	events []Event
}

func NewServer(ctx context.Context, logger *slog.Logger, tracer trace.Tracer, meter metric.Meter) *Server {
	group, groupCtx := errgroup.WithContext(ctx)

	return &Server{
		ctx:    groupCtx,
		group:  group,
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

func (s *Server) Run() error {
	if err := s.runEventServiceStart(); err != nil {
		return err
	}
	defer s.runEventServiceStop()

	if err := s.group.Wait(); err != nil {
		s.logger.ErrorContext(s.ctx, "exit reason: %s", err)
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
