package delivery

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/controlplane-com/libs-go/pkg/logging"
	"go.uber.org/zap/zapcore"
)

func init() { _ = logging.InitializeLogger(zapcore.InfoLevel) }

// --- in-memory fakes (no DB) ---

type fakeRecord struct {
	id string
	State
}

func (f *fakeRecord) GetID() string         { return f.id }
func (f *fakeRecord) DeliveryState() *State { return &f.State }

type fakeStore struct{ recs map[string]*fakeRecord }

func (s *fakeStore) GetByID(_ context.Context, id string) (*fakeRecord, error) {
	r, ok := s.recs[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return r, nil
}
func (s *fakeStore) Save(_ context.Context, _ *fakeRecord) error { return nil } // pointer mutations persist
func (s *fakeStore) MarkPushed(_ context.Context, id string) error {
	if r, ok := s.recs[id]; ok && r.PushedAt == nil {
		now := time.Now().UTC()
		r.PushedAt = &now
	}
	return nil
}
func (s *fakeStore) Claim(_ context.Context, id string, now, claimUntil time.Time) (*fakeRecord, bool, error) {
	r, ok := s.recs[id]
	if !ok {
		return nil, false, nil
	}
	expired := r.Status == StatusInProgress && r.NextRetryAt != nil && !r.NextRetryAt.After(now)
	if r.Status != StatusPending && r.Status != StatusFailed && !expired {
		return nil, false, nil
	}
	r.Status = StatusInProgress
	r.AttemptCount++
	r.LastAttemptAt = &now
	r.NextRetryAt = &claimUntil
	return r, true, nil
}
func (s *fakeStore) ListDue(_ context.Context, _ time.Time) ([]*fakeRecord, error) {
	return nil, nil
}

type fakeSender struct {
	err   error
	calls int
}

func (s *fakeSender) Send(_ context.Context, _ *fakeRecord) error { s.calls++; return s.err }

func engineFor(rec *fakeRecord, sender Sender[*fakeRecord], cfg Config) *Engine[*fakeRecord] {
	store := &fakeStore{recs: map[string]*fakeRecord{rec.id: rec}}
	return NewEngine[*fakeRecord](store, sender, cfg)
}

func TestProcess_Success(t *testing.T) {
	rec := &fakeRecord{id: "d1", State: State{Status: StatusPending}}
	sender := &fakeSender{}
	e := engineFor(rec, sender, Config{Name: "test", MaxRetries: 3})

	if err := e.Process(context.Background(), "d1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Status != StatusDelivered {
		t.Fatalf("expected delivered, got %q", rec.Status)
	}
	if rec.DeliveredAt == nil || rec.AttemptCount != 1 || sender.calls != 1 {
		t.Fatalf("unexpected state: deliveredAt=%v attempts=%d calls=%d", rec.DeliveredAt, rec.AttemptCount, sender.calls)
	}
}

func TestProcess_RetryableFailure(t *testing.T) {
	rec := &fakeRecord{id: "d1", State: State{Status: StatusPending}}
	sender := &fakeSender{err: errors.New("connection timeout")}
	e := engineFor(rec, sender, Config{Name: "test", MaxRetries: 3})

	if err := e.Process(context.Background(), "d1"); err == nil {
		t.Fatal("expected the send error to be returned")
	}
	if rec.Status != StatusFailed {
		t.Fatalf("expected failed, got %q", rec.Status)
	}
	if rec.NextRetryAt == nil || rec.AttemptCount != 1 {
		t.Fatalf("expected scheduled retry after 1 attempt, got attempts=%d next=%v", rec.AttemptCount, rec.NextRetryAt)
	}
}

func TestProcess_PermanentFailureFiresHook(t *testing.T) {
	rec := &fakeRecord{id: "d1", State: State{Status: StatusPending, AttemptCount: 3}} // exhausted after the claim bump
	sender := &fakeSender{err: errors.New("503 unavailable")}
	var hookID string
	var hookErr error
	e := engineFor(rec, sender, Config{Name: "test", MaxRetries: 3, OnPermanentFailure: func(d Delivery, err error) {
		hookID, hookErr = d.GetID(), err
	}})

	_ = e.Process(context.Background(), "d1")
	if rec.Status != StatusPermanentlyFailed {
		t.Fatalf("expected permanently_failed, got %q", rec.Status)
	}
	if hookID != "d1" || hookErr == nil {
		t.Fatalf("expected OnPermanentFailure to fire for d1, got id=%q err=%v", hookID, hookErr)
	}
}

func TestProcess_SkipsTerminal(t *testing.T) {
	rec := &fakeRecord{id: "d1", State: State{Status: StatusDelivered}}
	sender := &fakeSender{}
	e := engineFor(rec, sender, Config{Name: "test"})

	if err := e.Process(context.Background(), "d1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sender.calls != 0 {
		t.Fatalf("expected no send for an already-delivered record, got %d", sender.calls)
	}
}

func TestProcess_SkipsRecordClaimedByAnotherConsumer(t *testing.T) {
	future := time.Now().Add(5 * time.Minute)
	rec := &fakeRecord{id: "d1", State: State{Status: StatusInProgress, NextRetryAt: &future}}
	sender := &fakeSender{}
	e := engineFor(rec, sender, Config{Name: "test", MaxRetries: 3})

	if err := e.Process(context.Background(), "d1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sender.calls != 0 || rec.Status != StatusInProgress {
		t.Fatalf("a live claim must not be sent again: calls=%d status=%q", sender.calls, rec.Status)
	}
}

func TestProcess_ReclaimsExpiredClaim(t *testing.T) {
	past := time.Now().Add(-time.Minute)
	rec := &fakeRecord{id: "d1", State: State{Status: StatusInProgress, NextRetryAt: &past, AttemptCount: 1}}
	sender := &fakeSender{}
	e := engineFor(rec, sender, Config{Name: "test", MaxRetries: 3})

	if err := e.Process(context.Background(), "d1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sender.calls != 1 || rec.Status != StatusDelivered || rec.AttemptCount != 2 {
		t.Fatalf("an expired claim must be re-sent: calls=%d status=%q attempts=%d", sender.calls, rec.Status, rec.AttemptCount)
	}
}

func TestEnqueue_PollOnlyDeliversImmediately(t *testing.T) {
	rec := &fakeRecord{id: "d1", State: State{Status: StatusPending}}
	sender := &fakeSender{}
	e := engineFor(rec, sender, Config{Name: "test", MaxRetries: 3})

	if err := e.Enqueue("d1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sender.calls != 1 || rec.Status != StatusDelivered || rec.PushedAt == nil {
		t.Fatalf("expected an inline send: calls=%d status=%q pushedAt=%v", sender.calls, rec.Status, rec.PushedAt)
	}

	if err := e.Enqueue("d1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sender.calls != 1 {
		t.Fatalf("a delivered record must not be re-sent, calls=%d", sender.calls)
	}
}

func TestEnqueue_PollOnlyFailureIsLeftForThePoll(t *testing.T) {
	rec := &fakeRecord{id: "d1", State: State{Status: StatusPending}}
	sender := &fakeSender{err: errors.New("connection timeout")}
	e := engineFor(rec, sender, Config{Name: "test", MaxRetries: 3})

	if err := e.Enqueue("d1"); err != nil {
		t.Fatalf("a failed send must not fail the caller: %v", err)
	}
	if sender.calls != 1 || rec.Status != StatusFailed || rec.NextRetryAt == nil {
		t.Fatalf("expected a scheduled retry: calls=%d status=%q nextRetryAt=%v", sender.calls, rec.Status, rec.NextRetryAt)
	}
}
