package sync

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDoWithTimeout(t *testing.T) {
	t.Run("task completes before timeout", func(t *testing.T) {
		err := DoWithTimeout(func() {
			time.Sleep(10 * time.Millisecond)
		}, time.Second)
		require.NoError(t, err)
	})

	t.Run("task exceeds timeout", func(t *testing.T) {
		err := DoWithTimeout(func() {
			time.Sleep(time.Second)
		}, 10*time.Millisecond)
		require.ErrorIs(t, err, ErrTimeout)
	})
}

func TestDoErrWithTimeout(t *testing.T) {
	t.Run("task completes before timeout with no error", func(t *testing.T) {
		err := DoErrWithTimeout(func() error {
			time.Sleep(10 * time.Millisecond)
			return nil
		}, time.Second)
		require.NoError(t, err)
	})

	t.Run("task completes before timeout with error", func(t *testing.T) {
		taskErr := fmt.Errorf("task failed")
		err := DoErrWithTimeout(func() error {
			return taskErr
		}, time.Second)
		require.ErrorIs(t, err, taskErr)
	})

	t.Run("task exceeds timeout", func(t *testing.T) {
		err := DoErrWithTimeout(func() error {
			time.Sleep(time.Second)
			return nil
		}, 10*time.Millisecond)
		require.ErrorIs(t, err, ErrTimeout)
	})

	t.Run("task error is not a timeout", func(t *testing.T) {
		taskErr := fmt.Errorf("task failed")
		err := DoErrWithTimeout(func() error {
			return taskErr
		}, time.Second)
		require.ErrorIs(t, err, taskErr)
		require.False(t, errors.Is(err, ErrTimeout))
	})
}
