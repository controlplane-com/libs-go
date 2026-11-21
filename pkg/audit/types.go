// Package audit provides a small, reusable client for emitting audit EVENTS to
// the Control Plane audit service (POST /audit/org/{org}/auditctx/cpln). It is
// the emitter counterpart to go-libs/api/auditctx.go (which manages audit-context
// RESOURCES). The event shape mirrors the audit service's canonical model
// (audit/pkg/types/models.go) and the data-service TypeScript type
// (data-service/src/audit/types.ts) so the JSON is consumed unchanged.
package audit

// AuditEvent is the audit record posted to the audit service. Field names/JSON
// tags match audit/pkg/types.AuditEvent (the canonical server type) and the
// TS AuditEvent, so the emitted JSON is identical to what the TS service sends.
type AuditEvent struct {
	EventTime string    `json:"eventTime,omitempty"`
	RequestID string    `json:"requestId,omitempty"`
	Context   *Context  `json:"context,omitempty"`
	Subject   *Subject  `json:"subject,omitempty"`
	Resource  *Resource `json:"resource,omitempty"`
	Action    *Action   `json:"action,omitempty"`
	Result    *Result   `json:"result,omitempty"`
}

// Context carries request/environment facts. Mirrors types.ts EventContext and
// the subset of audit types.Context that the data-service emitter populates.
type Context struct {
	RemoteIP    string `json:"remoteIp,omitempty"`
	EventSource string `json:"eventSource,omitempty"`
	Org         string `json:"org,omitempty"`
	Data        any    `json:"data,omitempty"`
}

// Subject identifies the acting principal. Mirrors types.ts EventSubject.
type Subject struct {
	Name   string `json:"name,omitempty"`
	Claims any    `json:"claims,omitempty"`
	Data   any    `json:"data,omitempty"`
}

// Resource identifies the target of the action. Mirrors types.ts EventResource.
type Resource struct {
	ID   string `json:"id,omitempty"`
	Type string `json:"type,omitempty"`
	Name string `json:"name,omitempty"`
	Data any    `json:"data,omitempty"`
}

// Action describes what happened (create/edit/delete/reveal/…). Mirrors
// types.ts EventAction.
type Action struct {
	Type string `json:"type,omitempty"`
	Data any    `json:"data,omitempty"`
}

// Result is the outcome status/message. Mirrors types.ts EventResult.
type Result struct {
	Status  string `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
}
