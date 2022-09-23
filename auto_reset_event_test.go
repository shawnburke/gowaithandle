package gowaithandle

import (
	"math/rand"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAutoSimple(t *testing.T) {

	auto := AutoResetEvent{}

	res := <-auto.WaitOne(time.Millisecond)
	require.False(t, res)

	auto.Set()

	res = <-auto.WaitOne(time.Millisecond)
	require.True(t, res)

	res = <-auto.WaitOne(time.Millisecond)
	require.False(t, res)

	done := make(chan struct{})
	go func() {
		res = <-auto.WaitOne(time.Millisecond * 50)
		require.True(t, res)
		close(done)
	}()

	time.Sleep(time.Millisecond * 5)
	auto.Set()
	<-done
	require.True(t, res)

}

func TestAutoSignaled(t *testing.T) {

	auto := NewAutoResetEvent(true)

	res := <-auto.WaitOne(time.Millisecond)
	require.True(t, res)
}

func TestAutoMulti(t *testing.T) {
	auto := AutoResetEvent{}

	counter := int32(0)
	next := make(chan int)
	n := 100
	for i := 0; i < n; i++ {
		go func(i int) {
			// create jitter
			ms := time.Duration(10 + rand.Intn(20))
			time.Sleep(ms)

			res := <-auto.WaitOne(time.Second / 2)
			if res {
				atomic.AddInt32(&counter, 1)
			}
			next <- i
		}(i)
	}

	// do all but the last one
	for i := 0; i < n-1; i++ {
		auto.Set()
		<-next
		c := int(atomic.LoadInt32(&counter))
		require.Equal(t, i+1, c)
	}

	// last one should time out
	<-next
	c := int(atomic.LoadInt32(&counter))
	require.Equal(t, n-1, c)

}
