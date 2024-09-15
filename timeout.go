package gotick

import "time"

type jobWithTimeout struct {
	Job
	td time.Duration
}

func (t *jobWithTimeout) Timeout() time.Duration {
	return t.td
}

func (t *jobWithTimeout) Unwrap() Job {
	return t.Job
}

var _ Job = (*jobWithTimeout)(nil)
var _ Timeout = (*jobWithTimeout)(nil)

func NewJobWithTimeout(s Job, td time.Duration) Job {
	return &jobWithTimeout{s, td}
}
