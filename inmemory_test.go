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
	schedule := NewCalendarSchedule(time.Now())
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
	schedule := NewCalendarSchedule(time.Now())
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
	schedule := NewCalendarSchedule(time.Now())
	driver := newInMemoryDriver()

	scheduleID, err := driver.ScheduleJob(context.Background(), job, schedule)

	require.NoError(t, err)
	require.NotEmpty(t, scheduleID)

	err = driver.UnscheduleJobByScheduleID(context.Background(), scheduleID)

	assert.NoError(t, err)
}

func TestNextExecutionShouldReturnExecutionsOneByOne(t *testing.T) {
	job := newTestJob(uuid.NewString())

	schedule1 := NewCalendarSchedule(time.Now().Add(1 * time.Second))
	schedule2, err := NewSequenceSchedule(
		time.Now().Add(2*time.Second),
		time.Now().Add(3*time.Second),
		time.Now().Add(4*time.Second),
	)
	require.NoError(t, err)

	schedule3 := NewCalendarSchedule(time.Now())

	driver := newInMemoryDriver()

	scheduleID1, err := driver.ScheduleJob(context.Background(), job, schedule1)
	require.NoError(t, err)

	scheduleID2, err := driver.ScheduleJob(context.Background(), job, schedule2)
	require.NoError(t, err)

	scheduleID3, err := driver.ScheduleJob(context.Background(), job, schedule3)
	require.NoError(t, err)

	assertExecution := func(
		execution *JobPlannedExecution,
		schedule JobSchedule,
		scheduleID string,
		action func(*JobExecutionContext),
	) {
		assert.LessOrEqual(t, time.Time{}, execution.PlannedAt)
		assert.Equal(t, schedule, execution.JobScheduledExecution.Schedule)
		assert.Equal(t, scheduleID, execution.JobScheduledExecution.ScheduleID)
		assert.Equal(t, job, execution.JobScheduledExecution.Job)
		assert.NotEmpty(t, execution.ExecutionID)

		driver.OnJobExecutionInitiated(&JobExecutionContext{
			Execution: *execution,
		})

		jobCtx := &JobExecutionContext{
			Execution: *execution,
		}

		action(jobCtx)
	}

	assertExecutionWithExecuted := func(execution *JobPlannedExecution, schedule JobSchedule, scheduleID string) {
		assertExecution(execution, schedule, scheduleID, driver.OnJobExecuted)
	}

	assertExecutionWithSkipped := func(execution *JobPlannedExecution, schedule JobSchedule, scheduleID string) {
		assertExecution(execution, schedule, scheduleID, driver.OnJobExecutionSkipped)
	}

	execution := driver.NextExecution(context.Background())
	assertExecutionWithExecuted(execution, schedule3, scheduleID3)

	execution = driver.NextExecution(context.Background())
	assertExecutionWithSkipped(execution, schedule1, scheduleID1)

	execution = driver.NextExecution(context.Background())
	assertExecutionWithExecuted(execution, schedule2, scheduleID2)

	execution = driver.NextExecution(context.Background())
	assertExecutionWithSkipped(execution, schedule2, scheduleID2)

	execution = driver.NextExecution(context.Background())
	assertExecutionWithExecuted(execution, schedule2, scheduleID2)

	execution = driver.NextExecution(context.Background())
	assert.Nil(t, execution)
}

func TestNextExecutionShouldReturnExecutionsByCron(t *testing.T) {
	job := newTestJob(uuid.NewString())

	schedule, err := NewCronSchedule("0/1 * * * *")
	require.NoError(t, err)

	driver := newInMemoryDriver()

	_, err = driver.ScheduleJob(context.Background(), job, schedule)
	require.NoError(t, err)

	execution1 := driver.NextExecution(context.Background())

	driver.OnJobExecuted(&JobExecutionContext{
		Execution: *execution1,
	})

	execution2 := driver.NextExecution(context.Background())

	assert.Equal(t, 1*time.Minute, execution2.PlannedAt.Sub(execution1.PlannedAt))
}
