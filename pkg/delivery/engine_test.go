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
