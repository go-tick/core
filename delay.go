package gotick

import "time"

type jobScheduleWithMaxDelay struct {
	JobSchedule
	md time.Duration
}

func (d *jobScheduleWithMaxDelay) MaxDelay() time.Duration {
	return d.md
}

func (d *jobScheduleWithMaxDelay) Unwrap() JobSchedule {
	return d.JobSchedule
}

var _ JobSchedule = (*jobScheduleWithMaxDelay)(nil)
var _ MaxDelay = (*jobScheduleWithMaxDelay)(nil)

func NewJobScheduleWithMaxDelay(s JobSchedule, md time.Duration) JobSchedule {
	return &jobScheduleWithMaxDelay{s, md}
}
