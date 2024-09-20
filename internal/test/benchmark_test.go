package test

import (
	"context"
	"testing"
	"time"

	gotick "github.com/go-tick/core"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type benchmarkJob struct {
	id   gotick.JobID
	done chan any
}

type benchmarkJobFactory struct {
	job *benchmarkJob
}

func (b *benchmarkJob) Execute(*gotick.JobExecutionContext) {
	close(b.done)
}

func (b *benchmarkJobFactory) Create(jobID gotick.JobID) gotick.Job {
	return b.job
}

var _ gotick.JobFactory = (*benchmarkJobFactory)(nil)
var _ gotick.Job = (*benchmarkJob)(nil)

func newBenchmarkJob(id gotick.JobID) *benchmarkJob {
	return &benchmarkJob{id, make(chan any)}
}

func BenchmarkJobBetweenScheduleAndExecution(b *testing.B) {
	job := newBenchmarkJob(gotick.JobID(uuid.NewString()))
	jf := &benchmarkJobFactory{job}

	plannerCfg := gotick.DefaultPlannerConfig(gotick.WithJobFactory(jf))
	schedulerCfg := gotick.DefaultSchedulerConfig(
		gotick.WithIdlePollingInterval(0),
		gotick.WithDefaultPlannerFactory(plannerCfg),
	)

	scheduler, err := gotick.NewScheduler(schedulerCfg)
	require.NoError(b, err)

	err = scheduler.Start(context.Background())
	require.NoError(b, err)
	defer func() {
		err := scheduler.Stop()
		require.NoError(b, err)
	}()

	b.ResetTimer()
	for range b.N {
		schedule := gotick.NewCalendarSchedule(time.Now())

		_, err = scheduler.ScheduleJob(context.Background(), job.id, schedule)
		require.NoError(b, err)

		<-job.done
		job.done = make(chan any)
	}
}
