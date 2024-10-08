package gotick

import (
	"strings"
	"time"
)

type sequence struct {
	t []time.Time
}

func (s *sequence) First() time.Time {
	return s.t[0]
}

func (s *sequence) Next(t time.Time) *time.Time {
	for _, st := range s.t {
		if st.After(t) {
			return &st
		}
	}

	return nil
}

func (s *sequence) Schedule() string {
	r := make([]string, len(s.t))
	for i, t := range s.t {
		r[i] = t.Format(time.RFC3339)
	}

	return strings.Join(r, ",")
}

// NewSequenceSchedule creates a new JobSchedule based on the provided sequence of times.
// Sequence should be in ascending order.
func NewSequenceSchedule(t ...time.Time) (JobSchedule, error) {
	for i := len(t) - 1; i > 0; i-- {
		if t[i].Before(t[i-1]) {
			return nil, ErrInvalidSequenceSchedule
		}
	}

	return &sequence{t}, nil
}

// NewSequenceScheduleWithMaxDelay creates a new JobSchedule based on the provided sequence of times and max delay.
// Sequence should be in ascending order.
// Max delay is the maximum delay until the job should be executed. Otherwise, the job treated as delayed.
func NewSequenceScheduleWithMaxDelay(maxDelay time.Duration, t ...time.Time) (JobSchedule, error) {
	s, err := NewSequenceSchedule(t...)
	if err != nil {
		return nil, err
	}

	return NewJobScheduleWithMaxDelay(s, maxDelay), nil
}
