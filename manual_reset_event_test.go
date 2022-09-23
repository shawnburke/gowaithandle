package gowaithandle

import (
	"math/rand"
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

	// tests many threads using the event
	mre := NewManualResetEvent(true)

	finished := int32(0)

	wg := sync.WaitGroup{}
	n := 100

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// set a randome timeout and random start
			wait := time.Millisecond * time.Duration(rand.Intn(20))
			time.Sleep(wait)
			res := <-mre.WaitOne(time.Second / 2)
			if res {
				atomic.AddInt32(&finished, 1)
			}
		}()
	}
	wg.Wait()
	require.Equal(t, n, int(finished))

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
