package gotick

import (
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPlanShouldExecuteTheJob(t *testing.T) {
	jobID := uuid.NewString()
	job := newTestJob(jobID)
	subscriber := &plannerSubscriberMock{}

	timeout, cancel := newTestContext()
	defer cancel()

	ctx := &JobExecutionContext{
		Context: timeout,

		JobID:       jobID,
		ScheduleID:  uuid.NewString(),
		ExecutionID: uuid.NewString(),

		PlannedAt: time.Now(),

		ExecutionStatus: JobExecutionStatusInitiated,
	}

	planner, err := newPlanner(DefaultPlannerConfig(WithJobs(job)))
	require.NoError(t, err)

	planner.Subscribe(subscriber)

	done := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		wg.Wait()
		close(done)
	}()

	subscriber.On("OnBeforeJobExecution", ctx).Return().Run(func(args mock.Arguments) {
		assert.Equal(t, JobExecutionStatusPlanned, ctx.ExecutionStatus)
		wg.Done()
	})
	subscriber.On("OnJobExecuted", ctx).Return().Run(func(args mock.Arguments) {
		assert.Equal(t, JobExecutionStatusExecuted, ctx.ExecutionStatus)
		wg.Done()
	})

	err = planner.Start(timeout)
	require.NoError(t, err)

	planner.Plan(ctx)

	select {
	case <-done:
		assert.Equal(t, JobExecutionStatusExecuted, ctx.ExecutionStatus)
	case <-timeout.Done():
		require.Fail(t, "expected job to be executed within 2 seconds")
	}
}

func TestPlanShouldCallSubscriberAfterTimeout(t *testing.T) {
	jobID := uuid.NewString()
	job := newTestJob(jobID)
	subscriber := &plannerSubscriberMock{}

	timeout, cancel := newTestContext()
	defer cancel()

	ctx := &JobExecutionContext{
		Context: timeout,

		JobID:       jobID,
		ScheduleID:  uuid.NewString(),
		ExecutionID: uuid.NewString(),

		PlannedAt: time.Now(),

		ExecutionStatus: JobExecutionStatusInitiated,
	}

	planner, err := newPlanner(DefaultPlannerConfig(
		WithPlannerTimeout(100*time.Millisecond),
		WithJobs(job),
	))
	require.NoError(t, err)

	planner.Subscribe(subscriber)

	done := make(chan struct{})
	subscriber.On("OnJobExecutionUnplanned", ctx).Return().Run(func(args mock.Arguments) {
		close(done)
	})

	planner.Plan(ctx)
	planner.Plan(ctx)

	select {
	case <-done:
		assert.Equal(t, JobExecutionStatusUnplanned, ctx.ExecutionStatus)
	case <-timeout.Done():
		require.Fail(t, "expected job to be canceled within 2 seconds")
	}
}

func TestPlanShouldNotExecuteJobIfItsAheadOfTime(t *testing.T) {
	jobID := uuid.NewString()
	job := newTestJob(jobID)
	subscriber := &plannerSubscriberMock{}

	timeout, cancel := newTestContext()
	defer cancel()

	ctx := &JobExecutionContext{
		Context: timeout,

		JobID:       jobID,
		ScheduleID:  uuid.NewString(),
		ExecutionID: uuid.NewString(),

		PlannedAt: time.Now().Add(1 * time.Hour),

		ExecutionStatus: JobExecutionStatusInitiated,
	}

	planner, err := newPlanner(DefaultPlannerConfig(WithJobs(job)))
	require.NoError(t, err)

	planner.Subscribe(subscriber)

	err = planner.Start(timeout)
	require.NoError(t, err)

	planner.Plan(ctx)

	<-timeout.Done()
	assert.Equal(t, JobExecutionStatusPlanned, ctx.ExecutionStatus)
}

func TestStopShouldBeCalledWithoutErrorTwice(t *testing.T) {
	planner, err := newPlanner(DefaultPlannerConfig())
	require.NoError(t, err)

	timeout, cancel := newTestContext()
	defer cancel()

	err = planner.Start(timeout)
	require.NoError(t, err)

	err = planner.Stop()
	require.NoError(t, err)

	err = planner.Stop()
	require.NoError(t, err)
}
