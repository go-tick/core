package gotick

import (
	"context"
	"time"
)

type jobExecution struct {
	job      Job
	time     time.Time
	executed chan any
}

type planner struct {
	threads int
	jobs    chan jobExecution
	errs    chan error
}

func (p *planner) Errs() <-chan error {
	return p.errs
}

func (p *planner) Plan(ctx context.Context, job Job, t time.Time) (<-chan any, error) {
	executed := make(chan any, 1)
	select {
	case p.jobs <- jobExecution{job: job, time: t, executed: executed}:
		return executed, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (p *planner) Stop() error {
	close(p.jobs)
	close(p.errs)

	return nil
}

func (p *planner) Start(ctx context.Context) error {
	for i := 0; i < p.threads; i++ {
		go p.Executor(ctx)
	}

	return nil
}

func (p *planner) Executor(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case job := <-p.jobs:
			if time.Until(job.time) > 0 {
				select {
				case <-ctx.Done():
					return
				case <-time.After(time.Until(job.time)):
				}
			}

			err := job.job.Execute(ctx)
			if err != nil {
				p.errs <- err
			}

			job.executed <- struct{}{}
		}
	}
}

func newPlanner(threads int) Planner {
	return &planner{
		jobs:    make(chan jobExecution, threads),
		threads: threads,
		errs:    make(chan error),
	}
}
