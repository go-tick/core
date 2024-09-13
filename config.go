package gotick

import (
	"time"
)

type ScheduleDelayedStrategy int

const (
	ScheduleDelayedStrategySkip ScheduleDelayedStrategy = iota
	ScheduleDelayedStrategyPlan
)

type SchedulerConfiguration struct {
	maxPlanAhead        time.Duration
	idlePollingInterval time.Duration
	plannerFactory      func() Planner
	driverFactory       func() SchedulerDriver
	subscribers         []SchedulerSubscriber
	delayedStrategy     ScheduleDelayedStrategy
}

type SchedulerOption func(*SchedulerConfiguration)

// DefaultConfig returns the default configuration for the scheduler.
// Some options can be overridden by passing the appropriate SchedulerOption.
func DefaultConfig(options ...SchedulerOption) *SchedulerConfiguration {
	config := &SchedulerConfiguration{
		idlePollingInterval: 1 * time.Second,
		maxPlanAhead:        1 * time.Minute,
		subscribers:         make([]SchedulerSubscriber, 0),
		delayedStrategy:     ScheduleDelayedStrategyPlan,
	}

	config.plannerFactory = func() Planner {
		return newPlanner(1)
	}

	config.driverFactory = func() SchedulerDriver {
		driver := newInMemoryDriver()
		config.subscribers = append(config.subscribers, driver)
		return driver
	}

	for _, option := range options {
		option(config)
	}

	return config
}

// WithMaxPlanAhead sets the maximum time for which a job can be planned in advance.
// This is used to prevent the planner from planning too far ahead.
func WithMaxPlanAhead(maxPlanAhead time.Duration) SchedulerOption {
	return func(config *SchedulerConfiguration) {
		config.maxPlanAhead = maxPlanAhead
	}
}

// WithIdlePollingInterval sets the polling timeout.
// If there are no executions to plan the planner will sleep for this duration.
func WithIdlePollingInterval(idlePollingInterval time.Duration) SchedulerOption {
	return func(config *SchedulerConfiguration) {
		config.idlePollingInterval = idlePollingInterval
	}
}

// WithDefaultPlannerFactory sets the default planner factory with the given number of threads.
func WithDefaultPlannerFactory(threads uint) SchedulerOption {
	return func(config *SchedulerConfiguration) {
		config.plannerFactory = func() Planner {
			return newPlanner(threads)
		}
	}
}

// WithPlannerFactory sets the planner factory.
func WithPlannerFactory(factory func() Planner) SchedulerOption {
	return func(config *SchedulerConfiguration) {
		config.plannerFactory = factory
	}
}

// WithDriverFactory sets the driver factory.
func WithDriverFactory(factory func() SchedulerDriver) SchedulerOption {
	return func(config *SchedulerConfiguration) {
		config.driverFactory = factory
	}
}

// WithSubscribers adds the given subscribers to the scheduler.
func WithSubscribers(subscribers ...SchedulerSubscriber) SchedulerOption {
	return func(config *SchedulerConfiguration) {
		config.subscribers = append(config.subscribers, subscribers...)
	}
}

// WithDelayedStrategy sets the strategy to use when a job is delayed.
func WithDelayedStrategy(strategy ScheduleDelayedStrategy) SchedulerOption {
	return func(config *SchedulerConfiguration) {
		config.delayedStrategy = strategy
	}
}
