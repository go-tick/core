package gotick

import (
	"time"
)

type calendar struct {
	t time.Time
}

func (o *calendar) First() time.Time {
	return o.t
}

func (o *calendar) Next(t time.Time) *time.Time {
	if o.t.After(t) {
		return &o.t
	}

	return nil
}

func (o *calendar) Schedule() string {
	return o.t.Format(time.RFC3339)
}

// NewCalendarSchedule creates a new JobSchedule based on the provided time.
func NewCalendarSchedule(t time.Time) JobSchedule {
	return &calendar{t}
}

// NewCalendarScheduleWithMaxDelay creates a new JobSchedule based on the provided time and max delay.
// Max delay is the maximum delay until the job should be executed. Otherwise, the job treated as delayed.
func NewCalendarScheduleWithMaxDelay(t time.Time, maxDelay time.Duration) JobSchedule {
	c := NewCalendarSchedule(t)
	return NewJobScheduleWithMaxDelay(c, maxDelay)
}
