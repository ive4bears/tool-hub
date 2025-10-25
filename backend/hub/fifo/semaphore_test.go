package fifo

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewSemaphore(t *testing.T) {
	sem := NewSemaphore(5)
	assert.NotNil(t, sem)
	assert.Equal(t, uint(5), sem.avail)
	assert.Empty(t, sem.queue)
}

func TestSemaphore_AcquireRelease_Sequential(t *testing.T) {
	sem := NewSemaphore(2)
	ctx := context.Background()

	// Should acquire successfully
	err := sem.Acquire(ctx)
	assert.NoError(t, err)
	assert.Equal(t, uint(1), sem.avail)

	err = sem.Acquire(ctx)
	assert.NoError(t, err)
	assert.Equal(t, uint(0), sem.avail)

	// Release once
	sem.Release()
	assert.Equal(t, uint(1), sem.avail)

	// Release again
	sem.Release()
	assert.Equal(t, uint(2), sem.avail)
}

func TestSemaphore_AcquireBlocking(t *testing.T) {
	sem := NewSemaphore(1)
	ctx := context.Background()

	// Acquire the only permit
	err := sem.Acquire(ctx)
	assert.NoError(t, err)

	acquired := false
	var wg sync.WaitGroup
	wg.Add(1)

	// Try to acquire in goroutine - should block
	go func() {
		defer wg.Done()
		err := sem.Acquire(ctx)
		assert.NoError(t, err)
		acquired = true
	}()

	// Give goroutine time to block
	time.Sleep(50 * time.Millisecond)
	assert.False(t, acquired, "should be blocking")

	// Release to unblock
	sem.Release()
	wg.Wait()
	assert.True(t, acquired, "should have acquired after release")
}

func TestSemaphore_FIFOOrdering(t *testing.T) {
	sem := NewSemaphore(1)
	ctx := context.Background()

	// Acquire the only permit first
	err := sem.Acquire(ctx)
	assert.NoError(t, err)

	results := make([]int, 0)
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Use a channel to ensure goroutines start in order
	start := make(chan int, 5)

	// Queue up 5 goroutines - they should be queued in order
	for id := 0; id < 5; id++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Wait for signal to start
			id := <-start
			err := sem.Acquire(ctx)
			assert.NoError(t, err)
			mu.Lock()
			results = append(results, id)
			mu.Unlock()
			sem.Release()
		}()
	}

	// Send start signals in order to ensure FIFO queueing
	time.Sleep(10 * time.Millisecond) // Let goroutines reach the channel receive
	for id := 0; id < 5; id++ {
		start <- id
		time.Sleep(5 * time.Millisecond) // Small delay to ensure ordering
	}

	// Release initial permit to start the chain
	sem.Release()
	wg.Wait()

	// Verify FIFO ordering
	assert.Equal(t, []int{0, 1, 2, 3, 4}, results)
}

func TestSemaphore_ContextCancellation(t *testing.T) {
	sem := NewSemaphore(1)

	// Acquire the only permit
	ctx := context.Background()
	err := sem.Acquire(ctx)
	assert.NoError(t, err)

	// Try to acquire with canceled context
	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	err = sem.Acquire(canceledCtx)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
	assert.Empty(t, sem.queue, "queue should be empty after cancellation")
}

func TestSemaphore_ContextTimeout(t *testing.T) {
	sem := NewSemaphore(1)
	ctx := context.Background()

	// Acquire the only permit
	err := sem.Acquire(ctx)
	assert.NoError(t, err)

	// Try to acquire with timeout
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err = sem.Acquire(timeoutCtx)
	assert.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)
}

func TestSemaphore_ConcurrentAcquireRelease(t *testing.T) {
	sem := NewSemaphore(10)
	ctx := context.Background()

	const numGoroutines = 100
	var wg sync.WaitGroup
	successCount := int32(0)
	var mu sync.Mutex

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := sem.Acquire(ctx)
			if err == nil {
				mu.Lock()
				successCount++
				mu.Unlock()
				time.Sleep(10 * time.Millisecond)
				sem.Release()
			}
		}()
	}

	wg.Wait()
	assert.Equal(t, int32(numGoroutines), successCount)
	assert.Equal(t, uint(10), sem.avail, "all permits should be returned")
}

func TestSemaphore_ZeroCapacity(t *testing.T) {
	sem := NewSemaphore(0)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := sem.Acquire(ctx)
	assert.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)
}

func TestSemaphore_AvailableAndQueueLength(t *testing.T) {
	sem := NewSemaphore(3)
	ctx := context.Background()

	// Initially all permits available, no queue
	assert.Equal(t, uint(3), sem.Available())
	assert.Equal(t, 0, sem.QueueLength())

	// Acquire one
	err := sem.Acquire(ctx)
	assert.NoError(t, err)
	assert.Equal(t, uint(2), sem.Available())
	assert.Equal(t, 0, sem.QueueLength())

	// Acquire all remaining
	sem.Acquire(ctx)
	sem.Acquire(ctx)
	assert.Equal(t, uint(0), sem.Available())
	assert.Equal(t, 0, sem.QueueLength())

	// Try to acquire - should queue
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		sem.Acquire(ctx)
	}()

	time.Sleep(20 * time.Millisecond)
	assert.Equal(t, uint(0), sem.Available())
	assert.Equal(t, 1, sem.QueueLength())

	// Release to unblock
	sem.Release()
	wg.Wait()

	assert.Equal(t, uint(0), sem.Available())
	assert.Equal(t, 0, sem.QueueLength())

	// Release all
	sem.Release()
	sem.Release()
	sem.Release()
	assert.Equal(t, uint(3), sem.Available())
	assert.Equal(t, 0, sem.QueueLength())
}
