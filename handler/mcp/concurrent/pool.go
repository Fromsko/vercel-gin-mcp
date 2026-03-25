package concurrent

import (
	"context"

	"github.com/sourcegraph/conc/pool"
)

// Task represents a concurrent task
type Task func() error

// Pool wraps sourcegraph/conc pool for our use cases
type Pool struct {
	pool *pool.Pool
}

// ContextPool wraps sourcegraph/conc context pool for our use cases
type ContextPool struct {
	pool *pool.ContextPool
}

// NewPool creates a new goroutine pool with specified max concurrency
func NewPool(maxGoroutines int) *Pool {
	p := pool.New().WithMaxGoroutines(maxGoroutines)
	return &Pool{pool: p}
}

// NewContextPool creates a new context-aware goroutine pool
func NewContextPool(ctx context.Context, maxGoroutines int) *ContextPool {
	p := pool.New().WithContext(ctx).WithMaxGoroutines(maxGoroutines)
	return &ContextPool{pool: p}
}

// Go submits a task to the pool
func (p *Pool) Go(task Task) {
	p.pool.Go(func() { _ = task() })
}

// Wait waits for all tasks to complete
func (p *Pool) Wait() error {
	p.pool.Wait()
	return nil
}

// Go submits a task to the context pool
func (cp *ContextPool) Go(task Task) {
	cp.pool.Go(func(ctx context.Context) error { 
		return task()
	})
}

// Wait waits for all tasks to complete
func (cp *ContextPool) Wait() error {
	cp.pool.Wait()
	return nil
}