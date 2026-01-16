package group

import (
	"context"
	"sync"
	"time"
)

type group struct {
	err     error
	cancel  func(error)
	wg      sync.WaitGroup
	errOnce sync.Once
}

func WithContext(ctx context.Context) (*group, context.Context) {
	ctx, cancel := context.WithCancelCause(ctx)
	return &group{cancel: cancel}, ctx
}

func (g *group) Go(fn func() error) {
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		if err := fn(); err != nil {
			g.errOnce.Do(
				func() {
					g.err = err
					if g.cancel != nil {
						time.Sleep(100 * time.Millisecond)
						g.cancel(g.err)
					}
				})
		}
	}()
}

func (g *group) Wait() error {
	g.wg.Wait()
	if g.err != nil {
		g.cancel(g.err)
	}
	return g.err
}
