package context_test

import (
	"context"
	"testing"
	"time"

	kitctx "github.com/rudderlabs/rudder-go-kit/context"
)

func TestMergedContext(t *testing.T) {
	t.Run("cancel ctxA cancels merged context", func(t *testing.T) {
		ctxA, cancelA := context.WithCancel(context.Background())
		ctxB := context.Background()

		mergedCtx, mergedCancel := kitctx.MergedContext(ctxA, ctxB)
		defer mergedCancel()

		// Cancel ctxA
		cancelA()

		// Check that merged context is cancelled
		select {
		case <-mergedCtx.Done():
			// Success
		case <-time.After(100 * time.Millisecond):
			t.Error("Expected merged context to be cancelled when ctxA is cancelled")
		}
	})

	t.Run("cancel ctxB cancels merged context", func(t *testing.T) {
		ctxA := context.Background()
		ctxB, cancelB := context.WithCancel(context.Background())

		mergedCtx, mergedCancel := kitctx.MergedContext(ctxA, ctxB)
		defer mergedCancel()

		// Cancel ctxB
		cancelB()

		// Check that merged context is cancelled
		select {
		case <-mergedCtx.Done():
			// Success
		case <-time.After(100 * time.Millisecond):
			t.Error("Expected merged context to be cancelled when ctxB is cancelled")
		}
	})

	t.Run("cancel merged context directly", func(t *testing.T) {
		ctxA := context.Background()
		ctxB := context.Background()

		mergedCtx, mergedCancel := kitctx.MergedContext(ctxA, ctxB)

		// Cancel merged context directly
		mergedCancel()

		// Check that merged context is cancelled
		select {
		case <-mergedCtx.Done():
			// Success
		case <-time.After(100 * time.Millisecond):
			t.Error("Expected merged context to be cancelled when mergedCancel is called")
		}
	})

	t.Run("already cancelled ctxA", func(t *testing.T) {
		ctxA, cancelA := context.WithCancel(context.Background())
		ctxB := context.Background()

		// Cancel ctxA before merging
		cancelA()

		mergedCtx, mergedCancel := kitctx.MergedContext(ctxA, ctxB)
		defer mergedCancel()

		// Check that merged context is already cancelled
		select {
		case <-mergedCtx.Done():
			// Success
		default:
			t.Error("Expected merged context to be cancelled when ctxA was already cancelled")
		}
	})

	t.Run("already cancelled ctxB", func(t *testing.T) {
		ctxA := context.Background()
		ctxB, cancelB := context.WithCancel(context.Background())

		// Cancel ctxB before merging
		cancelB()

		mergedCtx, mergedCancel := kitctx.MergedContext(ctxA, ctxB)
		defer mergedCancel()

		// Check that merged context is already cancelled
		select {
		case <-mergedCtx.Done():
			// Success
		default:
			t.Error("Expected merged context to be cancelled when ctxB was already cancelled")
		}
	})

	t.Run("timeout ctxA cancels merged context", func(t *testing.T) {
		ctxA, cancelA := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancelA()
		ctxB := context.Background()

		mergedCtx, mergedCancel := kitctx.MergedContext(ctxA, ctxB)
		defer mergedCancel()

		// Wait for ctxA to timeout
		select {
		case <-mergedCtx.Done():
			// Success
		case <-time.After(100 * time.Millisecond):
			t.Error("Expected merged context to be cancelled when ctxA times out")
		}
	})

	t.Run("timeout ctxB cancels merged context", func(t *testing.T) {
		ctxA := context.Background()
		ctxB, cancelB := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancelB()

		mergedCtx, mergedCancel := kitctx.MergedContext(ctxA, ctxB)
		defer mergedCancel()

		// Wait for ctxB to timeout
		select {
		case <-mergedCtx.Done():
			// Success
		case <-time.After(100 * time.Millisecond):
			t.Error("Expected merged context to be cancelled when ctxB times out")
		}
	})
}
