package sync

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestEagerGroupWithLimit(t *testing.T) {
	g, ctx := NewEagerGroup(context.Background(), 2)
	var count atomic.Int64
	// One of the following three goroutines should DEFINITELY NOT be executed due to the limit of 2 and the context being cancelled.
	// The context should get cancelled automatically because the first two routines returned an error.
	g.Go(func() error {
		t.Log("one")
		count.Add(1)
		return fmt.Errorf("one")
	})
	g.Go(func() error {
		t.Log("two")
		count.Add(1)
		return fmt.Errorf("two")
	})
	g.Go(func() error {
		t.Log("three")
		count.Add(1)
		return fmt.Errorf("three")
	})
	require.Error(t, g.Wait(), "We expect group.Wait() to return an error")
	ok := true
	select {
	case <-ctx.Done():
		_, ok = <-ctx.Done()
	case <-time.After(time.Second):
	}
	require.False(t, ok, "We expect the context to be cancelled")
	require.True(t, 1 <= count.Load() && count.Load() <= 2, "We expect count to be between 1 and 2")
}

func TestEagerGroupWithNoLimit(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := NewEagerGroup(ctx, 0)
	funcCounter := &atomic.Int64{}

	go func() {
		for {
			if funcCounter.Load() > 10 {
				cancel()
				return
			}
		}
	}()

	for i := 0; i < 10000; i++ {
		g.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
			funcCounter.Add(1)
			return nil
		})
	}
	require.ErrorIs(t, g.Wait(), ctx.Err(), "We expect group.Wait() to return the context error")
	_, ok := <-ctx.Done()
	require.False(t, ok, "We expect the context to be cancelled")
	t.Log(funcCounter.Load(), "funcs executed")
	// We expect between 10 and 10000 funcs to be executed
	// because group tries to return early if context is cancelled
	require.Less(
		t,
		funcCounter.Load(),
		int64(10000),
		"Expected less than 1000 funcs to be executed",
	)
}

func TestNoInitEagerGroup(t *testing.T) {
	g := &EagerGroup{}
	f := func() error { return nil }
	require.Panics(
		t,
		func() { g.Go(f) },
		"We expect a panic when calling Go on a group that has not been initialized with NewEagerGroup",
	)
}
