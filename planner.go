package gotick

import (
	"context"
	"sync"
	"time"
)

type planner struct {
	cfg         *PlannerConfig
	jobs        chan *JobExecutionContext
	subscribers []PlannerSubscriber
	startOnce   sync.Once
	stopOnce    sync.Once
}

func (p *planner) Subscribe(subscriber PlannerSubscriber) {
	p.subscribers = append(p.subscribers, subscriber)
}

func (p *planner) Plan(ctx *JobExecutionContext) {
	callSubscribersOnJobExecutionUnplanned := func() {
		ctx.ExecutionStatus = JobExecutionStatusUnplanned
		p.callSubscribers(func(s PlannerSubscriber) {
			s.OnJobExecutionUnplanned(ctx)
		})
	}

	select {
	case <-ctx.Done():
		callSubscribersOnJobExecutionUnplanned()
	case <-time.After(p.cfg.planTimeout):
		callSubscribersOnJobExecutionUnplanned()
	case p.jobs <- ctx:
	}
}

func (p *planner) Start(ctx context.Context) error {
	p.startOnce.Do(func() { p.start(ctx) })
	return nil
}

func (p *planner) Stop() error {
	p.stopOnce.Do(p.stop)
	return nil
}

func (p *planner) executor(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case job := <-p.jobs:
			if job == nil {
				return
			}

			job.ExecutionStatus = JobExecutionStatusPlanned

			if time.Until(job.Execution.PlannedAt) > 0 {
				select {
				case <-ctx.Done():
					return
				case <-time.After(time.Until(job.Execution.PlannedAt)):
				}
			}

			job.StartedAt = time.Now()

			p.callSubscribers(func(s PlannerSubscriber) {
				s.OnBeforeJobExecution(job)
			})

			job.ExecutionStatus = JobExecutionStatusExecuting

			job.Execution.Job.Execute(job)

			job.ExecutionStatus = JobExecutionStatusExecuted
			job.ExecutedAt = time.Now()

			p.callSubscribers(func(s PlannerSubscriber) {
				s.OnJobExecuted(job)
			})
		}
	}
}

func (p *planner) callSubscribers(callback func(PlannerSubscriber)) {
	for _, subscriber := range p.subscribers {
		callback(subscriber)
	}
}

func (p *planner) start(ctx context.Context) {
	for range p.cfg.threads {
		go p.executor(ctx)
	}
}

func (p *planner) stop() {
	close(p.jobs)
}

func newPlanner(cfg *PlannerConfig) Planner {
	return &planner{
		cfg:         cfg,
		jobs:        make(chan *JobExecutionContext, cfg.threads),
		subscribers: make([]PlannerSubscriber, 0),
	}
}
