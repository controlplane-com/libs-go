package delivery

import (
	"context"
	"errors"
	"sort"
	"testing"
	"time"
)

type fakeOrderedRecord struct {
	fakeRecord
	Ordering
}

func (f *fakeOrderedRecord) DeliveryOrdering() *Ordering { return &f.Ordering }

func finished(status string) bool {
	return status == StatusDelivered || status == StatusPermanentlyFailed
}

// fakeOrderedStore mirrors GormStore's ordering rules in memory.
type fakeOrderedStore struct{ recs map[string]*fakeOrderedRecord }

func newOrderedStore(recs ...*fakeOrderedRecord) *fakeOrderedStore {
	s := &fakeOrderedStore{recs: map[string]*fakeOrderedRecord{}}
	for _, r := range recs {
		r.Status = StatusPending
		s.recs[r.id] = r
	}
	return s
}

func (s *fakeOrderedStore) blocked(r *fakeOrderedRecord) bool {
	for _, prev := range s.recs {
		if prev.Key == r.Key && prev.Seq < r.Seq && !finished(prev.Status) {
			return true
		}
	}
	return false
}

func (s *fakeOrderedStore) GetByID(_ context.Context, id string) (*fakeOrderedRecord, error) {
	return s.recs[id], nil
}
func (s *fakeOrderedStore) Save(_ context.Context, _ *fakeOrderedRecord) error { return nil }
func (s *fakeOrderedStore) MarkPushed(_ context.Context, _ string) error       { return nil }
func (s *fakeOrderedStore) Claim(_ context.Context, id string, now, claimUntil time.Time) (*fakeOrderedRecord, bool, error) {
	r := s.recs[id]
	if r == nil || (r.Status != StatusPending && r.Status != StatusFailed) || s.blocked(r) {
		return nil, false, nil
	}
	r.Status = StatusInProgress
	r.AttemptCount++
	r.NextRetryAt = &claimUntil
	return r, true, nil
}
func (s *fakeOrderedStore) ListDue(_ context.Context, _ time.Time) ([]*fakeOrderedRecord, error) {
	return nil, nil
}
func (s *fakeOrderedStore) NextInOrder(_ context.Context, d *fakeOrderedRecord) (string, error) {
	var next []*fakeOrderedRecord
	for _, r := range s.recs {
		if r.Key == d.Key && r.Seq > d.Seq && !finished(r.Status) {
			next = append(next, r)
		}
	}
	if len(next) == 0 {
		return "", nil
	}
	sort.Slice(next, func(i, j int) bool { return next[i].Seq < next[j].Seq })
	return next[0].id, nil
}

// recordingSender records send order and fails ids listed in errs.
type recordingSender struct {
	sent []string
	errs map[string]error
}

func (s *recordingSender) Send(_ context.Context, d *fakeOrderedRecord) error {
	s.sent = append(s.sent, d.id)
	return s.errs[d.id]
}

func ordered(id, key string, seq int64) *fakeOrderedRecord {
	return &fakeOrderedRecord{fakeRecord: fakeRecord{id: id}, Ordering: Ordering{Key: key, Seq: seq}}
}

func TestOrdered_NewerRecordWaitsThenFollowsTheOlderOne(t *testing.T) {
	first, second := ordered("a1", "a", 1), ordered("a2", "a", 2)
	sender := &recordingSender{}
	e := NewEngine[*fakeOrderedRecord](newOrderedStore(first, second), sender, Config{Name: "test", MaxRetries: 3})

	if err := e.Process(context.Background(), "a2"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sender.sent) != 0 || second.Status != StatusPending || second.AttemptCount != 0 {
		t.Fatalf("the newer record must be refused untouched: sent=%v status=%q attempts=%d", sender.sent, second.Status, second.AttemptCount)
	}

	if err := e.Process(context.Background(), "a1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := sender.sent; len(got) != 2 || got[0] != "a1" || got[1] != "a2" || second.Status != StatusDelivered {
		t.Fatalf("expected a1 then a2 via handoff, got %v (a2 status %q)", got, second.Status)
	}
}

func TestOrdered_RetryingRecordHoldsBackTheNext(t *testing.T) {
	first, second := ordered("a1", "a", 1), ordered("a2", "a", 2)
	sender := &recordingSender{errs: map[string]error{"a1": errors.New("connection timeout")}}
	e := NewEngine[*fakeOrderedRecord](newOrderedStore(first, second), sender, Config{Name: "test", MaxRetries: 3})

	_ = e.Process(context.Background(), "a1")
	_ = e.Process(context.Background(), "a2")
	if first.Status != StatusFailed || second.Status != StatusPending || len(sender.sent) != 1 {
		t.Fatalf("a retrying record must hold the next back: a1=%q a2=%q sent=%v", first.Status, second.Status, sender.sent)
	}
}

func TestOrdered_PermanentFailureReleasesTheNext(t *testing.T) {
	first, second := ordered("a1", "a", 1), ordered("a2", "a", 2)
	sender := &recordingSender{errs: map[string]error{"a1": errors.New("400 bad request")}}
	e := NewEngine[*fakeOrderedRecord](newOrderedStore(first, second), sender, Config{Name: "test", MaxRetries: 1})

	_ = e.Process(context.Background(), "a1")
	if first.Status != StatusPermanentlyFailed || second.Status != StatusDelivered {
		t.Fatalf("a permanent failure must release the next: a1=%q a2=%q", first.Status, second.Status)
	}
}

func TestOrdered_DifferentKeysDoNotBlockEachOther(t *testing.T) {
	a1, b1 := ordered("a1", "a", 1), ordered("b1", "b", 2)
	sender := &recordingSender{errs: map[string]error{"a1": errors.New("connection timeout")}}
	e := NewEngine[*fakeOrderedRecord](newOrderedStore(a1, b1), sender, Config{Name: "test", MaxRetries: 3})

	_ = e.Process(context.Background(), "a1")
	if err := e.Process(context.Background(), "b1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b1.Status != StatusDelivered {
		t.Fatalf("another key must not wait on a retrying record, b1=%q", b1.Status)
	}
}

func TestUnordered_DeliveryDoesNotHandOff(t *testing.T) {
	rec := &fakeRecord{id: "d1", State: State{Status: StatusPending}}
	sender := &fakeSender{}
	e := engineFor(rec, sender, Config{Name: "test", MaxRetries: 3})
	if _, ok := e.store.(orderedStore[*fakeRecord]); ok {
		t.Fatal("a store without NextInOrder must not be treated as ordered")
	}
	if err := e.Process(context.Background(), "d1"); err != nil || rec.Status != StatusDelivered || sender.calls != 1 {
		t.Fatalf("unordered delivery changed: err=%v status=%q calls=%d", err, rec.Status, sender.calls)
	}
}
