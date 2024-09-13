package gotick

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScheduleJobShouldReturnUniqueScheduleID(t *testing.T) {
	job := newTestJob(uuid.NewString())
	schedule := newFakeJobSchedule(time.Now())
	driver := newInMemoryDriver()

	scheduleID1, err1 := driver.ScheduleJob(context.Background(), job, schedule)
	scheduleID2, err2 := driver.ScheduleJob(context.Background(), job, schedule)

	assert.NoError(t, err1)
	assert.NoError(t, err2)

	assert.NotEmpty(t, scheduleID1)
	assert.NotEmpty(t, scheduleID2)

	assert.NotEqual(t, scheduleID1, scheduleID2)
}

func TestUnscheduleJobByJobIDShouldDoItEvenIfJobDoesNotExist(t *testing.T) {
	driver := newInMemoryDriver()

	err := driver.UnscheduleJobByJobID(context.Background(), uuid.NewString())

	assert.NoError(t, err)
}

func TestUnscheduleJobByJobIDShouldDoItSuccessfully(t *testing.T) {
	job := newTestJob(uuid.NewString())
	schedule := newFakeJobSchedule(time.Now())
	driver := newInMemoryDriver()

	scheduleID, err := driver.ScheduleJob(context.Background(), job, schedule)

	require.NoError(t, err)
	require.NotEmpty(t, scheduleID)

	err = driver.UnscheduleJobByJobID(context.Background(), job.ID())

	assert.NoError(t, err)
}

func TestUnscheduleJobByScheduleIDShouldDoItEvenIfJobDoesNotExist(t *testing.T) {
	driver := newInMemoryDriver()

	err := driver.UnscheduleJobByScheduleID(context.Background(), uuid.NewString())

	assert.NoError(t, err)
}

func TestUnscheduleJobByScheduleIDShouldDoItSuccessfully(t *testing.T) {
	job := newTestJob(uuid.NewString())
	schedule := newFakeJobSchedule(time.Now())
	driver := newInMemoryDriver()

	scheduleID, err := driver.ScheduleJob(context.Background(), job, schedule)

	require.NoError(t, err)
	require.NotEmpty(t, scheduleID)

	err = driver.UnscheduleJobByScheduleID(context.Background(), scheduleID)

	assert.NoError(t, err)
}

func TestNextExecutionShouldReturnExecutionsOneByOne(t *testing.T) {
	job := newTestJob(uuid.NewString())

	schedule1 := newFakeJobSchedule(time.Now())
	schedule2 := newFakeJobSchedule(time.Now().Add(1 * time.Second))
	schedule3 := newFakeJobSchedule(time.Now().Add(-2 * time.Second))

	driver := newInMemoryDriver()

	scheduleID1, err := driver.ScheduleJob(context.Background(), job, schedule1)
	require.NoError(t, err)

	scheduleID2, err := driver.ScheduleJob(context.Background(), job, schedule2)
	require.NoError(t, err)

	scheduleID3, err := driver.ScheduleJob(context.Background(), job, schedule3)
	require.NoError(t, err)

	assertCorrectExecution := func(execution *JobPlannedExecution, schedule *fakeJobSchedule, scheduleID string) {
		assert.Equal(t, *schedule.next, execution.PlannedAt)
		assert.Equal(t, schedule, execution.JobScheduledExecution.Schedule)
		assert.Equal(t, scheduleID, execution.JobScheduledExecution.ScheduleID)
		assert.Equal(t, job, execution.JobScheduledExecution.Job)
		assert.NotEmpty(t, execution.ExecutionID)
	}

	execution, err := driver.NextExecution(context.Background())

	require.NoError(t, err)
	assertCorrectExecution(execution, schedule3, scheduleID3)

	execution, err = driver.NextExecution(context.Background())

	require.NoError(t, err)
	assertCorrectExecution(execution, schedule1, scheduleID1)

	execution, err = driver.NextExecution(context.Background())

	require.NoError(t, err)
	assertCorrectExecution(execution, schedule2, scheduleID2)

	jobCtx := &JobExecutionContext{
		Execution: *execution,
	}

	driver.OnJobExecuted(jobCtx)

	execution, err = driver.NextExecution(context.Background())

	require.NoError(t, err)
	assertCorrectExecution(execution, schedule2, scheduleID2)

	jobCtx = &JobExecutionContext{
		Execution: *execution,
	}

	driver.OnJobExecutionSkipped(jobCtx)

	execution, err = driver.NextExecution(context.Background())

	require.NoError(t, err)
	assertCorrectExecution(execution, schedule2, scheduleID2)

	execution, err = driver.NextExecution(context.Background())

	require.NoError(t, err)
	assert.Nil(t, execution)
}

func TestNextExecutionShouldReturnExecutionsByCron(t *testing.T) {
	job := newTestJob(uuid.NewString())

	schedule, err := NewCron("0/1 * * * *")
	require.NoError(t, err)

	driver := newInMemoryDriver()

	_, err = driver.ScheduleJob(context.Background(), job, schedule)
	require.NoError(t, err)

	execution1, err := driver.NextExecution(context.Background())
	require.NoError(t, err)

	driver.OnJobExecuted(&JobExecutionContext{
		Execution: *execution1,
	})

	execution2, err := driver.NextExecution(context.Background())
	require.NoError(t, err)

	assert.Equal(t, execution2.PlannedAt.Sub(execution1.PlannedAt), 1*time.Minute)
}
