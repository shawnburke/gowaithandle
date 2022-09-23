package gowaithandle

import "time"

type WaitHandle interface {
	WaitOne(time.Duration) <-chan bool
}

type EventWaitHandle interface {
	WaitHandle
	Set() bool
	Reset() bool
}
