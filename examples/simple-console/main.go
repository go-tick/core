package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/misikdmytro/gotick"
	"golang.org/x/exp/rand"
)

type randomDelayJob struct {
	logger   *slog.Logger
	callback func()
}

type schedulerSubscriber struct {
	logger *slog.Logger
}

func (r *schedulerSubscriber) OnBeforeJobExecution(ctx *gotick.JobExecutionContext) {
	r.logger.Info(
		"on before job execution",
		slog.Any("ctx", ctx),
	)
}

func (r *schedulerSubscriber) OnBeforeJobExecutionPlanned(ctx *gotick.JobExecutionContext) {
	r.logger.Info(
		"on before job execution planned",
		slog.Any("ctx", ctx),
	)
}

func (r *schedulerSubscriber) OnJobExecuted(ctx *gotick.JobExecutionContext) {
	r.logger.Info(
		"on job executed",
		slog.Any("ctx", ctx),
	)
}

func (r *schedulerSubscriber) OnJobExecutionDelayed(ctx *gotick.JobExecutionContext) {
	r.logger.Info(
		"on job execution delayed",
		slog.Any("ctx", ctx),
	)
}

func (r *schedulerSubscriber) OnJobExecutionInitiated(ctx *gotick.JobExecutionContext) {
	r.logger.Info(
		"on job execution initiated",
		slog.Any("ctx", ctx),
	)
}

func (r *schedulerSubscriber) OnJobExecutionNotPlanned(ctx *gotick.JobExecutionContext) {
	r.logger.Info(
		"on job execution not planned",
		slog.Any("ctx", ctx),
	)
}

func (r *schedulerSubscriber) OnJobExecutionSkipped(ctx *gotick.JobExecutionContext) {
	r.logger.Info(
		"on job execution skipped",
		slog.Any("ctx", ctx),
	)
}

func (r *schedulerSubscriber) OnStart() {
	r.logger.Info("on started")
}

func (r *schedulerSubscriber) OnStop() {
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

func (r *randomDelayJob) ID() string {
	return "unique-id"
}

var _ gotick.Job = &randomDelayJob{}
var _ gotick.SchedulerSubscriber = &schedulerSubscriber{}

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
	subscriber := &schedulerSubscriber{logger}

	cfg := gotick.DefaultConfig(
		gotick.WithThreads(4),
		gotick.WithDefaultPlannerFactory(4),
		gotick.WithIdlePollingInterval(2*time.Second),
		gotick.WithMaxPlanAhead(5*time.Second),
		gotick.WithSubscribers(subscriber),
		gotick.WithDelayedStrategy(gotick.ScheduleDelayedStrategyPlan),
	)

	scheduler := gotick.NewScheduler(cfg)
	defer scheduler.Stop()

	err = scheduler.RegisterJob(job)
	panicOnErr(err)

	for i := range numOfJobs {
		sch := gotick.NewCalendarWithMaxDelay(time.Now().Add(time.Duration(i+1)*time.Second), 5*time.Second)

		scheduler.ScheduleJob(ctx, job.ID(), sch)
		panicOnErr(err)
	}

	err = scheduler.Start(ctx)
	panicOnErr(err)

	select {
	case <-done:
	case <-ctx.Done():
	}
}
