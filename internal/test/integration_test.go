package test

import (
	"context"
	"sync"
	"testing"
	"time"

	gotick "github.com/go-tick/core"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type jobWithDelay struct {
	id         string
	delay      time.Duration
	executions []*gotick.JobExecutionContext
	once       sync.Once
	done       chan any
}

type jobFactory struct {
	jobs []*jobWithDelay
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
	j.once.Do(func() { close(j.done) })
}

func (j *jobFactory) Create(jobID string) gotick.Job {
	for _, job := range j.jobs {
		if job.id == jobID {
			return job
		}
	}

	return nil
}

var _ gotick.Job = (*jobWithDelay)(nil)
var _ gotick.JobFactory = (*jobFactory)(nil)
var _ gotick.SchedulerObserver = (*schedulerTestSubscriber)(nil)

func newJobWithDelay(id string, delay time.Duration) *jobWithDelay {
	return &jobWithDelay{
		id:    id,
		delay: delay,
		done:  make(chan any),
	}
}

func newSchedulerTestSubscriber() *schedulerTestSubscriber {
	return &schedulerTestSubscriber{
		calls: &schedulerTestSubscriberCalls{},
	}
}

func TestJobShouldBeExecutedCorrectly(t *testing.T) {
	type testJobs struct {
		job             *jobWithDelay
		scheduleFactory func() gotick.JobSchedule
	}

	data := []struct {
		name            string
		skip            bool // if you want to skip the test, set this to true
		jobs            []testJobs
		plannerCfg      func([]*jobWithDelay) *gotick.PlannerConfig
		schedulerConfig func(*gotick.PlannerConfig) *gotick.SchedulerConfig
		deadline        time.Duration
		assertion       func([]testJobs, *schedulerTestSubscriber)
	}{
		{
			name: "single calendar job",
			jobs: []testJobs{
				{
					job: newJobWithDelay("job1", 0),
					scheduleFactory: func() gotick.JobSchedule {
						return gotick.NewCalendarSchedule(time.Now().Add(1 * time.Second))
					},
				},
			},
			plannerCfg: func(j []*jobWithDelay) *gotick.PlannerConfig {
				return gotick.DefaultPlannerConfig(gotick.WithJobFactory(&jobFactory{j}))
			},
			schedulerConfig: func(pc *gotick.PlannerConfig) *gotick.SchedulerConfig {
				return gotick.DefaultSchedulerConfig(gotick.WithDefaultPlannerFactory(pc))
			},
			deadline: 2 * time.Second,
			assertion: func(jf []testJobs, s *schedulerTestSubscriber) {
				// the job should be executed once at a specific time
				job := jf[0].job
				assert.Len(t, job.executions, 1)

				assert.LessOrEqual(t, job.executions[0].PlannedAt, job.executions[0].StartedAt)
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
			jobs: []testJobs{
				{
					job: newJobWithDelay("job1", 0),
					scheduleFactory: func() gotick.JobSchedule {
						c, err := gotick.NewCronSchedule("* * * * *")
						require.NoError(t, err)

						return c
					},
				},
			},
			plannerCfg: func(j []*jobWithDelay) *gotick.PlannerConfig {
				return gotick.DefaultPlannerConfig(gotick.WithJobFactory(&jobFactory{j}))
			},
			schedulerConfig: func(pc *gotick.PlannerConfig) *gotick.SchedulerConfig {
				return gotick.DefaultSchedulerConfig(
					gotick.WithIdlePollingInterval(1*time.Second),
					gotick.WithMaxPlanAhead(5*time.Second),
					gotick.WithInMemoryDriverFactory(
						gotick.DefaultInMemoryConfig(
							gotick.WithScheduleLockTimeout(1*time.Minute),
						),
					),
					gotick.WithDefaultPlannerFactory(pc),
				)
			},
			deadline: 1*time.Minute + 10*time.Second,
			assertion: func(jf []testJobs, s *schedulerTestSubscriber) {
				// the job should be executed at least once
				job := jf[0].job
				assert.LessOrEqual(t, 1, len(job.executions))

				plannedAt := make(map[time.Time]any)
				for _, execution := range job.executions {
					if _, ok := plannedAt[execution.PlannedAt]; ok {
						assert.Failf(t, "found two similar executions at %s", execution.PlannedAt.Format(time.RFC3339Nano))
					} else {
						plannedAt[execution.PlannedAt] = struct{}{}
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
			jobs: []testJobs{
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
			plannerCfg: func(j []*jobWithDelay) *gotick.PlannerConfig {
				return gotick.DefaultPlannerConfig(gotick.WithJobFactory(&jobFactory{j}))
			},
			schedulerConfig: func(pc *gotick.PlannerConfig) *gotick.SchedulerConfig {
				return gotick.DefaultSchedulerConfig(gotick.WithDefaultPlannerFactory(pc))
			},
			deadline: 30 * time.Second,
			assertion: func(jf []testJobs, s *schedulerTestSubscriber) {
				// the job should be executed in a sequence 5 times
				job := jf[0].job
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
			jobs: []testJobs{
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
						return gotick.NewJobScheduleWithMaxDelay(gotick.NewCalendarSchedule(time.Now().Add(3*time.Second)), 1*time.Second)
					},
				},
			},
			plannerCfg: func(j []*jobWithDelay) *gotick.PlannerConfig {
				return gotick.DefaultPlannerConfig(gotick.WithJobFactory(&jobFactory{j}))
			},
			schedulerConfig: func(pc *gotick.PlannerConfig) *gotick.SchedulerConfig {
				return gotick.DefaultSchedulerConfig(gotick.WithDefaultPlannerFactory(pc))
			},
			deadline: 15 * time.Second,
			assertion: func(jf []testJobs, s *schedulerTestSubscriber) {
				// the first job should be planned and executed after 10 seconds.
				// for all this time, this job will occupy the thread.
				// after 1 second, the second job should be planned for execution.
				// but because by default planner uses 1 thread, this thread is busy with the first job.
				// the third job should be still planned for execution, but later with delay.
				for _, j := range jf {
					job := j.job
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
		{
			name: "skipped job",
			jobs: []testJobs{
				{
					job: newJobWithDelay("job1", 5*time.Second),
					scheduleFactory: func() gotick.JobSchedule {
						return gotick.NewCalendarSchedule(time.Now().Add(1 * time.Second))
					},
				},
				{
					job: newJobWithDelay("job2", 5*time.Second),
					scheduleFactory: func() gotick.JobSchedule {
						return gotick.NewCalendarSchedule(time.Now().Add(2 * time.Second))
					},
				},
				{
					job: newJobWithDelay("job3", 0),
					scheduleFactory: func() gotick.JobSchedule {
						return gotick.NewJobScheduleWithMaxDelay(gotick.NewCalendarSchedule(time.Now().Add(3*time.Second)), 0)
					},
				},
			},
			plannerCfg: func(j []*jobWithDelay) *gotick.PlannerConfig {
				return gotick.DefaultPlannerConfig(
					gotick.WithPlannerTimeout(1*time.Second),
					gotick.WithJobFactory(&jobFactory{j}),
				)
			},
			schedulerConfig: func(pc *gotick.PlannerConfig) *gotick.SchedulerConfig {
				return gotick.DefaultSchedulerConfig(
					gotick.WithDelayedStrategy(gotick.ScheduleDelayedStrategySkip),
					gotick.WithDefaultPlannerFactory(pc),
				)
			},
			deadline: 5 * time.Second,
			assertion: func(jf []testJobs, s *schedulerTestSubscriber) {
				// the first job should be planned and executed after 5 seconds.
				// for all this time, this job will occupy the thread.
				// after 1 second, the second job should be planned for execution.
				// but because by default planner uses 1 thread, this thread is busy with the first job.
				// the third job should be skipped in 1 second as planner will not be able to plan it.
				job3 := jf[2].job
				assert.Len(t, job3.executions, 0)

				assert.GreaterOrEqual(t, 2, s.calls.NumOfOnBeforeJobExecutionCalls)
				assert.LessOrEqual(t, 2, s.calls.NumOfOnBeforeJobExecutionPlanCalls)
				assert.GreaterOrEqual(t, 2, s.calls.NumOfOnJobExecutedCalls)
				assert.Equal(t, 1, s.calls.NumOfOnJobExecutionDelayedCalls)
				assert.LessOrEqual(t, 3, s.calls.NumOfOnJobExecutionInitiatedCalls)
				assert.Equal(t, 1, s.calls.NumOfOnJobExecutionSkippedCalls)
				assert.LessOrEqual(t, 0, s.calls.NumOfOnJobExecutionUnplannedCalls)
				assert.Equal(t, 1, s.calls.NumOfOnStartCalls)
				assert.Equal(t, 1, s.calls.NumOfOnStopCalls)
			},
		},
		{
			name: "several planner threads",
			jobs: []testJobs{
				{
					job: newJobWithDelay("job1", 1*time.Second),
					scheduleFactory: func() gotick.JobSchedule {
						return gotick.NewCalendarSchedule(time.Now())
					},
				},
				{
					job: newJobWithDelay("job2", 1*time.Second),
					scheduleFactory: func() gotick.JobSchedule {
						return gotick.NewCalendarSchedule(time.Now())
					},
				},
				{
					job: newJobWithDelay("job3", 1*time.Second),
					scheduleFactory: func() gotick.JobSchedule {
						return gotick.NewCalendarSchedule(time.Now())
					},
				},
				{
					job: newJobWithDelay("job4", 1*time.Second),
					scheduleFactory: func() gotick.JobSchedule {
						return gotick.NewCalendarSchedule(time.Now())
					},
				},
			},
			plannerCfg: func(j []*jobWithDelay) *gotick.PlannerConfig {
				return gotick.DefaultPlannerConfig(
					gotick.WithPlannerThreads(4),
					gotick.WithJobFactory(&jobFactory{j}),
				)
			},
			schedulerConfig: func(pc *gotick.PlannerConfig) *gotick.SchedulerConfig {
				return gotick.DefaultSchedulerConfig(gotick.WithDefaultPlannerFactory(pc))
			},
			deadline: 3 * time.Second,
			assertion: func(jf []testJobs, s *schedulerTestSubscriber) {
				// all jobs should be executed as planner has 4 threads
				for _, j := range jf {
					job := j.job
					assert.Len(t, job.executions, 1)
				}

				assert.Equal(t, 4, s.calls.NumOfOnBeforeJobExecutionCalls)
				assert.Equal(t, 4, s.calls.NumOfOnBeforeJobExecutionPlanCalls)
				assert.Equal(t, 4, s.calls.NumOfOnJobExecutedCalls)
				assert.Equal(t, 0, s.calls.NumOfOnJobExecutionDelayedCalls)
				assert.Equal(t, 4, s.calls.NumOfOnJobExecutionInitiatedCalls)
				assert.Equal(t, 0, s.calls.NumOfOnJobExecutionSkippedCalls)
				assert.Equal(t, 0, s.calls.NumOfOnJobExecutionUnplannedCalls)
				assert.Equal(t, 1, s.calls.NumOfOnStartCalls)
				assert.Equal(t, 1, s.calls.NumOfOnStopCalls)
			},
		},
		{
			name: "several scheduler threads",
			jobs: []testJobs{
				{
					job: newJobWithDelay("job1", 1*time.Second),
					scheduleFactory: func() gotick.JobSchedule {
						return gotick.NewCalendarSchedule(time.Now())
					},
				},
				{
					job: newJobWithDelay("job2", 1*time.Second),
					scheduleFactory: func() gotick.JobSchedule {
						return gotick.NewCalendarSchedule(time.Now())
					},
				},
				{
					job: newJobWithDelay("job3", 1*time.Second),
					scheduleFactory: func() gotick.JobSchedule {
						return gotick.NewCalendarSchedule(time.Now())
					},
				},
				{
					job: newJobWithDelay("job4", 1*time.Second),
					scheduleFactory: func() gotick.JobSchedule {
						return gotick.NewCalendarSchedule(time.Now())
					},
				},
			},
			plannerCfg: func(j []*jobWithDelay) *gotick.PlannerConfig {
				return gotick.DefaultPlannerConfig(
					gotick.WithPlannerThreads(4),
					gotick.WithJobFactory(&jobFactory{j}),
				)
			},
			schedulerConfig: func(pc *gotick.PlannerConfig) *gotick.SchedulerConfig {
				return gotick.DefaultSchedulerConfig(
					gotick.WithThreads(4),
					gotick.WithDefaultPlannerFactory(pc),
				)
			},
			deadline: 3 * time.Second,
			assertion: func(jf []testJobs, s *schedulerTestSubscriber) {
				// all jobs should be executed exactly once
				for _, j := range jf {
					job := j.job
					assert.Len(t, job.executions, 1)
				}

				assert.Equal(t, 4, s.calls.NumOfOnBeforeJobExecutionCalls)
				assert.Equal(t, 4, s.calls.NumOfOnBeforeJobExecutionPlanCalls)
				assert.Equal(t, 4, s.calls.NumOfOnJobExecutedCalls)
				assert.Equal(t, 0, s.calls.NumOfOnJobExecutionDelayedCalls)
				assert.Equal(t, 4, s.calls.NumOfOnJobExecutionInitiatedCalls)
				assert.Equal(t, 0, s.calls.NumOfOnJobExecutionSkippedCalls)
				assert.Equal(t, 0, s.calls.NumOfOnJobExecutionUnplannedCalls)
				assert.Equal(t, 1, s.calls.NumOfOnStartCalls)
				assert.Equal(t, 1, s.calls.NumOfOnStopCalls)
			},
		},
		{
			name: "bad job timeout",
			jobs: []testJobs{
				{
					job: newJobWithDelay("job1", 3*time.Second),
					scheduleFactory: func() gotick.JobSchedule {
						return gotick.NewJobScheduleWithTimeout(gotick.NewCalendarSchedule(time.Now()), 1*time.Second)
					},
				},
			},
			plannerCfg: func(j []*jobWithDelay) *gotick.PlannerConfig {
				return gotick.DefaultPlannerConfig(
					gotick.WithPlannerThreads(2),
					gotick.WithJobFactory(&jobFactory{j}),
				)
			},
			schedulerConfig: func(pc *gotick.PlannerConfig) *gotick.SchedulerConfig {
				return gotick.DefaultSchedulerConfig(gotick.WithDefaultPlannerFactory(pc))
			},
			deadline: 6 * time.Second,
			assertion: func(jf []testJobs, s *schedulerTestSubscriber) {
				// because of the wrong timeout, the job will be executed several times
				for _, j := range jf {
					job := j.job
					assert.Less(t, 1, len(job.executions))
				}

				assert.LessOrEqual(t, 2, s.calls.NumOfOnBeforeJobExecutionCalls)
				assert.LessOrEqual(t, 2, s.calls.NumOfOnBeforeJobExecutionPlanCalls)
				assert.LessOrEqual(t, 1, s.calls.NumOfOnJobExecutedCalls)
				assert.Equal(t, 0, s.calls.NumOfOnJobExecutionDelayedCalls)
				assert.LessOrEqual(t, 2, s.calls.NumOfOnJobExecutionInitiatedCalls)
				assert.Equal(t, 0, s.calls.NumOfOnJobExecutionSkippedCalls)
				assert.Equal(t, 0, s.calls.NumOfOnJobExecutionUnplannedCalls)
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

			jobs := make([]*jobWithDelay, 0, len(d.jobs))
			for _, job := range d.jobs {
				jobs = append(jobs, job.job)
			}

			planerCfg := d.plannerCfg(jobs)
			schedulerCfg := d.schedulerConfig(planerCfg)

			scheduler, err := gotick.NewScheduler(schedulerCfg)
			require.NoError(t, err)

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
				_, err := scheduler.ScheduleJob(ctx, job.job.id, job.scheduleFactory())
				require.NoError(t, err)
			}

			err = scheduler.Start(ctx)
			require.NoError(t, err)

			time.Sleep(d.deadline)

			err = scheduler.Stop()
			require.NoError(t, err)

			if d.assertion != nil {
				d.assertion(d.jobs, subscriber)
			}
		})
	}
}

func TestJobShouldBeExecutedExactlyOnce(t *testing.T) {
	const iterations = 1000

	plannerCfg := gotick.DefaultPlannerConfig(gotick.WithPlannerThreads(16))

	jobs := make([]*jobWithDelay, iterations)
	jf := &jobFactory{jobs: jobs}
	for i := range iterations {
		jobs[i] = newJobWithDelay(uuid.NewString(), 0)
		gotick.WithJobFactory(jf)(plannerCfg)
	}

	scheduler, err := gotick.NewScheduler(gotick.DefaultSchedulerConfig(
		gotick.WithIdlePollingInterval(0),
		gotick.WithThreads(16),
		gotick.WithDefaultPlannerFactory(plannerCfg),
	))
	require.NoError(t, err)

	defer func() {
		err := scheduler.Stop()
		require.NoError(t, err)
	}()

	for i := range iterations {
		_, err := scheduler.ScheduleJob(context.Background(), jobs[i].id, gotick.NewCalendarSchedule(time.Now()))
		require.NoError(t, err)
	}

	err = scheduler.Start(context.Background())
	require.NoError(t, err)

	for i := range iterations {
		<-jobs[i].done
	}

	for _, job := range jobs {
		assert.Len(t, job.executions, 1)
	}
}
