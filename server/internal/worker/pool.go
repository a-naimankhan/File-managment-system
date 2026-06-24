package worker

import (
	"context"
	"log"
	"sync"
)

type Task interface {
	Execute(ctx context.Context) error
}

type Pool struct {
	tasks       chan Task
	closed      bool
	mu          sync.Mutex
	workerCount int
	wg          sync.WaitGroup
}

func NewPool(workerCount int) *Pool {
	return &Pool{
		tasks:       make(chan Task, workerCount),
		workerCount: workerCount,
	}
}

func (p *Pool) Start(ctx context.Context) {
	for i := 0; i < p.workerCount; i++ {
		p.wg.Add(1)
		go func(id int) {
			defer p.wg.Done()
			log.Printf("Worker %d starting", id)
			for {
				select {
				case task, ok := <-p.tasks:
					if !ok {

						return
					}
					task.Execute(ctx)
				case <-ctx.Done():
					return
				}
			}
		}(i)
	}
}

func (p *Pool) Submit(task Task) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		log.Println("worker is closed task is dropped")
		return
	}
	p.tasks <- task
}

func (p *Pool) Stop() {
	p.mu.Lock()
	if p.closed {
		log.Println("worker is closed task is dropped")
		return
	}
	p.closed = true
	close(p.tasks)
	p.mu.Unlock()

	p.wg.Wait()
	log.Println("worker pool stopped")
}
