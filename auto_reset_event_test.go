package gowaithandle

import (
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
	for i := 0; i < 4; i++ {
		go func(i int) {
			res := <-auto.WaitOne(time.Millisecond * 10)
			if res {
				atomic.AddInt32(&counter, 1)
			}
			next <- i
		}(i)
	}

	auto.Set()
	<-next
	require.Equal(t, 1, int(counter))
	auto.Set()
	<-next
	require.Equal(t, 2, int(counter))
	auto.Set()
	<-next
	require.Equal(t, 3, int(counter))

	// last one should time out
	<-next
	require.Equal(t, 3, int(counter))

}
