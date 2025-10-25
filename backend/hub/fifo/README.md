# FIFO Semaphore and GroupLimiter

This package provides two related, lightweight concurrency primitives:

1. A counting FIFO semaphore (`Semaphore`) with strict First-In-First-Out ordering for waiting goroutines.
2. A group-based limiter (`GroupLimiter`) that manages multiple FIFO semaphores keyed by a numeric group ID. This lets you control per-group concurrency while preserving FIFO ordering for waiters.

## Features

* FIFO Ordering: Goroutines are granted permits in the order they call `Acquire()`.
* Context Support: Acquire operations can be canceled via `context.Context`.
* Thread-Safe: All operations are safe for concurrent use.
* Grouped Concurrency: `GroupLimiter` manages semaphores per group ID so each group has its own concurrency limit.
* Zero Dependencies: Only uses the Go standard library.

## Semaphore Usage

Use `Semaphore` when you need a single FIFO semaphore.

```go
package main

import (
    "context"
    "tool-hub/backend/hub/fifo"
)

func main() {
    // Create a semaphore with 3 permits
    sem := fifo.NewSemaphore(3)

    ctx := context.Background()

    // Acquire a permit
    if err := sem.Acquire(ctx); err != nil {
        // Handle error (e.g., context canceled)
        return
    }

    // Do work...

    // Release the permit
    sem.Release()
}
```

## GroupLimiter (per-group concurrency control)

`GroupLimiter` manages semaphores per `groupID`. It does not require registration: callers pass the desired `maxConcurrent` when acquiring. A package-level default instance is provided for convenience.

Key methods:

* `NewGroupLimiter() *GroupLimiter` — create a fresh instance.
* `DefaultGroupLimiter` — a ready-to-use singleton instance.
* `(*GroupLimiter) Acquire(ctx context.Context, groupID uint, maxConcurrent uint) error` — acquire a permit for `groupID`; blocks until a permit is available or `ctx` is done.
* `(*GroupLimiter) Release(groupID uint)` — release a previously acquired permit for `groupID`.
* `(*GroupLimiter) Reset(groupID uint)` — remove the cached semaphore for `groupID` (next acquire recreates it).

Example using the default limiter:

```go
package main

import (
    "context"
    "time"
    "tool-hub/backend/hub/fifo"
)

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()

    // Acquire one of up to 3 concurrent slots for group 42
    if err := fifo.DefaultGroupLimiter.Acquire(ctx, 42, 3); err != nil {
        // handle ctx.Err()
        return
    }
    defer fifo.DefaultGroupLimiter.Release(42)

    // ... do work ...
}
```

## Notes and testing

* `Acquire` may return `context.DeadlineExceeded` or `context.Canceled` when the provided context is done.
* Prefer using `context.WithTimeout` in tests to avoid indefinite hangs.
* `GroupLimiter`'s `Reset` is useful when you update a group's concurrency limit and want the new setting to take effect.

## Implementation details

The `Semaphore` maintains:

* a count of available permits (`avail`)
* a FIFO queue of waiting goroutines (`queue`)
* a mutex to protect internal state (`mu`)

When `Acquire()` is called:

1. If permits are available, one is taken immediately.
2. Otherwise, the goroutine is added to the queue and blocks.
3. When a permit is released, the first waiting goroutine is woken.

This ensures strict FIFO ordering even under high contention.
