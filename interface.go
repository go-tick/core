package gotick

type Scheduler interface {
	ScheduleJob(job Job) error
	RemoveJob(jobID string) error
	Start() error
	Stop() error
}
