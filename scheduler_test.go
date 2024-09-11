package gotick_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/misikdmytro/gotick"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func schedulerTestContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 2*time.Second)
}

func TestRegisterJobShouldDoItSuccessfully(t *testing.T) {
	config, _, _ := NewTestConfig()
	scheduler := gotick.NewScheduler(config)

	id := uuid.NewString()

	err := scheduler.RegisterJob(NewTestJob(id))

	require.NoError(t, err)
}

func TestRegisterJobShouldFailIfIDIsNotUnique(t *testing.T) {
	config, _, _ := NewTestConfig()
	scheduler := gotick.NewScheduler(config)

	id := uuid.NewString()

	err := scheduler.RegisterJob(NewTestJob(id))

	require.NoError(t, err)

	err = scheduler.RegisterJob(NewTestJob(id))

	assert.Equal(t, err, gotick.ErrJobIDExists)
}

func TestUnscheduleJobByJobIDShouldDoIt(t *testing.T) {
	config, driver, _ := NewTestConfig()
	scheduler := gotick.NewScheduler(config)

	id := uuid.NewString()

	driver.On("UnscheduleJobByJobID", mock.Anything, id).Return(nil)

	err := scheduler.UnscheduleJobByJobID(context.Background(), id)
	assert.NoError(t, err)
}

func TestUnscheduleJobByJobIDShouldReturnErrorIfDriverFails(t *testing.T) {
	config, driver, _ := NewTestConfig()
	scheduler := gotick.NewScheduler(config)

	id := uuid.NewString()

	driver.On("UnscheduleJobByJobID", mock.Anything, id).Return(fmt.Errorf("error"))

	err := scheduler.UnscheduleJobByJobID(context.Background(), id)

	require.Error(t, err)
}

func TestUnscheduleJobByScheduleIDShouldDoIt(t *testing.T) {
	config, driver, _ := NewTestConfig()
	scheduler := gotick.NewScheduler(config)

	id := uuid.NewString()

	driver.On("UnscheduleJobByScheduleID", mock.Anything, id).Return(nil)

	err := scheduler.UnscheduleJobByScheduleID(context.Background(), id)
	assert.NoError(t, err)
}

func TestUnscheduleJobByScheduleIDShouldReturnErrorIfDriverFails(t *testing.T) {
	config, driver, _ := NewTestConfig()
	scheduler := gotick.NewScheduler(config)

	id := uuid.NewString()

	driver.On("UnscheduleJobByScheduleID", mock.Anything, id).Return(fmt.Errorf("error"))

	err := scheduler.UnscheduleJobByScheduleID(context.Background(), id)

	require.Error(t, err)
}

func TestScheduleJobShouldReturnErrorIfJobIsNotRegistered(t *testing.T) {
	config, _, _ := NewTestConfig()
	scheduler := gotick.NewScheduler(config)

	id := uuid.NewString()

	schedule, err := gotick.NewOnce(time.Now().Add(1 * time.Minute))
	require.NoError(t, err)

	_, err = scheduler.ScheduleJob(context.Background(), id, schedule)
	assert.Equal(t, gotick.ErrJobNotFound, err)
}

func TestScheduleJobShouldSucceed(t *testing.T) {
	config, driver, _ := NewTestConfig()
	scheduler := gotick.NewScheduler(config)

	jobID := uuid.NewString()
	scheduleID := uuid.NewString()

	job := NewTestJob(jobID)

	schedule, err := gotick.NewOnce(time.Now().Add(1 * time.Minute))
	require.NoError(t, err)

	driver.On("ScheduleJob", mock.Anything, job, schedule).Return(scheduleID, nil)

	err = scheduler.RegisterJob(job)
	require.NoError(t, err)

	id, err := scheduler.ScheduleJob(context.Background(), jobID, schedule)

	require.NoError(t, err)
	assert.Equal(t, scheduleID, id)
}

func TestScheduleJobShouldReturnErrorIfDriverFails(t *testing.T) {
	config, driver, _ := NewTestConfig()
	scheduler := gotick.NewScheduler(config)

	id := uuid.NewString()
	job := NewTestJob(id)

	schedule, err := gotick.NewOnce(time.Now().Add(1 * time.Minute))
	require.NoError(t, err)

	driver.On("ScheduleJob", mock.Anything, job, schedule).Return("", fmt.Errorf("error"))

	err = scheduler.RegisterJob(job)
	require.NoError(t, err)

	_, err = scheduler.ScheduleJob(context.Background(), id, schedule)
	require.Error(t, err)
}

func TestStartShouldReturnSubscriberError(t *testing.T) {
	config, driver, planner := NewTestConfig()
	scheduler := gotick.NewScheduler(config)
	subscriber := &schedulerSubscriberMock{}

	expected := fmt.Errorf("test")
	subscriber.On("OnStart").Return(expected)

	scheduler.Subscribe(subscriber)

	ctx, cancel := schedulerTestContext()
	defer cancel()

	err := scheduler.Start(ctx)

	require.Error(t, err)
	driver.AssertNotCalled(t, "NextExecution", mock.Anything, mock.Anything)
	planner.AssertNotCalled(t, "Start", mock.Anything)
}

func TestStartShouldReturnPlannerError(t *testing.T) {
	config, driver, planner := NewTestConfig()
	scheduler := gotick.NewScheduler(config)

	expected := fmt.Errorf("test")
	planner.On("Start", mock.Anything).Return(expected)

	ctx, cancel := schedulerTestContext()
	defer cancel()

	err := scheduler.Start(ctx)

	require.Error(t, err)
	driver.AssertNotCalled(t, "NextExecution", mock.Anything, mock.Anything)
}

func TestStartShouldExecuteJobIfThereIsSome(t *testing.T) {
	config, driver, planner := NewTestConfig()
	scheduler := gotick.NewScheduler(config)
	subscriber := &schedulerSubscriberMock{}

	scheduler.Subscribe(subscriber)

	id := uuid.NewString()
	job := NewTestJob(id)

	plannedTime := time.Now()

	jobExecution := &gotick.JobPlannedExecution{
		Job:       job,
		PlannedAt: plannedTime,
	}

	subscriber.On("OnStart").Return(nil)
	subscriber.On("OnBeforeJobExecution", mock.Anything).Return(nil)
	subscriber.On("OnJobExecuted", mock.Anything).Return(nil)

	driver.On("NextExecution", mock.Anything, mock.Anything).Return(jobExecution, nil).Times(1)
	driver.On("NextExecution", mock.Anything, mock.Anything).Return(nil, nil)

	planner.On("Start", mock.Anything).Return(nil)
	planner.On("Subscribe", mock.Anything).Return(nil)

	planner.On("Plan", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		ctx := args.Get(0).(*gotick.JobContext)
		err := job.Execute(ctx)
		require.NoError(t, err)

		planner.subscribers[0].OnJobExecuted(ctx)
	})

	ctx, cancel := schedulerTestContext()
	defer cancel()

	err := scheduler.Start(ctx)
	require.NoError(t, err)

	select {
	case <-ctx.Done():
		require.Fail(t, "expected job to be executed within 2 seconds")
	case <-job.done:
		assert.Equal(t, 1, len(job.executedAt))
		assert.LessOrEqual(t, job.executedAt[0], time.Now())
	}

	subscriber.AssertCalled(t, "OnStart")
	subscriber.AssertCalled(t, "OnBeforeJobExecution", mock.Anything)
	subscriber.AssertCalled(t, "OnJobExecuted", mock.Anything)

	planner.AssertCalled(t, "Start", mock.Anything)
	planner.AssertCalled(t, "Subscribe", scheduler)
}

func TestStartShouldSkipExecutionIfThisAheadOfPlaningTime(t *testing.T) {
	config, driver, planner := NewTestConfig(gotick.WithIdlePollingInterval(0))
	scheduler := gotick.NewScheduler(config)

	id := uuid.NewString()
	job := NewTestJob(id)

	plannedTime := time.Now().Add(config.MaxPlanAhead * 2)

	jobExecution := &gotick.JobPlannedExecution{
		Job:       job,
		PlannedAt: plannedTime,
	}

	driver.On("NextExecution", mock.Anything, mock.Anything).Return(jobExecution, nil).Times(1)

	closed := false
	proceed := make(chan struct{})
	driver.On("NextExecution", mock.Anything, mock.Anything).Return(nil, nil).Run(func(mock.Arguments) {
		if !closed {
			closed = true
			close(proceed)
		}
	})

	planner.On("Start", mock.Anything).Return(nil)
	planner.On("Subscribe", mock.Anything).Return(nil)

	ctx, cancel := schedulerTestContext()
	defer cancel()

	err := scheduler.Start(ctx)
	require.NoError(t, err)

	select {
	case <-ctx.Done():
	case <-proceed:
		driver.AssertCalled(t, "NextExecution", mock.Anything, mock.Anything)
		planner.AssertNotCalled(t, "Plan", mock.Anything, mock.Anything, mock.Anything)
	case <-job.done:
		require.Fail(t, "execution is not expected")
	}
}

func TestStartShouldPublishNextExecutionErrorToSubscribers(t *testing.T) {
	config, driver, planner := NewTestConfig()
	scheduler := gotick.NewScheduler(config)
	subscriber := &schedulerSubscriberMock{}

	scheduler.Subscribe(subscriber)

	subscriber.On("OnStart").Return(nil)

	expected := fmt.Errorf("test")
	proceed := make(chan struct{})
	subscriber.On("OnError", expected).Return().Run(func(mock.Arguments) {
		close(proceed)
	})

	driver.On("NextExecution", mock.Anything, mock.Anything).Return(nil, expected)

	planner.On("Start", mock.Anything).Return(nil)
	planner.On("Subscribe", mock.Anything).Return(nil)

	ctx, cancel := schedulerTestContext()
	defer cancel()

	err := scheduler.Start(ctx)
	require.NoError(t, err)

	select {
	case <-ctx.Done():
		require.Fail(t, "expected error to be published within 2 seconds")
	case <-proceed:
		subscriber.AssertCalled(t, "OnError", expected)
		planner.AssertNotCalled(t, "Plan", mock.Anything)
	}
}

func TestStartShouldStopIfNextExecutionFatalError(t *testing.T) {
	config, driver, planner := NewTestConfig()
	scheduler := gotick.NewScheduler(config)
	subscriber := &schedulerSubscriberMock{}

	scheduler.Subscribe(subscriber)

	subscriber.On("OnStart").Return(nil)
	subscriber.On("OnStop").Return(nil)

	expected := gotick.NewFatalError(fmt.Errorf("test"))
	proceed := make(chan struct{})
	subscriber.On("OnError", expected).Return().Run(func(mock.Arguments) {
		close(proceed)
	})

	driver.On("NextExecution", mock.Anything, mock.Anything).Return(nil, expected)

	planner.On("Start", mock.Anything).Return(nil)
	planner.On("Subscribe", mock.Anything).Return(nil)
	planner.On("Stop").Return(nil)

	ctx, cancel := schedulerTestContext()
	defer cancel()

	err := scheduler.Start(ctx)
	require.NoError(t, err)

	select {
	case <-ctx.Done():
		require.Fail(t, "expected error to be published within 2 seconds")
	case <-proceed:
		subscriber.AssertCalled(t, "OnError", expected)
		subscriber.AssertCalled(t, "OnStop")

		planner.AssertNotCalled(t, "Plan", mock.Anything)
		planner.AssertCalled(t, "Stop")
	}
}

func TestStartShouldPublishOnBeforeJobExecutionError(t *testing.T) {
	config, driver, planner := NewTestConfig()
	scheduler := gotick.NewScheduler(config)
	subscriber := &schedulerSubscriberMock{}

	scheduler.Subscribe(subscriber)

	id := uuid.NewString()
	job := NewTestJob(id)

	plannedTime := time.Now()

	jobExecution := &gotick.JobPlannedExecution{
		Job:       job,
		PlannedAt: plannedTime,
	}

	expected := fmt.Errorf("test")
	subscriber.On("OnBeforeJobExecution", mock.Anything).Return(expected)
	subscriber.On("OnStart").Return(nil)

	proceed := make(chan struct{})
	subscriber.On("OnError", expected).Return().Run(func(mock.Arguments) {
		close(proceed)
	})

	driver.On("NextExecution", mock.Anything, mock.Anything).Return(jobExecution, nil).Times(1)
	driver.On("NextExecution", mock.Anything, mock.Anything).Return(nil, nil)

	planner.On("Start", mock.Anything).Return(nil)
	planner.On("Subscribe", mock.Anything).Return(nil)
	planner.On("Plan", mock.Anything).Return(nil)

	ctx, cancel := schedulerTestContext()
	defer cancel()

	err := scheduler.Start(ctx)
	require.NoError(t, err)

	select {
	case <-ctx.Done():
		require.Fail(t, "expected job to be executed within 2 seconds")
	case <-proceed:
		subscriber.AssertCalled(t, "OnBeforeJobExecution", mock.Anything)
		subscriber.AssertCalled(t, "OnError", expected)
		planner.AssertCalled(t, "Plan", mock.Anything)
	}
}

func TestStartShouldStopOnFatalErrorOnBeforeExecution(t *testing.T) {
	config, driver, planner := NewTestConfig()
	scheduler := gotick.NewScheduler(config)
	subscriber := &schedulerSubscriberMock{}

	scheduler.Subscribe(subscriber)

	id := uuid.NewString()
	job := NewTestJob(id)

	plannedTime := time.Now()

	jobExecution := &gotick.JobPlannedExecution{
		Job:       job,
		PlannedAt: plannedTime,
	}

	expected := gotick.NewFatalError(fmt.Errorf("test"))
	subscriber.On("OnBeforeJobExecution", mock.Anything).Return(expected)
	subscriber.On("OnStart").Return(nil)
	subscriber.On("OnStop").Return(nil)

	proceed := make(chan struct{})
	subscriber.On("OnError", expected).Return().Run(func(mock.Arguments) {
		close(proceed)
	})

	driver.On("NextExecution", mock.Anything, mock.Anything).Return(jobExecution, nil).Times(1)

	planner.On("Start", mock.Anything).Return(nil)
	planner.On("Subscribe", mock.Anything).Return(nil)
	planner.On("Stop").Return(nil)

	ctx, cancel := schedulerTestContext()
	defer cancel()

	err := scheduler.Start(ctx)
	require.NoError(t, err)

	select {
	case <-ctx.Done():
		require.Fail(t, "expected job to be executed within 2 seconds")
	case <-proceed:
		subscriber.AssertCalled(t, "OnBeforeJobExecution", mock.Anything)
		subscriber.AssertCalled(t, "OnError", expected)
		subscriber.AssertCalled(t, "OnStop")

		planner.AssertNotCalled(t, "Plan", mock.Anything)
		planner.AssertCalled(t, "Stop")
	}
}

func TestStartShouldSkipJobExecutionIfJobLocked(t *testing.T) {
	config, driver, planner := NewTestConfig()
	scheduler := gotick.NewScheduler(config)
	subscriber := &schedulerSubscriberMock{}

	scheduler.Subscribe(subscriber)

	id := uuid.NewString()
	job := NewTestJob(id)

	plannedTime := time.Now()

	jobExecution := &gotick.JobPlannedExecution{
		Job:       job,
		PlannedAt: plannedTime,
	}

	expected := fmt.Errorf("test")
	subscriber.On("OnBeforeJobExecution", mock.Anything).Return(gotick.ErrJobLocked)
	subscriber.On("OnStart").Return(nil)

	driver.On("NextExecution", mock.Anything, mock.Anything).Return(jobExecution, nil).Times(1)

	proceed := make(chan struct{})
	driver.On("NextExecution", mock.Anything, mock.Anything).Return(nil, nil).Run(func(mock.Arguments) {
		close(proceed)
	})

	planner.On("Start", mock.Anything).Return(nil)
	planner.On("Subscribe", mock.Anything).Return(nil)

	ctx, cancel := schedulerTestContext()
	defer cancel()

	err := scheduler.Start(ctx)
	require.NoError(t, err)

	select {
	case <-ctx.Done():
		require.Fail(t, "expected job to be executed within 2 seconds")
	case <-proceed:
		subscriber.AssertNotCalled(t, "OnError", expected)
		planner.AssertNotCalled(t, "Plan", mock.Anything)
	}
}
