package gowaithandle

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestWaitGroup(t *testing.T) {
	wg := WaitGroup{}

	wg.Add(1)

	go func() {
		wg.Done()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	res := <-wg.WaitOne(ctx)
	require.True(t, res)
}

func TestWaitGroupTimeout(t *testing.T) {
	wg := WaitGroup{}
	wg.Add(1)
	res := <-wg.WaitOne(testTimeoutContext(time.Millisecond))
	require.False(t, res)
}
