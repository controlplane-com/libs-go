package schedule

import (
	"testing"
	"time"
)

// testItem is a simple implementation of Item for testing.
type testItem struct {
	id            string
	schedule      string
	nextExecution time.Time
	heapIndex     int
}

func (t *testItem) ID() string                      { return t.id }
func (t *testItem) Schedule() string                { return t.schedule }
func (t *testItem) NextExecution() time.Time        { return t.nextExecution }
func (t *testItem) SetNextExecution(next time.Time) { t.nextExecution = next }
func (t *testItem) HeapIndex() int                  { return t.heapIndex }
func (t *testItem) SetHeapIndex(idx int)            { t.heapIndex = idx }

func newTestItem(id string, nextExec time.Time) *testItem {
	return &testItem{
		id:            id,
		schedule:      "* * * * *",
		nextExecution: nextExec,
		heapIndex:     -1,
	}
}

func TestHeap_PushAndPeek(t *testing.T) {
	h := NewHeap[*testItem]()

	now := time.Now()
	item1 := newTestItem("item1", now.Add(10*time.Minute))
	item2 := newTestItem("item2", now.Add(5*time.Minute))
	item3 := newTestItem("item3", now.Add(15*time.Minute))

	h.Push(item1)
	h.Push(item2)
	h.Push(item3)

	// Peek should return the earliest item (item2)
	next, ok := h.Peek()
	if !ok {
		t.Fatal("expected to peek an item")
	}
	if next.ID() != "item2" {
		t.Errorf("expected item2 to be next, got %s", next.ID())
	}

	// Len should be 3
	if h.Len() != 3 {
		t.Errorf("expected len 3, got %d", h.Len())
	}
}

func TestHeap_Pop(t *testing.T) {
	h := NewHeap[*testItem]()

	now := time.Now()
	item1 := newTestItem("item1", now.Add(10*time.Minute))
	item2 := newTestItem("item2", now.Add(5*time.Minute))
	item3 := newTestItem("item3", now.Add(15*time.Minute))

	h.Push(item1)
	h.Push(item2)
	h.Push(item3)

	// Pop should return items in order of next execution
	popped, ok := h.Pop()
	if !ok {
		t.Fatal("expected to pop an item")
	}
	if popped.ID() != "item2" {
		t.Errorf("expected item2, got %s", popped.ID())
	}

	popped, ok = h.Pop()
	if !ok {
		t.Fatal("expected to pop an item")
	}
	if popped.ID() != "item1" {
		t.Errorf("expected item1, got %s", popped.ID())
	}

	popped, ok = h.Pop()
	if !ok {
		t.Fatal("expected to pop an item")
	}
	if popped.ID() != "item3" {
		t.Errorf("expected item3, got %s", popped.ID())
	}

	// Heap should be empty
	if h.Len() != 0 {
		t.Errorf("expected len 0, got %d", h.Len())
	}

	_, ok = h.Pop()
	if ok {
		t.Error("expected Pop to return false on empty heap")
	}
}

func TestHeap_PopIfDue(t *testing.T) {
	h := NewHeap[*testItem]()

	now := time.Now()
	item1 := newTestItem("item1", now.Add(10*time.Second))  // Future
	item2 := newTestItem("item2", now.Add(-5*time.Second))  // Past due
	item3 := newTestItem("item3", now.Add(-10*time.Second)) // Past due (older)

	h.Push(item1)
	h.Push(item2)
	h.Push(item3)

	// PopIfDue should return item3 first (earliest/most overdue)
	popped, ok := h.PopIfDue()
	if !ok {
		t.Fatal("expected to pop due item")
	}
	if popped.ID() != "item3" {
		t.Errorf("expected item3, got %s", popped.ID())
	}

	// PopIfDue should return item2 next
	popped, ok = h.PopIfDue()
	if !ok {
		t.Fatal("expected to pop due item")
	}
	if popped.ID() != "item2" {
		t.Errorf("expected item2, got %s", popped.ID())
	}

	// PopIfDue should NOT return item1 (it's in the future)
	_, ok = h.PopIfDue()
	if ok {
		t.Error("expected PopIfDue to return false for future item")
	}

	// item1 should still be in heap
	if h.Len() != 1 {
		t.Errorf("expected 1 item remaining, got %d", h.Len())
	}
}

func TestHeap_Remove(t *testing.T) {
	h := NewHeap[*testItem]()

	now := time.Now()
	item1 := newTestItem("item1", now.Add(10*time.Minute))
	item2 := newTestItem("item2", now.Add(5*time.Minute))
	item3 := newTestItem("item3", now.Add(15*time.Minute))

	h.Push(item1)
	h.Push(item2)
	h.Push(item3)

	// Remove item2 (the next one)
	ok := h.Remove("item2")
	if !ok {
		t.Error("expected Remove to return true")
	}

	// Next should now be item1
	next, ok := h.Peek()
	if !ok {
		t.Fatal("expected to peek an item")
	}
	if next.ID() != "item1" {
		t.Errorf("expected item1 to be next, got %s", next.ID())
	}

	// Len should be 2
	if h.Len() != 2 {
		t.Errorf("expected len 2, got %d", h.Len())
	}

	// Remove non-existent item
	ok = h.Remove("nonexistent")
	if ok {
		t.Error("expected Remove to return false for non-existent item")
	}
}

func TestHeap_Get(t *testing.T) {
	h := NewHeap[*testItem]()

	now := time.Now()
	item1 := newTestItem("item1", now.Add(10*time.Minute))

	h.Push(item1)

	// Get existing item
	got, ok := h.Get("item1")
	if !ok {
		t.Fatal("expected Get to return true")
	}
	if got.ID() != "item1" {
		t.Errorf("expected item1, got %s", got.ID())
	}

	// Get non-existent item
	_, ok = h.Get("nonexistent")
	if ok {
		t.Error("expected Get to return false for non-existent item")
	}
}

func TestHeap_PushUpdate(t *testing.T) {
	h := NewHeap[*testItem]()

	now := time.Now()
	item1 := newTestItem("item1", now.Add(10*time.Minute))
	item2 := newTestItem("item2", now.Add(5*time.Minute))

	h.Push(item1)
	h.Push(item2)

	// Next should be item2
	next, _ := h.Peek()
	if next.ID() != "item2" {
		t.Errorf("expected item2, got %s", next.ID())
	}

	// Update item1 to have an earlier time
	item1Updated := newTestItem("item1", now.Add(1*time.Minute))
	h.Push(item1Updated)

	// Len should still be 2 (update, not add)
	if h.Len() != 2 {
		t.Errorf("expected len 2, got %d", h.Len())
	}

	// Next should now be item1
	next, _ = h.Peek()
	if next.ID() != "item1" {
		t.Errorf("expected item1, got %s", next.ID())
	}
}

func TestHeap_HeapIndex(t *testing.T) {
	h := NewHeap[*testItem]()

	now := time.Now()
	item1 := newTestItem("item1", now.Add(10*time.Minute))
	item2 := newTestItem("item2", now.Add(5*time.Minute))

	// Before push, heap index should be -1
	if item1.HeapIndex() != -1 {
		t.Errorf("expected heap index -1 before push, got %d", item1.HeapIndex())
	}

	h.Push(item1)
	h.Push(item2)

	// After push, heap index should be >= 0
	if item1.HeapIndex() < 0 {
		t.Errorf("expected heap index >= 0 after push, got %d", item1.HeapIndex())
	}
	if item2.HeapIndex() < 0 {
		t.Errorf("expected heap index >= 0 after push, got %d", item2.HeapIndex())
	}

	// item2 should be at index 0 (next to pop)
	if item2.HeapIndex() != 0 {
		t.Errorf("expected item2 at index 0, got %d", item2.HeapIndex())
	}

	// After pop, heap index should be -1
	h.Pop()
	if item2.HeapIndex() != -1 {
		t.Errorf("expected heap index -1 after pop, got %d", item2.HeapIndex())
	}
}
