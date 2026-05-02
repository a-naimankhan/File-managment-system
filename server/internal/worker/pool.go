package worker

import (
	"context"
	"log"
)

type Task interface {
	Execute(ctx context.Context) error
}

type Pool struct {
	tasks       chan Task
	workerCount int
}

func NewPool(workerCount int) *Pool {
	return &Pool{
		tasks:       make(chan Task, workerCount),
		workerCount: workerCount,
	}
}

func (p *Pool) Start(ctx context.Context) {
	for i := 0; i < p.workerCount; i++ {
		go func(id int) {
			log.Printf("Worker %d starting", id)
			for {
				select {
				case task, ok := <-p.tasks:
					if !ok {
						log.Printf("Worker %d stopped , task channel closed", id)
						return
					}
					if err := task.Execute(ctx); err != nil {
						log.Printf("Worker %d failed", id)
					}
				case <-ctx.Done():
					log.Printf("Worker %d stopping via context", id)
					return
				}
			}
		}(i)
	}
}

func (p *Pool) Submit(task Task) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("worker pool is closed , task dropped : %v", r)
		}
	}()
	p.tasks <- task
}

func (p *Pool) Stop() {
	close(p.tasks)
}
