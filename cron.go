package gotick

import (
	"time"

	"github.com/go-tick/core/internal/utils"
	"github.com/robfig/cron/v3"
)

type cronStruct struct {
	createdAt time.Time
	s         string
	sch       cron.Schedule
}

func (c *cronStruct) First() time.Time {
	return c.sch.Next(c.createdAt)
}

func (c *cronStruct) Next(t time.Time) *time.Time {
	return utils.ToPointer(c.sch.Next(t))
}

func (c *cronStruct) Schedule() string {
	return c.s
}

// NewCronSchedule creates a new JobSchedule based on the provided cron string.
// If the cron string is invalid, an error is returned.
func NewCronSchedule(s string) (JobSchedule, error) {
	schedule, err := cron.ParseStandard(s)
	if err != nil {
		return nil, ErrInvalidCron
	}

	return &cronStruct{time.Now(), s, schedule}, nil
}
