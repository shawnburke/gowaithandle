package gowaithandle

import (
	"context"
	"sync"
	"sync/atomic"
)

type Semaphore struct {
	sync.Mutex
	pool    chan struct{}
	cap     int
	created int32
}

func NewSemaphore(size int) *Semaphore {
	return &Semaphore{
		cap: size,
	}
}

func (s *Semaphore) WaitOne(ctx context.Context) <-chan bool {
	if ctx == nil {
		ctx = context.Background()
	}

	done := make(chan bool)

	go func() {
		select {
		case <-ctx.Done():
			done <- false
			close(done)
		case s.getPool() <- struct{}{}:
			done <- true
			close(done)
		}
	}()
	return done
}

func (s *Semaphore) Release() int {
	p := s.getPool()
	c := len(p)
	if c > 0 {
		<-p
	}
	return c
}

func (s *Semaphore) Available() int {
	p := s.getPool()
	return cap(p) - len(p)
}

func (s *Semaphore) getPool() chan struct{} {
	if atomic.LoadInt32(&s.created) == 0 {
		s.Lock()
		if s.pool == nil {
			s.pool = make(chan struct{}, s.cap)
		}
		atomic.StoreInt32(&s.created, 1)
		s.Unlock()
	}
	return s.pool
}
