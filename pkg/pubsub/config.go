package pubsub

import "time"

// ClientConfig holds configuration for the Pub/Sub client.
type ClientConfig struct {
	ProjectID           string
	Region              string
	TopicName           string
	SubscriptionName    string
	DLQTopicName        string
	AckDeadline         time.Duration
	MinAckExtension     time.Duration
	MaxOutstandingMsgs  int
	MaxDeliveryAttempts int
	NumWorkers          int
}
