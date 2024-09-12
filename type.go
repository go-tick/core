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

func (j *JobContext) Clone() *JobContext {
	return &JobContext{
		Context:         j.Context,
		Job:             j.Job,
		Schedule:        j.Schedule,
		PlannedAt:       j.PlannedAt,
		StartedAt:       j.StartedAt,
		ExecutedAt:      j.ExecutedAt,
		ExecutionStatus: j.ExecutionStatus,
	}
}
