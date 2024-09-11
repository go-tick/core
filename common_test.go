package gotick_test

import (
	"context"
	"sync"
	"time"

	"github.com/misikdmytro/gotick"
	"github.com/stretchr/testify/mock"
)

type plannerMock struct {
	mock.Mock
	subscribers []gotick.PlannerSubscriber
}

type driverMock struct {
	mock.Mock
}

type schedulerSubscriberMock struct {
	mock.Mock
}

func (s *schedulerSubscriberMock) OnStart() error {
	args := s.Called()
	return args.Error(0)
}

func (s *schedulerSubscriberMock) OnStop() error {
	args := s.Called()
	return args.Error(0)
}

func (s *schedulerSubscriberMock) OnBeforeJobExecution(ctx *gotick.JobContext) error {
	args := s.Called(ctx)
	return args.Error(0)
}

func (s *schedulerSubscriberMock) OnError(err error) {
	s.Called(err)
}

func (s *schedulerSubscriberMock) OnJobExecuted(ctx *gotick.JobContext) error {
	args := s.Called(ctx)
	return args.Error(0)
}

func (d *driverMock) BeforeExecution(ctx gotick.JobContext) error {
	args := d.Called(ctx)
	return args.Error(0)
}

func (d *driverMock) Executed(ctx gotick.JobContext) error {
	args := d.Called(ctx)
	return args.Error(0)
}

func (d *driverMock) NextExecution(ctx context.Context, t time.Time) (*gotick.JobPlannedExecution, error) {
	args := d.Called(ctx, t)
	job := args.Get(0)
	if job == nil {
		return nil, args.Error(1)
	}

	return job.(*gotick.JobPlannedExecution), args.Error(1)
}

func (d *driverMock) UnscheduleJobByJobID(ctx context.Context, jobID string) error {
	args := d.Called(ctx, jobID)
	return args.Error(0)
}

func (d *driverMock) UnscheduleJobByScheduleID(ctx context.Context, scheduleID string) error {
	args := d.Called(ctx, scheduleID)
	return args.Error(0)
}

func (d *driverMock) ScheduleJob(ctx context.Context, job gotick.Job, schedule gotick.JobSchedule) (string, error) {
	args := d.Called(ctx, job, schedule)
	return args.String(0), args.Error(1)
}

func (p *plannerMock) Subscribe(subscriber gotick.PlannerSubscriber) {
	p.Called(subscriber)
	p.subscribers = append(p.subscribers, subscriber)
}

func (p *plannerMock) Plan(ctx *gotick.JobContext) error {
	args := p.Called(ctx)
	return args.Error(0)
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
var _ gotick.SchedulerSubscriber = (*schedulerSubscriberMock)(nil)

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

func (j *testJob) Execute(ctx *gotick.JobContext) error {
	j.lock.Lock()
	defer j.lock.Unlock()

	j.executedAt = append(j.executedAt, time.Now())
	close(j.done)

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
