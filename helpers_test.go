package gowaithandle

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestWaitAll(t *testing.T) {

	mre1 := &ManualResetEvent{}
	mre2 := &ManualResetEvent{}

	count := int32(0)
	done := make(chan struct{})
	go func() {
		result := <-WaitAll(testTimeoutContext(time.Second), mre1, mre2)
		require.True(t, result)
		atomic.StoreInt32(&count, 1)
		close(done)
	}()

	mre1.Set()
	require.Equal(t, 0, int(count))
	mre2.Set()
	<-done
	require.Equal(t, 1, int(count))

	mre1.Reset()
	// test timeout
	res := <-WaitAll(testTimeoutContext(time.Millisecond), mre1, mre2)
	require.False(t, res)
}

func TestWaitAny(t *testing.T) {

	mre1 := &ManualResetEvent{}
	mre2 := &ManualResetEvent{}

	count := int32(0)
	done := make(chan struct{})
	go func() {
		result := <-WaitAny(testTimeoutContext(time.Second), mre1, mre2)
		require.True(t, result)
		atomic.StoreInt32(&count, 1)
		close(done)
	}()

	mre1.Set()
	<-done
	require.Equal(t, 1, int(count))
	mre1.Reset()

	// test timeout
	res := <-WaitAny(testTimeoutContext(time.Millisecond), mre1, mre2)
	require.False(t, res)

}

type contextKey string

func testTimeoutContext(timeout time.Duration) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	// hack to avoid warning about cancel
	ctx = context.WithValue(ctx, contextKey("foo"), cancel)
	return ctx
}
