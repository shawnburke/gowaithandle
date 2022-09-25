package gowaithandle

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSemaphoreSimple(t *testing.T) {

	// Ensure the semaphore correctly works, rejects
	// waits when full, then correctly waits after a release

	s := NewSemaphore(2)

	// fill it up
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	res := <-s.WaitOne(ctx)
	require.True(t, res)

	res = <-s.WaitOne(ctx)
	require.True(t, res)

	// try again and expect it to timeout
	res = <-s.WaitOne(ctx)
	require.False(t, res)

	// release then add another item
	s.Release()

	ctx, cancel = context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	res = <-s.WaitOne(ctx)
	require.True(t, res)

}

func TestSemaphoreRelease(t *testing.T) {

	// ensure that once full, the semaphore
	// blocks and then resumes properly
	// when an item is released

	n := 3
	s := NewSemaphore(n)

	// fill it up
	for i := 0; i < n; i++ {
		<-s.WaitOne(context.Background())
	}

	require.Equal(t, 0, s.Available())

	wg := WaitGroup{}
	wg.Add(1)
	ok := false
	go func() {
		defer wg.Done()
		ok = <-s.WaitOne(context.Background())
	}()

	s.Release()
	wg.Wait()
	require.True(t, ok)
}

func TestSemaphoreTimeout(t *testing.T) {

	// ensure that timeout behavior is correct
	// with a full sempahore

	s := NewSemaphore(1)

	// fill it up
	res := <-s.WaitOne(context.Background())
	require.True(t, res)

	require.Equal(t, 0, s.Available())

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	// make sure it times out and does NOT take a slot
	res = <-s.WaitOne(ctx)
	require.False(t, res)

	require.Equal(t, 0, s.Available())

	// release and verify the cancelled channel put
	// did not happen
	s.Release()

	require.Equal(t, 1, s.Available())

}
