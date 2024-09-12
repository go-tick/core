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
	// Time for which a job can be planned in advance.
	maxPlanAhead time.Duration

	// Polling timeout.
	idlePollingInterval time.Duration

	// Planner factory function.
	plannerFactory func() Planner

	// Driver factory function.
	driverFactory func() SchedulerDriver

	// Default scheduler subscribers.
	subscribers []SchedulerSubscriber

	// Delayed strategy.
	delayedStrategy ScheduleDelayedStrategy
}

type SchedulerOption func(*SchedulerConfiguration)

func DefaultConfig(options ...SchedulerOption) SchedulerConfiguration {
	config := SchedulerConfiguration{
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
		option(&config)
	}

	return config
}

func WithMaxPlanAhead(maxPlanAhead time.Duration) SchedulerOption {
	return func(config *SchedulerConfiguration) {
		config.maxPlanAhead = maxPlanAhead
	}
}

func WithIdlePollingInterval(idlePollingInterval time.Duration) SchedulerOption {
	return func(config *SchedulerConfiguration) {
		config.idlePollingInterval = idlePollingInterval
	}
}

func WithDefaultPlannerFactory(threads uint) SchedulerOption {
	return func(config *SchedulerConfiguration) {
		config.plannerFactory = func() Planner {
			return newPlanner(threads)
		}
	}
}

func WithPlannerFactory(factory func() Planner) SchedulerOption {
	return func(config *SchedulerConfiguration) {
		config.plannerFactory = factory
	}
}

func WithDriverFactory(factory func() SchedulerDriver) SchedulerOption {
	return func(config *SchedulerConfiguration) {
		config.driverFactory = factory
	}
}

func WithSubscribers(subscribers ...SchedulerSubscriber) SchedulerOption {
	return func(config *SchedulerConfiguration) {
		config.subscribers = subscribers
	}
}

func WithDelayedStrategy(strategy ScheduleDelayedStrategy) SchedulerOption {
	return func(config *SchedulerConfiguration) {
		config.delayedStrategy = strategy
	}
}
