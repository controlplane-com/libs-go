package schedule

import (
	"container/heap"
	"sync"
	"time"
)

// Heap is a generic min-heap ordered by NextExecution time.
// It is safe for concurrent use.
type Heap[T Item] struct {
	mu    sync.RWMutex
	inner innerHeap[T]
	index map[string]T // ID -> item for O(1) lookup
}

// innerHeap implements heap.Interface for container/heap operations.
// This is an internal type that doesn't have locking - the outer Heap handles that.
type innerHeap[T Item] []T

func (h innerHeap[T]) Len() int { return len(h) }

func (h innerHeap[T]) Less(i, j int) bool {
	return h[i].NextExecution().Before(h[j].NextExecution())
}

func (h innerHeap[T]) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].SetHeapIndex(i)
	h[j].SetHeapIndex(j)
}

func (h *innerHeap[T]) Push(x any) {
	item := x.(T)
	item.SetHeapIndex(len(*h))
	*h = append(*h, item)
}

func (h *innerHeap[T]) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	var zero T
	old[n-1] = zero // avoid memory leak
	item.SetHeapIndex(-1)
	*h = old[0 : n-1]
	return item
}

// NewHeap creates a new empty heap.
func NewHeap[T Item]() *Heap[T] {
	h := &Heap[T]{
		inner: make(innerHeap[T], 0),
		index: make(map[string]T),
	}
	heap.Init(&h.inner)
	return h
}

// Push adds or updates an item in the heap.
// If an item with the same ID exists, it updates the entry and fixes the heap position.
func (h *Heap[T]) Push(item T) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if existing, ok := h.index[item.ID()]; ok {
		// Update existing entry
		idx := existing.HeapIndex()
		h.inner[idx] = item
		item.SetHeapIndex(idx)
		h.index[item.ID()] = item
		heap.Fix(&h.inner, idx)
	} else {
		heap.Push(&h.inner, item)
		h.index[item.ID()] = item
	}
}

// Remove removes an item by ID. Returns true if the item was found and removed.
func (h *Heap[T]) Remove(id string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	item, ok := h.index[id]
	if !ok {
		return false
	}

	idx := item.HeapIndex()
	if idx >= 0 && idx < len(h.inner) {
		heap.Remove(&h.inner, idx)
	}
	delete(h.index, id)
	return true
}

// Peek returns the next item without removing it from the heap.
// Returns the zero value of T and false if the heap is empty.
func (h *Heap[T]) Peek() (T, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if len(h.inner) == 0 {
		var zero T
		return zero, false
	}
	return h.inner[0], true
}

// Pop removes and returns the next item from the heap.
// Returns the zero value of T and false if the heap is empty.
func (h *Heap[T]) Pop() (T, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.inner) == 0 {
		var zero T
		return zero, false
	}

	item := heap.Pop(&h.inner).(T)
	delete(h.index, item.ID())
	return item, true
}

// PopIfDue removes and returns the next item if it's due (NextExecution <= now).
// Returns the zero value of T and false if the heap is empty or the next item is not yet due.
func (h *Heap[T]) PopIfDue() (T, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.inner) == 0 {
		var zero T
		return zero, false
	}

	next := h.inner[0]
	if time.Until(next.NextExecution()) > 0 {
		var zero T
		return zero, false
	}

	item := heap.Pop(&h.inner).(T)
	delete(h.index, item.ID())
	return item, true
}

// DueState reports whether an item with the given ID is currently in the heap
// and, if so, whether it is due (NextExecution <= now). An item that has been
// popped for execution is not in the heap and so reports present=false.
//
// NextExecution is read under the heap lock: indexed items are never mutated in
// place (updates replace the index entry via Push under the same lock), so this
// is safe to call concurrently with the scheduler loop.
func (h *Heap[T]) DueState(id string) (present, due bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	item, ok := h.index[id]
	if !ok {
		return false, false
	}
	return true, !item.NextExecution().After(time.Now())
}

// Len returns the number of items in the heap.
func (h *Heap[T]) Len() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.inner)
}

// Get returns the item with the given ID, or the zero value of T and false if not found.
func (h *Heap[T]) Get(id string) (T, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	item, ok := h.index[id]
	return item, ok
}

// Snapshot returns a copy of all items currently in the heap.
// The returned slice is in arbitrary heap order; callers should sort if a
// particular order is needed.
func (h *Heap[T]) Snapshot() []T {
	h.mu.RLock()
	defer h.mu.RUnlock()

	out := make([]T, len(h.inner))
	copy(out, h.inner)
	return out
}
