package gowaithandle

import (
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
		ctx, cancel := timeoutContext(time.Second)
		defer cancel()
		result := <-WaitAll(ctx, mre1, mre2)
		require.True(t, result)
		atomic.StoreInt32(&count, 1)
		close(done)
	}()

	mre1.Set()
	require.Equal(t, 0, int(count))
	mre2.Set()
	<-done
	require.Equal(t, 1, int(count))
}
