package gotick

import (
	"time"

	"github.com/misikdmytro/gotick/internal/utils"
)

type onceStruct struct {
	t time.Time
}

func (o *onceStruct) Next(t time.Time) *time.Time {
	if t.After(o.t) {
		// The time has already passed
		return nil
	}

	return utils.ToPointer(o.t)
}

func (o *onceStruct) Schedule() string {
	return o.t.Format(time.RFC3339)
}

func NewOnce(t time.Time) (JobSchedule, error) {
	if t.Before(time.Now()) {
		return nil, ErrPastTime
	}

	return &onceStruct{t}, nil
}
