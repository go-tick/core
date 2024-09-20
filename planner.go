package gotick

import (
	"context"
	"sync"
	"time"
)

type planner struct {
	cfg         *PlannerConfig
	executions  chan *JobExecutionContext
	subscribers []PlannerObserver
	startOnce   sync.Once
	stopOnce    sync.Once
}

func (p *planner) Subscribe(subscriber PlannerObserver) {
	p.subscribers = append(p.subscribers, subscriber)
}

func (p *planner) Plan(ctx *JobExecutionContext) {
	callSubscribersOnJobExecutionUnplanned := func() {
		ctx.ExecutionStatus = JobExecutionStatusUnplanned
		p.callSubscribers(func(s PlannerObserver) {
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
		case exec := <-p.executions:
			if exec == nil {
				return
			}

			exec.ExecutionStatus = JobExecutionStatusPlanned

			if time.Until(exec.PlannedAt) > 0 {
				select {
				case <-ctx.Done():
					return
				case <-time.After(time.Until(exec.PlannedAt)):
				}
			}

			exec.StartedAt = time.Now()

			p.callSubscribers(func(s PlannerObserver) {
				s.OnBeforeJobExecution(exec)
			})

			exec.ExecutionStatus = JobExecutionStatusExecuting

			job := p.cfg.jobFactory.Create(exec.JobID)
			job.Execute(exec)

			exec.ExecutionStatus = JobExecutionStatusExecuted
			exec.ExecutedAt = time.Now()

			p.callSubscribers(func(s PlannerObserver) {
				s.OnJobExecuted(exec)
			})
		}
	}
}

func (p *planner) callSubscribers(callback func(PlannerObserver)) {
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
	return &planner{
		cfg:         cfg,
		executions:  make(chan *JobExecutionContext, cfg.threads),
		subscribers: make([]PlannerObserver, 0),
	}, nil
}
