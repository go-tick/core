package gotick_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/misikdmytro/gotick"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRegisterJobShouldDoItSuccessfully(t *testing.T) {
	config, _, _ := newTestConfig()
	scheduler := gotick.NewScheduler(config)

	id := uuid.NewString()

	err := scheduler.RegisterJob(newTestJob(id, nil))

	require.NoError(t, err)
}

func TestRegisterJobShouldFailIfIDIsNotUnique(t *testing.T) {
	config, _, _ := newTestConfig()
	scheduler := gotick.NewScheduler(config)

	id := uuid.NewString()

	err := scheduler.RegisterJob(newTestJob(id, nil))

	require.NoError(t, err)

	err = scheduler.RegisterJob(newTestJob(id, nil))

	assert.Equal(t, err, gotick.ErrJobIDExists)
}

func TestUnscheduleJobByJobIDShouldDoIt(t *testing.T) {
	config, driver, _ := newTestConfig()
	scheduler := gotick.NewScheduler(config)

	id := uuid.NewString()

	driver.On("UnscheduleJobByJobID", mock.Anything, id).Return(nil)

	err := scheduler.UnscheduleJobByJobID(context.Background(), id)
	assert.NoError(t, err)
}

func TestUnscheduleJobByJobIDShouldReturnErrorIfDriverFails(t *testing.T) {
	config, driver, _ := newTestConfig()
	scheduler := gotick.NewScheduler(config)

	id := uuid.NewString()

	driver.On("UnscheduleJobByJobID", mock.Anything, id).Return(fmt.Errorf("error"))

	err := scheduler.UnscheduleJobByJobID(context.Background(), id)

	require.Error(t, err)
}

func TestUnscheduleJobByScheduleIDShouldDoIt(t *testing.T) {
	config, driver, _ := newTestConfig()
	scheduler := gotick.NewScheduler(config)

	id := uuid.NewString()

	driver.On("UnscheduleJobByScheduleID", mock.Anything, id).Return(nil)

	err := scheduler.UnscheduleJobByScheduleID(context.Background(), id)
	assert.NoError(t, err)
}

func TestUnscheduleJobByScheduleIDShouldReturnErrorIfDriverFails(t *testing.T) {
	config, driver, _ := newTestConfig()
	scheduler := gotick.NewScheduler(config)

	id := uuid.NewString()

	driver.On("UnscheduleJobByScheduleID", mock.Anything, id).Return(fmt.Errorf("error"))

	err := scheduler.UnscheduleJobByScheduleID(context.Background(), id)

	require.Error(t, err)
}

func TestScheduleJobShouldReturnErrorIfJobIsNotRegistered(t *testing.T) {
	config, _, _ := newTestConfig()
	scheduler := gotick.NewScheduler(config)

	id := uuid.NewString()

	schedule, err := gotick.NewOnce(time.Now().Add(1 * time.Minute))
	require.NoError(t, err)

	_, err = scheduler.ScheduleJob(context.Background(), id, schedule)
	assert.Equal(t, gotick.ErrJobNotFound, err)
}

func TestScheduleJobShouldSucceed(t *testing.T) {
	config, driver, _ := newTestConfig()
	scheduler := gotick.NewScheduler(config)

	jobID := uuid.NewString()
	scheduleID := uuid.NewString()

	job := newTestJob(jobID, nil)

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
	config, driver, _ := newTestConfig()
	scheduler := gotick.NewScheduler(config)

	id := uuid.NewString()
	job := newTestJob(id, nil)

	schedule, err := gotick.NewOnce(time.Now().Add(1 * time.Minute))
	require.NoError(t, err)

	driver.On("ScheduleJob", mock.Anything, job, schedule).Return("", fmt.Errorf("error"))

	err = scheduler.RegisterJob(job)
	require.NoError(t, err)

	_, err = scheduler.ScheduleJob(context.Background(), id, schedule)
	require.Error(t, err)
}

func TestStartShouldReturnSubscriberError(t *testing.T) {
	config, driver, planner := newTestConfig()
	scheduler := gotick.NewScheduler(config)
	subscriber := &schedulerSubscriberMock{}

	expected := fmt.Errorf("test")
	subscriber.On("OnStart").Return(expected)

	scheduler.Subscribe(subscriber)

	ctx, cancel := newTestContext()
	defer cancel()

	err := scheduler.Start(ctx)

	require.Error(t, err)
	driver.AssertNotCalled(t, "NextExecution", mock.Anything, mock.Anything)
	planner.AssertNotCalled(t, "Start", mock.Anything)
}

func TestStartShouldReturnPlannerError(t *testing.T) {
	config, driver, planner := newTestConfig()
	scheduler := gotick.NewScheduler(config)

	expected := fmt.Errorf("test")
	planner.On("Start", mock.Anything).Return(expected)

	ctx, cancel := newTestContext()
	defer cancel()

	err := scheduler.Start(ctx)

	require.Error(t, err)
	driver.AssertNotCalled(t, "NextExecution", mock.Anything, mock.Anything)
}

func TestStartShouldExecuteJobIfThereIsSome(t *testing.T) {
	config, driver, planner := newTestConfig()
	scheduler := gotick.NewScheduler(config)
	subscriber := &schedulerSubscriberMock{}

	scheduler.Subscribe(subscriber)

	id := uuid.NewString()
	job := newTestJob(id, nil)

	plannedTime := time.Now()

	sch, err := gotick.NewOnce(plannedTime)
	require.NoError(t, err)

	jobExecution := &gotick.JobPlannedExecution{
		Job:       job,
		PlannedAt: plannedTime,
		Schedule:  sch,
	}

	subscriber.On("OnStart").Return(nil)
	subscriber.On("OnBeforeJobExecution", mock.Anything).Return(nil)
	subscriber.On("OnBeforeJobPlanned", mock.Anything).Return(nil)
	subscriber.On("OnJobExecuted", mock.Anything).Return(nil)

	driver.On("NextExecution", mock.Anything, mock.Anything).Return(jobExecution, nil).Times(1)
	driver.On("NextExecution", mock.Anything, mock.Anything).Return(nil, nil)

	planner.On("Start", mock.Anything).Return(nil)
	planner.On("Stop").Return(nil)
	planner.On("Subscribe", scheduler).Return()

	var jobCtx *gotick.JobContext
	planner.On("Plan", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		jobCtx = args.Get(0).(*gotick.JobContext)

		assert.Equal(t, job, jobCtx.Job)
		assert.GreaterOrEqual(t, jobCtx.PlannedAt, plannedTime)
		assert.Equal(t, jobCtx.StartedAt, time.Time{})
		assert.Equal(t, jobCtx.ExecutedAt, time.Time{})
		assert.Equal(t, gotick.JobExecutionStatusInitiated, jobCtx.ExecutionStatus)
		assert.Equal(t, sch, jobCtx.Schedule)

		planner.subscribers[0].OnBeforeJobExecution(jobCtx)
		err := job.Execute(jobCtx)
		if err != nil {
			planner.subscribers[0].OnError(err)
		}

		planner.subscribers[0].OnJobExecuted(jobCtx)
	})

	ctx, cancel := newTestContext()
	defer cancel()

	err = scheduler.Start(ctx)
	require.NoError(t, err)

	select {
	case <-ctx.Done():
		require.Fail(t, "expected job to be executed within 2 seconds")
	case <-job.done:
	}

	subscriber.AssertCalled(t, "OnStart")
	subscriber.AssertNotCalled(t, "OnStop")
	subscriber.AssertCalled(t, "OnBeforeJobExecution", jobCtx)
	subscriber.AssertCalled(t, "OnBeforeJobPlanned", jobCtx)
	subscriber.AssertCalled(t, "OnJobExecuted", jobCtx)
	subscriber.AssertNotCalled(t, "OnError", mock.Anything)

	planner.AssertCalled(t, "Start", mock.Anything)
	planner.AssertNotCalled(t, "Stop")
	planner.AssertCalled(t, "Subscribe", scheduler)
	planner.AssertCalled(t, "Plan", jobCtx)
}

func TestStartShouldProceedBackgroundProcessIfNonFatalErrorOccured(t *testing.T) {
	config, driver, planner := newTestConfig(gotick.WithIdlePollingInterval(0))
	scheduler := gotick.NewScheduler(config)
	subscriber := &schedulerSubscriberMock{}

	scheduler.Subscribe(subscriber)

	id := uuid.NewString()
	job := newTestJob(id, nil)

	plannedTime := time.Now()

	jobExecution := &gotick.JobPlannedExecution{
		Job:       job,
		PlannedAt: plannedTime,
	}

	done := make(chan any)
	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		wg.Wait()
		close(done)
	}()

	expected := fmt.Errorf("test")

	subscriber.On("OnStart").Return(nil)
	subscriber.On("OnError", expected).Return().Run(func(args mock.Arguments) {
		wg.Done()
	})
	subscriber.On("OnBeforeJobExecution", mock.Anything).Return(nil)
	subscriber.On("OnBeforeJobPlanned", mock.Anything).Return(nil)
	subscriber.On("OnJobExecuted", mock.Anything).Return(nil)

	driver.On("NextExecution", mock.Anything, mock.Anything).Return(nil, expected).Times(1)
	driver.On("NextExecution", mock.Anything, mock.Anything).Return(jobExecution, nil).Times(1)
	driver.On("NextExecution", mock.Anything, mock.Anything).Return(nil, nil)

	planner.On("Start", mock.Anything).Return(nil)
	planner.On("Stop").Return(nil)
	planner.On("Subscribe", scheduler).Return()
	planner.On("Plan", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		wg.Done()
	})

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
	subscriber.AssertNotCalled(t, "OnBeforeJobExecution", mock.Anything)
	subscriber.AssertCalled(t, "OnBeforeJobPlanned", mock.Anything)
	subscriber.AssertNotCalled(t, "OnJobExecuted", mock.Anything)
	subscriber.AssertCalled(t, "OnError", expected)

	planner.AssertCalled(t, "Start", mock.Anything)
	planner.AssertNotCalled(t, "Stop")
	planner.AssertCalled(t, "Subscribe", scheduler)
	planner.AssertCalled(t, "Plan", mock.Anything)
}