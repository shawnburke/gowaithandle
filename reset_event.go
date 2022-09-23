package gowaithandle

import (
	"time"
)

func waitOne(sig chan struct{}, timeout time.Duration) <-chan bool {

	waiter := make(chan bool, 1)

	select {
	case <-time.After(timeout):
		waiter <- false
		close(waiter)
	case <-sig:
		waiter <- true
		close(waiter)
	}
	return waiter
}
