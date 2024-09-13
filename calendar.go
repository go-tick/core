package gotick

import (
	"time"

	"github.com/misikdmytro/gotick/internal/utils"
)

type calendar struct {
	t time.Time
}

func (o *calendar) Next(t time.Time) *time.Time {
	if t.After(o.t) {
		// The time has already passed
		return nil
	}

	return utils.ToPointer(o.t)
}

func (o *calendar) Schedule() string {
	return o.t.Format(time.RFC3339)
}

// NewCalendar creates a new JobSchedule based on the provided time.
// If the time is in the past, an error is returned.
func NewCalendar(t time.Time) (JobSchedule, error) {
	if t.Before(time.Now()) {
		return nil, ErrPastTime
	}

	return &calendar{t}, nil
}
