package server

import (
	"context"
	"fmt"
	"github.com/go-co-op/gocron/v2"
	"go.opentelemetry.io/otel/metric"
	"log/slog"
	"time"
)

type CronTask struct {
	Name string
	Cron string
	Fn   func(ctx context.Context) error
}

func (s *Server) Cron(tasks []CronTask) {
	s.group.Go(func() error {
		scheduler, err := gocron.NewScheduler()
		if err != nil {
			return err
		}

		for _, item := range tasks {
			_, err = scheduler.NewJob(
				gocron.CronJob(item.Cron, false),
				gocron.NewTask(func() {
					ctx, span := s.tracer.Start(s.ctx, "cron."+item.Name)
					defer span.End()

					histogram, _ := s.meter.Float64Histogram(
						fmt.Sprintf("cron.%s.duration", item.Name),
						metric.WithDescription("The duration of cron execution."),
						metric.WithUnit("s"),
					)

					start := time.Now()
					if err := item.Fn(ctx); err != nil {
						span.RecordError(err)
					}
					histogram.Record(ctx, time.Since(start).Seconds())
				}),
				gocron.WithName(item.Name),
			)
			if err != nil {
				return err
			}

			s.logger.InfoContext(s.ctx, "Schedule task", slog.String("name", item.Name), slog.String("cron", item.Cron))
		}

		if err := s.runEventWorkerStart(); err != nil {
			return err
		}
		defer s.runEventWorkerStop()

		scheduler.Start()
		<-s.ctx.Done()

		s.logger.InfoContext(s.ctx, "Cron server stopping")

		return scheduler.Shutdown()
	})
}
