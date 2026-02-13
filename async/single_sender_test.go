package async_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	"golang.org/x/sync/errgroup"

	"github.com/rudderlabs/rudder-go-kit/async"
)

func TestSingleSender(t *testing.T) {
	type valueOrError struct {
		value int
		err   error
	}

	send := func(ctx context.Context, s *async.SingleSender[valueOrError], times int) (sendCalls int) {
		defer s.Close()
		for i := range times {
			if ctx.Err() != nil {
				s.Send(valueOrError{err: ctx.Err()})
				return sendCalls
			}
			s.Send(valueOrError{value: i})
			sendCalls++
		}
		return sendCalls
	}

	receive := func(ch <-chan valueOrError, delay time.Duration) ([]int, []error) {
		var receivedValues []int
		var receivedErrors []error
		for v := range ch {
			time.Sleep(delay)
			if v.err != nil {
				receivedErrors = append(receivedErrors, v.err)
			} else {
				receivedValues = append(receivedValues, v.value)
			}
		}
		return receivedValues, receivedErrors
	}

	t.Run("receive all values from sender", func(t *testing.T) {
		defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
		s := &async.SingleSender[valueOrError]{}
		ctx, ch, _ := s.Begin(context.Background())
		defer s.Close()

		g := &errgroup.Group{}

		var sendCalls int
		g.Go(func() error {
			sendCalls = send(ctx, s, 10)
			return nil
		})

		var receivedValues []int
		var receivedErrors []error
		g.Go(func() error {
			receivedValues, receivedErrors = receive(ch, 0)
			return nil
		})

		_ = g.Wait()

		require.Equal(t, 10, sendCalls)
		require.Equal(t, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, receivedValues)
		require.Empty(t, receivedErrors)
	})

	t.Run("parent context is canceled", func(t *testing.T) {
		defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
		parentCtx, parentCtxCancel := context.WithCancel(context.Background())
		parentCtxCancel()
		s := &async.SingleSender[valueOrError]{}
		ctx, ch, _ := s.Begin(parentCtx)
		defer s.Close()

		g := &errgroup.Group{}

		var sendCalls int
		g.Go(func() error {
			sendCalls = send(ctx, s, 10)
			return nil
		})

		var receivedValues []int
		var receivedErrors []error

		g.Go(func() error {
			receivedValues, receivedErrors = receive(ch, 0)
			return nil
		})
		_ = g.Wait()

		require.Zero(t, sendCalls)
		require.Nil(t, receivedValues, "no values should be received")
		require.Equal(t, []error{context.Canceled}, receivedErrors)
	})

	t.Run("parent context is canceled after interaction has started", func(t *testing.T) {
		defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
		parentCtx, parentCtxCancel := context.WithCancel(context.Background())

		s := &async.SingleSender[valueOrError]{}
		ctx, ch, _ := s.Begin(parentCtx)
		defer s.Close()

		g := &errgroup.Group{}

		var sendCalls int
		g.Go(func() error {
			sendCalls = send(ctx, s, 1000)
			return nil
		})

		var receivedValues []int
		var receivedErrors []error

		g.Go(func() error {
			receivedValues, receivedErrors = receive(ch, 10*time.Millisecond)
			return nil
		})
		time.Sleep(time.Millisecond * 100)
		parentCtxCancel()
		_ = g.Wait()

		require.GreaterOrEqual(t, sendCalls, 1, "sender should have called send at least for 1 value")
		require.GreaterOrEqual(t, len(receivedValues), 1, "receiver should have called received at least for 1 value")
		require.Equal(t, []error{context.Canceled}, receivedErrors)
	})

	t.Run("try to send another value after sender is closed", func(t *testing.T) {
		defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
		s := async.SingleSender[valueOrError]{}
		_, ch, _ := s.Begin(context.Background())
		defer s.Close()

		g := &errgroup.Group{}

		g.Go(func() error {
			for i := range 10 {
				s.Send(valueOrError{value: i})
			}
			s.Close()
			s.Send(valueOrError{value: 10})
			return nil
		})

		var receivedValues []int
		var receivedErrors []error

		g.Go(func() error {
			receivedValues, receivedErrors = receive(ch, 0)
			return nil
		})
		_ = g.Wait()

		require.Equal(t, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, receivedValues)
		require.Empty(t, receivedErrors)
	})

	t.Run("receiver leaves before sender sends all values", func(t *testing.T) {
		defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
		s := async.SingleSender[valueOrError]{}
		ctx, ch, leave := s.Begin(context.Background())
		defer s.Close()

		g := &errgroup.Group{}

		var sendCalls int
		g.Go(func() error {
			sendCalls = send(ctx, &s, 10)
			return nil
		})

		var receivedValues []int
		var receivedErrors []error

		g.Go(func() error {
			for v := range ch {
				if v.err != nil {
					receivedErrors = append(receivedErrors, v.err)
				} else {
					receivedValues = append(receivedValues, v.value)
				}
				// leave after receiving 1 value
				leave()
				// make sure sender has time to try and send another value and figure out that context is canceled
				time.Sleep(100 * time.Millisecond)
			}
			return nil
		})
		_ = g.Wait()

		require.GreaterOrEqual(t, len(receivedValues), 1, "receiver should have received at least 1 value")
		require.LessOrEqual(t, len(receivedValues), 2, "receiver should have received at most 2 values")
		require.GreaterOrEqual(t, sendCalls, 1, "sender should have called send at least for 1 value")
		require.LessOrEqual(t, sendCalls, 2, "sender should have called send at most for 2 values, i.e. it should stop sending after receiver leaves")
	})

	t.Run("sender closes then sends again", func(t *testing.T) {
		s := async.SingleSender[valueOrError]{}
		_, ch, _ := s.Begin(context.Background())
		go func() {
			s.Close()
			s.Send(valueOrError{value: 1})
		}()

		var values []int
		for range ch {
			values = append(values, 1)
		}

		require.Empty(t, values)
	})

	t.Run("concurrent close operations", func(t *testing.T) {
		defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
		s := async.SingleSender[valueOrError]{}
		_, ch, _ := s.Begin(context.Background())

		// Start a goroutine to consume from the channel to prevent blocking
		done := make(chan struct{})
		go func() {
			for range ch {
				// consume values until channel is closed
			}
			close(done)
		}()

		// Multiple goroutines calling Close() simultaneously
		g := &errgroup.Group{}
		for range 10 {
			g.Go(func() error {
				s.Close()
				return nil
			})
		}

		_ = g.Wait()

		// Wait for the consumer goroutine to finish
		<-done
	})
}
