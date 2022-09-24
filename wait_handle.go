package gowaithandle

import (
	"context"
	"time"
)

type WaitHandle interface {
	WaitDuration(time.Duration) <-chan bool
	WaitOne(context.Context) <-chan bool
}

type EventWaitHandle interface {
	WaitHandle
	Set() bool
	Reset() bool
}
