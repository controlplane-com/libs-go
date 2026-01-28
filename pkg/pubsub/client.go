package pubsub

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/pubsub/v2"
	"cloud.google.com/go/pubsub/v2/apiv1/pubsubpb"
	"github.com/controlplane-com/libs-go/pkg/logging"
	"go.uber.org/zap"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
)

// PubSubClient implements the Client interface for GCP Pub/Sub.
type PubSubClient[T Message] struct {
	client             *pubsub.Client
	publisher          *publisherImpl[T]
	consumer           *consumerImpl[T]
	projectID          string
	region             string
	topicName          string
	subName            string
	dlqTopicName       string
	logger             *zap.SugaredLogger
	attributeExtractor AttributeExtractor[T]
	unmarshaler        Unmarshaler[T]
}

// NewClient creates a new Pub/Sub client with the specified configuration.
// It ensures the required topics and subscription exist, creating them if necessary.
//
// The attributeExtractor function is called when publishing messages to extract
// key-value attributes that are stored with the message.
//
// The unmarshaler function is called when consuming messages to deserialize
// the message data into the typed message struct.
func NewClient[T Message](
	ctx context.Context,
	cfg ClientConfig,
	attributeExtractor AttributeExtractor[T],
	unmarshaler Unmarshaler[T],
) (*PubSubClient[T], error) {
	logger := logging.Logger().Sugar()

	// Use regional endpoint for exactly-once guarantee
	endpoint := fmt.Sprintf("%s-pubsub.googleapis.com:443", cfg.Region)
	logger.Infof("Connecting to Pub/Sub regional endpoint: %s", endpoint)

	client, err := pubsub.NewClient(ctx, cfg.ProjectID,
		option.WithEndpoint(endpoint))
	if err != nil {
		return nil, fmt.Errorf("failed to create pubsub client: %w", err)
	}

	c := &PubSubClient[T]{
		client:             client,
		projectID:          cfg.ProjectID,
		region:             cfg.Region,
		topicName:          cfg.TopicName,
		subName:            cfg.SubscriptionName,
		dlqTopicName:       cfg.DLQTopicName,
		logger:             logger,
		attributeExtractor: attributeExtractor,
		unmarshaler:        unmarshaler,
	}

	// Create topic if not exists
	if err := c.ensureTopicExists(ctx, cfg.TopicName); err != nil {
		client.Close()
		return nil, err
	}

	// Create DLQ topic if not exists
	if err := c.ensureTopicExists(ctx, cfg.DLQTopicName); err != nil {
		client.Close()
		return nil, err
	}

	// Create subscription with exactly-once delivery if not exists
	if err := c.ensureSubscriptionExists(ctx, cfg); err != nil {
		client.Close()
		return nil, err
	}

	// Create publisher
	c.publisher = &publisherImpl[T]{
		pub:                client.Publisher(cfg.TopicName),
		logger:             logger,
		attributeExtractor: attributeExtractor,
	}

	// Create consumer
	c.consumer = &consumerImpl[T]{
		sub:                client.Subscriber(cfg.SubscriptionName),
		minAckExtension:    cfg.MinAckExtension,
		maxOutstandingMsgs: cfg.MaxOutstandingMsgs,
		numWorkers:         cfg.NumWorkers,
		logger:             logger,
		unmarshaler:        unmarshaler,
	}

	logger.Infof("Pub/Sub client initialized for project %s, topic %s, subscription %s",
		cfg.ProjectID, cfg.TopicName, cfg.SubscriptionName)

	return c, nil
}

// Publisher returns the publisher for sending messages.
func (c *PubSubClient[T]) Publisher() Publisher[T] {
	return c.publisher
}

// Consumer returns the consumer for receiving messages.
func (c *PubSubClient[T]) Consumer() Consumer[T] {
	return c.consumer
}

// Close closes the client and all underlying connections.
func (c *PubSubClient[T]) Close() error {
	if c.publisher != nil {
		c.publisher.Close()
	}
	return c.client.Close()
}

// ensureTopicExists creates the topic if it doesn't exist.
func (c *PubSubClient[T]) ensureTopicExists(ctx context.Context, topicName string) error {
	topicPath := fmt.Sprintf("projects/%s/topics/%s", c.projectID, topicName)

	// Check if topic exists
	_, err := c.client.TopicAdminClient.GetTopic(ctx, &pubsubpb.GetTopicRequest{
		Topic: topicPath,
	})
	if err == nil {
		c.logger.Debugf("Topic %s already exists", topicName)
		return nil
	}

	if status.Code(err) != codes.NotFound {
		return fmt.Errorf("failed to check topic existence: %w", err)
	}

	// Topic doesn't exist, create it
	_, err = c.client.TopicAdminClient.CreateTopic(ctx, &pubsubpb.Topic{
		Name: topicPath,
	})
	if err != nil {
		if status.Code(err) == codes.AlreadyExists {
			c.logger.Infof("Topic %s already exists (race condition)", topicName)
			return nil
		}
		return fmt.Errorf("failed to create topic: %w", err)
	}

	c.logger.Infof("Created topic %s", topicName)
	return nil
}

// ensureSubscriptionExists creates the subscription with exactly-once delivery if it doesn't exist.
func (c *PubSubClient[T]) ensureSubscriptionExists(ctx context.Context, cfg ClientConfig) error {
	subPath := fmt.Sprintf("projects/%s/subscriptions/%s", c.projectID, cfg.SubscriptionName)
	topicPath := fmt.Sprintf("projects/%s/topics/%s", c.projectID, cfg.TopicName)
	dlqTopicPath := fmt.Sprintf("projects/%s/topics/%s", c.projectID, cfg.DLQTopicName)

	// Check if subscription exists
	_, err := c.client.SubscriptionAdminClient.GetSubscription(ctx, &pubsubpb.GetSubscriptionRequest{
		Subscription: subPath,
	})
	if err == nil {
		c.logger.Debugf("Subscription %s already exists", cfg.SubscriptionName)
		return nil
	}

	if status.Code(err) != codes.NotFound {
		return fmt.Errorf("failed to check subscription existence: %w", err)
	}

	// Subscription doesn't exist, create it with exactly-once delivery
	_, err = c.client.SubscriptionAdminClient.CreateSubscription(ctx, &pubsubpb.Subscription{
		Name:                      subPath,
		Topic:                     topicPath,
		AckDeadlineSeconds:        int32(cfg.AckDeadline.Seconds()),
		EnableExactlyOnceDelivery: true, // Critical for exactly-once semantics
		DeadLetterPolicy: &pubsubpb.DeadLetterPolicy{
			DeadLetterTopic:     dlqTopicPath,
			MaxDeliveryAttempts: int32(cfg.MaxDeliveryAttempts),
		},
		RetryPolicy: &pubsubpb.RetryPolicy{
			MinimumBackoff: durationpb.New(10 * time.Second),
			MaximumBackoff: durationpb.New(600 * time.Second),
		},
		// Retain messages for 7 days for debugging
		MessageRetentionDuration: durationpb.New(7 * 24 * time.Hour),
	})
	if err != nil {
		if status.Code(err) == codes.AlreadyExists {
			c.logger.Infof("Subscription %s already exists (race condition)", cfg.SubscriptionName)
			return nil
		}
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	c.logger.Infof("Created subscription %s with exactly-once delivery enabled", cfg.SubscriptionName)
	return nil
}
