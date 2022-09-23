package gowaithandle

import (
	"sync/atomic"
	"time"
)

func WaitAll(timeout time.Duration, ev ...WaitHandle) <-chan bool {
	return waitMultipleCore(timeout, modeAll, ev...)
}

func WaitAny(timeout time.Duration, ev ...WaitHandle) <-chan bool {
	return waitMultipleCore(timeout, modeAny, ev...)
}

type mode int

const (
	modeAny mode = 0
	modeAll mode = 1
)

func waitMultipleCore(timeout time.Duration, mode mode, ev ...WaitHandle) <-chan bool {
	done := make(chan bool)
	exit := make(chan struct{})

	finish := func(result bool) {
		done <- result
		close(exit)
		close(done)
	}

	count := int32(len(ev))
	for _, ewh := range ev {
		go func(h WaitHandle) {
			res := <-h.WaitOne(timeout)
			if !res {
				if mode == modeAll {
					finish(false)
				}
				return
			}

			if mode == modeAny || atomic.AddInt32(&count, 1) == count {
				finish(true)
			}
		}(ewh)
	}

	go func() {
		select {
		case <-exit:
			// success!
		case <-time.After(timeout):
			finish(false)
		}
	}()

	return done
}
