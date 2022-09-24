package gowaithandle

import (
	"context"
)

func waitOne(ctx context.Context, sig chan struct{}) <-chan bool {

	waiter := make(chan bool, 1)

	select {
	case <-ctx.Done():
		waiter <- false
		close(waiter)
	case <-sig:
		waiter <- true
		close(waiter)
	}
	return waiter
}
