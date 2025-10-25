package fifo

import (
	"context"
	"sync"
)

// Semaphore provides a counting semaphore with FIFO ordering for goroutine access control.
// When permits are unavailable, goroutines are queued and woken in the order they called Acquire.
type Semaphore struct {
	mu    sync.Mutex
	queue []chan struct{} // FIFO queue of waiting goroutines
	avail uint            // number of available permits
}

// NewSemaphore creates a semaphore with the specified number of initial permits.
// maxConcurrent determines how many goroutines can hold permits simultaneously.
func NewSemaphore(maxConcurrent uint) *Semaphore {
	return &Semaphore{
		queue: make([]chan struct{}, 0),
		avail: maxConcurrent,
	}
}

// Acquire tries to acquire one permit in FIFO order.
// If a permit is available, it returns immediately.
// Otherwise, it queues the goroutine and blocks until:
//   - a permit becomes available (returns nil), or
//   - the context is canceled (returns ctx.Err())
func (s *Semaphore) Acquire(ctx context.Context) error {
	s.mu.Lock()
	if s.avail > 0 {
		s.avail--
		s.mu.Unlock()
		return nil
	}

	waiter := make(chan struct{})
	s.queue = append(s.queue, waiter)
	s.mu.Unlock()

	select {
	case <-waiter: // got permit
		return nil
	case <-ctx.Done(): // canceled
		s.mu.Lock()
		// Remove from queue if still waiting
		for i, c := range s.queue {
			if c == waiter {
				s.queue = append(s.queue[:i], s.queue[i+1:]...)
				break
			}
		}
		s.mu.Unlock()
		return ctx.Err()
	}
}

// Release releases one permit and wakes the next waiting goroutine in FIFO order.
// If there are goroutines waiting in the queue, the first one is woken.
// Otherwise, the permit count is incremented for future use.
func (s *Semaphore) Release() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.queue) > 0 {
		waiter := s.queue[0]
		s.queue = s.queue[1:]
		close(waiter)
	} else {
		s.avail++
	}
}

// Available returns the number of permits currently available (not held by any goroutine).
// This is primarily useful for testing and debugging.
func (s *Semaphore) Available() uint {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.avail
}

// QueueLength returns the number of goroutines waiting to acquire a permit.
// This is primarily useful for testing and debugging.
func (s *Semaphore) QueueLength() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.queue)
}
