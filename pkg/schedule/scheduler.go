package schedule

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/controlplane-com/libs-go/pkg/logging"
)

// rejectedSchedules caches schedules that have been determined to be impossible
// (e.g., February 31st). Since these schedules will never become valid, we cache
// the rejection to avoid expensive Next() calls on repeated Add attempts.
var rejectedSchedules sync.Map

// ClearRejectedSchedulesCache clears the cache of rejected schedules.
// This is primarily useful for testing.
func ClearRejectedSchedulesCache() {
	rejectedSchedules.Range(func(key, _ any) bool {
		rejectedSchedules.Delete(key)
		return true
	})
}

// IsScheduleRejected returns true if the given schedule has been cached as rejected.
// This is primarily useful for testing.
func IsScheduleRejected(schedule string) bool {
	_, rejected := rejectedSchedules.Load(schedule)
	return rejected
}

// Executor is called when a scheduled item is due for execution.
// The executor should NOT reschedule the item - that's handled by the scheduler.
type Executor[T Item] func(ctx context.Context, item T) error

// Config configures the generic scheduler behavior.
type Config struct {
	// MissedWindow is how late an item can be before skipping execution (default: 5m)
	MissedWindow time.Duration
	// IdleInterval is how long to sleep when heap is empty (default: 1m)
	IdleInterval time.Duration
}

// DefaultConfig returns the default scheduler configuration.
func DefaultConfig() Config {
	return Config{
		MissedWindow: 5 * time.Minute,
		IdleInterval: time.Minute,
	}
}

// Scheduler is a generic cron-based scheduler.
type Scheduler[T Item] struct {
	heap         *Heap[T]
	config       Config
	executor     Executor[T]
	isLeader     func() bool
	wakeupCh     chan struct{}
	cronParser   cron.Parser
	itemResolver func(T) T // optional callback to get latest item version before rescheduling
}

// NewScheduler creates a new generic scheduler.
func NewScheduler[T Item](executor Executor[T], config Config) *Scheduler[T] {
	return &Scheduler[T]{
		heap:     NewHeap[T](),
		config:   config,
		executor: executor,
		isLeader: func() bool { return true }, // default to always execute
		wakeupCh: make(chan struct{}, 1),
		// Standard cron parser with 5 fields (minute, hour, day-of-month, month, day-of-week)
		cronParser: cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow),
	}
}

// SetIsLeader sets the leadership check callback.
// Only the leader actually executes items; non-leaders still maintain the heap.
func (s *Scheduler[T]) SetIsLeader(fn func() bool) {
	s.isLeader = fn
}

// SetItemResolver sets the item resolver callback.
// When set, the scheduler calls this before rescheduling an item to get the latest version.
// This helps avoid race conditions where an item is updated while being executed.
func (s *Scheduler[T]) SetItemResolver(fn func(T) T) {
	s.itemResolver = fn
}

// validateNextExecution checks if the schedule can produce a valid next execution time.
// robfig/cron returns zero time if no valid time exists within its search window (~5 years).
func validateNextExecution(next time.Time, schedule string) error {
	if next.IsZero() {
		return fmt.Errorf("schedule '%s' will never fire (e.g., February 31st)", schedule)
	}
	return nil
}

// ValidateSchedule checks if a cron schedule expression is valid.
// Returns nil if valid, or an error describing why it's invalid.
// Use this to pre-validate schedules before calling Add().
func (s *Scheduler[T]) ValidateSchedule(schedule string) error {
	// Check cache for previously rejected schedules to avoid expensive Next() calls
	if _, rejected := rejectedSchedules.Load(schedule); rejected {
		return fmt.Errorf("schedule '%s' will never fire (e.g., February 31st)", schedule)
	}

	sched, err := s.cronParser.Parse(schedule)
	if err != nil {
		return fmt.Errorf("failed to parse cron expression '%s': %w", schedule, err)
	}
	next := sched.Next(time.Now())
	if err := validateNextExecution(next, schedule); err != nil {
		// Cache this rejection for future lookups
		rejectedSchedules.Store(schedule, struct{}{})
		return err
	}
	return nil
}

// Add adds or updates an item in the schedule.
// The item's Schedule() must return a valid cron expression.
func (s *Scheduler[T]) Add(item T) error {
	schedule := item.Schedule()

	// Check cache for previously rejected schedules to avoid expensive Next() calls
	if _, rejected := rejectedSchedules.Load(schedule); rejected {
		return fmt.Errorf("schedule '%s' will never fire (e.g., February 31st)", schedule)
	}

	sched, err := s.cronParser.Parse(schedule)
	if err != nil {
		return fmt.Errorf("failed to parse cron expression '%s': %w", schedule, err)
	}
	next := sched.Next(time.Now())
	if err := validateNextExecution(next, schedule); err != nil {
		// Cache this rejection for future lookups
		rejectedSchedules.Store(schedule, struct{}{})
		return err
	}
	item.SetNextExecution(next)
	s.heap.Push(item)
	s.tryWakeup()
	return nil
}

// Remove removes an item by ID.
func (s *Scheduler[T]) Remove(id string) {
	s.heap.Remove(id)
}

// Get returns an item by ID, or nil and false if not found.
func (s *Scheduler[T]) Get(id string) (T, bool) {
	return s.heap.Get(id)
}

// Run executes the scheduler loop until context is cancelled.
func (s *Scheduler[T]) Run(ctx context.Context) error {
	logger := logging.LoggerWithContext(ctx)
	logger.Info("Starting scheduler loop")

	for {
		next, ok := s.heap.Peek()
		var sleepDuration time.Duration
		if !ok {
			sleepDuration = s.config.IdleInterval
		} else {
			sleepDuration = time.Until(next.NextExecution())
			if sleepDuration < 0 {
				sleepDuration = 0
			}
		}

		logger.Debugf("Scheduler sleeping for %v", sleepDuration)

		select {
		case <-ctx.Done():
			logger.Info("Scheduler loop stopped by context")
			return ctx.Err()
		case <-s.wakeupCh:
			logger.Debug("Scheduler woken up")
			continue
		case <-time.After(sleepDuration):
			s.processDueItems(ctx)
		}
	}
}

// processDueItems processes all items that are currently due.
func (s *Scheduler[T]) processDueItems(ctx context.Context) {
	now := time.Now()
	for {
		item, ok := s.heap.PopIfDue()
		if !ok {
			return
		}
		s.executeAndReschedule(ctx, item, now)
	}
}

// executeAndReschedule handles a single item, ensuring it's always rescheduled even on panic.
func (s *Scheduler[T]) executeAndReschedule(ctx context.Context, item T, now time.Time) {
	logger := logging.LoggerWithContext(ctx)

	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("Panic while processing item %s: %v", item.ID(), r)
		}

		// Get the latest version of the item if resolver is set
		// This prevents race conditions where an update arrives during execution
		itemToReschedule := item
		if s.itemResolver != nil {
			itemToReschedule = s.itemResolver(item)
		}

		// Always reschedule based on current time to ensure next execution is in the future
		schedule := itemToReschedule.Schedule()

		// Check cache for previously rejected schedules
		if _, rejected := rejectedSchedules.Load(schedule); rejected {
			logger.Errorf("Removing item %s from schedule: schedule '%s' will never fire (cached)",
				itemToReschedule.ID(), schedule)
			return
		}

		sched, err := s.cronParser.Parse(schedule)
		if err != nil {
			logger.Errorf("Failed to reschedule item %s: failed to parse cron expression '%s': %v",
				itemToReschedule.ID(), schedule, err)
			return
		}
		next := sched.Next(time.Now())
		if err := validateNextExecution(next, schedule); err != nil {
			// Cache this rejection for future lookups
			rejectedSchedules.Store(schedule, struct{}{})
			logger.Errorf("Removing item %s from schedule: %v", itemToReschedule.ID(), err)
			return // Don't push back to heap - effectively removes the item
		}
		itemToReschedule.SetNextExecution(next)
		s.heap.Push(itemToReschedule)
		logger.Debugf("Rescheduled item %s for next execution at %v", itemToReschedule.ID(), itemToReschedule.NextExecution())
	}()

	// Skip if missed by too much
	if timeMissed := now.Sub(item.NextExecution()); timeMissed > s.config.MissedWindow {
		logger.Warnf("Skipping missed execution for item %s (missed by %v)", item.ID(), timeMissed)
		return
	}

	// Only leader executes
	if !s.isLeader() {
		logger.Debugf("Not leader, skipping execution for item %s", item.ID())
		return
	}

	// Execute
	if err := s.executor(ctx, item); err != nil {
		logger.Errorf("Failed to execute item %s: %v", item.ID(), err)
	}
}

// tryWakeup sends a non-blocking signal to wake up the scheduler loop.
func (s *Scheduler[T]) tryWakeup() {
	select {
	case s.wakeupCh <- struct{}{}:
	default:
		// Channel already has a pending wakeup
	}
}

// Count returns the number of scheduled items.
func (s *Scheduler[T]) Count() int {
	return s.heap.Len()
}
