package gotick

import (
	"context"
	"time"
)

type scheduler struct {
	cfg        SchedulerConfiguration
	ctx        context.Context
	cancel     context.CancelFunc
	jf         map[string]Job
	driver     SchedulerDriver
	errs       chan error
	executions chan Job
}

func (s *scheduler) RegisterJob(job Job) error {
	id := job.ID()
	if _, ok := s.jf[id]; ok {
		return ErrJobIDExists
	}

	s.jf[id] = job
	return nil
}

func (s *scheduler) RemoveJob(ctx context.Context, jobID string) error {
	return s.driver.RemoveJob(ctx, jobID)
}

func (s *scheduler) ScheduleJob(ctx context.Context, jobID string, schedule JobSchedule) error {
	if job, ok := s.jf[jobID]; !ok {
		return ErrJobNotFound
	} else {
		s.driver.StoreJob(ctx, job, schedule)
		return nil
	}
}

func (s *scheduler) Start(ctx context.Context) error {
	s.ctx, s.cancel = context.WithCancel(ctx)

	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				// poll for next job
				job, t, err := s.driver.NextExecution(ctx, time.Now())
				if err != nil {
					s.errs <- err
					continue
				}

				if time.Until(t) > s.cfg.MaxPlanAhead {
					select {
					case <-time.After(s.cfg.PollInterval):
						continue
					case <-ctx.Done():
						return
					}
				}

				s.executions <- job
				go func() {
					if err := job.Execute(ctx); err != nil {
						s.errs <- err
					}

					<-s.executions
				}()
			}
		}
	}(s.ctx)

	return nil
}

func (s *scheduler) Stop() error {
	s.cancel()
	return nil
}

func (s *scheduler) Errs() <-chan error {
	return s.errs
}

func NewScheduler() Scheduler {
	cfg := SchedulerConfiguration{
		MaxJobs:      -1,
		PollInterval: 1 * time.Second,
		Threads:      1,
		MaxPlanAhead: 1 * time.Minute,
	}

	return &scheduler{
		jf:         map[string]Job{},
		driver:     nil,
		cfg:        cfg,
		errs:       make(chan error),
		executions: make(chan Job, cfg.Threads),
	}
}
