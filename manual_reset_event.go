package gowaithandle

import (
	"sync"
	"sync/atomic"
	"time"
)

// ManualResetEvent blocks all threads until
// signaled, then lets them all through
type ManualResetEvent struct {
	sync.RWMutex
	set      chan struct{}
	signaled int32
}

var _ EventWaitHandle = &ManualResetEvent{}

func NewManualResetEvent(signaled bool) *ManualResetEvent {
	mre := &ManualResetEvent{}

	if signaled {
		mre.signaled = 1
	}

	return mre
}

func (mre *ManualResetEvent) WaitOne(timeout time.Duration) <-chan bool {

	waiter := make(chan bool, 1)

	select {
	case <-time.After(timeout):
		waiter <- false
		close(waiter)
	case <-mre.getSignal():
		waiter <- true
		close(waiter)
	}
	return waiter
}

func (mre *ManualResetEvent) getSignal() chan struct{} {

	if atomic.LoadInt32(&mre.signaled) == 1 {
		c := make(chan struct{})
		close(c)
		return c
	}

	if mre.set != nil {
		mre.Lock()
		defer mre.Unlock()
		if mre.set == nil {
			mre.set = make(chan struct{})
		}
	}
	return mre.set
}

func (mre *ManualResetEvent) Set() bool {
	if atomic.CompareAndSwapInt32(&mre.signaled, 0, 1) {
		mre.Lock()
		defer mre.Unlock()
		if mre.set != nil {
			close(mre.set)
			mre.set = nil
		}
		return true
	}
	return false
}

func (mre *ManualResetEvent) Reset() bool {
	return atomic.CompareAndSwapInt32(&mre.signaled, 1, 0)
}
