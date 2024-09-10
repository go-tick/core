package gotick

import (
	"context"
	"time"
)

type Job interface {
	ID() string
	Execute(JobContext) error
}

type JobSchedule interface {
	Schedule() string
	Next(time.Time) *time.Time
}

type Scheduler interface {
	RegisterJob(job Job) error
	ScheduleJob(ctx context.Context, jobID string, schedule JobSchedule) error
	UnscheduleJob(ctx context.Context, jobID string) error
	Start(context.Context) error
	Stop() error
	Errs() <-chan error
}

type Planner interface {
	Plan(JobContext) error
	Start(context.Context) error
	Stop() error
	Errs() <-chan error
}

type SchedulerDriver interface {
	ScheduleJob(ctx context.Context, job Job, schedule JobSchedule) error
	UnscheduleJob(ctx context.Context, jobID string) error
	NextExecution(context.Context, time.Time) (*JobPlannedExecution, error)

	BeforeExecution(JobContext) error
	Executed(JobContext) error
}
