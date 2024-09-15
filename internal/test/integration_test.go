package test

import (
	"context"
	"testing"
	"time"

	"github.com/misikdmytro/gotick"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type jobWithDelay struct {
	id         string
	delay      time.Duration
	executions []*gotick.JobExecutionContext
}

type schedulerTestSubscriberCalls struct {
	NumOfOnBeforeJobExecutionCalls     int
	NumOfOnBeforeJobExecutionPlanCalls int
	NumOfOnJobExecutedCalls            int
	NumOfOnJobExecutionDelayedCalls    int
	NumOfOnJobExecutionInitiatedCalls  int
	NumOfOnJobExecutionSkippedCalls    int
	NumOfOnJobExecutionUnplannedCalls  int
	NumOfOnStartCalls                  int
	NumOfOnStopCalls                   int
}

type schedulerTestSubscriber struct {
	mock.Mock
	calls *schedulerTestSubscriberCalls
}

func (s *schedulerTestSubscriber) OnBeforeJobExecution(ctx *gotick.JobExecutionContext) {
	s.Called(ctx)
	s.calls.NumOfOnBeforeJobExecutionCalls++
}

func (s *schedulerTestSubscriber) OnBeforeJobExecutionPlan(ctx *gotick.JobExecutionContext) {
	s.Called(ctx)
	s.calls.NumOfOnBeforeJobExecutionPlanCalls++
}

func (s *schedulerTestSubscriber) OnJobExecuted(ctx *gotick.JobExecutionContext) {
	s.Called(ctx)
	s.calls.NumOfOnJobExecutedCalls++
}

func (s *schedulerTestSubscriber) OnJobExecutionDelayed(ctx *gotick.JobExecutionContext) {
	s.Called(ctx)
	s.calls.NumOfOnJobExecutionDelayedCalls++
}

func (s *schedulerTestSubscriber) OnJobExecutionInitiated(ctx *gotick.JobExecutionContext) {
	s.Called(ctx)
	s.calls.NumOfOnJobExecutionInitiatedCalls++
}

func (s *schedulerTestSubscriber) OnJobExecutionSkipped(ctx *gotick.JobExecutionContext) {
	s.Called(ctx)
	s.calls.NumOfOnJobExecutionSkippedCalls++
}

func (s *schedulerTestSubscriber) OnJobExecutionUnplanned(ctx *gotick.JobExecutionContext) {
	s.Called(ctx)
	s.calls.NumOfOnJobExecutionUnplannedCalls++
}

func (s *schedulerTestSubscriber) OnStart() {
	s.Called()
	s.calls.NumOfOnStartCalls++
}

func (s *schedulerTestSubscriber) OnStop() {
	s.Called()
	s.calls.NumOfOnStopCalls++
}

func (j *jobWithDelay) Execute(ctx *gotick.JobExecutionContext) {
	time.Sleep(j.delay)
	j.executions = append(j.executions, ctx)
}

func (j *jobWithDelay) ID() string {
	return j.id
}

var _ gotick.Job = (*jobWithDelay)(nil)
var _ gotick.SchedulerSubscriber = (*schedulerTestSubscriber)(nil)

func newJobWithDelay(id string, delay time.Duration) *jobWithDelay {
	return &jobWithDelay{
		id:    id,
		delay: delay,
	}
}

func newSchedulerTestSubscriber() *schedulerTestSubscriber {
	return &schedulerTestSubscriber{
		calls: &schedulerTestSubscriberCalls{},
	}
}

func TestJobShouldBeExecutedCorrectly(t *testing.T) {
	type jobFactory struct {
		job             gotick.Job
		scheduleFactory func() gotick.JobSchedule
	}

	data := []struct {
		name            string
		skip            bool // if you want to skip the test, set this to true
		jobs            []jobFactory
		schedulerConfig gotick.SchedulerConfig
		deadline        time.Duration
		assertion       func([]jobFactory, *schedulerTestSubscriber)
	}{
		{
			name: "single calendar job",
			jobs: []jobFactory{
				{
					job: newJobWithDelay("job1", 0),
					scheduleFactory: func() gotick.JobSchedule {
						return gotick.NewCalendarSchedule(time.Now().Add(1 * time.Second))
					},
				},
			},
			schedulerConfig: *gotick.DefaultSchedulerConfig(),
			deadline:        2 * time.Second,
			assertion: func(jf []jobFactory, s *schedulerTestSubscriber) {
				job := jf[0].job.(*jobWithDelay)
				assert.Len(t, job.executions, 1)

				assert.Equal(t, job.executions[0].Execution.PlannedAt, job.executions[0].Execution.Schedule.First())
				assert.LessOrEqual(t, job.executions[0].Execution.PlannedAt, job.executions[0].StartedAt)
				assert.LessOrEqual(t, job.executions[0].StartedAt, job.executions[0].ExecutedAt)
				assert.Equal(t, gotick.JobExecutionStatusExecuted, job.executions[0].ExecutionStatus)

				assert.Equal(t, 1, s.calls.NumOfOnBeforeJobExecutionCalls)
				assert.Equal(t, 1, s.calls.NumOfOnBeforeJobExecutionPlanCalls)
				assert.Equal(t, 1, s.calls.NumOfOnJobExecutedCalls)
				assert.Equal(t, 0, s.calls.NumOfOnJobExecutionDelayedCalls)
				assert.Equal(t, 1, s.calls.NumOfOnJobExecutionInitiatedCalls)
				assert.Equal(t, 0, s.calls.NumOfOnJobExecutionSkippedCalls)
				assert.Equal(t, 0, s.calls.NumOfOnJobExecutionUnplannedCalls)
				assert.Equal(t, 1, s.calls.NumOfOnStartCalls)
				assert.Equal(t, 1, s.calls.NumOfOnStopCalls)
			},
		},
		{
			name: "single cron job",
			jobs: []jobFactory{
				{
					job: newJobWithDelay("job1", 0),
					scheduleFactory: func() gotick.JobSchedule {
						c, err := gotick.NewCronSchedule("* * * * *")
						require.NoError(t, err)

						return c
					},
				},
			},
			schedulerConfig: *gotick.DefaultSchedulerConfig(
				gotick.WithIdlePollingInterval(1*time.Second),
				gotick.WithMaxPlanAhead(5*time.Second),
				gotick.WithInMemoryDriverFactory(
					gotick.DefaultInMemoryConfig(
						gotick.WithScheduleLockTimeout(1*time.Minute),
					),
				),
			),
			deadline: 1*time.Minute + 10*time.Second,
			assertion: func(jf []jobFactory, s *schedulerTestSubscriber) {
				job := jf[0].job.(*jobWithDelay)
				assert.LessOrEqual(t, 1, len(job.executions))

				plannedAt := make(map[time.Time]any)
				for _, execution := range job.executions {
					if _, ok := plannedAt[execution.Execution.PlannedAt]; ok {
						assert.Failf(t, "found two similar exeuctions at %s", execution.Execution.PlannedAt.Format(time.RFC3339))
					} else {
						plannedAt[execution.Execution.PlannedAt] = struct{}{}
					}
				}

				assert.LessOrEqual(t, 1, s.calls.NumOfOnBeforeJobExecutionCalls)
				assert.LessOrEqual(t, 1, s.calls.NumOfOnBeforeJobExecutionPlanCalls)
				assert.LessOrEqual(t, 1, s.calls.NumOfOnJobExecutedCalls)
				assert.Equal(t, 0, s.calls.NumOfOnJobExecutionDelayedCalls)
				assert.LessOrEqual(t, 1, s.calls.NumOfOnJobExecutionInitiatedCalls)
				assert.Equal(t, 0, s.calls.NumOfOnJobExecutionSkippedCalls)
				assert.Equal(t, 0, s.calls.NumOfOnJobExecutionUnplannedCalls)
				assert.Equal(t, 1, s.calls.NumOfOnStartCalls)
				assert.Equal(t, 1, s.calls.NumOfOnStopCalls)
			},
		},
		{
			name: "single seq job",
			jobs: []jobFactory{
				{
					job: newJobWithDelay("job1", 0),
					scheduleFactory: func() gotick.JobSchedule {
						c, err := gotick.NewSequenceSchedule(
							time.Now().Add(1*time.Second),
							time.Now().Add(2*time.Second),
							time.Now().Add(3*time.Second),
							time.Now().Add(5*time.Second),
							time.Now().Add(15*time.Second),
						)
						require.NoError(t, err)

						return c
					},
				},
			},
			schedulerConfig: *gotick.DefaultSchedulerConfig(),
			deadline:        30 * time.Second,
			assertion: func(jf []jobFactory, s *schedulerTestSubscriber) {
				job := jf[0].job.(*jobWithDelay)
				assert.Len(t, job.executions, 5)

				assert.Equal(t, 5, s.calls.NumOfOnBeforeJobExecutionCalls)
				assert.Equal(t, 5, s.calls.NumOfOnBeforeJobExecutionPlanCalls)
				assert.Equal(t, 5, s.calls.NumOfOnJobExecutedCalls)
				assert.Equal(t, 0, s.calls.NumOfOnJobExecutionDelayedCalls)
				assert.Equal(t, 5, s.calls.NumOfOnJobExecutionInitiatedCalls)
				assert.Equal(t, 0, s.calls.NumOfOnJobExecutionSkippedCalls)
				assert.Equal(t, 0, s.calls.NumOfOnJobExecutionUnplannedCalls)
				assert.Equal(t, 1, s.calls.NumOfOnStartCalls)
				assert.Equal(t, 1, s.calls.NumOfOnStopCalls)
			},
		},
		{
			name: "delayed job",
			jobs: []jobFactory{
				{
					job: newJobWithDelay("job1", 10*time.Second),
					scheduleFactory: func() gotick.JobSchedule {
						return gotick.NewCalendarSchedule(time.Now().Add(1 * time.Second))
					},
				},
				{
					job: newJobWithDelay("job2", 0),
					scheduleFactory: func() gotick.JobSchedule {
						return gotick.NewCalendarSchedule(time.Now().Add(2 * time.Second))
					},
				},
				{
					job: newJobWithDelay("job3", 0),
					scheduleFactory: func() gotick.JobSchedule {
						return gotick.NewCalendarScheduleWithMaxDelay(time.Now().Add(3*time.Second), 5*time.Second)
					},
				},
			},
			schedulerConfig: *gotick.DefaultSchedulerConfig(),
			deadline:        15 * time.Second,
			assertion: func(jf []jobFactory, s *schedulerTestSubscriber) {
				for _, j := range jf {
					job := j.job.(*jobWithDelay)
					assert.Len(t, job.executions, 1)
				}

				assert.Equal(t, 3, s.calls.NumOfOnBeforeJobExecutionCalls)
				assert.LessOrEqual(t, 3, s.calls.NumOfOnBeforeJobExecutionPlanCalls)
				assert.Equal(t, 3, s.calls.NumOfOnJobExecutedCalls)
				assert.LessOrEqual(t, 1, s.calls.NumOfOnJobExecutionDelayedCalls)
				assert.LessOrEqual(t, 3, s.calls.NumOfOnJobExecutionInitiatedCalls)
				assert.Equal(t, 0, s.calls.NumOfOnJobExecutionSkippedCalls)
				assert.LessOrEqual(t, 1, s.calls.NumOfOnJobExecutionUnplannedCalls)
				assert.Equal(t, 1, s.calls.NumOfOnStartCalls)
				assert.Equal(t, 1, s.calls.NumOfOnStopCalls)
			},
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			t.Parallel()
			if d.skip {
				t.SkipNow()
			}

			scheduler := gotick.NewScheduler(&d.schedulerConfig)

			subscriber := newSchedulerTestSubscriber()
			subscriber.On("OnBeforeJobExecution", mock.Anything).Return()
			subscriber.On("OnBeforeJobExecutionPlan", mock.Anything).Return()
			subscriber.On("OnJobExecuted", mock.Anything).Return()
			subscriber.On("OnJobExecutionDelayed", mock.Anything).Return()
			subscriber.On("OnJobExecutionInitiated", mock.Anything).Return()
			subscriber.On("OnJobExecutionSkipped", mock.Anything).Return()
			subscriber.On("OnJobExecutionUnplanned", mock.Anything).Return()
			subscriber.On("OnStart").Return()
			subscriber.On("OnStop").Return()

			scheduler.Subscribe(subscriber)

			ctx := context.Background()

			for _, job := range d.jobs {
				err := scheduler.RegisterJob(job.job)
				require.NoError(t, err)

				_, err = scheduler.ScheduleJob(ctx, job.job.ID(), job.scheduleFactory())
				require.NoError(t, err)
			}

			err := scheduler.Start(ctx)
			require.NoError(t, err)

			time.Sleep(d.deadline)
			scheduler.Stop()

			if d.assertion != nil {
				d.assertion(d.jobs, subscriber)
			}
		})
	}
}
