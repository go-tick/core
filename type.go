package gotick

import (
	"context"
	"time"
)

type JobExecutionStatus int

const (
	JobExecutionStatusInitiated JobExecutionStatus = 1 << iota
	JobExecutionStatusDelayed
	JobExecutionStatusSkipped
	JobExecutionStatusPlanned
	JobExecutionStatusUnplanned
	JobExecutionStatusExecuting
	JobExecutionStatusExecuted
)

type JobScheduledExecution struct {
	Job        Job
	Schedule   JobSchedule
	ScheduleID string
}

type JobPlannedExecution struct {
	JobScheduledExecution

	ExecutionID string
	PlannedAt   time.Time
}

type JobExecutionContext struct {
	context.Context

	Execution JobPlannedExecution

	StartedAt  time.Time
	ExecutedAt time.Time

	ExecutionStatus JobExecutionStatus
}
