package gowaithandle

import (
	"context"
)

type WaitHandle interface {
	WaitOne(context.Context) <-chan bool
}

type EventWaitHandle interface {
	WaitHandle
	Set() bool
	Reset() bool
}
