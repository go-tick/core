package gotick

import "time"

type jobScheduleWithMaxDelay struct {
	JobSchedule
	md time.Duration
}

type scheduleWithTimeout struct {
	JobSchedule
	td time.Duration
}

func (d *jobScheduleWithMaxDelay) MaxDelay() time.Duration {
	return d.md
}

func (d *jobScheduleWithMaxDelay) Unwrap() JobSchedule {
	return d.JobSchedule
}

func (t *scheduleWithTimeout) Timeout() time.Duration {
	return t.td
}

func (t *scheduleWithTimeout) Unwrap() JobSchedule {
	return t.JobSchedule
}

var _ JobSchedule = (*jobScheduleWithMaxDelay)(nil)
var _ MaxDelay = (*jobScheduleWithMaxDelay)(nil)
var _ JobSchedule = (*scheduleWithTimeout)(nil)
var _ Timeout = (*scheduleWithTimeout)(nil)

func NewJobScheduleWithMaxDelay(s JobSchedule, md time.Duration) JobSchedule {
	return &jobScheduleWithMaxDelay{s, md}
}

func NewJobScheduleWithTimeout(s JobSchedule, td time.Duration) JobSchedule {
	return &scheduleWithTimeout{s, td}
}

func MaxDelayFromJobSchedule(s JobSchedule) (time.Duration, bool) {
	if d, ok := s.(MaxDelay); ok {
		return d.MaxDelay(), true
	}

	if u, ok := s.(interface{ Unwrap() JobSchedule }); ok {
		return MaxDelayFromJobSchedule(u.Unwrap())
	}

	return 0, false
}

func TimeoutFromJobSchedule(s JobSchedule) (time.Duration, bool) {
	if t, ok := s.(Timeout); ok {
		return t.Timeout(), true
	}

	if u, ok := s.(interface{ Unwrap() JobSchedule }); ok {
		return TimeoutFromJobSchedule(u.Unwrap())
	}

	return 0, false
}
