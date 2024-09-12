package gotick

import (
	"context"
	"time"
)

type BackgroundService interface {
	Start(context.Context) error
	Stop() error
}
type Job interface {
	ID() string
	Execute(*JobContext)
}

type JobSchedule interface {
	Schedule() string
	Next(time.Time) *time.Time
}
type SchedulerSubscriber interface {
	OnStart()
	OnStop()

	OnBeforeJobPlanned(*JobContext)
	OnBeforeJobExecution(*JobContext)
	OnJobExecuted(*JobContext)

	OnError(error)
}

type Scheduler interface {
	BackgroundService

	Subscribe(SchedulerSubscriber)
	RegisterJob(job Job) error
	ScheduleJob(ctx context.Context, jobID string, schedule JobSchedule) (string, error)
	UnscheduleJobByJobID(ctx context.Context, jobID string) error
	UnscheduleJobByScheduleID(ctx context.Context, scheduleID string) error
}

type PlannerSubscriber interface {
	OnBeforeJobExecution(*JobContext)
	OnJobExecuted(*JobContext)
	OnError(error)
}

type Planner interface {
	BackgroundService

	Subscribe(PlannerSubscriber)
	Plan(*JobContext) error
}

type SchedulerDriver interface {
	ScheduleJob(ctx context.Context, job Job, schedule JobSchedule) (string, error)
	UnscheduleJobByJobID(ctx context.Context, jobID string) error
	UnscheduleJobByScheduleID(ctx context.Context, scheduleID string) error
	NextExecution(context.Context) (*JobPlannedExecution, error)
}
