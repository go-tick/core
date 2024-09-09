package gotick

import "time"

type Job interface {
	ID() string
	Execute() error
}

type JobSchedule interface {
	Schedule() string
	Next(time.Time) *time.Time
}
