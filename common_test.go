package gotick_test

import (
	"context"
	"sync"
	"time"

	"github.com/misikdmytro/gotick"
	"github.com/stretchr/testify/mock"
)

func newTestContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 3*time.Second)
}

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

type plannerSubscriberMock struct {
	mock.Mock
}

func (p *plannerSubscriberMock) OnError(err error) {
	p.Called(err)
}

func (p *plannerSubscriberMock) OnBeforeJobExecution(ctx *gotick.JobContext) {
	p.Called(ctx)
}

func (p *plannerSubscriberMock) OnJobExecuted(ctx *gotick.JobContext) {
	p.Called(ctx)
}

func (s *schedulerSubscriberMock) OnStart() error {
	args := s.Called()
	return args.Error(0)
}

func (s *schedulerSubscriberMock) OnStop() error {
	args := s.Called()
	return args.Error(0)
}

func (s *schedulerSubscriberMock) OnBeforeJobPlanned(ctx *gotick.JobContext) error {
	args := s.Called(ctx)
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
var _ gotick.PlannerSubscriber = (*plannerSubscriberMock)(nil)

func newTestConfig(options ...gotick.SchedulerOption) (gotick.SchedulerConfiguration, *driverMock, *plannerMock) {
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

type TestJob struct {
	id   string
	lock sync.Mutex
	done chan any
	err  error
}

func (j *TestJob) ID() string {
	return j.id
}

func (j *TestJob) Execute(ctx *gotick.JobContext) error {
	j.lock.Lock()
	defer j.lock.Unlock()

	close(j.done)

	return j.err
}

var _ gotick.Job = (*TestJob)(nil)

func newTestJob(id string, err error) *TestJob {
	return &TestJob{
		id:   id,
		done: make(chan any),
		err:  err,
	}
}
