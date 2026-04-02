package schedule

import "time"

// Item represents anything that can be scheduled.
// Implementations must be pointer types to support mutation of heap index.
type Item interface {
	// ID returns a unique identifier for this item.
	ID() string
	// Schedule returns the cron expression for this item.
	Schedule() string
	// NextExecution returns when this item should next execute.
	NextExecution() time.Time
	// SetNextExecution updates the next execution time.
	SetNextExecution(time.Time)
	// HeapIndex returns the current index in the heap (-1 if not in heap).
	HeapIndex() int
	// SetHeapIndex updates the heap index.
	SetHeapIndex(int)
}
