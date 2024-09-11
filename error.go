package gotick

import (
	"errors"
	"fmt"
)

var (
	ErrPastTime    = fmt.Errorf("time is in the past")
	ErrInvalidCron = fmt.Errorf("invalid cron schedule")
	ErrJobIDExists = fmt.Errorf("job ID already exists")
	ErrJobNotFound = fmt.Errorf("job not found")
)

type fatalError struct {
	Err error
}

func (f fatalError) Error() string {
	return f.Err.Error()
}

func NewFatalError(err error) error {
	return &fatalError{Err: err}
}

func isFatalError(err error) bool {
	var ferr *fatalError
	return errors.As(err, &ferr)
}
