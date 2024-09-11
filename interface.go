package gotick

import (
	"context"
	"time"
)

type Job interface {
	ID() string
	Execute(*JobContext) error
}

type JobSchedule interface {
	Schedule() string
	Next(time.Time) *time.Time
}

type SchedulerSubscriber interface {
	OnStart() error
	OnStop() error

	OnBeforeJobPlanned(*JobContext) error
	OnBeforeJobExecution(*JobContext) error
	OnJobExecuted(*JobContext) error

	OnError(error)
}

type Scheduler interface {
	Subscribe(SchedulerSubscriber)
	RegisterJob(job Job) error
	ScheduleJob(ctx context.Context, jobID string, schedule JobSchedule) (string, error)
	UnscheduleJobByJobID(ctx context.Context, jobID string) error
	UnscheduleJobByScheduleID(ctx context.Context, scheduleID string) error
	Start(context.Context) error
	Stop() error
}

type PlannerSubscriber interface {
	OnBeforeJobExecution(*JobContext)
	OnJobExecuted(*JobContext)
	OnError(error)
}

type Planner interface {
	Subscribe(PlannerSubscriber)
	Plan(*JobContext) error
	Start(context.Context) error
	Stop() error
}

type SchedulerDriver interface {
	ScheduleJob(ctx context.Context, job Job, schedule JobSchedule) (string, error)
	UnscheduleJobByJobID(ctx context.Context, jobID string) error
	UnscheduleJobByScheduleID(ctx context.Context, scheduleID string) error
	NextExecution(context.Context, time.Time) (*JobPlannedExecution, error)
}
