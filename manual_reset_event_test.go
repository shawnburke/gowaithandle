package gowaithandle

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestManualSimple(t *testing.T) {

	mre := NewManualResetEvent(false)

	// test timeout
	signaled := <-mre.WaitOne(time.Millisecond)
	require.False(t, signaled)

	wg := sync.WaitGroup{}
	wg.Add(1)
	// test signal
	go func() {
		defer wg.Done()
		s := <-mre.WaitOne(time.Millisecond * 5)
		require.True(t, s)
		signaled = s
	}()
	mre.Set()
	wg.Wait()
	require.True(t, signaled)

	// test reset
	signaled = <-mre.WaitOne(time.Millisecond)
	require.True(t, signaled)
	mre.Reset()
	signaled = <-mre.WaitOne(time.Millisecond)
	require.False(t, signaled)

	mre = NewManualResetEvent(true)
	res := <-mre.WaitOne(time.Millisecond)
	require.True(t, res)
}

func TestManualMultiSignaled(t *testing.T) {

	mre := NewManualResetEvent(true)

	finished := int32(0)

	wg := sync.WaitGroup{}
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res := <-mre.WaitOne(time.Millisecond)
			if res {
				atomic.AddInt32(&finished, 1)
			}
		}()
	}
	wg.Wait()
	require.Equal(t, 4, int(finished))

}

func TestManualMultiNotSignaled(t *testing.T) {
	mre := NewManualResetEvent(true)

	finished := int32(0)

	wg := sync.WaitGroup{}
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res := <-mre.WaitOne(time.Second)
			if res {
				atomic.AddInt32(&finished, 1)
			}
		}()
	}
	mre.Set()
	wg.Wait()
	require.Equal(t, 4, int(finished))

}
