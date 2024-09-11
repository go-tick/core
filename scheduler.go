package gotick

import (
	"context"
	"errors"
	"sync"
	"time"
)

type scheduler struct {
	cfg         SchedulerConfiguration
	cancel      context.CancelFunc
	driver      SchedulerDriver
	planner     Planner
	registry    map[string]Job
	errs        chan error
	subscribers []SchedulerSubscriber
	stopOnce    sync.Once
}

func (s *scheduler) OnError(err error) {
	s.onError(err)
}

func (s *scheduler) OnBeforeJobExecution(ctx *JobContext) {
	err := s.callSubscribers(func(subscriber SchedulerSubscriber) error {
		return subscriber.OnBeforeJobExecution(ctx)
	})

	if err != nil {
		s.onError(err)
	}
}

func (s *scheduler) OnJobExecuted(ctx *JobContext) {
	err := s.callSubscribers(func(subscriber SchedulerSubscriber) error {
		return subscriber.OnJobExecuted(ctx)
	})

	if err != nil {
		s.onError(err)
	}
}

func (s *scheduler) Subscribe(subscriber SchedulerSubscriber) {
	s.subscribers = append(s.subscribers, subscriber)
}

func (s *scheduler) RegisterJob(job Job) error {
	id := job.ID()
	if _, ok := s.registry[id]; ok {
		return ErrJobIDExists
	}

	s.registry[id] = job
	return nil
}

func (s *scheduler) UnscheduleJobByJobID(ctx context.Context, jobID string) error {
	return s.driver.UnscheduleJobByJobID(ctx, jobID)
}

func (s *scheduler) UnscheduleJobByScheduleID(ctx context.Context, schedulerID string) error {
	return s.driver.UnscheduleJobByScheduleID(ctx, schedulerID)
}

func (s *scheduler) ScheduleJob(ctx context.Context, jobID string, schedule JobSchedule) (string, error) {
	if job, ok := s.registry[jobID]; !ok {
		return "", ErrJobNotFound
	} else {
		return s.driver.ScheduleJob(ctx, job, schedule)
	}
}

func (s *scheduler) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	err := s.callSubscribers(func(subscriber SchedulerSubscriber) error {
		return subscriber.OnStart()
	})
	if err != nil {
		return err
	}

	err = s.planner.Start(ctx)
	if err != nil {
		return err
	}

	s.planner.Subscribe(s)

	go s.background(ctx)
	go s.errors(ctx)

	return nil
}

func (s *scheduler) Stop() (err error) {
	s.stopOnce.Do(func() {
		err = s.stop()
	})

	return
}

func (s *scheduler) errors(ctx context.Context) {
	for {
		select {
		case err := <-s.errs:
			s.callSubscribers(func(subscriber SchedulerSubscriber) error {
				go subscriber.OnError(err)
				return nil
			})
		case <-ctx.Done():
			return
		}
	}
}

func (s *scheduler) background(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// poll for next job
			plan, err := s.driver.NextExecution(ctx, time.Now())
			if err != nil || plan == nil || time.Until(plan.PlannedAt) > s.cfg.MaxPlanAhead {
				if err != nil {
					s.onError(err)
				}

				select {
				case <-ctx.Done():
					return
				case <-time.After(s.cfg.IdlePollingInterval):
				}
			}

			if plan == nil || time.Until(plan.PlannedAt) > s.cfg.MaxPlanAhead {
				continue
			}

			jobCtx := &JobContext{
				Context:         ctx,
				Job:             plan.Job,
				Schedule:        plan.Schedule,
				PlannedAt:       plan.PlannedAt,
				ExecutionStatus: JobExecutionStatusInitiated,
			}

			planExecution := true
			err = s.callSubscribers(func(subscriber SchedulerSubscriber) error {
				err := subscriber.OnBeforeJobPlanned(jobCtx)
				if errors.Is(err, ErrSkipExecution) {
					planExecution = false
					return nil
				}

				return err
			})
			if err != nil {
				s.onError(err)
			}

			if planExecution {
				err = s.planner.Plan(jobCtx)
				if err != nil {
					s.errs <- err
				}
			}
		}
	}
}

func (s *scheduler) onError(err error) {
	var e any = err
	if e, ok := e.(interface{ Unwrap() []error }); ok {
		errs := e.Unwrap()
		for _, err := range errs {
			s.onError(err)
		}
	} else {
		s.errs <- err

		if isFatalError(err) {
			s.cancel()
		}
	}
}

func (s *scheduler) callSubscribers(callback func(SchedulerSubscriber) error) error {
	errs := make([]error, 0, len(s.subscribers))
	for _, subscriber := range s.subscribers {
		err := callback(subscriber)

		if err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (s *scheduler) stop() error {
	s.cancel()

	err := s.planner.Stop()
	if err != nil {
		return err
	}

	return s.callSubscribers(func(subscriber SchedulerSubscriber) error {
		return subscriber.OnStop()
	})
}

var _ PlannerSubscriber = &scheduler{}

func NewScheduler(cfg SchedulerConfiguration) Scheduler {
	return &scheduler{
		cfg:         cfg,
		driver:      cfg.DriverFactory(),
		planner:     cfg.PlannerFactory(),
		registry:    make(map[string]Job),
		errs:        make(chan error),
		subscribers: make([]SchedulerSubscriber, 0),
	}
}
