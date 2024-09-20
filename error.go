package gotick

import (
	"fmt"
)

var (
	ErrInvalidCron             = fmt.Errorf("invalid cron schedule")
	ErrInvalidSequenceSchedule = fmt.Errorf("invalid sequence schedule")
)
