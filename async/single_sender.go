package async

import (
	"context"
)

// SingleSender is a helper for sending and receiving values to and from a channel between 2 separate goroutines, a sending and a receiving goroutine, while at the same time supporting the following scenarios:
//  1. The sending goroutine in case the parent context is canceled should be able to notify the receiver goroutine about the error through the channel.
//  2. The receiving goroutine should be able to stop listening from the channel (a.k.a. leave) at any point.
//  3. The sending goroutine shouldn't be blocked trying to send to the channel when the receiver has left it.
//  4. Receiver's departure should act as a context cancellation signal to the sending goroutine, i.e. it should stop working.
type SingleSender[T any] struct {
	ctx           context.Context
	ctxCancel     context.CancelFunc
	sendCtx       context.Context
	sendCtxCancel context.CancelFunc
	ch            chan T
	closed        bool
}

// Begin creates a new channel and returns it along with a context for the sending goroutine to use and a function for the receiving goroutine to be able to leave the "conversation" if needed.
func (s *SingleSender[T]) Begin(parentCtx context.Context) (ctx context.Context, ch <-chan T, leave func()) {
	s.ctx, s.ctxCancel = context.WithCancel(parentCtx)
	s.ch = make(chan T)
	s.sendCtx, s.sendCtxCancel = context.WithCancel(context.Background())
	return s.ctx, s.ch, s.sendCtxCancel
}

// Send tries to send a value to the channel. If the channel is closed, or the receiving goroutine has left it does nothing.
func (s *SingleSender[T]) Send(value T) {
	closed := s.closed
	if closed { // don't send to a closed channel
		return
	}
	select {
	case <-s.sendCtx.Done():
		s.ctxCancel()
		return
	case s.ch <- value:
	}
}

// Close the channel and cancel all related contexts.
func (s *SingleSender[T]) Close() {
	if s.closed {
		return
	}
	s.closed = true
	s.ctxCancel()
	s.sendCtxCancel()
	close(s.ch)
}
