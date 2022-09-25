package gowaithandle

import (
	"context"
	"sync/atomic"
)

// ManualResetEvent blocks all threads until
// signaled, then lets them all through when in that state
type ManualResetEvent struct {
	set        chan struct{}
	setVersion int64
	signaled   int32
	version    int64
}

var _ EventWaitHandle = &ManualResetEvent{}

// NewManualResetEvent creates the handle specifying
// the sginal state.
func NewManualResetEvent(signaled bool) *ManualResetEvent {
	mre := &ManualResetEvent{}

	if signaled {
		mre.signaled = 1
	}

	return mre
}

// WaitOne waits until the handle is signaled, then allows the caller to
// proceed. Context can be used to set timeout or deadline or perform cancellation.
// The return channel will have true if the handle was signaled, or false if the context
// timed out or was cancelled.
func (mre *ManualResetEvent) WaitOne(ctx context.Context) <-chan bool {
	return waitOne(ctx, mre.getSignal(), nil)
}

// Set signals the handle to allow any blocked callers to proceed.
func (mre *ManualResetEvent) Set() bool {
	if atomic.CompareAndSwapInt32(&mre.signaled, 0, 1) {
		if mre.set != nil {
			close(mre.set)
		}
		return true
	}
	return false
}

// Reset sets the handle to the non-signaled state, such that
// subsequent calls to WaitOne will block until Set is called.
func (mre *ManualResetEvent) Reset() bool {
	return atomic.CompareAndSwapInt32(&mre.signaled, 1, 0)
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
