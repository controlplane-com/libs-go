package schedule

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestScheduler_Add(t *testing.T) {
	executor := func(ctx context.Context, item *testItem) error {
		return nil
	}

	s := NewScheduler(executor, DefaultConfig())

	item := &testItem{
		id:        "test1",
		schedule:  "* * * * *", // Every minute
		heapIndex: -1,
	}

	err := s.Add(item)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Item should be in the heap
	if s.Count() != 1 {
		t.Errorf("expected count 1, got %d", s.Count())
	}

	// NextExecution should be set
	if item.NextExecution().IsZero() {
		t.Error("expected NextExecution to be set")
	}

	// NextExecution should be in the future (within ~1 minute)
	if item.NextExecution().Before(time.Now()) {
		t.Error("expected NextExecution to be in the future")
	}
}

func TestScheduler_AddInvalidCron(t *testing.T) {
	executor := func(ctx context.Context, item *testItem) error {
		return nil
	}

	s := NewScheduler(executor, DefaultConfig())

	item := &testItem{
		id:        "test1",
		schedule:  "invalid cron",
		heapIndex: -1,
	}

	err := s.Add(item)
	if err == nil {
		t.Error("expected error for invalid cron expression")
	}
}

func TestScheduler_Remove(t *testing.T) {
	executor := func(ctx context.Context, item *testItem) error {
		return nil
	}

	s := NewScheduler(executor, DefaultConfig())

	item := &testItem{
		id:        "test1",
		schedule:  "* * * * *",
		heapIndex: -1,
	}

	s.Add(item)
	if s.Count() != 1 {
		t.Fatalf("expected count 1, got %d", s.Count())
	}

	s.Remove("test1")
	if s.Count() != 0 {
		t.Errorf("expected count 0, got %d", s.Count())
	}
}

func TestScheduler_Get(t *testing.T) {
	executor := func(ctx context.Context, item *testItem) error {
		return nil
	}

	s := NewScheduler(executor, DefaultConfig())

	item := &testItem{
		id:        "test1",
		schedule:  "* * * * *",
		heapIndex: -1,
	}

	s.Add(item)

	got, ok := s.Get("test1")
	if !ok {
		t.Fatal("expected to get item")
	}
	if got.ID() != "test1" {
		t.Errorf("expected test1, got %s", got.ID())
	}

	_, ok = s.Get("nonexistent")
	if ok {
		t.Error("expected Get to return false for nonexistent item")
	}
}

func TestScheduler_ExecutesOnSchedule(t *testing.T) {
	var mu sync.Mutex
	var executed []string

	executor := func(ctx context.Context, item *testItem) error {
		mu.Lock()
		executed = append(executed, item.ID())
		mu.Unlock()
		return nil
	}

	config := Config{
		MissedWindow: 5 * time.Minute,
		IdleInterval: 100 * time.Millisecond,
	}

	s := NewScheduler(executor, config)

	// Add an item that's already past due
	item := &testItem{
		id:            "test1",
		schedule:      "* * * * *",
		nextExecution: time.Now().Add(-1 * time.Second), // Past due
		heapIndex:     -1,
	}

	s.heap.Push(item)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	// Run scheduler in background
	go s.Run(ctx)

	// Wait for execution
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	if len(executed) == 0 {
		t.Error("expected item to be executed")
	} else if executed[0] != "test1" {
		t.Errorf("expected test1 to be executed, got %s", executed[0])
	}
	mu.Unlock()
}

func TestScheduler_ReschedulesAfterExecution(t *testing.T) {
	executedCh := make(chan struct{}, 1)
	doneCh := make(chan struct{})

	executor := func(ctx context.Context, item *testItem) error {
		select {
		case executedCh <- struct{}{}:
		default:
		}
		return nil
	}

	config := Config{
		MissedWindow: 5 * time.Minute,
		IdleInterval: 10 * time.Millisecond,
	}

	s := NewScheduler(executor, config)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		s.Run(ctx)
		close(doneCh)
	}()

	// Add an item that's past due (use Add to trigger wakeup)
	item := &testItem{
		id:            "test1",
		schedule:      "* * * * *",
		nextExecution: time.Now().Add(-1 * time.Second),
		heapIndex:     -1,
	}
	s.heap.Push(item)
	s.tryWakeup()

	// Wait for execution signal
	select {
	case <-executedCh:
		// Item was executed
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for item to be executed")
	}

	// Give scheduler time to reschedule, then stop it
	time.Sleep(10 * time.Millisecond)
	cancel()
	s.tryWakeup() // Wake scheduler so it sees cancellation promptly

	// Wait for scheduler goroutine to fully exit
	select {
	case <-doneCh:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("scheduler did not stop promptly after cancel")
	}

	// Now safe to check item state - scheduler is stopped
	if s.Count() != 1 {
		t.Errorf("expected item to be rescheduled, count is %d", s.Count())
	}

	got, ok := s.Get("test1")
	if !ok {
		t.Fatal("expected to get item")
	}
	if got.NextExecution().Before(time.Now()) {
		t.Error("expected next execution to be in the future after reschedule")
	}
}

func TestScheduler_SkipsMissedExecutions(t *testing.T) {
	var execCount int32
	doneCh := make(chan struct{})

	executor := func(ctx context.Context, item *testItem) error {
		atomic.AddInt32(&execCount, 1)
		return nil
	}

	config := Config{
		MissedWindow: 1 * time.Second, // Very short missed window
		IdleInterval: 50 * time.Millisecond,
	}

	s := NewScheduler(executor, config)

	// Record original next execution time
	originalNextExec := time.Now().Add(-10 * time.Second)

	// Add an item that's way past due (beyond missed window)
	item := &testItem{
		id:            "test1",
		schedule:      "* * * * *",
		nextExecution: originalNextExec, // Missed by 10s
		heapIndex:     -1,
	}

	s.heap.Push(item)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		s.Run(ctx)
		close(doneCh)
	}()

	// Wait for the scheduler to process the item
	// We detect this by waiting for the item's next execution to change
	deadline := time.Now().Add(300 * time.Millisecond)
	processed := false
	for time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
		// Check if item was rescheduled by looking at whether nextExecution changed
		// We can't read item directly due to races, but we can infer from count temporarily dropping
		// Actually, let's just wait a bit for scheduler to process
	}

	// Give scheduler enough time to process, then stop it
	time.Sleep(100 * time.Millisecond)
	cancel()
	<-doneCh
	_ = processed

	// Now safe to check state
	if atomic.LoadInt32(&execCount) != 0 {
		t.Errorf("expected missed execution to be skipped, but got %d executions", execCount)
	}

	// Item should still be rescheduled
	if s.Count() != 1 {
		t.Errorf("expected item to be rescheduled, count is %d", s.Count())
	}

	// Verify it was rescheduled to a future time
	got, ok := s.Get("test1")
	if !ok {
		t.Fatal("expected item to be in heap")
	}
	if got.NextExecution().Before(time.Now()) {
		t.Error("expected next execution to be in the future after reschedule")
	}
	// Also verify it actually changed from the original
	if got.NextExecution().Equal(originalNextExec) {
		t.Error("expected next execution to be updated from original")
	}
}

func TestScheduler_LeadershipCheck(t *testing.T) {
	var execCount int32
	doneCh := make(chan struct{})

	executor := func(ctx context.Context, item *testItem) error {
		atomic.AddInt32(&execCount, 1)
		return nil
	}

	config := Config{
		MissedWindow: 5 * time.Minute,
		IdleInterval: 50 * time.Millisecond,
	}

	s := NewScheduler(executor, config)

	// Set as non-leader
	s.SetIsLeader(func() bool { return false })

	// Record original next execution time
	originalNextExec := time.Now().Add(-1 * time.Second)

	// Add an item that's past due
	item := &testItem{
		id:            "test1",
		schedule:      "* * * * *",
		nextExecution: originalNextExec,
		heapIndex:     -1,
	}

	s.heap.Push(item)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		s.Run(ctx)
		close(doneCh)
	}()

	// Give scheduler enough time to process, then stop it
	time.Sleep(100 * time.Millisecond)
	cancel()
	<-doneCh

	// Now safe to check state
	if atomic.LoadInt32(&execCount) != 0 {
		t.Errorf("expected non-leader to skip execution, but got %d executions", execCount)
	}

	// Item should still be rescheduled (even non-leaders reschedule)
	if s.Count() != 1 {
		t.Errorf("expected item to be rescheduled by non-leader, count is %d", s.Count())
	}

	// Verify it was rescheduled to a future time
	got, ok := s.Get("test1")
	if !ok {
		t.Fatal("expected item to be in heap")
	}
	if got.NextExecution().Before(time.Now()) {
		t.Error("expected next execution to be in the future after reschedule")
	}
	// Also verify it actually changed from the original
	if got.NextExecution().Equal(originalNextExec) {
		t.Error("expected next execution to be updated from original")
	}
}

func TestScheduler_WakeupOnAdd(t *testing.T) {
	var executed []string
	var mu sync.Mutex

	executor := func(ctx context.Context, item *testItem) error {
		mu.Lock()
		executed = append(executed, item.ID())
		mu.Unlock()
		return nil
	}

	config := Config{
		MissedWindow: 5 * time.Minute,
		IdleInterval: 1 * time.Hour, // Very long idle (shouldn't be reached)
	}

	s := NewScheduler(executor, config)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	go s.Run(ctx)

	// Wait a bit, then add an item that's past due
	time.Sleep(50 * time.Millisecond)

	item := &testItem{
		id:            "test1",
		schedule:      "* * * * *",
		nextExecution: time.Now().Add(-1 * time.Second),
		heapIndex:     -1,
	}
	s.heap.Push(item)
	s.tryWakeup() // Manually wake up since we bypassed Add

	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	if len(executed) == 0 {
		t.Error("expected wakeup to trigger execution")
	}
	mu.Unlock()
}

func TestScheduler_ItemResolver(t *testing.T) {
	// Track which item was rescheduled
	var rescheduledSchedule string
	var mu sync.Mutex
	doneCh := make(chan struct{})

	executor := func(ctx context.Context, item *testItem) error {
		// Simulate work that takes some time
		time.Sleep(10 * time.Millisecond)
		return nil
	}

	config := Config{
		MissedWindow: 5 * time.Minute,
		IdleInterval: 50 * time.Millisecond,
	}

	s := NewScheduler(executor, config)

	// Create a "latest items" map to simulate the regulator's behavior
	latestItems := make(map[string]*testItem)
	var latestMu sync.RWMutex

	// Set item resolver to return the latest version
	s.SetItemResolver(func(item *testItem) *testItem {
		latestMu.RLock()
		defer latestMu.RUnlock()
		if latest, ok := latestItems[item.ID()]; ok {
			mu.Lock()
			rescheduledSchedule = latest.Schedule()
			mu.Unlock()
			return latest
		}
		mu.Lock()
		rescheduledSchedule = item.Schedule()
		mu.Unlock()
		return item
	})

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		s.Run(ctx)
		close(doneCh)
	}()

	// Add an item that's past due with schedule "*/5 * * * *"
	originalItem := &testItem{
		id:            "test1",
		schedule:      "*/5 * * * *", // Every 5 minutes
		nextExecution: time.Now().Add(-1 * time.Second),
		heapIndex:     -1,
	}
	s.heap.Push(originalItem)
	s.tryWakeup()

	// Simulate an update arriving during execution - update the latest items map
	// with a new schedule "*/10 * * * *"
	time.Sleep(5 * time.Millisecond) // Give executor time to start
	updatedItem := &testItem{
		id:        "test1",
		schedule:  "*/10 * * * *", // Every 10 minutes
		heapIndex: -1,
	}
	latestMu.Lock()
	latestItems["test1"] = updatedItem
	latestMu.Unlock()

	// Wait for execution and rescheduling to complete
	time.Sleep(100 * time.Millisecond)
	cancel()
	<-doneCh

	// Verify the updated schedule was used for rescheduling
	mu.Lock()
	if rescheduledSchedule != "*/10 * * * *" {
		t.Errorf("expected rescheduled schedule to be '*/10 * * * *', got '%s'", rescheduledSchedule)
	}
	mu.Unlock()

	// Verify item is still in the heap
	if s.Count() != 1 {
		t.Errorf("expected 1 item in heap, got %d", s.Count())
	}
}

func TestScheduler_AddImpossibleSchedule(t *testing.T) {
	// Clear the cache to ensure test isolation
	ClearRejectedSchedulesCache()

	testCases := []struct {
		name     string
		schedule string
		wantErr  bool
	}{
		{"Feb 31", "1 1 31 2 *", true},
		{"Feb 30", "0 0 30 2 *", true},
		{"Apr 31", "0 0 31 4 *", true},
		{"Jun 31", "0 0 31 6 *", true},
		{"Sep 31", "0 0 31 9 *", true},
		{"Nov 31", "0 0 31 11 *", true},
		{"Feb 29 (valid - leap years)", "0 0 29 2 *", false},
		{"Jan 31 (valid)", "0 0 31 1 *", false},
		{"Every minute (valid)", "* * * * *", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			executor := func(ctx context.Context, item *testItem) error { return nil }
			s := NewScheduler(executor, DefaultConfig())
			item := &testItem{id: "test", schedule: tc.schedule, heapIndex: -1}

			err := s.Add(item)
			if tc.wantErr && err == nil {
				t.Errorf("expected error for schedule %q", tc.schedule)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error for schedule %q: %v", tc.schedule, err)
			}
		})
	}
}

func TestScheduler_ImpossibleScheduleCache(t *testing.T) {
	// Clear the cache before testing
	ClearRejectedSchedulesCache()

	executor := func(ctx context.Context, item *testItem) error { return nil }
	s := NewScheduler(executor, DefaultConfig())

	impossibleSchedule := "1 1 31 2 *" // February 31st

	// Verify schedule is not in cache initially
	if IsScheduleRejected(impossibleSchedule) {
		t.Error("schedule should not be in cache initially")
	}

	// First Add should reject and cache
	item := &testItem{id: "test1", schedule: impossibleSchedule, heapIndex: -1}
	err := s.Add(item)
	if err == nil {
		t.Fatal("expected error for impossible schedule")
	}

	// Verify schedule is now in cache
	if !IsScheduleRejected(impossibleSchedule) {
		t.Error("schedule should be in cache after rejection")
	}

	// Second Add should use cache and still reject
	item2 := &testItem{id: "test2", schedule: impossibleSchedule, heapIndex: -1}
	err = s.Add(item2)
	if err == nil {
		t.Fatal("expected error for cached impossible schedule")
	}

	// Valid schedules should not be cached
	validSchedule := "* * * * *"
	if IsScheduleRejected(validSchedule) {
		t.Error("valid schedule should not be in cache")
	}

	// Clean up
	ClearRejectedSchedulesCache()
}
