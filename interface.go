package gotick

import (
	"context"
	"time"
)

type Job interface {
	ID() string
	Execute(context.Context) error
}

type JobSchedule interface {
	Schedule() string
	Next(time.Time) *time.Time
}

type Scheduler interface {
	RegisterJob(job Job) error
	ScheduleJob(ctx context.Context, jobID string, schedule JobSchedule) error
	RemoveJob(ctx context.Context, jobID string) error
	Start(context.Context) error
	Stop() error
	Errs() <-chan error
}

type Planner interface {
	Plan(context.Context, Job, time.Time) (<-chan any, error)
	Start(context.Context) error
	Stop() error
	Errs() <-chan error
}

type SchedulerDriver interface {
	ScheduleJob(ctx context.Context, job Job, schedule JobSchedule) error
	RemoveJob(ctx context.Context, jobID string) error
	NextExecution(context.Context, time.Time) (Job, time.Time, error)
	Executed(context.Context, Job, time.Time) error
}
