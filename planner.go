package gotick

import (
	"context"
	"time"
)

type planner struct {
	threads     uint
	jobs        chan *JobContext
	errs        chan error
	subscribers []PlannerSubscriber
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

func (p *planner) Stop() error {
	close(p.jobs)

	return nil
}

func (p *planner) Start(ctx context.Context) error {
	for range p.threads {
		go p.executor(ctx)
	}

	go p.errsListener(ctx)

	return nil
}

func (p *planner) errsListener(ctx context.Context) {
	for {
		select {
		case err := <-p.errs:
			for _, s := range p.subscribers {
				go s.OnError(err)
			}
		case <-ctx.Done():
			return
		}
	}
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
			err := job.Job.Execute(job)
			if err != nil {
				p.errs <- err
				job.ExecutionStatus = JobExecutionStatusFailed
			} else {
				job.ExecutionStatus = JobExecutionStatusExecuted
			}

			job.ExecutedAt = time.Now()
			for _, s := range p.subscribers {
				go s.OnJobExecuted(job)
			}
		}
	}
}

func newPlanner(threads uint) Planner {
	return &planner{
		jobs:        make(chan *JobContext, threads),
		threads:     threads,
		errs:        make(chan error),
		subscribers: make([]PlannerSubscriber, 0),
	}
}
