package gowaithandle

import (
	"context"
	"sync/atomic"
)

// WaitAll waits for all of the handles to be signaled.
// When successful, true is put onto the return channel, otherwise
// false in the case of a timeout or other failure of any of the handles.
func WaitAll(ctx context.Context, ev ...WaitHandle) <-chan bool {
	return waitMultipleCore(ctx, modeAll, ev...)
}

// WaitAny waits for any of the handles to be signaled, then returns true
// on the return channel
func WaitAny(ctx context.Context, ev ...WaitHandle) <-chan bool {
	return waitMultipleCore(ctx, modeAny, ev...)
}

type mode int

const (
	modeAny mode = 0
	modeAll mode = 1
)

func waitMultipleCore(ctx context.Context, mode mode, ev ...WaitHandle) <-chan bool {
	done := make(chan bool)
	exit := make(chan struct{})
	finished := int32(0)

	finish := func(result bool) {
		if atomic.CompareAndSwapInt32(&finished, 0, 1) {
			done <- result
			close(exit)
			close(done)
		}
	}

	count := int32(len(ev))
	for _, ewh := range ev {
		go func(h WaitHandle) {

			select {
			case <-exit:
				return
			case res := <-h.WaitOne(ctx):
				if !res {
					if mode == modeAll {
						finish(false)
					}
					return
				}

				if mode == modeAny || atomic.AddInt32(&count, -1) == 0 {
					finish(true)
				}
			}
		}(ewh)
	}

	go func() {
		select {
		case <-exit:
			// success!
		case <-ctx.Done():
			finish(false)
		}
	}()

	return done
}

// waitOne is a helper function that takes a context and a signal channel and returns
// a bool channel.
//
// The channel will contain TRUE if the signal channel receives a message.  Otherwise,
// it will contain FALSE if the context times out or is cancelled.
func waitOne(ctx context.Context, sig chan struct{}, notify func(bool)) <-chan bool {

	if ctx == nil {
		ctx = context.Background()
	}
	waiter := make(chan bool, 1)

	finish := func(result bool) {
		waiter <- result
		close(waiter)
		if notify != nil {
			notify(result)
		}
	}

	go func() {
		select {
		case <-ctx.Done():
			finish(false)
		case <-sig:
			finish(true)
		}
	}()
	return waiter
}
