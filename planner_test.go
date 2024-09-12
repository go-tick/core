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

	ctx := &JobContext{
		Context:         timeout,
		Job:             job,
		PlannedAt:       time.Now(),
		ExecutionStatus: JobExecutionStatusInitiated,
	}

	planner := newPlanner(1)
	planner.Subscribe(subscriber)

	done := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		wg.Wait()
		close(done)
	}()

	subscriber.On("OnBeforeJobExecution", ctx).Return().Run(func(args mock.Arguments) {
		assert.Equal(t, JobExecutionStatusExecuting, ctx.ExecutionStatus)
		wg.Done()
	})
	subscriber.On("OnJobExecuted", ctx).Return().Run(func(args mock.Arguments) {
		assert.Equal(t, JobExecutionStatusExecuted, ctx.ExecutionStatus)
		wg.Done()
	})

	err := planner.Start(timeout)
	require.NoError(t, err)

	err = planner.Plan(ctx)
	require.NoError(t, err)

	select {
	case <-done:
		assert.Equal(t, JobExecutionStatusExecuted, ctx.ExecutionStatus)
	case <-timeout.Done():
		require.Fail(t, "expected job to be executed within 2 seconds")
	}
}

func TestPlanShouldNotExecuteJobIfItsAheadOfTime(t *testing.T) {
	id := uuid.NewString()
	job := newTestJob(id)
	subscriber := &plannerSubscriberMock{}

	timeout, cancel := newTestContext()
	defer cancel()

	ctx := &JobContext{
		Context:         timeout,
		Job:             job,
		PlannedAt:       time.Now().Add(10 * time.Minute),
		ExecutionStatus: JobExecutionStatusInitiated,
	}

	planner := newPlanner(1)
	planner.Subscribe(subscriber)

	err := planner.Start(timeout)
	require.NoError(t, err)

	err = planner.Plan(ctx)
	require.NoError(t, err)

	<-timeout.Done()
	assert.Equal(t, JobExecutionStatusPlanned, ctx.ExecutionStatus)
}

func TestStopShouldBeCalledWithoutErrorTwice(t *testing.T) {
	planner := newPlanner(1)

	timeout, cancel := newTestContext()
	defer cancel()

	err := planner.Start(timeout)
	require.NoError(t, err)

	err = planner.Stop()
	require.NoError(t, err)

	err = planner.Stop()
	require.NoError(t, err)
}
