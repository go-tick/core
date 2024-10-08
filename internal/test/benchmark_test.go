package test

import (
	"context"
	"testing"
	"time"

	"github.com/go-tick/core"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type benchmarkJob struct {
	id   string
	done chan any
}

func (b *benchmarkJob) Execute(*gotick.JobExecutionContext) {
	close(b.done)
}

func (b *benchmarkJob) ID() string {
	return b.id
}

var _ gotick.Job = (*benchmarkJob)(nil)

func newBenchmarkJob(id string) *benchmarkJob {
	return &benchmarkJob{id, make(chan any)}
}

func BenchmarkJobBetweenScheduleAndExecution(b *testing.B) {
	scheduler := gotick.NewScheduler(gotick.DefaultSchedulerConfig(
		gotick.WithIdlePollingInterval(0),
	))
	err := scheduler.Start(context.Background())
	require.NoError(b, err)
	defer func() {
		err := scheduler.Stop()
		require.NoError(b, err)
	}()

	b.ResetTimer()
	for range b.N {
		job := newBenchmarkJob(uuid.NewString())
		schedule := gotick.NewCalendarSchedule(time.Now())

		err = scheduler.RegisterJob(job)
		require.NoError(b, err)

		_, err = scheduler.ScheduleJob(context.Background(), job.ID(), schedule)
		require.NoError(b, err)

		<-job.done
	}
}
