package gotick

import "time"

type scheduleWithTimeout struct {
	JobSchedule
	td time.Duration
}

func (t *scheduleWithTimeout) Timeout() time.Duration {
	return t.td
}

func (t *scheduleWithTimeout) Unwrap() JobSchedule {
	return t.JobSchedule
}

var _ JobSchedule = (*scheduleWithTimeout)(nil)
var _ Timeout = (*scheduleWithTimeout)(nil)

func NewScheduleWithTimeout(s JobSchedule, td time.Duration) JobSchedule {
	return &scheduleWithTimeout{s, td}
}
