package gowaithandle

import (
	"sync"
	"time"
)

// AutoResetEvent blocks all threads until
// signaled, then lets them all through
type AutoResetEvent struct {
	sync.RWMutex
	signals chan bool
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

	waiter := make(chan bool, 1)

	select {
	case <-time.After(timeout):
		waiter <- false
		close(waiter)
	case <-are.getSignals():
		waiter <- true
		close(waiter)
	}
	return waiter
}

func (are *AutoResetEvent) getSignals() chan bool {
	are.Lock()
	defer are.Unlock()
	if are.signals == nil {
		are.signals = make(chan bool, 1)
	}
	return are.signals
}

func (are *AutoResetEvent) Set() bool {

	// this will hit default if buffer is full
	select {
	case are.getSignals() <- true:
		return true
	default:
		return false
	}
}

func (are *AutoResetEvent) Reset() bool {
	return true
}
