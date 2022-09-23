package gowaithandle

import (
	"sync"
	"sync/atomic"
	"time"
)

// AutoResetEvent blocks all threads until
// signaled, then lets them all through
type AutoResetEvent struct {
	sync.RWMutex
	signals chan struct{}
	created int32
}

var _ EventWaitHandle = &AutoResetEvent{}

func NewAutoResetEvent(signaled bool) *AutoResetEvent {
	are := &AutoResetEvent{}
	if signaled {
		are.Set()
	}
	return are
}

func (are *AutoResetEvent) WaitOne(timeout time.Duration) <-chan bool {
	return waitOne(are.getSignals(), timeout)
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

func (are *AutoResetEvent) Set() bool {

	// this will hit default if buffer is full
	select {
	case are.getSignals() <- struct{}{}:
		return true
	default:
		return false
	}
}

func (are *AutoResetEvent) Reset() bool {
	return true
}
