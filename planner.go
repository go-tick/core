package gotick

import (
	"context"
	"sync"
	"time"
)

type planner struct {
	jobs        map[string]Job
	cfg         *PlannerConfig
	executions  chan *JobExecutionContext
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
	case p.executions <- ctx:
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
		case job := <-p.executions:
			if job == nil {
				return
			}

			job.ExecutionStatus = JobExecutionStatusPlanned

			if time.Until(job.PlannedAt) > 0 {
				select {
				case <-ctx.Done():
					return
				case <-time.After(time.Until(job.PlannedAt)):
				}
			}

			job.StartedAt = time.Now()

			p.callSubscribers(func(s PlannerSubscriber) {
				s.OnBeforeJobExecution(job)
			})

			job.ExecutionStatus = JobExecutionStatusExecuting

			p.jobs[job.JobID].Execute(job)

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
	close(p.executions)
}

func newPlanner(cfg *PlannerConfig) (Planner, error) {
	jr := make(map[string]Job)
	for _, job := range cfg.jobs {
		jobID := job.ID()
		if _, ok := jr[jobID]; ok {
			return nil, ErrDuplicateJobID
		}

		jr[job.ID()] = job
	}

	return &planner{
		jobs:        jr,
		cfg:         cfg,
		executions:  make(chan *JobExecutionContext, cfg.threads),
		subscribers: make([]PlannerSubscriber, 0),
	}, nil
}
