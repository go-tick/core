package gotick

import (
	"context"
	"time"
)

type planner struct {
	threads uint
	jobs    chan JobContext
	errs    chan error
}

func (p *planner) Errs() <-chan error {
	return p.errs
}

func (p *planner) Plan(ctx JobContext) error {
	select {
	case p.jobs <- ctx:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (p *planner) Stop() error {
	close(p.jobs)
	close(p.errs)

	return nil
}

func (p *planner) Start(ctx context.Context) error {
	for range p.threads {
		go p.executor(ctx)
	}

	return nil
}

func (p *planner) executor(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case job := <-p.jobs:
			if time.Until(job.PlannedAt) > 0 {
				select {
				case <-ctx.Done():
					return
				case <-time.After(time.Until(job.PlannedAt)):
				}
			}

			job.ExecutedAt = time.Now()
			err := job.Job.Execute(job)
			if err != nil {
				p.errs <- err
			}

			job.executed <- struct{}{}
		}
	}
}

func newPlanner(threads uint) Planner {
	return &planner{
		jobs:    make(chan JobContext, threads),
		threads: threads,
		errs:    make(chan error, 1),
	}
}
