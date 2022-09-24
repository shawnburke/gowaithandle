package gowaithandle

import (
	"context"
	"sync"
	"time"
)

type WaitGroup struct {
	sync.WaitGroup
}

var _ WaitHandle = &WaitGroup{}

func (wg *WaitGroup) WaitDuration(timeout time.Duration) <-chan bool {
	ctx, _ := timeoutContext(timeout)
	return wg.WaitOne(ctx)
}

func (wg *WaitGroup) WaitOne(ctx context.Context) <-chan bool {
	done := make(chan bool, 1)
	exit := make(chan struct{})
	go func() {
		wg.WaitGroup.Wait()
		done <- true
		close(exit)
	}()

	go func() {
		defer close(done)
		select {
		case <-exit:
			return
		case <-ctx.Done():
			done <- false
			return
		}
	}()

	return done
}
