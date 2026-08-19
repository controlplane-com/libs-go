package delivery

import (
	"context"
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
	// ListDue returns records that need pushing or (re)processing as of now:
	// unpushed outbox rows, due retries, pushed-but-never-consumed rows, and
	// stuck in_progress rows.
	ListDue(ctx context.Context, now time.Time) ([]D, error)
}

// GormStore is a generic gorm-backed Store over any model embedding State. A new
// pipeline gets persistence for free by passing a newRecord factory:
//
//	NewGormStore(dbRw, dbRo, func() *Foo { return &Foo{} })
type GormStore[D Delivery] struct {
	dbRw      *gorm.DB
	dbRo      *gorm.DB
	newRecord func() D
}

func NewGormStore[D Delivery](dbRw, dbRo *gorm.DB, newRecord func() D) *GormStore[D] {
	return &GormStore[D]{dbRw: dbRw, dbRo: dbRo, newRecord: newRecord}
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

func (g *GormStore[D]) ListDue(ctx context.Context, now time.Time) ([]D, error) {
	stuckThreshold := now.Add(-1 * time.Minute)
	var out []D
	err := g.dbRo.WithContext(ctx).
		Model(g.newRecord()).
		Where("status NOT IN ?", []string{StatusDelivered, StatusPermanentlyFailed}).
		Where(`(
			pushed_at IS NULL
			OR (next_retry_at IS NOT NULL AND next_retry_at <= ?)
			OR (status = ? AND next_retry_at IS NULL AND pushed_at <= ?)
			OR (status = ? AND COALESCE(last_attempt_at, created_at) <= ?)
		)`,
			now, StatusPending, stuckThreshold, StatusInProgress, stuckThreshold).
		Find(&out).Error
	return out, err
}
