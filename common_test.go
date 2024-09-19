package gotick

import (
	"context"
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

func (s *schedulerSubscriberMock) OnBeforeJobExecutionPlan(ctx *JobExecutionContext) {
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

func (d *driverMock) Start(ctx context.Context) error {
	args := d.Called(ctx)
	return args.Error(0)
}

func (d *driverMock) Stop() error {
	args := d.Called()
	return args.Error(0)
}

func (d *driverMock) BeforeExecution(ctx JobExecutionContext) error {
	args := d.Called(ctx)
	return args.Error(0)
}

func (d *driverMock) Executed(ctx JobExecutionContext) error {
	args := d.Called(ctx)
	return args.Error(0)
}

func (d *driverMock) NextExecution(ctx context.Context) *NextExecutionResult {
	args := d.Called(ctx)
	job := args.Get(0)
	if job == nil {
		return nil
	}

	return job.(*NextExecutionResult)
}

func (d *driverMock) UnscheduleJobByJobID(ctx context.Context, jobID string) error {
	args := d.Called(ctx, jobID)
	return args.Error(0)
}

func (d *driverMock) UnscheduleJobByScheduleID(ctx context.Context, scheduleID string) error {
	args := d.Called(ctx, scheduleID)
	return args.Error(0)
}

func (d *driverMock) ScheduleJob(ctx context.Context, jobID string, schedule JobSchedule) (string, error) {
	args := d.Called(ctx, jobID, schedule)
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
		WithDriverFactory(func(*SchedulerConfig) (SchedulerDriver, error) {
			return driver, nil
		}),
		WithPlannerFactory(func(*SchedulerConfig) (Planner, error) {
			return planner, nil
		}),
	)

	return DefaultSchedulerConfig(options...), driver, planner
}

type testJob struct {
	id string
}

func (j *testJob) ID() string {
	return j.id
}

func (j *testJob) Execute(ctx *JobExecutionContext) {
}

var _ Job = (*testJob)(nil)

func newTestJob(id string) *testJob {
	return &testJob{
		id: id,
	}
}
