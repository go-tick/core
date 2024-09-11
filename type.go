package gotick

import (
	"context"
	"time"
)

type JobExecutionStatus int

const (
	JobExecutionStatusInitiated JobExecutionStatus = 1 << iota
	JobExecutionStatusPlanned
	JobExecutionStatusExecuting
	JobExecutionStatusExecuted
	JobExecutionStatusDelayed
	JobExecutionStatusSkipped
	JobExecutionStatusFailed
)

type JobPlannedExecution struct {
	Job       Job
	Schedule  JobSchedule
	PlannedAt time.Time
}

type JobContext struct {
	context.Context
	Job      Job
	Schedule JobSchedule

	PlannedAt  time.Time
	StartedAt  time.Time
	ExecutedAt time.Time

	ExecutionStatus JobExecutionStatus
}
