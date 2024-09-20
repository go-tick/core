package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"time"

	gotick "github.com/go-tick/core"
	"github.com/google/uuid"
	"golang.org/x/exp/rand"
)

type randomDelayJob struct {
	logger   *slog.Logger
	callback func()
}

type randomDelayJobFactory struct {
	jobs map[string]*randomDelayJob
}

type schedulerObserver struct {
	logger *slog.Logger
}

func (r *schedulerObserver) OnBeforeJobExecution(ctx *gotick.JobExecutionContext) {
	r.logger.Info(
		"on before job execution",
		slog.Any("ctx", ctx),
	)
}

func (r *schedulerObserver) OnBeforeJobExecutionPlan(ctx *gotick.JobExecutionContext) {
	r.logger.Info(
		"on before job execution planned",
		slog.Any("ctx", ctx),
	)
}

func (r *schedulerObserver) OnJobExecuted(ctx *gotick.JobExecutionContext) {
	r.logger.Info(
		"on job executed",
		slog.Any("ctx", ctx),
	)
}

func (r *schedulerObserver) OnJobExecutionDelayed(ctx *gotick.JobExecutionContext) {
	r.logger.Info(
		"on job execution delayed",
		slog.Any("ctx", ctx),
	)
}

func (r *schedulerObserver) OnJobExecutionInitiated(ctx *gotick.JobExecutionContext) {
	r.logger.Info(
		"on job execution initiated",
		slog.Any("ctx", ctx),
	)
}

func (r *schedulerObserver) OnJobExecutionUnplanned(ctx *gotick.JobExecutionContext) {
	r.logger.Info(
		"on job execution unplanned",
		slog.Any("ctx", ctx),
	)
}

func (r *schedulerObserver) OnJobExecutionSkipped(ctx *gotick.JobExecutionContext) {
	r.logger.Info(
		"on job execution skipped",
		slog.Any("ctx", ctx),
	)
}

func (r *schedulerObserver) OnStart() {
	r.logger.Info("on started")
}

func (r *schedulerObserver) OnStop() {
	r.logger.Info("on stopped")
}

func (r *randomDelayJob) Execute(ctx *gotick.JobExecutionContext) {
	rnd := rand.New(rand.NewSource(uint64(time.Now().UnixNano())))
	delay := time.Duration(rnd.Intn(10000)) * time.Millisecond

	r.logger.Info(
		"job execution started",
		slog.Any("ctx", ctx),
		slog.Duration("delay", delay),
	)

	time.Sleep(delay)

	r.logger.Info(
		"job executed",
		slog.Any("ctx", ctx),
	)

	r.callback()
}

func (r *randomDelayJobFactory) Create(jobID string) gotick.Job {
	return r.jobs[jobID]
}

var _ gotick.Job = (*randomDelayJob)(nil)
var _ gotick.JobFactory = (*randomDelayJobFactory)(nil)
var _ gotick.SchedulerObserver = (*schedulerObserver)(nil)

func main() {
	const numOfJobs = 20

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	panicOnErr := func(err error) {
		if err != nil {
			panic(err)
		}
	}

	f, err := os.Create("logs.txt")
	panicOnErr(err)
	defer f.Close()

	var wg sync.WaitGroup
	done := make(chan struct{})

	wg.Add(numOfJobs)

	go func() {
		wg.Wait()
		close(done)
	}()

	logger := slog.New(slog.NewTextHandler(f, nil))

	job := &randomDelayJob{logger, wg.Done}
	jf := &randomDelayJobFactory{make(map[string]*randomDelayJob)}

	jobID := uuid.NewString()
	jf.jobs[jobID] = job

	subscriber := &schedulerObserver{logger}

	plannerCfg := gotick.DefaultPlannerConfig(
		gotick.WithPlannerThreads(2),
		gotick.WithPlannerTimeout(3*time.Second),
		gotick.WithJobFactory(jf),
	)

	driverCfg := gotick.DefaultInMemoryConfig(gotick.WithScheduleLockTimeout(20 * time.Second))

	schedulerCfg := gotick.DefaultSchedulerConfig(
		gotick.WithThreads(4),
		gotick.WithIdlePollingInterval(2*time.Second),
		gotick.WithMaxPlanAhead(5*time.Second),
		gotick.WithSubscribers(subscriber),
		gotick.WithDelayedStrategy(gotick.ScheduleDelayedStrategyPlan),
		gotick.WithDefaultPlannerFactory(plannerCfg),
		gotick.WithInMemoryDriverFactory(driverCfg),
	)

	scheduler, err := gotick.NewScheduler(schedulerCfg)
	panicOnErr(err)
	defer scheduler.Stop()

	for i := range numOfJobs {
		sch := gotick.NewJobScheduleWithMaxDelay(
			gotick.NewCalendarSchedule(time.Now().Add(time.Duration(i+1)*time.Second)),
			5*time.Second,
		)

		scheduler.ScheduleJob(ctx, jobID, sch)
		panicOnErr(err)
	}

	err = scheduler.Start(ctx)
	panicOnErr(err)

	select {
	case <-done:
	case <-ctx.Done():
	}
}
