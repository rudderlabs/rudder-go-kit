package context

import (
	"context"
	"sync"
)

// MergedContext creates a context with the first context as parent, but is cancelled when the second context is cancelled.
// calling the returned CancelFunc will also cancel the merged context and stop the associated goroutine.
func MergedContext(ctxA, ctxB context.Context) (context.Context, context.CancelFunc) {
	// Create a cancellable context from ctxA
	mergedCtx, cancel := context.WithCancel(ctxA)

	// Check if ctxB is already done, if so cancel immediately
	if ctxB.Err() != nil {
		cancel()
		return mergedCtx, func() {}
	}

	var wg sync.WaitGroup

	wg.Go(func() {
		select {
		case <-ctxB.Done():
			cancel() // When ctxB is cancelled, cancel the merged context
		case <-mergedCtx.Done():
			// If mergedCtx is done, just exit the goroutine
		}
	})

	return mergedCtx, func() {
		cancel()
		wg.Wait() // Ensure the goroutine has finished before returning
	}
}
