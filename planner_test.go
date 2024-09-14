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
	id := uuid.NewString()
	job := newTestJob(id)
	subscriber := &plannerSubscriberMock{}

	timeout, cancel := newTestContext()
	defer cancel()

	ctx := &JobExecutionContext{
		Context: timeout,
		Execution: JobPlannedExecution{
			JobScheduledExecution: JobScheduledExecution{
				Job: job,
			},
			PlannedAt: time.Now(),
		},
		ExecutionStatus: JobExecutionStatusInitiated,
	}

	planner := newPlanner(DefaultPlannerConfig())
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

	planner.Start(timeout)

	planner.Plan(ctx)

	select {
	case <-done:
		assert.Equal(t, JobExecutionStatusExecuted, ctx.ExecutionStatus)
	case <-timeout.Done():
		require.Fail(t, "expected job to be executed within 2 seconds")
	}
}

func TestPlanShouldCallSubscriberAfterTimeout(t *testing.T) {
	id := uuid.NewString()
	job := newTestJob(id)
	subscriber := &plannerSubscriberMock{}

	timeout, cancel := newTestContext()
	defer cancel()

	ctx := &JobExecutionContext{
		Context: timeout,
		Execution: JobPlannedExecution{
			JobScheduledExecution: JobScheduledExecution{
				Job: job,
			},
			PlannedAt: time.Now(),
		},
		ExecutionStatus: JobExecutionStatusInitiated,
	}

	planner := newPlanner(DefaultPlannerConfig(
		WithPlannerTimeout(100 * time.Millisecond),
	))
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
		require.Fail(t, "expected job to be cancelled within 2 seconds")
	}
}

func TestPlanShouldNotExecuteJobIfItsAheadOfTime(t *testing.T) {
	id := uuid.NewString()
	job := newTestJob(id)
	subscriber := &plannerSubscriberMock{}

	timeout, cancel := newTestContext()
	defer cancel()

	ctx := &JobExecutionContext{
		Context: timeout,
		Execution: JobPlannedExecution{
			JobScheduledExecution: JobScheduledExecution{
				Job: job,
			},
			PlannedAt: time.Now().Add(5 * time.Second),
		},
		ExecutionStatus: JobExecutionStatusInitiated,
	}

	planner := newPlanner(DefaultPlannerConfig())
	planner.Subscribe(subscriber)

	planner.Start(timeout)
	require.NoError(t, err)

	planner.Plan(ctx)

	<-timeout.Done()
	assert.Equal(t, JobExecutionStatusPlanned, ctx.ExecutionStatus)
}

func TestStopShouldBeCalledWithoutErrorTwice(t *testing.T) {
	planner := newPlanner(DefaultPlannerConfig())

	timeout, cancel := newTestContext()
	defer cancel()

	err := planner.Start(timeout)
	require.NoError(t, err)

	err = planner.Stop()
	require.NoError(t, err)

	err = planner.Stop()
	require.NoError(t, err)
}
