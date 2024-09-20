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

type NextExecutionResult struct {
	JobID JobID

	Schedule   JobSchedule
	ScheduleID string

	PlannedAt time.Time
}

type JobExecutionContext struct {
	context.Context

	JobID       JobID
	ScheduleID  string
	ExecutionID string

	Schedule JobSchedule

	PlannedAt  time.Time
	StartedAt  time.Time
	ExecutedAt time.Time

	ExecutionStatus JobExecutionStatus
}
