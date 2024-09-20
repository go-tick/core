package gotick

import (
	"context"
	"errors"
	"slices"
	"sync"
	"time"

	"github.com/google/uuid"
)

type scheduler struct {
	cfg         *SchedulerConfig
	cancel      context.CancelFunc
	driver      SchedulerDriver
	planner     Planner
	subscribers []SchedulerObserver
	startOnce   sync.Once
	stopOnce    sync.Once
}

func (s *scheduler) OnJobExecutionUnplanned(ctx *JobExecutionContext) {
	s.callSubscribers(func(subscriber SchedulerObserver) {
		subscriber.OnJobExecutionUnplanned(ctx)
	})
}

func (s *scheduler) OnBeforeJobExecution(ctx *JobExecutionContext) {
	s.callSubscribers(func(subscriber SchedulerObserver) {
		subscriber.OnBeforeJobExecution(ctx)
	})
}

func (s *scheduler) OnJobExecuted(ctx *JobExecutionContext) {
	s.callSubscribers(func(subscriber SchedulerObserver) {
		subscriber.OnJobExecuted(ctx)
	})
}

func (s *scheduler) Subscribe(subscriber SchedulerObserver) {
	s.subscribers = append(s.subscribers, subscriber)
}

func (s *scheduler) UnscheduleJobByJobID(ctx context.Context, jobID string) error {
	return s.driver.UnscheduleJobByJobID(ctx, jobID)
}

func (s *scheduler) UnscheduleJobByScheduleID(ctx context.Context, scheduleID string) error {
	return s.driver.UnscheduleJobByScheduleID(ctx, scheduleID)
}

func (s *scheduler) ScheduleJob(ctx context.Context, jobID string, schedule JobSchedule) (string, error) {
	return s.driver.ScheduleJob(ctx, jobID, schedule)
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
			plan := s.driver.NextExecution(ctx)
			if plan == nil || time.Until(plan.PlannedAt) > s.cfg.maxPlanAhead {
				select {
				case <-ctx.Done():
					return
				case <-time.After(s.cfg.idlePollingInterval):
				}
			}

			if plan == nil || time.Until(plan.PlannedAt) > s.cfg.maxPlanAhead {
				continue
			}

			jobCtx := &JobExecutionContext{
				Context: ctx,

				JobID:       plan.JobID,
				ScheduleID:  plan.ScheduleID,
				ExecutionID: uuid.NewString(),

				Schedule: plan.Schedule,

				PlannedAt: plan.PlannedAt,

				ExecutionStatus: JobExecutionStatusInitiated,
			}

			s.callSubscribers(func(subscriber SchedulerObserver) {
				subscriber.OnJobExecutionInitiated(jobCtx)
			})

			if s.isJobDelayed(plan) {
				jobCtx.ExecutionStatus = JobExecutionStatusDelayed

				s.callSubscribers(func(subscriber SchedulerObserver) {
					subscriber.OnJobExecutionDelayed(jobCtx)
				})

				if s.cfg.delayedStrategy == ScheduleDelayedStrategySkip {
					jobCtx.ExecutionStatus = JobExecutionStatusSkipped

					s.callSubscribers(func(subscriber SchedulerObserver) {
						subscriber.OnJobExecutionSkipped(jobCtx)
					})

					continue
				}
			}

			s.callSubscribers(func(subscriber SchedulerObserver) {
				subscriber.OnBeforeJobExecutionPlan(jobCtx)
			})

			s.planner.Plan(jobCtx)
		}
	}
}

func (s *scheduler) callSubscribers(callback func(SchedulerObserver)) {
	for _, subscriber := range s.subscribers {
		callback(subscriber)
	}
}

func (s *scheduler) start(ctx context.Context) (err error) {
	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	s.callSubscribers(func(subscriber SchedulerObserver) {
		subscriber.OnStart()
	})

	err = s.planner.Start(ctx)
	if err != nil {
		return
	}
	s.planner.Subscribe(s)

	err = s.driver.Start(ctx)
	if err != nil {
		return
	}

	for range s.cfg.threads {
		go s.background(ctx)
	}

	return
}

func (s *scheduler) stop() error {
	s.cancel()

	err1 := s.driver.Stop()
	err2 := s.planner.Stop()
	s.callSubscribers(func(subscriber SchedulerObserver) {
		subscriber.OnStop()
	})

	return errors.Join(err1, err2)
}

func (s *scheduler) isJobDelayed(plan *NextExecutionResult) bool {
	if time.Since(plan.PlannedAt) < 0 {
		return false
	}

	// the algorithm is next:
	// if schedule has max delay, check if the delay is exceeded.
	// otherwise check if the next execution after the planned time is in the past.
	schedule := plan.Schedule
	if maxDelay, ok := MaxDelayFromJobSchedule(schedule); ok {
		return time.Since(plan.PlannedAt) > maxDelay
	}

	next := schedule.Next(plan.PlannedAt)
	return next != nil && next.Before(time.Now())
}

var _ PlannerObserver = &scheduler{}

func NewScheduler(cfg *SchedulerConfig) (Scheduler, error) {
	driver, err := cfg.driverFactory(cfg)
	if err != nil {
		return nil, err
	}

	planner, err := cfg.plannerFactory(cfg)
	if err != nil {
		return nil, err
	}

	return &scheduler{
		cfg:         cfg,
		driver:      driver,
		planner:     planner,
		subscribers: slices.Clone(cfg.subscribers),
		cancel:      func() {},
	}, nil
}
