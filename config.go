package gotick

import (
	"time"
)

type ScheduleDelayedStrategy int

const (
	ScheduleDelayedStrategySkip ScheduleDelayedStrategy = iota
	ScheduleDelayedStrategyPlan
)

type Option[T any] func(*T)

type SchedulerConfig struct {
	maxPlanAhead        time.Duration
	idlePollingInterval time.Duration
	plannerFactory      func(*SchedulerConfig) (Planner, error)
	driverFactory       func(*SchedulerConfig) (SchedulerDriver, error)
	subscribers         []SchedulerSubscriber
	delayedStrategy     ScheduleDelayedStrategy
	threads             uint
}

type InMemoryDriverConfig struct {
	lockTimeout time.Duration
}

type PlannerConfig struct {
	jobs []Job

	threads     uint
	planTimeout time.Duration
}

// DefaultSchedulerConfig returns the default configuration for the scheduler.
// Some options can be overridden by passing the appropriate options.
func DefaultSchedulerConfig(options ...Option[SchedulerConfig]) *SchedulerConfig {
	config := &SchedulerConfig{
		idlePollingInterval: 1 * time.Second,
		maxPlanAhead:        1 * time.Minute,
		subscribers:         make([]SchedulerSubscriber, 0),
		delayedStrategy:     ScheduleDelayedStrategyPlan,
		threads:             1,
	}

	WithDefaultPlannerFactory(DefaultPlannerConfig())(config)
	WithInMemoryDriverFactory(DefaultInMemoryConfig())(config)

	for _, option := range options {
		option(config)
	}

	return config
}

// WithMaxPlanAhead sets the maximum time for which a job can be planned in advance.
// This is used to prevent the planner from planning too far ahead.
func WithMaxPlanAhead(maxPlanAhead time.Duration) Option[SchedulerConfig] {
	return func(config *SchedulerConfig) {
		config.maxPlanAhead = maxPlanAhead
	}
}

// WithIdlePollingInterval sets the polling timeout.
// If there are no executions to plan the planner will sleep for this duration.
func WithIdlePollingInterval(idlePollingInterval time.Duration) Option[SchedulerConfig] {
	return func(config *SchedulerConfig) {
		config.idlePollingInterval = idlePollingInterval
	}
}

// WithPlannerFactory sets the planner factory.
func WithPlannerFactory(factory func(*SchedulerConfig) (Planner, error)) Option[SchedulerConfig] {
	return func(config *SchedulerConfig) {
		config.plannerFactory = factory
	}
}

// WithDefaultPlannerFactory sets the default planner factory.
func WithDefaultPlannerFactory(cfg *PlannerConfig) Option[SchedulerConfig] {
	return WithPlannerFactory(func(*SchedulerConfig) (Planner, error) {
		return newPlanner(cfg)
	})
}

// WithDriverFactory sets the driver factory.
func WithDriverFactory(factory func(*SchedulerConfig) (SchedulerDriver, error)) Option[SchedulerConfig] {
	return func(config *SchedulerConfig) {
		config.driverFactory = factory
	}
}

// WithInMemoryDriverFactory sets the in-memory driver factory.
func WithInMemoryDriverFactory(cfg *InMemoryDriverConfig) Option[SchedulerConfig] {
	return WithDriverFactory(func(config *SchedulerConfig) (SchedulerDriver, error) {
		driver, err := newInMemoryDriver(cfg)
		if err != nil {
			return nil, err
		}

		config.subscribers = append(config.subscribers, driver)
		return driver, nil
	})
}

// WithSubscribers adds the given subscribers to the scheduler.
func WithSubscribers(subscribers ...SchedulerSubscriber) Option[SchedulerConfig] {
	return func(config *SchedulerConfig) {
		config.subscribers = append(config.subscribers, subscribers...)
	}
}

// WithDelayedStrategy sets the strategy to use when a job is delayed.
func WithDelayedStrategy(strategy ScheduleDelayedStrategy) Option[SchedulerConfig] {
	return func(config *SchedulerConfig) {
		config.delayedStrategy = strategy
	}
}

// WithThreads sets the number of threads to use in the scheduler for next execution evaluations.
func WithThreads(threads uint) Option[SchedulerConfig] {
	return func(config *SchedulerConfig) {
		config.threads = threads
	}
}

// DefaultInMemoryConfig returns the default configuration for the in-memory driver.
// Some options can be overridden by passing the appropriate options.
func DefaultInMemoryConfig(options ...Option[InMemoryDriverConfig]) *InMemoryDriverConfig {
	config := &InMemoryDriverConfig{
		lockTimeout: 1 * time.Hour,
	}

	for _, option := range options {
		option(config)
	}

	return config
}

// WithScheduleLockTimeout sets the timeout for the in-memory driver for the schedule to be locked after job is planned for next execution.
// In case job is not finished after the timeout, race condition may occur.
func WithScheduleLockTimeout(timeout time.Duration) Option[InMemoryDriverConfig] {
	return func(config *InMemoryDriverConfig) {
		config.lockTimeout = timeout
	}
}

// DefaultPlannerConfig returns the default configuration for the planner.
// Some options can be overridden by passing the appropriate options.
func DefaultPlannerConfig(options ...Option[PlannerConfig]) *PlannerConfig {
	config := &PlannerConfig{
		threads:     1,
		planTimeout: 5 * time.Second,
	}

	for _, option := range options {
		option(config)
	}

	return config
}

// WithPlannerThreads sets the number of threads to use in the planner for next execution evaluations.
func WithPlannerThreads(threads uint) Option[PlannerConfig] {
	return func(config *PlannerConfig) {
		config.threads = threads
	}
}

// WithPlannerTimeout sets the timeout for the planner to plan a job.
func WithPlannerTimeout(timeout time.Duration) Option[PlannerConfig] {
	return func(config *PlannerConfig) {
		config.planTimeout = timeout
	}
}

// WithJobs registers the given jobs in the planner.
// Required for the planner to be able to plan the jobs.
func WithJobs(jobs ...Job) Option[PlannerConfig] {
	return func(config *PlannerConfig) {
		config.jobs = append(config.jobs, jobs...)
	}
}
