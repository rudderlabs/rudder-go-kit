package sync

import (
	"context"
	"sync"
)

// A EagerGroup is a collection of goroutines working on subtasks that are part of
// the same overall task.
//
// Use NewEagerGroup to create a new group.
type EagerGroup struct {
	ctx     context.Context
	cancel  context.CancelCauseFunc
	wg      sync.WaitGroup
	sem     chan struct{}
	errOnce sync.Once
	err     error
}

// NewEagerGroup returns a new eager group and an associated Context derived from ctx.
//
// The derived Context is canceled the first time a function passed to Go
// returns a non-nil error or the first time Wait returns, whichever occurs
// first.
//
// limit < 1 means no limit on the number of active goroutines.
func NewEagerGroup(ctx context.Context, limit int) (*EagerGroup, context.Context) {
	ctx, cancel := context.WithCancelCause(ctx)
	g := &EagerGroup{
		ctx:    ctx,
		cancel: cancel,
	}
	if limit > 0 {
		g.sem = make(chan struct{}, limit)
	}
	return g, ctx
}

// Go calls the given function in a new goroutine.
// It blocks until the new goroutine can be added without the number of
// active goroutines in the group exceeding the configured limit.
//
// The first call to return a non-nil error cancels the group's context.
// The error will be returned by Wait.
//
// If the group was created by calling NewEagerGroup with limit < 1, there is no
// limit on the number of active goroutines.
//
// If the group's context is canceled, routines that have not executed yet due to the limit won't be executed.
// Additionally, there is a best effort not to execute `f()` once the context is canceled
// and that happens whether or not a limit has been specified.
func (g *EagerGroup) Go(f func() error) {
	if err := g.ctx.Err(); err != nil {
		g.errOnce.Do(func() {
			g.err = g.ctx.Err()
			g.cancel(g.err)
		})
		return
	}

	if g.sem != nil {
		select {
		case <-g.ctx.Done():
			g.errOnce.Do(func() {
				g.err = g.ctx.Err()
				g.cancel(g.err)
			})
			return
		case g.sem <- struct{}{}:
		}
	}

	g.wg.Add(1)
	go func() {
		err := g.ctx.Err()
		if err == nil {
			err = f()
		}
		if err != nil {
			g.errOnce.Do(func() {
				g.err = err
				g.cancel(g.err)
			})
		}
		if g.sem != nil {
			<-g.sem
		}
		g.wg.Done()
	}()
}

// Wait blocks until all function calls from the Go method have returned, then
// returns the first non-nil error (if any) from them.
func (g *EagerGroup) Wait() error {
	g.wg.Wait()
	g.cancel(g.err)
	return g.err
}
