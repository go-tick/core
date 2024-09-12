package gotick

import "time"

type SchedulerConfiguration struct {
	// Time for which a job can be planned in advance.
	maxPlanAhead time.Duration

	// Polling timeout.
	idlePollingInterval time.Duration

	// Planner factory function.
	plannerFactory func() Planner

	// Driver factory function.
	driverFactory func() SchedulerDriver
}

type SchedulerOption func(*SchedulerConfiguration)

func DefaultConfig(options ...SchedulerOption) SchedulerConfiguration {
	config := SchedulerConfiguration{
		idlePollingInterval: 1 * time.Second,
		maxPlanAhead:        1 * time.Minute,
	}

	config.plannerFactory = func() Planner {
		return newPlanner(1)
	}

	config.driverFactory = func() SchedulerDriver {
		return nil
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
