package gotick

import (
	"context"
	"sync"
	"time"
)

type planner struct {
	threads     uint
	jobs        chan *JobContext
	subscribers []PlannerSubscriber
	startOnce   sync.Once
	stopOnce    sync.Once
}

func (p *planner) Subscribe(subscriber PlannerSubscriber) {
	p.subscribers = append(p.subscribers, subscriber)
}

func (p *planner) Plan(ctx *JobContext) (res error) {
	defer func() {
		if err := recover(); err != nil {
			res = err.(error)
		}
	}()

	select {
	case p.jobs <- ctx:
		return nil
	case <-ctx.Done():
		return ctx.Err()
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
			job.ExecutionStatus = JobExecutionStatusPlanned

			if time.Until(job.PlannedAt) > 0 {
				select {
				case <-ctx.Done():
					return
				case <-time.After(time.Until(job.PlannedAt)):
				}
			}

			job.StartedAt = time.Now()
			job.ExecutionStatus = JobExecutionStatusExecuting

			p.callSubscribers(func(s PlannerSubscriber) {
				s.OnBeforeJobExecution(job.Clone())
			})

			job.Job.Execute(job.Clone())
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

func NewPlanner(threads uint) Planner {
	return &planner{
		jobs:        make(chan *JobContext, threads),
		threads:     threads,
		subscribers: make([]PlannerSubscriber, 0),
	}
}
