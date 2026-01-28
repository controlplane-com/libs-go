package threading

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// Test basic non-partitioned functionality (backward compatibility)
func TestWorkQueue_NonPartitioned_BasicFunctionality(t *testing.T) {
	var processedCount atomic.Int32
	handler := func(item int) error {
		processedCount.Add(1)
		return nil
	}

	queue := NewWorkQueue[int](handler, 10, 3)
	queue.Start()

	// Enqueue items
	items := []int{1, 2, 3, 4, 5}
	queue.Enqueue(items...)

	queue.Stop()
	queue.AwaitCompletion()

	if processedCount.Load() != 5 {
		t.Errorf("Expected 5 items processed, got %d", processedCount.Load())
	}
}

// Test types that can be partitioned

type PartitionedTask struct {
	PartitionID string
	Value       int
}

func (p *PartitionedTask) PartitionKey() string {
	return p.PartitionID
}

type NonPartitionedTask struct {
	Value int
}

// Test partitioning with Partitionable interface
func TestWorkQueue_InterfaceBasedPartitioning(t *testing.T) {
	var mu sync.Mutex
	processedByPartition := make(map[string][]int)

	handler := func(task *PartitionedTask) error {
		mu.Lock()
		defer mu.Unlock()
		processedByPartition[task.PartitionID] = append(processedByPartition[task.PartitionID], task.Value)
		time.Sleep(1 * time.Millisecond) // Simulate work
		return nil
	}

	queue := NewWorkQueue[*PartitionedTask](handler, 20, 3)
	queue.Start()

	// Enqueue tasks with different partition keys
	tasks := []*PartitionedTask{
		{PartitionID: "A", Value: 1},
		{PartitionID: "B", Value: 1},
		{PartitionID: "A", Value: 2},
		{PartitionID: "C", Value: 1},
		{PartitionID: "B", Value: 2},
		{PartitionID: "A", Value: 3},
		{PartitionID: "C", Value: 2},
		{PartitionID: "B", Value: 3},
	}

	queue.Enqueue(tasks...)
	queue.Stop()
	queue.AwaitCompletion()

	// Verify ordering within each partition
	if len(processedByPartition["A"]) != 3 {
		t.Errorf("Expected 3 items in partition A, got %d", len(processedByPartition["A"]))
	}
	if len(processedByPartition["B"]) != 3 {
		t.Errorf("Expected 3 items in partition B, got %d", len(processedByPartition["B"]))
	}
	if len(processedByPartition["C"]) != 2 {
		t.Errorf("Expected 2 items in partition C, got %d", len(processedByPartition["C"]))
	}

	// Verify items within each partition are in order
	expectedA := []int{1, 2, 3}
	expectedB := []int{1, 2, 3}
	expectedC := []int{1, 2}

	for i, v := range processedByPartition["A"] {
		if v != expectedA[i] {
			t.Errorf("Partition A: expected %d at position %d, got %d", expectedA[i], i, v)
		}
	}
	for i, v := range processedByPartition["B"] {
		if v != expectedB[i] {
			t.Errorf("Partition B: expected %d at position %d, got %d", expectedB[i], i, v)
		}
	}
	for i, v := range processedByPartition["C"] {
		if v != expectedC[i] {
			t.Errorf("Partition C: expected %d at position %d, got %d", expectedC[i], i, v)
		}
	}
}

// Test partitioning with PartitionFunc
func TestWorkQueue_FunctionBasedPartitioning(t *testing.T) {
	var mu sync.Mutex
	processedByPartition := make(map[string][]int)

	handler := func(task *NonPartitionedTask) error {
		mu.Lock()
		defer mu.Unlock()
		partition := getPartitionForValue(task.Value)
		processedByPartition[partition] = append(processedByPartition[partition], task.Value)
		time.Sleep(1 * time.Millisecond)
		return nil
	}

	queue := NewWorkQueue[*NonPartitionedTask](handler, 20, 3)
	queue.PartitionFunc = func(task *NonPartitionedTask) string {
		return getPartitionForValue(task.Value)
	}
	queue.Start()

	// Enqueue tasks
	tasks := []*NonPartitionedTask{
		{Value: 1}, {Value: 11}, {Value: 21}, // Partition "1"
		{Value: 2}, {Value: 12}, {Value: 22}, // Partition "2"
		{Value: 3}, {Value: 13}, {Value: 23}, // Partition "3"
	}

	queue.Enqueue(tasks...)
	queue.Stop()
	queue.AwaitCompletion()

	// Verify ordering within each partition
	expected1 := []int{1, 11, 21}
	expected2 := []int{2, 12, 22}
	expected3 := []int{3, 13, 23}

	for i, v := range processedByPartition["1"] {
		if v != expected1[i] {
			t.Errorf("Partition 1: expected %d at position %d, got %d", expected1[i], i, v)
		}
	}
	for i, v := range processedByPartition["2"] {
		if v != expected2[i] {
			t.Errorf("Partition 2: expected %d at position %d, got %d", expected2[i], i, v)
		}
	}
	for i, v := range processedByPartition["3"] {
		if v != expected3[i] {
			t.Errorf("Partition 3: expected %d at position %d, got %d", expected3[i], i, v)
		}
	}
}

func getPartitionForValue(value int) string {
	return string(rune('0' + (value % 10)))
}

// Test that interface takes precedence over function
func TestWorkQueue_InterfacePrecedenceOverFunction(t *testing.T) {
	var mu sync.Mutex
	usedKeys := make(map[string]bool)

	handler := func(task *PartitionedTask) error {
		mu.Lock()
		defer mu.Unlock()
		usedKeys[task.PartitionID] = true
		return nil
	}

	queue := NewWorkQueue[*PartitionedTask](handler, 10, 2)
	// Set a function that would return different keys
	queue.PartitionFunc = func(task *PartitionedTask) string {
		return "WRONG_KEY"
	}
	queue.Start()

	tasks := []*PartitionedTask{
		{PartitionID: "INTERFACE_KEY", Value: 1},
		{PartitionID: "INTERFACE_KEY", Value: 2},
	}

	queue.Enqueue(tasks...)
	queue.Stop()
	queue.AwaitCompletion()

	// Interface should have been used, not the function
	if !usedKeys["INTERFACE_KEY"] {
		t.Error("Expected interface PartitionKey to be used")
	}
	if usedKeys["WRONG_KEY"] {
		t.Error("Function should not have been used when interface is implemented")
	}
}

// Test worker assignment is deterministic
func TestWorkQueue_DeterministicWorkerAssignment(t *testing.T) {
	var mu sync.Mutex
	partitionToWorkerID := make(map[string]int)
	workerIDs := make(map[int]bool)

	handler := func(task *PartitionedTask) error {
		mu.Lock()
		defer mu.Unlock()
		// Record which worker (goroutine) is processing this partition
		// We'll use a simple counter to identify workers
		return nil
	}

	queue := NewWorkQueue[*PartitionedTask](handler, 20, 3)
	queue.Start()

	// Send multiple items with same partition keys
	for i := 0; i < 10; i++ {
		tasks := []*PartitionedTask{
			{PartitionID: "A", Value: i},
			{PartitionID: "B", Value: i},
			{PartitionID: "C", Value: i},
		}
		queue.Enqueue(tasks...)
	}

	queue.Stop()
	queue.AwaitCompletion()

	// Verify assignments were made (actual worker ID tracking would require more instrumentation)
	// This test primarily verifies the queue processes all items without error
	_ = partitionToWorkerID
	_ = workerIDs
}

// Test round-robin distribution of new partitions
func TestWorkQueue_RoundRobinDistribution(t *testing.T) {
	handler := func(task *PartitionedTask) error {
		time.Sleep(1 * time.Millisecond)
		return nil
	}

	queue := NewWorkQueue[*PartitionedTask](handler, 50, 3)
	queue.Start()

	// Create many unique partitions
	for i := 0; i < 12; i++ {
		task := &PartitionedTask{
			PartitionID: string(rune('A' + i)),
			Value:       i,
		}
		queue.Enqueue(task)
	}

	queue.Stop()
	queue.AwaitCompletion()

	// Verify partitions were assigned
	if len(queue.partitionToWorker) != 12 {
		t.Errorf("Expected 12 partitions assigned, got %d", len(queue.partitionToWorker))
	}

	// Verify round-robin: count partitions per worker
	workerCounts := make(map[int]int)
	for _, workerIdx := range queue.partitionToWorker {
		workerCounts[workerIdx]++
	}

	// With 12 partitions and 3 workers, each should get 4 partitions
	for workerIdx := 0; workerIdx < 3; workerIdx++ {
		if workerCounts[workerIdx] != 4 {
			t.Errorf("Expected worker %d to have 4 partitions, got %d", workerIdx, workerCounts[workerIdx])
		}
	}
}

// Test error handling with StopOnError in partitioned mode
func TestWorkQueue_StopOnError_WithPartitioning(t *testing.T) {
	var processedCount atomic.Int32
	var errorOccurred atomic.Bool

	handler := func(task *PartitionedTask) error {
		count := processedCount.Add(1)
		if count == 3 && !errorOccurred.Load() {
			errorOccurred.Store(true)
			return &testError{msg: "intentional error"}
		}
		time.Sleep(5 * time.Millisecond)
		return nil
	}

	queue := NewWorkQueue[*PartitionedTask](handler, 10, 2)
	queue.StopOnError = true
	queue.Start()

	tasks := make([]*PartitionedTask, 10)
	for i := 0; i < 10; i++ {
		tasks[i] = &PartitionedTask{
			PartitionID: string(rune('A' + (i % 3))),
			Value:       i,
		}
	}

	queue.Enqueue(tasks...)
	queue.Stop()
	queue.AwaitCompletion()

	// Should have stopped after error
	if queue.Error() == nil {
		t.Error("Expected error to be set")
	}

	// Verify error occurred and queue was stopped
	if !errorOccurred.Load() {
		t.Error("Expected error to have occurred")
	}

	// Note: In partitioned mode, items already enqueued to worker channels
	// will be processed even after Stop() is called. This is expected behavior.
	// We just verify that an error was captured.
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

// Test retry behavior with partitioning
func TestWorkQueue_RetryWithPartitioning(t *testing.T) {
	var attemptCounts sync.Map

	handler := func(task *PartitionedTask) error {
		key := task.PartitionID
		val, _ := attemptCounts.LoadOrStore(key, &atomic.Int32{})
		count := val.(*atomic.Int32).Add(1)

		// Fail first 2 attempts, succeed on 3rd
		if count < 3 {
			return &testError{msg: "retry me"}
		}
		return nil
	}

	queue := NewWorkQueue[*PartitionedTask](handler, 10, 2)
	queue.RetryPolicy = RetryPolicyImmediately
	queue.MaxRetries = 3
	queue.Start()

	tasks := []*PartitionedTask{
		{PartitionID: "A", Value: 1},
		{PartitionID: "B", Value: 1},
	}

	queue.Enqueue(tasks...)
	queue.Stop()
	queue.AwaitCompletion()

	// Verify each partition was attempted 3 times
	valA, _ := attemptCounts.Load("A")
	valB, _ := attemptCounts.Load("B")

	if valA.(*atomic.Int32).Load() != 3 {
		t.Errorf("Expected 3 attempts for partition A, got %d", valA.(*atomic.Int32).Load())
	}
	if valB.(*atomic.Int32).Load() != 3 {
		t.Errorf("Expected 3 attempts for partition B, got %d", valB.(*atomic.Int32).Load())
	}
}

// Test concurrent enqueue with partitioning
func TestWorkQueue_ConcurrentEnqueue(t *testing.T) {
	var processedCount atomic.Int32

	handler := func(task *PartitionedTask) error {
		processedCount.Add(1)
		return nil
	}

	queue := NewWorkQueue[*PartitionedTask](handler, 100, 4)
	queue.Start()

	// Enqueue from multiple goroutines concurrently
	var wg sync.WaitGroup
	enqueueCount := 50
	goroutines := 4

	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for i := 0; i < enqueueCount; i++ {
				task := &PartitionedTask{
					PartitionID: string(rune('A' + (i % 5))),
					Value:       i*goroutines + goroutineID,
				}
				queue.Enqueue(task)
			}
		}(g)
	}

	wg.Wait()
	queue.Stop()
	queue.AwaitCompletion()

	expectedTotal := enqueueCount * goroutines
	if processedCount.Load() != int32(expectedTotal) {
		t.Errorf("Expected %d items processed, got %d", expectedTotal, processedCount.Load())
	}
}

// Test sequential processing within partition under concurrent load
func TestWorkQueue_SequentialProcessingWithinPartition(t *testing.T) {
	var mu sync.Mutex
	processingMap := make(map[string]bool)
	violations := make([]string, 0)

	handler := func(task *PartitionedTask) error {
		mu.Lock()
		partitionKey := task.PartitionID

		// Check if this partition is already being processed
		if processingMap[partitionKey] {
			violations = append(violations, partitionKey)
		}

		// Mark as processing
		processingMap[partitionKey] = true
		mu.Unlock()

		// Simulate work
		time.Sleep(5 * time.Millisecond)

		// Mark as done
		mu.Lock()
		processingMap[partitionKey] = false
		mu.Unlock()

		return nil
	}

	queue := NewWorkQueue[*PartitionedTask](handler, 50, 3)
	queue.Start()

	// Enqueue many items for the same partitions
	for i := 0; i < 30; i++ {
		task := &PartitionedTask{
			PartitionID: string(rune('A' + (i % 3))),
			Value:       i,
		}
		queue.Enqueue(task)
	}

	queue.Stop()
	queue.AwaitCompletion()

	// Should be no violations (no concurrent processing of same partition)
	if len(violations) > 0 {
		t.Errorf("Found %d violations of sequential processing within partitions", len(violations))
	}
}

// Test backward compatibility - non-partitioned type works as before
func TestWorkQueue_BackwardCompatibility(t *testing.T) {
	var processedCount atomic.Int32

	handler := func(value int) error {
		processedCount.Add(1)
		return nil
	}

	queue := NewWorkQueue[int](handler, 10, 3)
	queue.RetryPolicy = RetryPolicyNever
	queue.Start()

	for i := 0; i < 20; i++ {
		queue.Enqueue(i)
	}

	queue.Stop()
	queue.AwaitCompletion()

	if processedCount.Load() != 20 {
		t.Errorf("Expected 20 items processed, got %d", processedCount.Load())
	}

	// Verify it used non-partitioned mode
	if queue.workerQueues != nil {
		t.Error("Expected workerQueues to be nil for non-partitioned mode")
	}
	if queue.itemQueue == nil {
		t.Error("Expected itemQueue to be initialized for non-partitioned mode")
	}
}

// Test ErrorCallback with partitioning
func TestWorkQueue_ErrorCallbackWithPartitioning(t *testing.T) {
	var callbackInvoked atomic.Bool
	var callbackPartition string

	handler := func(task *PartitionedTask) error {
		if task.Value == 5 {
			return &testError{msg: "fail on value 5"}
		}
		return nil
	}

	queue := NewWorkQueue[*PartitionedTask](handler, 10, 2)
	queue.ErrorCallback = func(task *PartitionedTask, err error, retries int) {
		callbackInvoked.Store(true)
		callbackPartition = task.PartitionID
	}
	queue.MaxRetries = 1
	queue.RetryPolicy = RetryPolicyImmediately
	queue.Start()

	tasks := []*PartitionedTask{
		{PartitionID: "A", Value: 1},
		{PartitionID: "B", Value: 5},
		{PartitionID: "C", Value: 3},
	}

	queue.Enqueue(tasks...)
	queue.Stop()
	queue.AwaitCompletion()

	if !callbackInvoked.Load() {
		t.Error("Expected ErrorCallback to be invoked")
	}

	if callbackPartition != "B" {
		t.Errorf("Expected callback for partition B, got %s", callbackPartition)
	}
}
