package gowaithandle

import (
	"context"
	"sync/atomic"
	"time"
)

func timeoutContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	ctx := context.Background()
	cancel := func() {}

	if timeout > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
	}
	return ctx, cancel
}

func WaitAll(ctx context.Context, ev ...WaitHandle) <-chan bool {
	return waitMultipleCore(ctx, modeAll, ev...)
}

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
