package workerpool

import (
	"context"
	"sync"
)

type Job interface {
	Do(ctx context.Context) error
}

type JobFunc struct {
	f        func(context.Context) error
	executer []func()
}

type Workerpool struct {
	ctx  context.Context
	jobs chan Job
	errs chan error
	wg   sync.WaitGroup
}

func NewWorkerPool(ctx context.Context, wn int) *Workerpool {
	pool := &Workerpool{
		ctx:  ctx,
		jobs: make(chan Job, wn),
		errs: make(chan error),
	}

	for w := 0; w <= wn; w++ {
		pool.wg.Add(1)
		go func() {
			defer pool.wg.Done()
			pool.work()
		}()
	}
	return pool
}

func (wp *Workerpool) work() {
	for {
		select {
		case job := <-wp.jobs:
			err := job.Do(wp.ctx)
			if err != nil {
				wp.errs <- err
			}
		case <-wp.ctx.Done():
			return
		}
	}
}

func (wp *Workerpool) Submit(j Job) {
	wp.jobs <- j
}

func (wp *Workerpool) Errors() <-chan error {
	return wp.errs
}

func (wp *Workerpool) Close() error {
	wp.wg.Wait()
	return <-wp.errs
}

func (j *JobFunc) Do(ctx context.Context) error {
	defer func() {
		for _, exec := range j.executer {
			exec()
		}
	}()
	return j.f(ctx)
}

func NewJob(f func(ctx context.Context) error, e ...func()) *JobFunc {
	return &JobFunc{
		f:        f,
		executer: e,
	}
}
