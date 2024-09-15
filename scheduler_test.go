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

func TestRegisterJobShouldDoItSuccessfully(t *testing.T) {
	config, _, _ := newTestConfig()
	scheduler := NewScheduler(config)

	id := uuid.NewString()

	err := scheduler.RegisterJob(newTestJob(id))

	require.NoError(t, err)
}

func TestRegisterJobShouldFailIfIDIsNotUnique(t *testing.T) {
	config, _, _ := newTestConfig()
	scheduler := NewScheduler(config)

	id := uuid.NewString()

	err := scheduler.RegisterJob(newTestJob(id))

	require.NoError(t, err)

	err = scheduler.RegisterJob(newTestJob(id))

	assert.Equal(t, err, ErrJobIDExists)
}

func TestUnscheduleJobByJobIDShouldDoIt(t *testing.T) {
	config, driver, _ := newTestConfig()
	scheduler := NewScheduler(config)

	id := uuid.NewString()

	driver.On("UnscheduleJobByJobID", mock.Anything, id).Return(nil)

	err := scheduler.UnscheduleJobByJobID(context.Background(), id)
	assert.NoError(t, err)
}

func TestUnscheduleJobByJobIDShouldReturnErrorIfDriverFails(t *testing.T) {
	config, driver, _ := newTestConfig()
	scheduler := NewScheduler(config)

	id := uuid.NewString()

	driver.On("UnscheduleJobByJobID", mock.Anything, id).Return(fmt.Errorf("error"))

	err := scheduler.UnscheduleJobByJobID(context.Background(), id)

	require.Error(t, err)
}

func TestUnscheduleJobByScheduleIDShouldDoIt(t *testing.T) {
	config, driver, _ := newTestConfig()
	scheduler := NewScheduler(config)

	id := uuid.NewString()

	driver.On("UnscheduleJobByScheduleID", mock.Anything, id).Return(nil)

	err := scheduler.UnscheduleJobByScheduleID(context.Background(), id)
	assert.NoError(t, err)
}

func TestUnscheduleJobByScheduleIDShouldReturnErrorIfDriverFails(t *testing.T) {
	config, driver, _ := newTestConfig()
	scheduler := NewScheduler(config)

	id := uuid.NewString()

	driver.On("UnscheduleJobByScheduleID", mock.Anything, id).Return(fmt.Errorf("error"))

	err := scheduler.UnscheduleJobByScheduleID(context.Background(), id)

	require.Error(t, err)
}

func TestScheduleJobShouldReturnErrorIfJobIsNotRegistered(t *testing.T) {
	config, _, _ := newTestConfig()
	scheduler := NewScheduler(config)

	id := uuid.NewString()
	schedule := NewCalendarSchedule(time.Now())

	_, err := scheduler.ScheduleJob(context.Background(), id, schedule)
	assert.Equal(t, ErrJobNotFound, err)
}

func TestScheduleJobShouldSucceed(t *testing.T) {
	config, driver, _ := newTestConfig()
	scheduler := NewScheduler(config)

	jobID := uuid.NewString()
	scheduleID := uuid.NewString()

	job := newTestJob(jobID)
	schedule := NewCalendarSchedule(time.Now())

	driver.On("ScheduleJob", mock.Anything, job, schedule).Return(scheduleID, nil)

	err := scheduler.RegisterJob(job)
	require.NoError(t, err)

	id, err := scheduler.ScheduleJob(context.Background(), jobID, schedule)

	require.NoError(t, err)
	assert.Equal(t, scheduleID, id)
}

func TestScheduleJobShouldReturnErrorIfDriverFails(t *testing.T) {
	config, driver, _ := newTestConfig()
	scheduler := NewScheduler(config)

	id := uuid.NewString()
	job := newTestJob(id)
	schedule := NewCalendarSchedule(time.Now())

	driver.On("ScheduleJob", mock.Anything, job, schedule).Return("", fmt.Errorf("error"))

	err := scheduler.RegisterJob(job)
	require.NoError(t, err)

	_, err = scheduler.ScheduleJob(context.Background(), id, schedule)
	require.Error(t, err)
}

func TestStartShouldReturnPlannerError(t *testing.T) {
	config, driver, planner := newTestConfig()
	scheduler := NewScheduler(config)

	expected := fmt.Errorf("test")
	planner.On("Start", mock.Anything).Return(expected)
	planner.On("Subscribe", scheduler).Return()

	driver.On("NextExecution", mock.Anything).Return(nil, nil)

	ctx, cancel := newTestContext()
	defer cancel()

	err := scheduler.Start(ctx)
	require.Error(t, err)
}

func TestStartShouldExecuteJobIfThereIsSome(t *testing.T) {
	config, driver, planner := newTestConfig()
	scheduler := NewScheduler(config)
	subscriber := &schedulerSubscriberMock{}

	scheduler.Subscribe(subscriber)

	id := uuid.NewString()
	job := newTestJob(id)

	plannedTime := time.Now()

	sch := NewCalendarScheduleWithMaxDelay(time.Now(), 1*time.Minute)
	jobExecution := &JobPlannedExecution{
		JobScheduledExecution: JobScheduledExecution{
			Job:        job,
			Schedule:   sch,
			ScheduleID: uuid.NewString(),
		},
		PlannedAt:   plannedTime,
		ExecutionID: uuid.NewString(),
	}

	assertJobContext := func(jobCtx *JobExecutionContext, expectedStatus JobExecutionStatus) {
		assert.Equal(t, job, jobCtx.Execution.Job)
		assert.GreaterOrEqual(t, jobCtx.Execution.PlannedAt, plannedTime)
		assert.Equal(t, jobCtx.StartedAt, time.Time{})
		assert.Equal(t, jobCtx.ExecutedAt, time.Time{})
		assert.Equal(t, expectedStatus, jobCtx.ExecutionStatus)
		assert.Equal(t, sch, jobCtx.Execution.Schedule)
		assert.Equal(t, jobExecution.ExecutionID, jobCtx.Execution.ExecutionID)
		assert.Equal(t, jobExecution.ScheduleID, jobCtx.Execution.ScheduleID)
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

	driver.On("NextExecution", mock.Anything).Return(jobExecution).Times(1)
	driver.On("NextExecution", mock.Anything).Return(nil)

	planner.On("Start", mock.Anything).Return(nil)
	planner.On("Subscribe", scheduler).Return()

	planner.On("Plan", mock.Anything).Return().Run(func(args mock.Arguments) {
		jobCtx := args.Get(0).(*JobExecutionContext)
		assertJobContext(jobCtx, JobExecutionStatusInitiated)

		planner.subscribers[0].OnJobExecutionUnplanned(jobCtx)
		planner.subscribers[0].OnBeforeJobExecution(jobCtx)
		job.Execute(jobCtx)
		planner.subscribers[0].OnJobExecuted(jobCtx)
	})

	ctx, cancel := newTestContext()
	defer cancel()

	err := scheduler.Start(ctx)
	require.NoError(t, err)

	select {
	case <-ctx.Done():
		require.Fail(t, "expected job to be executed within 2 seconds")
	case <-job.done:
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
	scheduler := NewScheduler(config)
	subscriber := &schedulerSubscriberMock{}

	scheduler.Subscribe(subscriber)

	id := uuid.NewString()
	job := newTestJob(id)

	plannedTime := time.Now().Add(-1 * time.Second)
	sch := NewCalendarScheduleWithMaxDelay(plannedTime, 0)

	jobExecution := &JobPlannedExecution{
		JobScheduledExecution: JobScheduledExecution{
			Job:      job,
			Schedule: sch,
		},
		PlannedAt: plannedTime,
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

	driver.On("NextExecution", mock.Anything).Return(jobExecution).Times(1)
	driver.On("NextExecution", mock.Anything).Return(nil)

	planner.On("Start", mock.Anything).Return(nil)
	planner.On("Subscribe", scheduler).Return()

	ctx, cancel := newTestContext()
	defer cancel()

	err := scheduler.Start(ctx)
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
	scheduler := NewScheduler(config)
	subscriber := &schedulerSubscriberMock{}

	scheduler.Subscribe(subscriber)

	id := uuid.NewString()
	job := newTestJob(id)

	plannedTime := time.Now().Add(-1 * time.Second)
	sch := NewCalendarScheduleWithMaxDelay(plannedTime, 0)

	jobExecution := &JobPlannedExecution{
		JobScheduledExecution: JobScheduledExecution{
			Job:      job,
			Schedule: sch,
		},
		PlannedAt: plannedTime,
	}

	subscriber.On("OnStart").Return()
	subscriber.On("OnJobExecutionInitiated", mock.Anything).Return()
	subscriber.On("OnJobExecutionDelayed", mock.Anything).Return()
	subscriber.On("OnBeforeJobExecutionPlan", mock.Anything).Return()

	driver.On("NextExecution", mock.Anything).Return(jobExecution).Times(1)
	driver.On("NextExecution", mock.Anything).Return(nil)

	planner.On("Start", mock.Anything).Return(nil)
	planner.On("Subscribe", scheduler).Return()
	planner.On("Plan", mock.Anything).Return().Run(func(args mock.Arguments) {
		job.Execute(args.Get(0).(*JobExecutionContext))
	})

	ctx, cancel := newTestContext()
	defer cancel()

	err := scheduler.Start(ctx)
	require.NoError(t, err)

	select {
	case <-ctx.Done():
		require.Fail(t, "expected job to be executed within 2 seconds")
	case <-job.done:
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
	config, _, planner := newTestConfig()
	scheduler := NewScheduler(config)
	subscriber := &schedulerSubscriberMock{}

	scheduler.Subscribe(subscriber)

	subscriber.On("OnStop").Return().Once()
	planner.On("Stop").Return(nil).Once()

	err := scheduler.Stop()
	require.NoError(t, err)

	err = scheduler.Stop()
	require.NoError(t, err)
}
