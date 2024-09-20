package gotick

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUnscheduleJobByJobIDShouldDoIt(t *testing.T) {
	config, driver, _ := newTestConfig()
	scheduler, err := NewScheduler(config)
	require.NoError(t, err)

	id := uuid.NewString()

	driver.On("UnscheduleJobByJobID", mock.Anything, id).Return(nil)

	err = scheduler.UnscheduleJobByJobID(context.Background(), id)
	assert.NoError(t, err)
}

func TestUnscheduleJobByJobIDShouldReturnErrorIfDriverFails(t *testing.T) {
	config, driver, _ := newTestConfig()
	scheduler, err := NewScheduler(config)
	require.NoError(t, err)

	id := uuid.NewString()

	driver.On("UnscheduleJobByJobID", mock.Anything, id).Return(fmt.Errorf("error"))

	err = scheduler.UnscheduleJobByJobID(context.Background(), id)

	require.Error(t, err)
}

func TestUnscheduleJobByScheduleIDShouldDoIt(t *testing.T) {
	config, driver, _ := newTestConfig()
	scheduler, err := NewScheduler(config)
	require.NoError(t, err)

	id := uuid.NewString()

	driver.On("UnscheduleJobByScheduleID", mock.Anything, id).Return(nil)

	err = scheduler.UnscheduleJobByScheduleID(context.Background(), id)
	assert.NoError(t, err)
}

func TestUnscheduleJobByScheduleIDShouldReturnErrorIfDriverFails(t *testing.T) {
	config, driver, _ := newTestConfig()
	scheduler, err := NewScheduler(config)
	require.NoError(t, err)

	id := uuid.NewString()

	driver.On("UnscheduleJobByScheduleID", mock.Anything, id).Return(fmt.Errorf("error"))

	err = scheduler.UnscheduleJobByScheduleID(context.Background(), id)

	require.Error(t, err)
}

func TestScheduleJobShouldSucceed(t *testing.T) {
	config, driver, _ := newTestConfig()
	scheduler, err := NewScheduler(config)
	require.NoError(t, err)

	jobID := uuid.NewString()
	scheduleID := uuid.NewString()

	schedule := NewCalendarSchedule(time.Now())

	driver.On("ScheduleJob", mock.Anything, jobID, schedule).Return(scheduleID, nil)

	id, err := scheduler.ScheduleJob(context.Background(), jobID, schedule)

	require.NoError(t, err)
	assert.Equal(t, scheduleID, id)
}

func TestScheduleJobShouldReturnErrorIfDriverFails(t *testing.T) {
	config, driver, _ := newTestConfig()
	scheduler, err := NewScheduler(config)
	require.NoError(t, err)

	jobID := uuid.NewString()
	schedule := NewCalendarSchedule(time.Now())

	driver.On("ScheduleJob", mock.Anything, jobID, schedule).Return("", fmt.Errorf("error"))

	_, err = scheduler.ScheduleJob(context.Background(), jobID, schedule)
	require.Error(t, err)
}

func TestStartShouldReturnPlannerError(t *testing.T) {
	config, driver, planner := newTestConfig()
	scheduler, err := NewScheduler(config)
	require.NoError(t, err)

	expected := fmt.Errorf("test")
	planner.On("Start", mock.Anything).Return(expected)
	planner.On("Subscribe", scheduler).Return()

	driver.On("NextExecution", mock.Anything).Return(nil, nil)

	ctx, cancel := newTestContext()
	defer cancel()

	err = scheduler.Start(ctx)
	require.Error(t, err)
}

func TestStartShouldExecuteJobIfThereIsSome(t *testing.T) {
	config, driver, planner := newTestConfig()
	scheduler, err := NewScheduler(config)
	require.NoError(t, err)

	subscriber := &schedulerSubscriberMock{}

	scheduler.Subscribe(subscriber)

	jobID := uuid.NewString()
	plannedTime := time.Now()

	sch := NewJobScheduleWithMaxDelay(NewCalendarSchedule(time.Now()), 5*time.Second)
	jobExecution := &NextExecutionResult{
		JobID:      jobID,
		Schedule:   sch,
		ScheduleID: uuid.NewString(),
		PlannedAt:  plannedTime,
	}

	assertJobContext := func(ctx *JobExecutionContext, expectedStatus JobExecutionStatus) {
		assert.Equal(t, jobID, ctx.JobID)
		assert.GreaterOrEqual(t, ctx.PlannedAt, plannedTime)
		assert.Equal(t, ctx.StartedAt, time.Time{})
		assert.Equal(t, ctx.ExecutedAt, time.Time{})
		assert.Equal(t, expectedStatus, ctx.ExecutionStatus)
		assert.Equal(t, jobExecution.ScheduleID, ctx.ScheduleID)
		assert.NotEmpty(t, ctx.ExecutionID)
	}

	subscriber.On("OnStart").Return()
	subscriber.On("OnJobExecutionUnplanned", mock.Anything).Return().Run(func(args mock.Arguments) {
		jobCtx := args.Get(0).(*JobExecutionContext)
		assertJobContext(jobCtx, JobExecutionStatusInitiated)
	})

	subscriber.On("OnJobExecutionInitiated", mock.Anything).Return().Run(func(args mock.Arguments) {
		jobCtx := args.Get(0).(*JobExecutionContext)
		assertJobContext(jobCtx, JobExecutionStatusInitiated)
	})

	subscriber.On("OnBeforeJobExecutionPlan", mock.Anything).Return().Run(func(args mock.Arguments) {
		jobCtx := args.Get(0).(*JobExecutionContext)
		assertJobContext(jobCtx, JobExecutionStatusInitiated)
	})

	subscriber.On("OnBeforeJobExecution", mock.Anything).Return()
	subscriber.On("OnJobExecuted", mock.Anything).Return()

	driver.On("Start", mock.Anything).Return(nil)
	driver.On("NextExecution", mock.Anything).Return(jobExecution).Times(1)
	driver.On("NextExecution", mock.Anything).Return(nil)

	planner.On("Start", mock.Anything).Return(nil)
	planner.On("Subscribe", scheduler).Return()

	done := make(chan any)
	planner.On("Plan", mock.Anything).Return().Run(func(args mock.Arguments) {
		jobCtx := args.Get(0).(*JobExecutionContext)
		assertJobContext(jobCtx, JobExecutionStatusInitiated)

		planner.subscribers[0].OnJobExecutionUnplanned(jobCtx)
		planner.subscribers[0].OnBeforeJobExecution(jobCtx)
		planner.subscribers[0].OnJobExecuted(jobCtx)

		close(done)
	})

	ctx, cancel := newTestContext()
	defer cancel()

	err = scheduler.Start(ctx)
	require.NoError(t, err)

	select {
	case <-ctx.Done():
		require.Fail(t, "expected job to be executed within 2 seconds")
	case <-done:
	}

	subscriber.AssertCalled(t, "OnStart")
	subscriber.AssertNotCalled(t, "OnStop")
	subscriber.AssertCalled(t, "OnJobExecutionInitiated", mock.Anything)
	subscriber.AssertNotCalled(t, "OnJobExecutionDelayed", mock.Anything)
	subscriber.AssertNotCalled(t, "OnJobExecutionSkipped", mock.Anything)
	subscriber.AssertCalled(t, "OnJobExecutionUnplanned", mock.Anything)
	subscriber.AssertCalled(t, "OnBeforeJobExecutionPlan", mock.Anything)
	subscriber.AssertCalled(t, "OnBeforeJobExecution", mock.Anything)
	subscriber.AssertCalled(t, "OnJobExecuted", mock.Anything)

	planner.AssertCalled(t, "Start", mock.Anything)
	planner.AssertNotCalled(t, "Stop")
	planner.AssertCalled(t, "Subscribe", scheduler)
	planner.AssertCalled(t, "Plan", mock.Anything)
}

func TestStartShouldSkipDelayedJob(t *testing.T) {
	config, driver, planner := newTestConfig(WithDelayedStrategy(ScheduleDelayedStrategySkip))
	scheduler, err := NewScheduler(config)
	require.NoError(t, err)

	subscriber := &schedulerSubscriberMock{}

	scheduler.Subscribe(subscriber)

	jobID := uuid.NewString()

	plannedTime := time.Now().Add(-1 * time.Second)
	sch := NewJobScheduleWithMaxDelay(NewCalendarSchedule(plannedTime), 0)

	jobExecution := &NextExecutionResult{
		JobID:      jobID,
		Schedule:   sch,
		ScheduleID: uuid.NewString(),
		PlannedAt:  plannedTime,
	}

	done := make(chan any)
	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		wg.Wait()
		close(done)
	}()

	subscriber.On("OnStart").Return()
	subscriber.On("OnJobExecutionInitiated", mock.Anything).Return()
	subscriber.On("OnJobExecutionDelayed", mock.Anything).Return().Run(func(args mock.Arguments) {
		jobCtx := args.Get(0).(*JobExecutionContext)
		assert.Equal(t, JobExecutionStatusDelayed, jobCtx.ExecutionStatus)

		wg.Done()
	})

	subscriber.On("OnJobExecutionSkipped", mock.Anything).Return().Run(func(args mock.Arguments) {
		jobCtx := args.Get(0).(*JobExecutionContext)
		assert.Equal(t, JobExecutionStatusSkipped, jobCtx.ExecutionStatus)

		wg.Done()
	})

	driver.On("Start", mock.Anything).Return(nil)
	driver.On("NextExecution", mock.Anything).Return(jobExecution).Times(1)
	driver.On("NextExecution", mock.Anything).Return(nil)

	planner.On("Start", mock.Anything).Return(nil)
	planner.On("Subscribe", scheduler).Return()

	ctx, cancel := newTestContext()
	defer cancel()

	err = scheduler.Start(ctx)
	require.NoError(t, err)

	select {
	case <-ctx.Done():
		require.Fail(t, "expected job to be executed within 2 seconds")
	case <-done:
	}

	subscriber.AssertCalled(t, "OnStart")
	subscriber.AssertNotCalled(t, "OnStop")
	subscriber.AssertCalled(t, "OnJobExecutionInitiated", mock.Anything)
	subscriber.AssertCalled(t, "OnJobExecutionDelayed", mock.Anything)
	subscriber.AssertCalled(t, "OnJobExecutionSkipped", mock.Anything)
	subscriber.AssertNotCalled(t, "OnJobExecutionUnplanned", mock.Anything)
	subscriber.AssertNotCalled(t, "OnBeforeJobExecutionPlan", mock.Anything)
	subscriber.AssertNotCalled(t, "OnBeforeJobExecution", mock.Anything)
	subscriber.AssertNotCalled(t, "OnJobExecuted", mock.Anything)

	planner.AssertCalled(t, "Start", mock.Anything)
	planner.AssertNotCalled(t, "Stop")
	planner.AssertCalled(t, "Subscribe", scheduler)
	planner.AssertNotCalled(t, "Plan")
}

func TestStartShouldProceedDelayed(t *testing.T) {
	config, driver, planner := newTestConfig(WithDelayedStrategy(ScheduleDelayedStrategyPlan))
	scheduler, err := NewScheduler(config)
	require.NoError(t, err)

	subscriber := &schedulerSubscriberMock{}

	scheduler.Subscribe(subscriber)

	jobID := uuid.NewString()

	plannedTime := time.Now().Add(-1 * time.Second)
	sch := NewJobScheduleWithMaxDelay(NewCalendarSchedule(plannedTime), 0)

	jobExecution := &NextExecutionResult{
		JobID:      jobID,
		Schedule:   sch,
		ScheduleID: uuid.NewString(),
		PlannedAt:  plannedTime,
	}

	subscriber.On("OnStart").Return()
	subscriber.On("OnJobExecutionInitiated", mock.Anything).Return()
	subscriber.On("OnJobExecutionDelayed", mock.Anything).Return()
	subscriber.On("OnBeforeJobExecutionPlan", mock.Anything).Return()

	driver.On("Start", mock.Anything).Return(nil)
	driver.On("NextExecution", mock.Anything).Return(jobExecution).Times(1)
	driver.On("NextExecution", mock.Anything).Return(nil)

	planner.On("Start", mock.Anything).Return(nil)
	planner.On("Subscribe", scheduler).Return()

	done := make(chan any)
	planner.On("Plan", mock.Anything).Return().Run(func(args mock.Arguments) {
		close(done)
	})

	ctx, cancel := newTestContext()
	defer cancel()

	err = scheduler.Start(ctx)
	require.NoError(t, err)

	select {
	case <-ctx.Done():
		require.Fail(t, "expected job to be executed within 2 seconds")
	case <-done:
	}

	subscriber.AssertCalled(t, "OnStart")
	subscriber.AssertNotCalled(t, "OnStop")
	subscriber.AssertCalled(t, "OnJobExecutionInitiated", mock.Anything)
	subscriber.AssertCalled(t, "OnJobExecutionDelayed", mock.Anything)
	subscriber.AssertNotCalled(t, "OnJobExecutionSkipped", mock.Anything)
	subscriber.AssertNotCalled(t, "OnJobExecutionUnplanned", mock.Anything)
	subscriber.AssertCalled(t, "OnBeforeJobExecutionPlan", mock.Anything)
	subscriber.AssertNotCalled(t, "OnBeforeJobExecution", mock.Anything)
	subscriber.AssertNotCalled(t, "OnJobExecuted", mock.Anything)

	planner.AssertCalled(t, "Start", mock.Anything)
	planner.AssertNotCalled(t, "Stop")
	planner.AssertCalled(t, "Subscribe", scheduler)
	planner.AssertCalled(t, "Plan", mock.Anything)
}

func TestStopShouldBeCalledOnce(t *testing.T) {
	config, driver, planner := newTestConfig()
	scheduler, err := NewScheduler(config)
	require.NoError(t, err)

	subscriber := &schedulerSubscriberMock{}

	scheduler.Subscribe(subscriber)

	subscriber.On("OnStop").Return().Once()
	planner.On("Stop").Return(nil).Once()
	driver.On("Stop").Return(nil).Once()

	err = scheduler.Stop()
	require.NoError(t, err)

	err = scheduler.Stop()
	require.NoError(t, err)
}

func TestJobScheduleWrappersCanBeCombined(t *testing.T) {
	cron, err := NewCronSchedule("0 0 1 1 *")
	require.NoError(t, err)

	calendar := NewCalendarSchedule(time.Now().Add(1 * time.Minute))

	sequence, err := NewSequenceSchedule(time.Now(), time.Now().Add(1*time.Minute))
	require.NoError(t, err)

	expectedTimeout := 1 * time.Second
	expectedMaxDelay := 2 * time.Second

	data := []struct {
		name     string
		schedule JobSchedule
		wrap     func(JobSchedule) JobSchedule
	}{
		{
			name:     "cron (timeout into max delay)",
			schedule: cron,
			wrap: func(js JobSchedule) JobSchedule {
				return NewJobScheduleWithMaxDelay(NewJobScheduleWithTimeout(js, expectedTimeout), expectedMaxDelay)
			},
		},
		{
			name:     "calendar (timeout into max delay)",
			schedule: calendar,
			wrap: func(js JobSchedule) JobSchedule {
				return NewJobScheduleWithMaxDelay(NewJobScheduleWithTimeout(js, expectedTimeout), expectedMaxDelay)
			},
		},
		{
			name:     "sequence (timeout into max delay)",
			schedule: sequence,
			wrap: func(js JobSchedule) JobSchedule {
				return NewJobScheduleWithMaxDelay(NewJobScheduleWithTimeout(js, expectedTimeout), expectedMaxDelay)
			},
		},
		{
			name:     "cron (max delay into timeout)",
			schedule: cron,
			wrap: func(js JobSchedule) JobSchedule {
				return NewJobScheduleWithTimeout(NewJobScheduleWithMaxDelay(js, expectedMaxDelay), expectedTimeout)
			},
		},
		{
			name:     "calendar (max delay into timeout)",
			schedule: calendar,
			wrap: func(js JobSchedule) JobSchedule {
				return NewJobScheduleWithTimeout(NewJobScheduleWithMaxDelay(js, expectedMaxDelay), expectedTimeout)
			},
		},
		{
			name:     "sequence (max delay into timeout)",
			schedule: sequence,
			wrap: func(js JobSchedule) JobSchedule {
				return NewJobScheduleWithTimeout(NewJobScheduleWithMaxDelay(js, expectedMaxDelay), expectedTimeout)
			},
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			sch := d.wrap(d.schedule)
			require.NotNil(t, sch)

			assert.Equal(t, d.schedule.Schedule(), sch.Schedule())
			assert.Equal(t, d.schedule.First(), sch.First())
			assert.Equal(t, d.schedule.Next(time.Now().Add(1*time.Minute)), sch.Next(time.Now().Add(1*time.Minute)))

			result, ok := TimeoutFromJobSchedule(sch)
			require.True(t, ok)
			assert.Equal(t, expectedTimeout, result)

			result, ok = MaxDelayFromJobSchedule(sch)
			require.True(t, ok)
			assert.Equal(t, expectedMaxDelay, result)
		})
	}
}
