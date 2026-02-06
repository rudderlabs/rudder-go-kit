package sync

import (
	"errors"
	"time"
)

// ErrTimeout is returned when a task does not complete within the specified duration.
var ErrTimeout = errors.New("task timed out")

// DoWithTimeout runs the given task function and enforces a timeout.
// If the task does not complete within the specified duration, it returns [ErrTimeout].
// If the task completes in time, it returns nil.
//
// # Warning
//
// On timeout the goroutine running the task is not cancelled or killed;
// callers must ensure the task eventually returns to avoid goroutine leaks.

func DoWithTimeout(task func(), timeout time.Duration) error {
	return DoErrWithTimeout(func() error {
		task()
		return nil
	}, timeout)
}

// DoErrWithTimeout runs the given task function and enforces a timeout.
// If the task does not complete within the specified duration, it returns an error(ErrTimeout). Dangling goroutines are not killed.
// If the task completes in time, it returns the error returned by the task function (which may be nil).
func DoErrWithTimeout(task func() error, timeout time.Duration) error {
	done := make(chan struct{})
	var taskErr error
	go func() {
		taskErr = task()
		close(done)
	}()

	select {
	case <-done:
		return taskErr
	case <-time.After(timeout):
		return ErrTimeout
	}
}
