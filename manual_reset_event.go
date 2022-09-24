package gowaithandle

import (
	"context"
	"sync/atomic"
	"time"
)

// ManualResetEvent blocks all threads until
// signaled, then lets them all through
type ManualResetEvent struct {
	set        chan struct{}
	setVersion int64
	signaled   int32
	version    int64
}

var _ EventWaitHandle = &ManualResetEvent{}

func NewManualResetEvent(signaled bool) *ManualResetEvent {
	mre := &ManualResetEvent{}

	if signaled {
		mre.signaled = 1
	}

	return mre
}

func (mre *ManualResetEvent) WaitDuration(timeout time.Duration) <-chan bool {
	ctx, _ := timeoutContext(timeout)
	return mre.WaitOne(ctx)
}

func (mre *ManualResetEvent) WaitOne(ctx context.Context) <-chan bool {
	return waitOne(ctx, mre.getSignal())
}

func (mre *ManualResetEvent) getSignal() chan struct{} {

	if atomic.LoadInt32(&mre.signaled) == 1 {
		c := make(chan struct{})
		close(c)
		return c
	}

	// if the version and set version are the same,
	// it means we need a new signal channel so we create
	// one and increment the version
	v := atomic.LoadInt64(&mre.version)
	sv := atomic.LoadInt64(&mre.setVersion)
	next := v + 1
	if v == sv && atomic.CompareAndSwapInt64(&mre.version, v, next) {
		mre.set = make(chan struct{})
		atomic.StoreInt64(&mre.setVersion, next)
	}
	return mre.set
}

func (mre *ManualResetEvent) Set() bool {
	if atomic.CompareAndSwapInt32(&mre.signaled, 0, 1) {
		if mre.set != nil {
			close(mre.set)
		}
		return true
	}
	return false
}

func (mre *ManualResetEvent) Reset() bool {
	return atomic.CompareAndSwapInt32(&mre.signaled, 1, 0)
}
