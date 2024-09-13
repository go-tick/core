package gotick

import (
	"fmt"
)

var (
	ErrInvalidCron = fmt.Errorf("invalid cron schedule")
	ErrJobIDExists = fmt.Errorf("job ID already exists")
	ErrJobNotFound = fmt.Errorf("job not found")
)
