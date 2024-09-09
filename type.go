package gotick

import "time"

type SchedulerConfiguration struct {
	// The maximum number of jobs that can be scheduled at any given time. -1 means no limit.
	MaxJobs int

	// How often the scheduler should check for jobs to run.
	PollInterval time.Duration

	// Threads to use for running jobs.
	Threads int

	// Time for which a job can be planned in advance.
	MaxPlanAhead time.Duration
}
