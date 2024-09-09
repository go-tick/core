package gotick

import (
	"time"

	"github.com/misikdmytro/gotick/internal/utils"
	"github.com/robfig/cron/v3"
)

type cronStruct struct {
	s        string
	schedule cron.Schedule
}

func (c *cronStruct) Next(t time.Time) *time.Time {
	return utils.ToPointer(c.schedule.Next(t))
}

func (c *cronStruct) Schedule() string {
	return c.s
}

func NewCron(s string) (JobSchedule, error) {
	schedule, err := cron.ParseStandard(s)
	if err != nil {
		return nil, ErrInvalidCron
	}

	return &cronStruct{s, schedule}, nil
}
