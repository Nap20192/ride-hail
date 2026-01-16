package conc

import (
	"fmt"
	"sync"
	"sync/atomic"
)

var (
	ErrInvalidPoolSize = fmt.Errorf("invalid pool size")
	ErrPoolClosed      = fmt.Errorf("pool is closed")
)

type Pool[T any] struct {
	numWorkers atomic.Int64
	active     atomic.Int64
	closed     atomic.Bool
	wg         sync.WaitGroup
}

func NewWorkerPool[T any](numWorkers int64) *Pool[T] {
	var p Pool[T]
	p.numWorkers.Store(numWorkers)
	p.active.Store(0)
	p.closed.Store(false)
	return &p
}

func (p *Pool[T]) Resize(newSize int64) error {
	if newSize < 1 {
		return ErrInvalidPoolSize
	}

	if newSize == p.numWorkers.Load() {
		return fmt.Errorf("pool size is already %d", newSize)
	}

	if p.closed.Load() {
		return ErrPoolClosed
	}

	p.numWorkers.Store(newSize)

	fmt.Printf("Resized pool to %d workers\n", newSize)
	return nil
}

func (p *Pool[T]) Work(task T, handler func(T)) {
	for {
		active := p.active.Load()
		if active < p.numWorkers.Load() {
			if p.active.CompareAndSwap(active, active+1) {
				p.wg.Add(1)

				go func() {
					defer p.wg.Done()
					defer p.active.Add(-1)
					handler(task)
				}()
				return
			}
		}
	}
}

func (p *Pool[T]) Wait() {
	p.wg.Wait()
	p.closed.Store(true)
}

func (p *Pool[T]) ActiveCount() int64 {
	return p.active.Load()
}
