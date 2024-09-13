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

func (j *JobExecutionContext) Clone() *JobExecutionContext {
	return &JobExecutionContext{
		Context: j.Context,
		Execution: JobPlannedExecution{
			JobScheduledExecution: JobScheduledExecution{
				Job:        j.Execution.Job,
				Schedule:   j.Execution.Schedule,
				ScheduleID: j.Execution.ScheduleID,
			},
			ExecutionID: j.Execution.ExecutionID,
			PlannedAt:   j.Execution.PlannedAt,
		},
		StartedAt:       j.StartedAt,
		ExecutedAt:      j.ExecutedAt,
		ExecutionStatus: j.ExecutionStatus,
	}
}
