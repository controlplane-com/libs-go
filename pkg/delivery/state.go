// Package delivery is a reusable engine for durable, idempotent, retried,
// audited, Pub/Sub-backed delivery pipelines. A pipeline supplies a record model
// (embedding State), a Store (persistence) and a Sender (the actual delivery);
// the Engine owns the leader-elected discovery loop, the Pub/Sub consumer, and
// the claim -> send -> finalize lifecycle with retry/backoff and an error-history
// audit trail. Record creation/enqueue stays pipeline-specific (domain logic).
package delivery

import "time"

// Delivery statuses. A record advances pending -> in_progress -> delivered, or
// in_progress -> failed (a retry is scheduled) -> ... -> permanently_failed.
const (
	StatusPending           = "pending"
	StatusInProgress        = "in_progress"
	StatusDelivered         = "delivered"
	StatusFailed            = "failed"
	StatusPermanentlyFailed = "permanently_failed"
)

// ErrorEntry is one entry in a delivery's structured error history.
type ErrorEntry struct {
	AttemptTimestamp string `json:"attemptTimestamp"`
	Message          string `json:"message"`
	// Detail is the underlying error exactly as the Sender returned it. Message is
	// only a classification summary and is deliberately generic ("Authentication or
	// authorization failed"); Detail is what actually went wrong, and is the field
	// to read when diagnosing a failed delivery.
	Detail        string `json:"detail,omitempty"`
	AttemptNumber int    `json:"attemptNumber"`
	ErrorType     string `json:"errorType"`
}

// State holds the delivery-lifecycle fields the engine manages. Embed it
// ANONYMOUSLY in a pipeline's record model so gorm flattens the columns and JSON
// promotes the fields to the parent object (preserving existing API/UI shapes):
//
//	type Foo struct {
//	    Id string `gorm:"primaryKey;column:id"`
//	    delivery.State
//	    // ... pipeline-specific payload columns ...
//	}
//	func (f *Foo) GetID() string                  { return f.Id }
//	func (f *Foo) DeliveryState() *delivery.State  { return &f.State }
type State struct {
	Status       string `gorm:"type:text;not null;default:'pending';column:status" json:"status"`
	AttemptCount int    `gorm:"type:integer;not null;default:0;column:attempt_count" json:"attemptCount"`
	// PushedAt records the first time this record was submitted into the delivery
	// system (published to the queue, or handed to the leader's inline processor
	// in poll-only mode). It is nil for a freshly-created outbox row that the
	// initiating action committed but that has not yet reached the delivery
	// system. The leader-elected discovery loop is responsible for pushing every
	// unpushed row and stamping this column, so a crash between the initiating
	// commit and the push can never lose the message.
	PushedAt      *time.Time   `gorm:"type:timestamptz;column:pushed_at" json:"pushedAt,omitempty"`
	DeliveredAt   *time.Time   `gorm:"type:timestamptz;column:delivered_at" json:"deliveredAt,omitempty"`
	NextRetryAt   *time.Time   `gorm:"type:timestamptz;column:next_retry_at" json:"nextRetryAt,omitempty"`
	LastAttemptAt *time.Time   `gorm:"type:timestamptz;column:last_attempt_at" json:"lastAttemptAt,omitempty"`
	LastErrorType *string      `gorm:"type:text;column:last_error_type" json:"lastErrorType,omitempty"`
	ErrorMessages []ErrorEntry `gorm:"type:jsonb;column:error_messages;serializer:json" json:"errorMessages,omitempty"`
	CreatedAt     time.Time    `gorm:"type:timestamptz;not null;default:now();column:created_at" json:"createdAt"`
	UpdatedAt     time.Time    `gorm:"type:timestamptz;not null;default:now();column:updated_at" json:"updatedAt"`
}

// Delivery is implemented by each pipeline's record model so the engine can drive
// the shared lifecycle while the model carries its own payload columns.
type Delivery interface {
	GetID() string
	DeliveryState() *State
}

// appendError appends an entry to the structured error history. detail is the
// underlying error text and may be empty.
func (s *State) appendError(errorType, message, detail string) {
	s.ErrorMessages = append(s.ErrorMessages, ErrorEntry{
		AttemptTimestamp: time.Now().UTC().Format(time.RFC3339),
		Message:          message,
		Detail:           detail,
		AttemptNumber:    s.AttemptCount,
		ErrorType:        errorType,
	})
}
