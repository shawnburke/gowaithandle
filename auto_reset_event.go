package gowaithandle

import (
	"context"
	"sync"
	"sync/atomic"
)

// AutoResetEvent blocks all threads until
// signaled, then lets one of them through and
// then immediately resets. This allows "pulsing" access to one
// thread at a time.
type AutoResetEvent struct {
	sync.RWMutex
	signals  chan struct{}
	created  int32
	signaled int32
}

var _ EventWaitHandle = &AutoResetEvent{}

// NewAutoResetEvent creates a new instance in the specified
// signal state. If set in a signaled state, it means the first
// waiter will be let through, then the event will flip to non-signaled.
func NewAutoResetEvent(signaled bool) *AutoResetEvent {
	are := &AutoResetEvent{}
	if signaled {
		are.Set()
	}
	return are
}

// WaitOne will wait on this handle. The return channel will
// receive a true if the handle has been signaled, or false if the context
// hits its timeout or deadline, or is cancelled.
func (are *AutoResetEvent) WaitOne(ctx context.Context) <-chan bool {
	return waitOne(ctx, are.getSignals(), func(res bool) {
		atomic.AddInt32(&are.signaled, -1)
	})
}

// Set signals the handle to let a single thread proceed and then immediately
// returns to the non-signaled state.  That is, if 3 threads are in WaitOne,
// one will be allowed to proceed for each call to Set.
func (are *AutoResetEvent) Set() bool {

	// this will hit default if buffer is full
	select {
	case are.getSignals() <- struct{}{}:
		atomic.AddInt32(&are.signaled, 1)
		return true
	default:
		return false
	}
}

// Reset immediately sets the event to a non-signaled state
func (are *AutoResetEvent) Reset() bool {
	if atomic.LoadInt32(&are.signaled) > 0 {
		<-are.getSignals()
		return true
	}
	return false
}

func (are *AutoResetEvent) getSignals() chan struct{} {

	// use a lock here but only in the default
	// non-created state, which allows us
	// to lazily create the channel but only take
	// the lock once to avoid contention in the future
	if atomic.LoadInt32(&are.created) == 0 {
		are.Lock()
		if are.signals == nil {
			are.signals = make(chan struct{}, 1)
		}
		atomic.StoreInt32(&are.created, 1)
		are.Unlock()
	}
	return are.signals
}
