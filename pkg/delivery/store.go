package delivery

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Store persists delivery records. The engine only needs these operations;
// pipeline-specific creation/listing lives in the pipeline's own repository.
type Store[D Delivery] interface {
	GetByID(ctx context.Context, id string) (D, error)
	Save(ctx context.Context, d D) error
	// MarkPushed stamps pushed_at (once) to record that the record has been
	// submitted into the delivery system. It is a targeted column update, NOT a
	// full Save, so it never clobbers a concurrent consumer's status write; and
	// it only sets the column when still null, so retry re-pushes are no-ops.
	MarkPushed(ctx context.Context, id string) error
	// Claim atomically moves a claimable record (pending, failed, or in_progress whose
	// claim has expired) to in_progress and returns it; false means another consumer
	// holds it, it is terminal, or (ordered records) an earlier record in its order is unfinished.
	Claim(ctx context.Context, id string, now, claimUntil time.Time) (D, bool, error)
	// ListDue returns records that need pushing or (re)processing as of now:
	// unpushed outbox rows, due retries, pushed-but-never-consumed rows, and
	// stuck in_progress rows.
	ListDue(ctx context.Context, now time.Time) ([]D, error)
}

// claimTimeout is how long a claimed record stays in_progress before it may be re-claimed;
// it must exceed the longest send, including Pub/Sub ack deadlines.
const claimTimeout = 10 * time.Minute

// GormStore is a generic gorm-backed Store over any model embedding State. A new
// pipeline gets persistence for free by passing a newRecord factory:
//
//	NewGormStore(dbRw, dbRo, func() *Foo { return &Foo{} })
type GormStore[D Delivery] struct {
	dbRw      *gorm.DB
	dbRo      *gorm.DB
	newRecord func() D
	// firstInOrder is set for OrderedDelivery models: it holds a record back while
	// an earlier record with the same ordering key is unfinished.
	firstInOrder string
}

func NewGormStore[D Delivery](dbRw, dbRo *gorm.DB, newRecord func() D) *GormStore[D] {
	g := &GormStore[D]{dbRw: dbRw, dbRo: dbRo, newRecord: newRecord}
	if _, ordered := any(newRecord()).(OrderedDelivery); ordered {
		stmt := &gorm.Statement{DB: dbRw}
		if err := stmt.Parse(newRecord()); err != nil {
			panic(fmt.Sprintf("delivery: cannot parse ordered record model: %v", err))
		}
		table := stmt.Quote(stmt.Schema.Table)
		g.firstInOrder = fmt.Sprintf(`NOT EXISTS (
			SELECT 1 FROM %[1]s prev
			WHERE prev.ordering_key = %[1]s.ordering_key
			AND prev.ordering_seq < %[1]s.ordering_seq
			AND prev.status NOT IN ('%[2]s', '%[3]s'))`, table, StatusDelivered, StatusPermanentlyFailed)
	}
	return g
}

// inOrder restricts q to records with no unfinished earlier record in their order; a no-op when unordered.
func (g *GormStore[D]) inOrder(q *gorm.DB) *gorm.DB {
	if g.firstInOrder == "" {
		return q
	}
	return q.Where(g.firstInOrder)
}

// NextInOrder returns the id of the next unfinished record after d in d's order,
// or "" when there is none or d is unordered.
func (g *GormStore[D]) NextInOrder(ctx context.Context, d D) (string, error) {
	o, ok := any(d).(OrderedDelivery)
	if !ok {
		return "", nil
	}
	ord := o.DeliveryOrdering()
	var ids []string
	err := g.dbRw.WithContext(ctx).
		Model(g.newRecord()).
		Where("ordering_key = ? AND ordering_seq > ? AND status NOT IN ?", ord.Key, ord.Seq, []string{StatusDelivered, StatusPermanentlyFailed}).
		Order("ordering_seq").
		Limit(1).
		Pluck("id", &ids).Error
	if err != nil || len(ids) == 0 {
		return "", err
	}
	return ids[0], nil
}

func (g *GormStore[D]) GetByID(ctx context.Context, id string) (D, error) {
	rec := g.newRecord()
	if err := g.dbRo.WithContext(ctx).First(rec, "id = ?", id).Error; err != nil {
		var zero D
		return zero, err
	}
	return rec, nil
}

func (g *GormStore[D]) Save(ctx context.Context, d D) error {
	d.DeliveryState().UpdatedAt = time.Now().UTC()
	return g.dbRw.WithContext(ctx).Save(d).Error
}

func (g *GormStore[D]) MarkPushed(ctx context.Context, id string) error {
	now := time.Now().UTC()
	return g.dbRw.WithContext(ctx).
		Model(g.newRecord()).
		Where("id = ? AND pushed_at IS NULL", id).
		Updates(map[string]any{"pushed_at": now, "updated_at": now}).Error
}

func (g *GormStore[D]) Claim(ctx context.Context, id string, now, claimUntil time.Time) (D, bool, error) {
	var zero D
	res := g.inOrder(g.dbRw.WithContext(ctx).Model(g.newRecord())).
		Where("id = ?", id).
		Where(`(
			status IN ?
			OR (status = ? AND (next_retry_at <= ? OR (next_retry_at IS NULL AND COALESCE(last_attempt_at, created_at) <= ?)))
		)`, []string{StatusPending, StatusFailed}, StatusInProgress, now, now.Add(-claimTimeout)).
		Updates(map[string]any{
			"status":          StatusInProgress,
			"attempt_count":   gorm.Expr("attempt_count + 1"),
			"last_attempt_at": now,
			"next_retry_at":   claimUntil,
			"updated_at":      now,
		})
	if res.Error != nil {
		return zero, false, res.Error
	}
	if res.RowsAffected == 0 {
		return zero, false, nil
	}
	rec := g.newRecord()
	if err := g.dbRw.WithContext(ctx).First(rec, "id = ?", id).Error; err != nil {
		return zero, false, err
	}
	return rec, true, nil
}

func (g *GormStore[D]) ListDue(ctx context.Context, now time.Time) ([]D, error) {
	stuckThreshold := now.Add(-1 * time.Minute)
	var out []D
	err := g.inOrder(g.dbRo.WithContext(ctx).Model(g.newRecord())).
		Where("status NOT IN ?", []string{StatusDelivered, StatusPermanentlyFailed}).
		Where(`(
			pushed_at IS NULL
			OR (next_retry_at IS NOT NULL AND next_retry_at <= ?)
			OR (status = ? AND next_retry_at IS NULL AND pushed_at <= ?)
			OR (status = ? AND next_retry_at IS NULL AND COALESCE(last_attempt_at, created_at) <= ?)
		)`,
			now, StatusPending, stuckThreshold, StatusInProgress, now.Add(-claimTimeout)).
		Find(&out).Error
	return out, err
}
