package gotick

import "fmt"

var (
	ErrPastTime    = fmt.Errorf("time is in the past")
	ErrInvalidCron = fmt.Errorf("invalid cron schedule")
)
