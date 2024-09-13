package gotick

import (
	"context"
	"sync"
	"time"
)

type planner struct {
	threads     uint
	jobs        chan *JobExecutionContext
	subscribers []PlannerSubscriber
	startOnce   sync.Once
	stopOnce    sync.Once
}

func (p *planner) Subscribe(subscriber PlannerSubscriber) {
	p.subscribers = append(p.subscribers, subscriber)
}

func (p *planner) Plan(ctx *JobExecutionContext) {
	select {
	case <-ctx.Done():
		p.callSubscribers(func(s PlannerSubscriber) {
			s.OnJobExecutionNotPlanned(ctx.Clone())
		})
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
				s.OnBeforeJobExecution(job.Clone())
			})

			job.ExecutionStatus = JobExecutionStatusExecuting

			job.Execution.Job.Execute(job.Clone())

			job.ExecutionStatus = JobExecutionStatusExecuted
			job.ExecutedAt = time.Now()

			p.callSubscribers(func(s PlannerSubscriber) {
				s.OnJobExecuted(job.Clone())
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
	for range p.threads {
		go p.executor(ctx)
	}
}

func (p *planner) stop() {
	close(p.jobs)
}

func newPlanner(threads uint) Planner {
	return &planner{
		jobs:        make(chan *JobExecutionContext, threads),
		threads:     threads,
		subscribers: make([]PlannerSubscriber, 0),
	}
}
