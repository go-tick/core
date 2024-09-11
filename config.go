package gotick

import "time"

type SchedulerConfiguration struct {
	// Time for which a job can be planned in advance.
	MaxPlanAhead time.Duration

	// Polling timeout.
	IdlePollingInterval time.Duration

	// Planner factory function.
	PlannerFactory func() Planner

	// Driver factory function.
	DriverFactory func() SchedulerDriver
}

type SchedulerOption func(*SchedulerConfiguration)

func DefaultConfig(options ...SchedulerOption) SchedulerConfiguration {
	config := SchedulerConfiguration{
		IdlePollingInterval: 1 * time.Second,
		MaxPlanAhead:        1 * time.Minute,
	}

	config.PlannerFactory = func() Planner {
		return NewPlanner(1)
	}

	config.DriverFactory = func() SchedulerDriver {
		return nil
	}

	for _, option := range options {
		option(&config)
	}

	return config
}

func WithMaxPlanAhead(maxPlanAhead time.Duration) SchedulerOption {
	return func(config *SchedulerConfiguration) {
		config.MaxPlanAhead = maxPlanAhead
	}
}

func WithIdlePollingInterval(idlePollingInterval time.Duration) SchedulerOption {
	return func(config *SchedulerConfiguration) {
		config.IdlePollingInterval = idlePollingInterval
	}
}

func WithDefaultPlannerFactory(threads uint) SchedulerOption {
	return func(config *SchedulerConfiguration) {
		config.PlannerFactory = func() Planner {
			return NewPlanner(threads)
		}
	}
}

func WithPlannerFactory(factory func() Planner) SchedulerOption {
	return func(config *SchedulerConfiguration) {
		config.PlannerFactory = factory
	}
}

func WithDriverFactory(factory func() SchedulerDriver) SchedulerOption {
	return func(config *SchedulerConfiguration) {
		config.DriverFactory = factory
	}
}
