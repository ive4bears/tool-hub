package fifo

import (
	"context"
	"sync"
)

// GroupLimiter manages semaphores for concurrent groups to control concurrency.
// Each group ID has an associated FIFO semaphore that limits the number of concurrent operations.
// The limiter caches semaphores to avoid repeated semaphore creation.
type GroupLimiter struct {
	mu     sync.RWMutex
	groups map[uint]*Semaphore // maps group ID to its semaphore
}

// DefaultGroupLimiter is a ready-to-use singleton instance for package-level convenience.
var DefaultGroupLimiter = NewGroupLimiter()

// NewGroupLimiter creates a new instance of GroupLimiter.
func NewGroupLimiter() *GroupLimiter {
	return &GroupLimiter{
		groups: make(map[uint]*Semaphore),
	}
}

// getSemaphoreFor returns the semaphore for the given group ID.
// If the semaphore doesn't exist in the cache, it creates a new one with the specified maxConcurrent.
// This method is thread-safe and handles the case where maxConcurrent is 0 (defaults to 1).
func (l *GroupLimiter) getSemaphoreFor(groupID uint, maxConcurrent uint) *Semaphore {
	l.mu.Lock()
	defer l.mu.Unlock()

	if sem, exists := l.groups[groupID]; exists {
		return sem
	}

	if maxConcurrent == 0 {
		maxConcurrent = 1
	}

	sem := NewSemaphore(maxConcurrent)
	l.groups[groupID] = sem
	return sem
}

// Acquire acquires a permit from the semaphore for the given group.
// This blocks until a permit is available or the context is canceled.
// Returns an error if the context is canceled.
func (l *GroupLimiter) Acquire(ctx context.Context, groupID uint, maxConcurrent uint) error {
	sem := l.getSemaphoreFor(groupID, maxConcurrent)
	return sem.Acquire(ctx)
}

// Release releases a previously acquired permit for the given group.
// This is a no-op if the group doesn't exist in the cache.
// It's safe to call Release even if the semaphore doesn't exist.
func (l *GroupLimiter) Release(groupID uint) {
	l.mu.RLock()
	sem, exists := l.groups[groupID]
	l.mu.RUnlock()

	if exists && sem != nil {
		sem.Release()
	}
}

// Reset removes the cached semaphore for a group, forcing it to be recreated on next Acquire.
// This is useful when the group's maxConcurrent setting has been updated.
func (l *GroupLimiter) Reset(groupID uint) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.groups, groupID)
}
