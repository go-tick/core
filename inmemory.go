package gotick

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

type ScheduleID string
type ExecutionID string

type InMemoryDriver interface {
	SchedulerDriver
	SchedulerSubscriber
}

type inMemoryDriver struct {
	schedule          map[ScheduleID]JobScheduledExecution
	lastExecutions    map[ScheduleID]time.Time
	currentExecutions map[ExecutionID]ScheduleID
	lock              sync.Mutex
}

func (i *inMemoryDriver) OnJobExecutionDelayed(*JobExecutionContext) {
}

func (i *inMemoryDriver) OnJobExecutionInitiated(*JobExecutionContext) {
}

func (i *inMemoryDriver) OnJobExecutionSkipped(ctx *JobExecutionContext) {
	i.onJobExecuted(ctx)
}

func (i *inMemoryDriver) OnBeforeJobExecution(*JobExecutionContext) {
}

func (i *inMemoryDriver) OnBeforeJobExecutionPlanned(*JobExecutionContext) {
}

func (i *inMemoryDriver) OnError(error) {
}

func (i *inMemoryDriver) OnJobExecuted(ctx *JobExecutionContext) {
	i.onJobExecuted(ctx)
}

func (i *inMemoryDriver) OnStart() {
}

func (i *inMemoryDriver) OnStop() {
}

func (i *inMemoryDriver) NextExecution(ctx context.Context) (execution *JobPlannedExecution, err error) {
	i.lock.Lock()
	defer i.lock.Unlock()

	currentlyExecutingScheduleIDs := make(map[ScheduleID]any)
	for _, scheduleID := range i.currentExecutions {
		currentlyExecutingScheduleIDs[scheduleID] = struct{}{}
	}

	toUnschedule := make([]ScheduleID, 0)

	for scheduleID, schedule := range i.schedule {
		if _, ok := currentlyExecutingScheduleIDs[scheduleID]; !ok {
			// find job last execution time and calculate the next one based on it
			// if job was never executed, calculate the next one based on the current time.
			// e.g. for cron 0/5 * * * * (every 5th minute) if last time job was last executed at 12:05:00
			// the next execution will be planned at 12:10:00
			from, ok := i.lastExecutions[scheduleID]
			if !ok {
				from = time.Now()
			}

			next := schedule.Schedule.Next(from)

			// if next is nil, it means that the job is not scheduled anymore
			// so we should remove it from the schedule
			if next == nil {
				toUnschedule = append(toUnschedule, scheduleID)
				continue
			}

			// if execution is nil or next is before the planned execution time
			// we should execute current job next
			if execution == nil || next.Before(execution.PlannedAt) {
				execution = &JobPlannedExecution{
					JobScheduledExecution: schedule,
					ExecutionID:           uuid.NewString(),
					PlannedAt:             *next,
				}
			}
		}
	}

	for _, scheduleID := range toUnschedule {
		delete(i.schedule, scheduleID)
	}

	// if we have found an execution, we should mark it as currently executing
	if execution != nil {
		i.currentExecutions[ExecutionID(execution.ExecutionID)] = ScheduleID(execution.JobScheduledExecution.ScheduleID)
	}

	return
}

func (i *inMemoryDriver) ScheduleJob(ctx context.Context, job Job, schedule JobSchedule) (string, error) {
	i.lock.Lock()
	defer i.lock.Unlock()

	scheduleID := uuid.NewString()
	i.schedule[ScheduleID(scheduleID)] = JobScheduledExecution{
		Job:        job,
		Schedule:   schedule,
		ScheduleID: scheduleID,
	}

	return scheduleID, nil
}

func (i *inMemoryDriver) UnscheduleJobByJobID(ctx context.Context, jobID string) error {
	i.lock.Lock()
	defer i.lock.Unlock()

	scheduleIDs := make([]ScheduleID, 0)
	for scheduleID, schedule := range i.schedule {
		if schedule.Job.ID() == jobID {
			scheduleIDs = append(scheduleIDs, scheduleID)
		}
	}

	for _, scheduleID := range scheduleIDs {
		delete(i.schedule, scheduleID)
	}

	return nil
}

func (i *inMemoryDriver) UnscheduleJobByScheduleID(ctx context.Context, scheduleID string) error {
	i.lock.Lock()
	defer i.lock.Unlock()

	delete(i.schedule, ScheduleID(scheduleID))
	return nil
}

func (i *inMemoryDriver) onJobExecuted(ctx *JobExecutionContext) {
	i.lock.Lock()
	defer i.lock.Unlock()

	delete(i.currentExecutions, ExecutionID(ctx.Execution.ExecutionID))
	i.lastExecutions[ScheduleID(ctx.Execution.ScheduleID)] = ctx.Execution.PlannedAt
}

func newInMemoryDriver() InMemoryDriver {
	return &inMemoryDriver{
		schedule:          make(map[ScheduleID]JobScheduledExecution),
		lastExecutions:    make(map[ScheduleID]time.Time),
		currentExecutions: make(map[ExecutionID]ScheduleID),
	}
}
