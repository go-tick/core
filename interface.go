package gotick

import (
	"context"
	"time"
)

// BackgroundService is an interface that represents a service that can be started and stopped in the background.
type BackgroundService interface {
	// Start starts the service in the background.
	Start(context.Context) error

	// Stop stops the service.
	Stop() error
}

// ErrorObserver is an interface that represents an observer for errors.
type ErrorObserver interface {
	// OnError is called when an error occurs.
	OnError(error)
}

// BackgroundServiceObserver is an interface that represents an observer for a background service.
type BackgroundServiceObserver interface {
	// OnStart is called when the service is started.
	OnStart()

	// OnStop is called when the service is stopped.
	OnStop()
}

// Publisher is an interface that represents an event publisher.
type Publisher[TObserver any] interface {
	// Subscribe subscribes to the publisher updates.
	Subscribe(TObserver)
}

// Job represents a job that can be executed by a scheduler.
type Job interface {
	// Execute executes the job.
	Execute(*JobExecutionContext)
}

// JobFactory represents a factory that can create a job instance.
type JobFactory interface {
	// Create creates a new job instance with the provided job ID.
	Create(jobID string) Job
}

// JobSchedule represents a schedule for a job.
type JobSchedule interface {
	// Schedule returns the string representation of the schedule.
	Schedule() string

	// First returns the first time the job should be executed.
	First() time.Time

	// Next returns the next time the job should be executed after the provided time.
	// Nil is returned if the job should not be executed anymore.
	Next(time.Time) *time.Time
}

// MaxDelay represents an interface that exposes the maximum delay (used as extension for a schedule).
type MaxDelay interface {
	// MaxDelay returns the maximum delay.
	MaxDelay() time.Duration
}

// Timeout represents an interface that exposes the timeout (used as extension for a schedule).
type Timeout interface {
	// Timeout returns the timeout.
	Timeout() time.Duration
}

// PlannerObserver is an interface that represents a subscriber to a planner.
type PlannerObserver interface {
	// OnJobExecutionUnplanned is called when a job execution was not planned successfully.
	OnJobExecutionUnplanned(*JobExecutionContext)

	// OnBeforeJobExecution is called before a job execution.
	OnBeforeJobExecution(*JobExecutionContext)

	// OnJobExecuted is called when a job is executed.
	OnJobExecuted(*JobExecutionContext)
}

type Planner interface {
	// BackgroundService starts/stops the planner.
	BackgroundService

	// Publisher allows to subscribe to planner events.
	Publisher[PlannerObserver]

	// Plan plans a job execution.
	Plan(*JobExecutionContext)
}

// SchedulerObserver is an interface that represents a subscriber to a scheduler.
type SchedulerObserver interface {
	// PlannerSubscriber is observer of planner events via scheduler.
	PlannerObserver

	// BackgroundServiceObserver is observer of scheduler start/stop events.
	BackgroundServiceObserver

	// OnJobExecutionInitiated is called when a job execution is initiated.
	OnJobExecutionInitiated(*JobExecutionContext)

	// OnJobExecutionDelayed is called when a job execution is delayed.
	OnJobExecutionDelayed(*JobExecutionContext)

	// OnJobExecutionSkipped is called when a job execution is skipped.
	OnJobExecutionSkipped(*JobExecutionContext)

	// OnBeforeJobExecutionPlan is called before a job execution is planned.
	OnBeforeJobExecutionPlan(*JobExecutionContext)
}

// Scheduler is an interface that represents a job scheduler.
type Scheduler interface {
	// BackgroundService starts/stops the scheduler.
	BackgroundService

	// Publisher allows to subscribe to scheduler events.
	Publisher[SchedulerObserver]

	// ScheduleJob schedules a job with the provided schedule.
	// Returns the schedule ID of the scheduled job.
	ScheduleJob(ctx context.Context, jobID string, schedule JobSchedule) (string, error)

	// UnscheduleJobByJobID unschedules a job by its ID.
	UnscheduleJobByJobID(ctx context.Context, jobID string) error

	// UnscheduleJobByScheduleID unschedules a job by its schedule ID.
	UnscheduleJobByScheduleID(ctx context.Context, scheduleID string) error
}

// SchedulerDriver is an interface that represents a driver (storage) for a scheduler that can schedule jobs, unschedule jobs and plans job executions.
type SchedulerDriver interface {
	// BackgroundService starts/stops the driver.
	BackgroundService

	// ScheduleJob schedules a job with the provided schedule.
	ScheduleJob(ctx context.Context, jobID string, schedule JobSchedule) (string, error)

	// UnscheduleJobByJobID unschedules a job by its ID.
	UnscheduleJobByJobID(ctx context.Context, jobID string) error

	// UnscheduleJobByScheduleID unschedules a job by its schedule ID.
	UnscheduleJobByScheduleID(ctx context.Context, scheduleID string) error

	// NextExecution returns the next job execution that should be executed.
	// Nil is returned if currently there are no more job executions.
	NextExecution(context.Context) *NextExecutionResult
}
