package gotick_test

import (
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/misikdmytro/gotick"
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

	ctx := &gotick.JobContext{
		Context:         timeout,
		Job:             job,
		PlannedAt:       time.Now(),
		ExecutionStatus: gotick.JobExecutionStatusInitiated,
	}

	planner := gotick.NewPlanner(1)
	planner.Subscribe(subscriber)

	done := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		wg.Wait()
		close(done)
	}()

	subscriber.On("OnBeforeJobExecution", ctx).Return().Run(func(args mock.Arguments) {
		assert.Equal(t, gotick.JobExecutionStatusExecuting, ctx.ExecutionStatus)
		wg.Done()
	})
	subscriber.On("OnJobExecuted", ctx).Return().Run(func(args mock.Arguments) {
		assert.Equal(t, gotick.JobExecutionStatusExecuted, ctx.ExecutionStatus)
		wg.Done()
	})

	planner.Start(timeout)
	planner.Plan(ctx)

	select {
	case <-done:
		assert.Equal(t, gotick.JobExecutionStatusExecuted, ctx.ExecutionStatus)
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

	ctx := &gotick.JobContext{
		Context:         timeout,
		Job:             job,
		PlannedAt:       time.Now().Add(10 * time.Minute),
		ExecutionStatus: gotick.JobExecutionStatusInitiated,
	}

	planner := gotick.NewPlanner(1)
	planner.Subscribe(subscriber)

	planner.Start(timeout)
	planner.Plan(ctx)

	<-timeout.Done()
	assert.Equal(t, gotick.JobExecutionStatusPlanned, ctx.ExecutionStatus)
}
