package queue

import (
	"context"
	"sync"
)

type Task func() error

type Queue struct {
	tasks   chan Task
	workers int
	wg      sync.WaitGroup
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewQueue(workers int) *Queue {
	ctx, cancel := context.WithCancel(context.Background())
	q := &Queue{
		tasks:   make(chan Task, 100), // Buffer for 100 tasks
		workers: workers,
		ctx:     ctx,
		cancel:  cancel,
	}
	q.start()
	return q
}

func (q *Queue) start() {
	for i := 0; i < q.workers; i++ {
		q.wg.Add(1)
		go q.worker()
	}
}

func (q *Queue) worker() {
	defer q.wg.Done()
	for {
		select {
		case task := <-q.tasks:
			if task != nil {
				_ = task() // Execute task, errors handled within task
			}
		case <-q.ctx.Done():
			return
		}
	}
}

func (q *Queue) Submit(task Task) error {
	select {
	case q.tasks <- task:
		return nil
	case <-q.ctx.Done():
		return q.ctx.Err()
	}
}

func (q *Queue) SubmitSync(task Task) error {
	done := make(chan error, 1)

	wrappedTask := func() error {
		err := task()
		done <- err
		return err
	}

	if err := q.Submit(wrappedTask); err != nil {
		return err
	}

	return <-done
}

func (q *Queue) Shutdown() {
	q.cancel()
	close(q.tasks)
	q.wg.Wait()
}
