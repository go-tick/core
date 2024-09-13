package gotick

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/misikdmytro/gotick/internal/utils"
)

type scheduleID string
type executionID string

type inMemoryDriver struct {
	schedule          map[scheduleID]JobScheduledExecution
	lastExecutions    map[scheduleID]time.Time
	currentExecutions map[executionID]scheduleID
	lock              sync.Mutex
}

func (i *inMemoryDriver) OnJobExecutionDelayed(*JobExecutionContext) {
}

func (i *inMemoryDriver) OnJobExecutionInitiated(ctx *JobExecutionContext) {
	i.lock.Lock()
	defer i.lock.Unlock()

	i.currentExecutions[executionID(ctx.Execution.ExecutionID)] = scheduleID(ctx.Execution.JobScheduledExecution.ScheduleID)
}

func (i *inMemoryDriver) OnJobExecutionSkipped(ctx *JobExecutionContext) {
	i.onJobExecuted(ctx)
}

func (i *inMemoryDriver) OnBeforeJobExecution(*JobExecutionContext) {
}

func (i *inMemoryDriver) OnJobExecutionNotPlanned(ctx *JobExecutionContext) {
	i.lock.Lock()
	defer i.lock.Unlock()

	delete(i.currentExecutions, executionID(ctx.Execution.ExecutionID))
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

func (i *inMemoryDriver) NextExecution(ctx context.Context) (execution *JobPlannedExecution) {
	i.lock.Lock()
	defer i.lock.Unlock()

	currentlyExecutingScheduleIDs := make(map[scheduleID]any)
	for _, scheduleID := range i.currentExecutions {
		currentlyExecutingScheduleIDs[scheduleID] = struct{}{}
	}

	toUnschedule := make([]scheduleID, 0)

	for scheduleID, schedule := range i.schedule {
		if _, ok := currentlyExecutingScheduleIDs[scheduleID]; !ok {
			// find job last execution time and calculate the next one based on it
			// if job was never executed, use the first execution time provided by shcedule.
			// e.g. for cron 0/5 * * * * (every 5th minute) if last time job was last executed at 12:05:00
			// the next execution will be planned at 12:10:00
			var next *time.Time
			if from, ok := i.lastExecutions[scheduleID]; !ok {
				next = schedule.Schedule.Next(from)
			} else {
				next = utils.ToPointer(schedule.Schedule.First())
			}

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

	return
}

func (i *inMemoryDriver) ScheduleJob(ctx context.Context, job Job, schedule JobSchedule) (string, error) {
	i.lock.Lock()
	defer i.lock.Unlock()

	id := uuid.NewString()
	i.schedule[scheduleID(id)] = JobScheduledExecution{
		Job:        job,
		Schedule:   schedule,
		ScheduleID: id,
	}

	return id, nil
}

func (i *inMemoryDriver) UnscheduleJobByJobID(ctx context.Context, jobID string) error {
	i.lock.Lock()
	defer i.lock.Unlock()

	scheduleIDs := make([]scheduleID, 0)
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

func (i *inMemoryDriver) UnscheduleJobByScheduleID(ctx context.Context, id string) error {
	i.lock.Lock()
	defer i.lock.Unlock()

	delete(i.schedule, scheduleID(id))
	return nil
}

func (i *inMemoryDriver) onJobExecuted(ctx *JobExecutionContext) {
	i.lock.Lock()
	defer i.lock.Unlock()

	delete(i.currentExecutions, executionID(ctx.Execution.ExecutionID))
	i.lastExecutions[scheduleID(ctx.Execution.ScheduleID)] = ctx.Execution.PlannedAt
}

func newInMemoryDriver() *inMemoryDriver {
	return &inMemoryDriver{
		schedule:          make(map[scheduleID]JobScheduledExecution),
		lastExecutions:    make(map[scheduleID]time.Time),
		currentExecutions: make(map[executionID]scheduleID),
	}
}
