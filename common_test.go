package gotick

import (
	"context"
	"sync"
	"time"

	"github.com/stretchr/testify/mock"
)

func newTestContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 3*time.Second)
}

type plannerMock struct {
	mock.Mock
	subscribers []PlannerSubscriber
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

func (p *plannerSubscriberMock) OnBeforeJobExecution(ctx *JobContext) {
	p.Called(ctx)
}

func (p *plannerSubscriberMock) OnJobExecuted(ctx *JobContext) {
	p.Called(ctx)
}

func (s *schedulerSubscriberMock) OnStart() {
	s.Called()
}

func (s *schedulerSubscriberMock) OnStop() {
	s.Called()
}

func (s *schedulerSubscriberMock) OnBeforeJobPlanned(ctx *JobContext) {
	s.Called(ctx)
}

func (s *schedulerSubscriberMock) OnBeforeJobExecution(ctx *JobContext) {
	s.Called(ctx)
}

func (s *schedulerSubscriberMock) OnError(err error) {
	s.Called(err)
}

func (s *schedulerSubscriberMock) OnJobExecuted(ctx *JobContext) {
	s.Called(ctx)
}

func (d *driverMock) BeforeExecution(ctx JobContext) error {
	args := d.Called(ctx)
	return args.Error(0)
}

func (d *driverMock) Executed(ctx JobContext) error {
	args := d.Called(ctx)
	return args.Error(0)
}

func (d *driverMock) NextExecution(ctx context.Context) (*JobPlannedExecution, error) {
	args := d.Called(ctx)
	job := args.Get(0)
	if job == nil {
		return nil, args.Error(1)
	}

	return job.(*JobPlannedExecution), args.Error(1)
}

func (d *driverMock) UnscheduleJobByJobID(ctx context.Context, jobID string) error {
	args := d.Called(ctx, jobID)
	return args.Error(0)
}

func (d *driverMock) UnscheduleJobByScheduleID(ctx context.Context, scheduleID string) error {
	args := d.Called(ctx, scheduleID)
	return args.Error(0)
}

func (d *driverMock) ScheduleJob(ctx context.Context, job Job, schedule JobSchedule) (string, error) {
	args := d.Called(ctx, job, schedule)
	return args.String(0), args.Error(1)
}

func (p *plannerMock) Subscribe(subscriber PlannerSubscriber) {
	p.Called(subscriber)
	p.subscribers = append(p.subscribers, subscriber)
}

func (p *plannerMock) Plan(ctx *JobContext) error {
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

var _ Planner = (*plannerMock)(nil)
var _ SchedulerDriver = (*driverMock)(nil)
var _ SchedulerSubscriber = (*schedulerSubscriberMock)(nil)
var _ PlannerSubscriber = (*plannerSubscriberMock)(nil)

func newTestConfig(options ...SchedulerOption) (SchedulerConfiguration, *driverMock, *plannerMock) {
	driver, planner := new(driverMock), new(plannerMock)
	options = append(
		options,
		WithDriverFactory(func() SchedulerDriver {
			return driver
		}),
		WithPlannerFactory(func() Planner {
			return planner
		}),
	)

	return DefaultConfig(options...), driver, planner
}

type TestJob struct {
	id   string
	lock sync.Mutex
	done chan any
}

func (j *TestJob) ID() string {
	return j.id
}

func (j *TestJob) Execute(ctx *JobContext) {
	j.lock.Lock()
	defer j.lock.Unlock()

	close(j.done)
}

var _ Job = (*TestJob)(nil)

func newTestJob(id string) *TestJob {
	return &TestJob{
		id:   id,
		done: make(chan any),
	}
}
