package gowaithandle

import (
	"sync"
	"time"
)

// AutoResetEvent blocks all threads until
// signaled, then lets them all through
type AutoResetEvent struct {
	sync.RWMutex
	signals chan struct{}
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
	are.Lock()
	defer are.Unlock()
	if are.signals == nil {
		are.signals = make(chan struct{}, 1)
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
