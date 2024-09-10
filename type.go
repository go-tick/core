package gotick

import (
	"context"
	"time"
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
	ExecutedAt time.Time

	executed chan any
}
