package gotick

import (
	"fmt"
)

var (
	ErrInvalidCron             = fmt.Errorf("invalid cron schedule")
	ErrDuplicateJobID          = fmt.Errorf("duplicate job ID")
	ErrInvalidSequenceSchedule = fmt.Errorf("invalid sequence schedule")
)
