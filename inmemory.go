package gotick

import (
	"context"
	"sync"
	"time"

	"github.com/go-tick/core/internal/utils"
	"github.com/google/uuid"
)

type scheduleID string

type inMemoryDriver struct {
	cfg      *InMemoryDriverConfig
	schedule map[scheduleID]*entry
	lock     sync.Mutex
}

type entry struct {
	jobID          string
	schedule       JobSchedule
	lastExecutedAt time.Time
	lock           utils.Lock
}

func (i *inMemoryDriver) OnJobExecutionDelayed(*JobExecutionContext) {
}

func (i *inMemoryDriver) OnJobExecutionInitiated(ctx *JobExecutionContext) {
}

func (i *inMemoryDriver) OnJobExecutionSkipped(ctx *JobExecutionContext) {
	i.onJobExecuted(ctx, true)
}

func (i *inMemoryDriver) OnBeforeJobExecution(*JobExecutionContext) {
}

func (i *inMemoryDriver) OnJobExecutionUnplanned(ctx *JobExecutionContext) {
	i.onJobExecuted(ctx, false)
}

func (i *inMemoryDriver) OnBeforeJobExecutionPlan(*JobExecutionContext) {
}

func (i *inMemoryDriver) OnError(error) {
}

func (i *inMemoryDriver) OnJobExecuted(ctx *JobExecutionContext) {
	i.onJobExecuted(ctx, true)
}

func (i *inMemoryDriver) OnStart() {
}

func (i *inMemoryDriver) OnStop() {
}

func (i *inMemoryDriver) Start(context.Context) error {
	// do nothing
	return nil
}

func (i *inMemoryDriver) Stop() error {
	// do nothing
	return nil
}

func (i *inMemoryDriver) NextExecution(ctx context.Context) (result *NextExecutionResult) {
	i.lock.Lock()
	defer i.lock.Unlock()

	toUnschedule := make([]scheduleID, 0)
	for sid, schedule := range i.schedule {
		// if schedule is locked, it means the job is taken by some other thread
		// so we should skip it
		if locked := schedule.lock.TryLock(); !locked {
			continue
		}

		// find job last execution time and calculate the next one based on it
		// if job was never executed, use the first execution time provided by shcedule.
		// e.g. for cron 0/5 * * * * (every 5th minute) if last time job was last executed at 12:05:00
		// the next execution will be planned at 12:10:00
		var next *time.Time
		if schedule.lastExecutedAt.IsZero() {
			next = utils.ToPointer(schedule.schedule.First())
		} else {
			next = schedule.schedule.Next(schedule.lastExecutedAt)
		}

		// if next is nil, it means that the job is not scheduled anymore
		// so we should remove it from the schedule
		if next == nil {
			toUnschedule = append(toUnschedule, sid)
			continue
		}

		// if execution is nil or next is before the planned execution time
		// we should execute current job next
		if result == nil || next.Before(result.PlannedAt) {
			if result != nil {
				i.schedule[scheduleID(result.ScheduleID)].lock.Unlock()
			}

			result = &NextExecutionResult{
				JobID: schedule.jobID,

				Schedule:   schedule.schedule,
				ScheduleID: string(sid),

				PlannedAt: *next,
			}
		} else {
			schedule.lock.Unlock()
		}
	}

	for _, scheduleID := range toUnschedule {
		i.unscheduleJobByScheduleIDWithoutLock(scheduleID)
	}

	return
}

func (i *inMemoryDriver) ScheduleJob(ctx context.Context, jobID string, schedule JobSchedule) (string, error) {
	i.lock.Lock()
	defer i.lock.Unlock()

	timeout := i.cfg.lockTimeout
	if st, ok := schedule.(Timeout); ok {
		timeout = st.Timeout()
	}

	id := uuid.NewString()
	i.schedule[scheduleID(id)] = &entry{
		jobID:    jobID,
		schedule: schedule,
		lock:     utils.NewLockWithTimeout(timeout),
	}

	return id, nil
}

func (i *inMemoryDriver) UnscheduleJobByJobID(ctx context.Context, jobID string) error {
	i.lock.Lock()
	defer i.lock.Unlock()

	scheduleIDs := make([]scheduleID, 0)
	for scheduleID, schedule := range i.schedule {
		if schedule.jobID == jobID {
			scheduleIDs = append(scheduleIDs, scheduleID)
		}
	}

	for _, scheduleID := range scheduleIDs {
		i.unscheduleJobByScheduleIDWithoutLock(scheduleID)
	}

	return nil
}

func (i *inMemoryDriver) UnscheduleJobByScheduleID(ctx context.Context, id string) error {
	i.lock.Lock()
	defer i.lock.Unlock()

	i.unscheduleJobByScheduleIDWithoutLock(scheduleID(id))
	return nil
}

func (i *inMemoryDriver) onJobExecuted(ctx *JobExecutionContext, success bool) {
	i.lock.Lock()
	defer i.lock.Unlock()

	if schedule, ok := i.schedule[scheduleID(ctx.ScheduleID)]; ok {
		defer schedule.lock.Unlock()
		if success {
			schedule.lastExecutedAt = ctx.PlannedAt
		}
	}
}

func (i *inMemoryDriver) unscheduleJobByScheduleIDWithoutLock(scheduleID scheduleID) {
	if schedule, ok := i.schedule[scheduleID]; ok {
		schedule.lock.Unlock()
	}
	delete(i.schedule, scheduleID)
}

func newInMemoryDriver(cfg *InMemoryDriverConfig) (*inMemoryDriver, error) {
	return &inMemoryDriver{
		cfg:      cfg,
		schedule: make(map[scheduleID]*entry),
	}, nil
}
