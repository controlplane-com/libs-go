package threading

import (
	"context"
	"sync"
	"time"
)

// WorkQueue provides a worker pool for processing items concurrently.
//
// Basic Usage (Non-Partitioned):
//
//	queue := threading.NewWorkQueue[int](func(item int) error {
//	    // Process item
//	    return nil
//	}, 100, 10) // bufferSize=100, workerCount=10
//	queue.Start()
//	queue.Enqueue(1, 2, 3, 4, 5)
//	queue.Stop()
//	queue.AwaitCompletion()
//
// Partitioned Usage (Interface-based):
//
//	type Task struct {
//	    AccountID string
//	    Data      []byte
//	}
//	func (t *Task) PartitionKey() string { return t.AccountID }
//
//	queue := threading.NewWorkQueue[*Task](handler, 100, 10)
//	queue.Start()
//	// All tasks with the same AccountID will be processed by the same worker in order
//
// Partitioned Usage (Function-based):
//
//	type Task struct {
//	    OrgID string
//	    Data  []byte
//	}
//
//	queue := threading.NewWorkQueue[*Task](handler, 100, 10)
//	queue.PartitionFunc = func(t *Task) string { return t.OrgID }
//	queue.Start()
//	// All tasks with the same OrgID will be processed by the same worker in order
//
// Partitioning Behavior:
//   - When partitioning is enabled, items with the same partition key are guaranteed
//     to be processed by the same worker in FIFO order.
//   - Partitions are discovered dynamically and assigned to workers using round-robin.
//   - Different partitions may be interleaved on the same worker.
//   - Partitioning is enabled when items implement Partitionable OR PartitionFunc is set.
//   - If both are present, the Partitionable interface takes precedence.
//
// Configuration:
//   - StopOnError: If true, stops processing on first error (default: false)
//   - RetryPolicy: RetryPolicyNever, RetryPolicyImmediately, or RetryPolicyAfterDuration
//   - MaxRetries: Maximum number of retry attempts (default: 0)
//   - RetryAfterDuration: Duration to wait between retries when using RetryPolicyAfterDuration
//   - ErrorCallback: Called when an item fails after all retries are exhausted
//   - PartitionFunc: Optional function to extract partition keys for ordered processing

type WorkItem[T any] struct {
	Item    T
	Retries int
}

type Handler[T any] func(T) error
type ErrorCallback[T any] func(T, error, int)

type RetryPolicy int32

const RetryPolicyNever RetryPolicy = 0
const RetryPolicyAfterDuration RetryPolicy = 1
const RetryPolicyImmediately RetryPolicy = 2

// Partitionable is an optional interface that work items can implement
// to provide partition keys for ordered processing within partitions.
// If a type implements this interface, items with the same partition key
// will always be processed by the same worker in FIFO order.
type Partitionable interface {
	PartitionKey() string
}

func NewWorkQueue[T any](handler Handler[T], bufferSize int, workerCount int) *WorkQueue[T] {
	return &WorkQueue[T]{
		handler:     handler,
		bufferSize:  bufferSize,
		workerCount: workerCount,
		RetryPolicy: RetryPolicyNever,
		m: &sync.Mutex{},
	}
}

type WorkQueue[T any] struct {
	itemQueue            chan *WorkItem[T]
	workerCount          int
	bufferSize           int
	ctx                  context.Context
	wg                   sync.WaitGroup
	running              bool
	handler              Handler[T]
	errorDuringExecution error
	StopOnError          bool
	RetryPolicy          RetryPolicy
	RetryAfterDuration   time.Duration
	MaxRetries           int
	ErrorCallback        ErrorCallback[T]
	m                    *sync.Mutex

	// Partitioning support (optional)
	// PartitionFunc is an optional function to extract partition keys from items.
	// When set (or when items implement Partitionable), items with the same
	// partition key will be processed by the same worker in FIFO order.
	PartitionFunc func(T) string

	// Internal partitioning state
	workerQueues      []chan *WorkItem[T] // One channel per worker (used when partitioning enabled)
	partitionToWorker map[string]int      // Maps partition key to assigned worker index
	partitionMutex    sync.RWMutex        // Protects partitionToWorker map
	nextWorkerIndex   int                 // Round-robin counter for assigning new partitions
}

func (q *WorkQueue[T]) Start() {
	q.running = true
	q.ctx = context.Background()
	q.errorDuringExecution = nil
	q.wg.Add(q.workerCount)

	// Check if partitioning is enabled
	if q.isPartitioningEnabled() {
		// Initialize worker queues (one per worker)
		q.workerQueues = make([]chan *WorkItem[T], q.workerCount)
		for i := 0; i < q.workerCount; i++ {
			q.workerQueues[i] = make(chan *WorkItem[T], q.bufferSize)
		}
		// Initialize partition tracking
		q.partitionToWorker = make(map[string]int)
		q.nextWorkerIndex = 0

		// Start workers with their indices
		for i := 0; i < q.workerCount; i++ {
			go q.worker(q.ctx, i)
		}
	} else {
		// Non-partitioned mode: use single shared queue
		q.itemQueue = make(chan *WorkItem[T], q.bufferSize)
		for i := 0; i < q.workerCount; i++ {
			go q.worker(q.ctx, -1) // -1 indicates non-partitioned mode
		}
	}
}

func (q *WorkQueue[T]) Enqueue(items ...T) {
	for _, item := range items {
		workItem := &WorkItem[T]{Item: item}

		// Check if partitioning is enabled
		partitionKey, hasPartition := q.getPartitionKey(item)
		if !hasPartition {
			// Non-partitioned mode: send to shared queue
			q.itemQueue <- workItem
			continue
		}

		// Partitioned mode: route to appropriate worker queue
		workerIndex := q.getOrAssignWorker(partitionKey)
		q.workerQueues[workerIndex] <- workItem
	}
}

// getOrAssignWorker returns the worker index for a partition key,
// assigning a new worker if this is the first time seeing this partition.
func (q *WorkQueue[T]) getOrAssignWorker(partitionKey string) int {
	// Fast path: check if partition is already assigned (read lock)
	q.partitionMutex.RLock()
	if workerIndex, exists := q.partitionToWorker[partitionKey]; exists {
		q.partitionMutex.RUnlock()
		return workerIndex
	}
	q.partitionMutex.RUnlock()

	// Slow path: assign partition to a worker (write lock)
	q.partitionMutex.Lock()
	defer q.partitionMutex.Unlock()

	// Double-check: another goroutine might have assigned it while we waited
	if workerIndex, exists := q.partitionToWorker[partitionKey]; exists {
		return workerIndex
	}

	// Assign partition to next worker (round-robin)
	workerIndex := q.nextWorkerIndex % q.workerCount
	q.partitionToWorker[partitionKey] = workerIndex
	q.nextWorkerIndex++
	return workerIndex
}

func (q *WorkQueue[T]) Stop() {
	q.m.Lock()
	defer q.m.Unlock()
	if !q.running {
		return
	}

	// Close appropriate channels based on mode
	if q.workerQueues != nil {
		// Partitioned mode: close all worker queues
		for _, queue := range q.workerQueues {
			close(queue)
		}
	} else {
		// Non-partitioned mode: close shared queue
		close(q.itemQueue)
	}

	q.running = false
}

func (q *WorkQueue[T]) IsRunning() bool {
	return q.running
}

func (q *WorkQueue[T]) AwaitCompletion() {
	q.wg.Wait()
}

func (q *WorkQueue[T]) Error() error {
	return q.errorDuringExecution
}

// getPartitionKey extracts the partition key from an item.
// It first checks if the item implements the Partitionable interface,
// then falls back to PartitionFunc if set.
// Returns (key, true) if a partition key is available, ("", false) otherwise.
func (q *WorkQueue[T]) getPartitionKey(item T) (string, bool) {
	// Try Partitionable interface first
	if partitionable, ok := any(item).(Partitionable); ok {
		return partitionable.PartitionKey(), true
	}
	// Try PartitionFunc second
	if q.PartitionFunc != nil {
		return q.PartitionFunc(item), true
	}
	// No partitioning
	return "", false
}

// isPartitioningEnabled returns true if partitioning is configured
func (q *WorkQueue[T]) isPartitioningEnabled() bool {
	// Check if type implements Partitionable by attempting to convert a zero value
	var zero T
	if _, ok := any(zero).(Partitionable); ok {
		return true
	}
	// Check if PartitionFunc is set
	return q.PartitionFunc != nil
}

func (q *WorkQueue[T]) worker(ctx context.Context, workerIndex int) {
	defer q.wg.Done()

	// Determine which channel to read from
	var itemChan <-chan *WorkItem[T]
	if workerIndex >= 0 {
		// Partitioned mode: read from this worker's dedicated queue
		itemChan = q.workerQueues[workerIndex]
	} else {
		// Non-partitioned mode: read from shared queue
		itemChan = q.itemQueue
	}

	// Process items from the channel
	for wi := range itemChan {
		for q.workOnItem(wi) {
			wi.Retries++
		}
	}
}

func (q *WorkQueue[T]) workOnItem(wi *WorkItem[T]) (retry bool) {
	err := q.handler(wi.Item)
	if err == nil {
		return false
	}

	q.errorDuringExecution = err
	if q.StopOnError {
		q.Stop()
		// If we're not retrying and there's an error callback, invoke it
		if q.ErrorCallback != nil {
			q.ErrorCallback(wi.Item, err, wi.Retries)
		}
		return false
	}
	shouldRetry := q.shouldRetry(wi)
	// If we're not retrying and there's an error callback, invoke it
	if !shouldRetry && q.ErrorCallback != nil {
		q.ErrorCallback(wi.Item, err, wi.Retries)
	}
	return shouldRetry
}

func (q *WorkQueue[T]) shouldRetry(wi *WorkItem[T]) bool {
	if q.MaxRetries > 0 && q.MaxRetries == wi.Retries {
		return false
	}
	switch q.RetryPolicy {
	case RetryPolicyNever:
		return false
	case RetryPolicyAfterDuration:
		time.Sleep(q.RetryAfterDuration)
		return true
	case RetryPolicyImmediately:
		return true
	default:
		return false
	}
}

func ProcessInWorkQueue[TIn, TOut any](in []TIn, workerFunc func(TIn) ([]TOut, error), numWorkers int, bufferSize int, maxRetries int) ([]TOut, error) {
	m := &sync.Mutex{}
	var out []TOut
	q := NewWorkQueue[TIn](func(in TIn) error {
		newOut, err := workerFunc(in)
		if err != nil {
			return err
		}
		m.Lock()
		defer m.Unlock()
		out = append(out, newOut...)
		return nil
	}, bufferSize, numWorkers)
	q.MaxRetries = maxRetries
	q.RetryPolicy = RetryPolicyAfterDuration
	q.RetryAfterDuration = time.Millisecond * 100
	q.Start()
	q.Enqueue(in...)
	q.Stop()
	q.AwaitCompletion()
	err := q.Error()
	if err != nil {
		return nil, err
	}
	return out, nil
}
