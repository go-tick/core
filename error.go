package gotick

import (
	"fmt"
)

var (
	ErrPastTime    = fmt.Errorf("time is in the past")
	ErrInvalidCron = fmt.Errorf("invalid cron schedule")
	ErrJobIDExists = fmt.Errorf("job ID already exists")
	ErrJobNotFound = fmt.Errorf("job not found")
)
