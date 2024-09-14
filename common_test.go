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

func (p *plannerSubscriberMock) OnJobExecutionUnplanned(ctx *JobExecutionContext) {
	p.Called(ctx)
}

func (p *plannerSubscriberMock) OnBeforeJobExecution(ctx *JobExecutionContext) {
	p.Called(ctx)
}

func (p *plannerSubscriberMock) OnJobExecuted(ctx *JobExecutionContext) {
	p.Called(ctx)
}

func (s *schedulerSubscriberMock) OnStart() {
	s.Called()
}

func (s *schedulerSubscriberMock) OnStop() {
	s.Called()
}

func (s *schedulerSubscriberMock) OnJobExecutionInitiated(ctx *JobExecutionContext) {
	s.Called(ctx)
}

func (s *schedulerSubscriberMock) OnJobExecutionDelayed(ctx *JobExecutionContext) {
	s.Called(ctx)
}

func (s *schedulerSubscriberMock) OnJobExecutionSkipped(ctx *JobExecutionContext) {
	s.Called(ctx)
}

func (s *schedulerSubscriberMock) OnBeforeJobExecutionPlanned(ctx *JobExecutionContext) {
	s.Called(ctx)
}

func (s *schedulerSubscriberMock) OnJobExecutionUnplanned(ctx *JobExecutionContext) {
	s.Called(ctx)
}

func (s *schedulerSubscriberMock) OnBeforeJobExecution(ctx *JobExecutionContext) {
	s.Called(ctx)
}

func (s *schedulerSubscriberMock) OnJobExecuted(ctx *JobExecutionContext) {
	s.Called(ctx)
}

func (d *driverMock) BeforeExecution(ctx JobExecutionContext) error {
	args := d.Called(ctx)
	return args.Error(0)
}

func (d *driverMock) Executed(ctx JobExecutionContext) error {
	args := d.Called(ctx)
	return args.Error(0)
}

func (d *driverMock) NextExecution(ctx context.Context) *JobPlannedExecution {
	args := d.Called(ctx)
	job := args.Get(0)
	if job == nil {
		return nil
	}

	return job.(*JobPlannedExecution)
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

func (p *plannerMock) Plan(ctx *JobExecutionContext) {
	p.Called(ctx)
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

func newTestConfig(options ...Option[SchedulerConfig]) (*SchedulerConfig, *driverMock, *plannerMock) {
	driver, planner := new(driverMock), new(plannerMock)
	options = append(
		options,
		WithDriverFactory(func(*SchedulerConfig) SchedulerDriver {
			return driver
		}),
		WithPlannerFactory(func(*SchedulerConfig) Planner {
			return planner
		}),
	)

	return DefaultSchedulerConfig(options...), driver, planner
}

type TestJob struct {
	id   string
	lock sync.Mutex
	done chan any
}

func (j *TestJob) ID() string {
	return j.id
}

func (j *TestJob) Execute(ctx *JobExecutionContext) {
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
