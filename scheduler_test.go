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

type plannerMock struct {
	mock.Mock
}

type driverMock struct {
	mock.Mock
}

func (d *driverMock) Executed(ctx context.Context, job gotick.Job, t time.Time) error {
	args := d.Called(ctx, job, t)
	return args.Error(0)
}

func (d *driverMock) NextExecution(ctx context.Context, t time.Time) (gotick.Job, time.Time, error) {
	args := d.Called(ctx, t)
	job := args.Get(0)
	if job == nil {
		return nil, args.Get(1).(time.Time), args.Error(2)
	}

	return job.(gotick.Job), args.Get(1).(time.Time), args.Error(2)
}

func (d *driverMock) UnscheduleJob(ctx context.Context, jobID string) error {
	args := d.Called(ctx, jobID)
	return args.Error(0)
}

func (d *driverMock) ScheduleJob(ctx context.Context, job gotick.Job, schedule gotick.JobSchedule) error {
	args := d.Called(ctx, job, schedule)
	return args.Error(0)
}

func (p *plannerMock) Errs() <-chan error {
	return p.Called().Get(0).(<-chan error)
}

func (p *plannerMock) Plan(ctx context.Context, job gotick.Job, t time.Time) (<-chan any, error) {
	args := p.Called(ctx, job, t)
	res := args.Get(0)
	if res == nil {
		return nil, args.Error(1)
	}

	return res.(<-chan any), args.Error(1)
}

func (p *plannerMock) Start(ctx context.Context) error {
	args := p.Called(ctx)
	return args.Error(0)
}

func (p *plannerMock) Stop() error {
	args := p.Called()
	return args.Error(0)
}

var _ gotick.Planner = (*plannerMock)(nil)
var _ gotick.SchedulerDriver = (*driverMock)(nil)

func NewTestConfig(options ...gotick.SchedulerOption) (gotick.SchedulerConfiguration, *driverMock, *plannerMock) {
	driver, planner := new(driverMock), new(plannerMock)
	options = append(
		options,
		gotick.WithDriverFactory(func() gotick.SchedulerDriver {
			return driver
		}),
		gotick.WithPlannerFactory(func() gotick.Planner {
			return planner
		}),
	)

	return gotick.DefaultConfig(options...), driver, planner
}

type testJob struct {
	id         string
	lock       sync.Mutex
	executedAt []time.Time
	done       chan any
}

func (j *testJob) ID() string {
	return j.id
}

func (j *testJob) Execute(ctx context.Context) error {
	j.lock.Lock()
	defer j.lock.Unlock()

	j.executedAt = append(j.executedAt, time.Now())
	j.done <- struct{}{}

	return nil
}

var _ gotick.Job = (*testJob)(nil)

func NewTestJob(id string) *testJob {
	return &testJob{
		id:         id,
		executedAt: make([]time.Time, 0, 1),
		done:       make(chan any),
	}
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

func TestUnscheduleJobShouldDoIt(t *testing.T) {
	config, driver, _ := NewTestConfig()
	scheduler := gotick.NewScheduler(config)

	id := uuid.NewString()

	driver.On("UnscheduleJob", mock.Anything, id).Return(nil)

	err := scheduler.UnscheduleJob(context.Background(), id)
	assert.NoError(t, err)
}

func TestUnscheduleJobShouldReturnErrorIfDriverFails(t *testing.T) {
	config, driver, _ := NewTestConfig()
	scheduler := gotick.NewScheduler(config)

	id := uuid.NewString()

	driver.On("UnscheduleJob", mock.Anything, id).Return(fmt.Errorf("error"))

	err := scheduler.UnscheduleJob(context.Background(), id)

	require.Error(t, err)
}

func TestScheduleJobShouldReturnErrorIfJobIsNotRegistered(t *testing.T) {
	config, _, _ := NewTestConfig()
	scheduler := gotick.NewScheduler(config)

	id := uuid.NewString()

	schedule, err := gotick.NewOnce(time.Now().Add(1 * time.Minute))
	require.NoError(t, err)

	err = scheduler.ScheduleJob(context.Background(), id, schedule)
	assert.Equal(t, gotick.ErrJobNotFound, err)
}

func TestScheduleJobShouldSucceed(t *testing.T) {
	config, driver, _ := NewTestConfig()
	scheduler := gotick.NewScheduler(config)

	id := uuid.NewString()
	job := NewTestJob(id)

	schedule, err := gotick.NewOnce(time.Now().Add(1 * time.Minute))
	require.NoError(t, err)

	driver.On("ScheduleJob", mock.Anything, job, schedule).Return(nil)

	err = scheduler.RegisterJob(job)
	require.NoError(t, err)

	err = scheduler.ScheduleJob(context.Background(), id, schedule)
	assert.NoError(t, err)
}

func TestScheduleJobShouldReturnErrorIfDriverFails(t *testing.T) {
	config, driver, _ := NewTestConfig()
	scheduler := gotick.NewScheduler(config)

	id := uuid.NewString()
	job := NewTestJob(id)

	schedule, err := gotick.NewOnce(time.Now().Add(1 * time.Minute))
	require.NoError(t, err)

	driver.On("ScheduleJob", mock.Anything, job, schedule).Return(fmt.Errorf("error"))

	err = scheduler.RegisterJob(job)
	require.NoError(t, err)

	err = scheduler.ScheduleJob(context.Background(), id, schedule)
	require.Error(t, err)
}

func TestStartShouldExecuteJobIfThereIsSome(t *testing.T) {
	config, driver, planner := NewTestConfig()
	scheduler := gotick.NewScheduler(config)

	id := uuid.NewString()
	job := NewTestJob(id)

	plannedTime := time.Now()

	driver.On("NextExecution", mock.Anything, mock.Anything).Return(job, plannedTime, nil).Times(1)
	driver.On("NextExecution", mock.Anything, mock.Anything).Return(nil, time.Time{}, nil)

	executed := make(chan any, 1)
	driver.On("Executed", mock.Anything, job, plannedTime).Return(nil).Run(func(args mock.Arguments) {
		executed <- struct{}{}
	})

	plan := make(chan any, 1)
	var res <-chan any = plan
	planner.On("Plan", mock.Anything, job, plannedTime).Return(res, nil).Run(func(args mock.Arguments) {
		err := job.Execute(args.Get(0).(context.Context))
		require.NoError(t, err)

		plan <- struct{}{}
	})

	planner.On("Errs").Return(make(<-chan error))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := scheduler.Start(ctx)
	require.NoError(t, err)

	select {
	case <-ctx.Done():
		require.Fail(t, "expected job to be executed within 5 seconds")
	case <-job.done:
		assert.Equal(t, 1, len(job.executedAt))
		assert.LessOrEqual(t, job.executedAt[0], time.Now())
	}

	select {
	case <-executed:
		return
	case <-ctx.Done():
		require.Fail(t, "expected job to be executed within 5 seconds")
	}
}

func TestStartShouldSkipExecutionIfThisAheadOfPlaningTime(t *testing.T) {
	config, driver, planner := NewTestConfig()
	scheduler := gotick.NewScheduler(config)

	id := uuid.NewString()
	job := NewTestJob(id)

	plannedTime := time.Now().Add(config.MaxPlanAhead * 2)

	driver.On("NextExecution", mock.Anything, mock.Anything).Return(job, plannedTime, nil).Times(1)
	driver.On("NextExecution", mock.Anything, mock.Anything).Return(nil, time.Time{}, nil)

	planner.On("Errs").Return(make(<-chan error))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := scheduler.Start(ctx)
	require.NoError(t, err)

	select {
	case <-ctx.Done():
		driver.AssertCalled(t, "NextExecution", mock.Anything, mock.Anything)
		planner.AssertNotCalled(t, "Plan", mock.Anything, mock.Anything, mock.Anything)
	case <-job.done:
		require.Fail(t, "execution is not expected")
	}
}

func TestStartShouldPublishErrorsFromDriverNextExecution(t *testing.T) {
	config, driver, planner := NewTestConfig()
	scheduler := gotick.NewScheduler(config)

	expected := fmt.Errorf("test")
	driver.On("NextExecution", mock.Anything, mock.Anything).Return(nil, time.Time{}, expected)

	planner.On("Errs").Return(make(<-chan error))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := scheduler.Start(ctx)
	require.NoError(t, err)

	select {
	case <-ctx.Done():
		require.Fail(t, "expected error to be published")
	case err := <-scheduler.Errs():
		assert.Equal(t, err, expected)
	}
}

func TestStartShouldPublishErrorsFromPlannerPlan(t *testing.T) {
	config, driver, planner := NewTestConfig()
	scheduler := gotick.NewScheduler(config)

	id := uuid.NewString()
	job := NewTestJob(id)

	plannedTime := time.Now()

	driver.On("NextExecution", mock.Anything, mock.Anything).Return(job, plannedTime, nil).Times(1)
	driver.On("NextExecution", mock.Anything, mock.Anything).Return(nil, time.Time{}, nil)

	expected := fmt.Errorf("test")
	planner.On("Plan", mock.Anything, job, plannedTime).Return(nil, expected)

	planner.On("Errs").Return(make(<-chan error))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := scheduler.Start(ctx)
	require.NoError(t, err)

	select {
	case <-ctx.Done():
		require.Fail(t, "expected error to be published")
	case err := <-scheduler.Errs():
		assert.Equal(t, err, expected)
	}
}

func TestStartShouldPublishErrorsFromPlannerErrorChannel(t *testing.T) {
	config, driver, planner := NewTestConfig()
	scheduler := gotick.NewScheduler(config)

	driver.On("NextExecution", mock.Anything, mock.Anything).Return(nil, time.Time{}, nil)

	expected := fmt.Errorf("test")
	errs := make(chan error, 1)
	errs <- expected
	var res <-chan error = errs
	planner.On("Errs").Return(res)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := scheduler.Start(ctx)
	require.NoError(t, err)

	select {
	case <-ctx.Done():
		require.Fail(t, "expected error to be published")
	case err := <-scheduler.Errs():
		assert.Equal(t, err, expected)
	}
}

func TestStartShouldNotBlockExecutionIfNobodySubscribedToErrs(t *testing.T) {
	config, driver, planner := NewTestConfig(gotick.WithPollInterval(1 * time.Second))
	scheduler := gotick.NewScheduler(config)

	expected := fmt.Errorf("test")
	numOfCalls := 0
	driver.On("NextExecution", mock.Anything, mock.Anything).Return(nil, time.Time{}, expected).Run(func(args mock.Arguments) {
		numOfCalls = numOfCalls + 1
	})

	planner.On("Errs").Return(make(<-chan error))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := scheduler.Start(ctx)
	require.NoError(t, err)

	<-ctx.Done()

	assert.Greater(t, numOfCalls, 1)
}
