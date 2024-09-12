package gotick

import (
	"context"
	"sync"
	"time"
)

type scheduler struct {
	cfg         SchedulerConfiguration
	cancel      context.CancelFunc
	driver      SchedulerDriver
	planner     Planner
	registry    map[string]Job
	subscribers []SchedulerSubscriber
	startOnce   sync.Once
	stopOnce    sync.Once
}

func (s *scheduler) OnError(err error) {
	s.onError(err)
}

func (s *scheduler) OnBeforeJobExecution(ctx *JobContext) {
	s.callSubscribers(func(subscriber SchedulerSubscriber) {
		subscriber.OnBeforeJobExecution(ctx)
	})
}

func (s *scheduler) OnJobExecuted(ctx *JobContext) {
	s.callSubscribers(func(subscriber SchedulerSubscriber) {
		subscriber.OnJobExecuted(ctx)
	})
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

func (s *scheduler) UnscheduleJobByScheduleID(ctx context.Context, scheduleID string) error {
	return s.driver.UnscheduleJobByScheduleID(ctx, scheduleID)
}

func (s *scheduler) ScheduleJob(ctx context.Context, jobID string, schedule JobSchedule) (string, error) {
	if job, ok := s.registry[jobID]; !ok {
		return "", ErrJobNotFound
	} else {
		return s.driver.ScheduleJob(ctx, job, schedule)
	}
}

func (s *scheduler) Start(ctx context.Context) (err error) {
	s.startOnce.Do(func() { err = s.start(ctx) })
	return
}

func (s *scheduler) Stop() (err error) {
	s.stopOnce.Do(func() { err = s.stop() })
	return
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

			s.callSubscribers(func(subscriber SchedulerSubscriber) {
				subscriber.OnBeforeJobPlanned(jobCtx.Clone())
			})

			err = s.planner.Plan(jobCtx)
			if err != nil {
				s.onError(err)
			}
		}
	}
}

func (s *scheduler) onError(err error) {
	s.callSubscribers(func(subscriber SchedulerSubscriber) {
		subscriber.OnError(err)
	})
}

func (s *scheduler) callSubscribers(callback func(SchedulerSubscriber)) {
	for _, subscriber := range s.subscribers {
		callback(subscriber)
	}
}

func (s *scheduler) start(ctx context.Context) (err error) {
	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	s.callSubscribers(func(subscriber SchedulerSubscriber) {
		subscriber.OnStart()
	})

	err = s.planner.Start(ctx)
	s.planner.Subscribe(s)

	go s.background(ctx)

	return
}

func (s *scheduler) stop() (err error) {
	s.cancel()

	err = s.planner.Stop()
	s.callSubscribers(func(subscriber SchedulerSubscriber) {
		subscriber.OnStop()
	})

	return
}

var _ PlannerSubscriber = &scheduler{}

func NewScheduler(cfg SchedulerConfiguration) Scheduler {
	return &scheduler{
		cfg:         cfg,
		driver:      cfg.DriverFactory(),
		planner:     cfg.PlannerFactory(),
		registry:    make(map[string]Job),
		subscribers: make([]SchedulerSubscriber, 0),
		cancel:      func() {},
	}
}
