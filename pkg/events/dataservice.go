package events

import "time"

type DataServiceEventType string

const (
	DataServiceEventTypeCreated DataServiceEventType = "created"
	DataServiceEventTypeUpdated DataServiceEventType = "updated"
	DataServiceEventTypeExists  DataServiceEventType = "exists"
	DataServiceEventTypeDeleted DataServiceEventType = "deleted"
)

type DataServiceEvent[T any] struct {
	Id        string               `json:"id,omitempty"`
	Timestamp time.Time            `json:"timestamp,omitempty"`
	EventType DataServiceEventType `json:"eventType,omitempty"`
	Data      T                    `json:"data,omitempty"`
	Extra     map[string]any       `json:"extra,omitempty"`
	Subject   string               `json:"subject,omitempty"`
}
