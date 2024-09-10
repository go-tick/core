package gotick

import (
	"context"
	"errors"
	"time"
)

type scheduler struct {
	cfg      SchedulerConfiguration
	cancel   context.CancelFunc
	driver   SchedulerDriver
	planner  Planner
	registry map[string]Job
	errs     chan error
}

func (s *scheduler) RegisterJob(job Job) error {
	id := job.ID()
	if _, ok := s.registry[id]; ok {
		return ErrJobIDExists
	}

	s.registry[id] = job
	return nil
}

func (s *scheduler) UnscheduleJob(ctx context.Context, jobID string) error {
	return s.driver.UnscheduleJob(ctx, jobID)
}

func (s *scheduler) ScheduleJob(ctx context.Context, jobID string, schedule JobSchedule) error {
	if job, ok := s.registry[jobID]; !ok {
		return ErrJobNotFound
	} else {
		return s.driver.ScheduleJob(ctx, job, schedule)
	}
}

func (s *scheduler) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	err := s.planner.Start(ctx)

	go s.execution(ctx)
	go s.errsListener(ctx)

	return err
}

func (s *scheduler) Stop() error {
	s.cancel()
	plannerErr := s.planner.Stop()
	close(s.errs)

	return plannerErr
}

func (s *scheduler) Errs() <-chan error {
	return s.errs
}

func (s *scheduler) errsListener(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			s.publishErr(ctx.Err())
			return
		case err := <-s.planner.Errs():
			s.publishErr(err)
		}
	}
}

func (s *scheduler) execution(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// poll for next job
			plan, err := s.driver.NextExecution(ctx, time.Now())
			if err != nil || plan == nil || time.Until(plan.PlannedAt) > s.cfg.MaxPlanAhead {
				if err != nil {
					s.publishErr(err)
				}

				select {
				case <-time.After(s.cfg.PollInterval):
					continue
				case <-ctx.Done():
					return
				}
			}

			if time.Until(plan.PlannedAt) > s.cfg.MaxPlanAhead {
				continue
			}

			jobCtx := JobContext{
				Context:   ctx,
				Job:       plan.Job,
				Schedule:  plan.Schedule,
				PlannedAt: plan.PlannedAt,
				executed:  make(chan any, 1),
			}

			err = s.driver.BeforeExecution(jobCtx)
			if err != nil {
				if errors.Is(err, ErrJobLocked) {
					continue
				}

				s.publishErr(err)
			}

			err = s.planner.Plan(jobCtx)
			if err != nil {
				s.publishErr(err)
			} else {
				go func() {
					select {
					case <-jobCtx.executed:
						err = s.driver.Executed(jobCtx)
						if err != nil {
							s.publishErr(err)
						}
					case <-ctx.Done():
						return
					}
				}()
			}
		}
	}
}

func (s *scheduler) publishErr(err error) {
	select {
	case s.errs <- err:
		return
	default:
		// nobody listens to channel
		return
	}
}

func NewScheduler(cfg SchedulerConfiguration) Scheduler {
	return &scheduler{
		cfg:      cfg,
		driver:   cfg.DriverFactory(),
		planner:  cfg.PlannerFactory(),
		registry: make(map[string]Job),
		errs:     make(chan error),
	}
}
