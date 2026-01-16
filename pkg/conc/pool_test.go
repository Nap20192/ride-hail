package conc

import (
	"context"
	"testing"
	"time"
)

func TestPool(t *testing.T) {
	pool := NewWorkerPool[int](3)

	tasks := make([]int, 100)
	for i := 0; i < 100; i++ {
		tasks[i] = i
	}

	go func(p *Pool[int]) {
		time.Sleep(5 * time.Second)
		err := p.Resize(4)
		if err != nil {
			t.Logf("Resize error: %v", err)
		}
	}(pool)

	go func(p *Pool[int]) {
		time.Sleep(8 * time.Second)
		err := p.Resize(12)
		if err != nil {
			t.Logf("Resize error: %v", err)
		}
	}(pool)
	ticker := time.NewTicker(1 * time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	go func(p *Pool[int]) {
		for {
			select {
			case <-ticker.C:
				t.Logf("Active workers: %d", p.ActiveCount())
			case <-ctx.Done():
				return
			}
		}
	}(pool)

	handler := func(task int) {
		time.Sleep(2 * time.Second)
	}

	for _, task := range tasks {
		pool.Work(task, handler)
	}

	pool.Wait()
	cancel()
}
