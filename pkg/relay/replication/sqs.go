package replication

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/controlplane-com/libs-go/pkg/logging"
	"go.uber.org/zap"
	"hash/fnv"
	"log"
)

const (
	DefaultGroupCount       = 10
	DefaultRetryMaxAttempts = 10
)

var l *zap.SugaredLogger
var configMap = map[string]*sqsConfig{}

type groupConfig struct {
	field string
	count int
	hash  bool
}

type sqsConfig struct {
	queueUrl           string
	credentialsPath    string
	group              groupConfig
	deduplicationField string
	retryMaxAttempts   int
	client             *sqs.Client
}

func logger() *zap.SugaredLogger {
	if l == nil {
		l = logging.Logger().Sugar()
	}
	return l
}
func initializeSqs(list DestinationList) error {
	for _, spec := range list {
		if spec.Kind == DestinationKindSQS {
			c, err := parseSqsConfig(spec)
			if err != nil {
				return err
			}
			configMap[spec.Name] = c
		}
	}
	return nil
}

func (g groupConfig) name(payload map[string]any) string {
	fieldValue := assertParam[string](payload, g.field)
	if g.hash {
		hasher := fnv.New32a()
		_, _ = hasher.Write([]byte(fieldValue))
		return fmt.Sprintf("group_%d", int(hasher.Sum32())%g.count)
	}
	return fieldValue
}

func newGroupConfig(params map[string]any) groupConfig {
	return groupConfig{
		field: assertParam[string](params, "field"),
		count: assertParamWithDefault[int](params, "count", DefaultGroupCount),
		hash:  assertParam[bool](params, "hash"),
	}
}

func getSqsConfig(spec *DestinationSpec) *sqsConfig {
	if spec == nil {
		return nil
	}
	c, ok := configMap[spec.Name]
	if !ok {
		return nil
	}
	return c
}

func parseSqsConfig(spec *DestinationSpec) (*sqsConfig, error) {
	if spec == nil {
		return nil, errors.New("spec is nil")
	}
	params := spec.Parameters
	credentialsPath := assertParam[string](params, "credentialsPath")
	if credentialsPath == "" {
		return nil, fmt.Errorf("missing credentialsPathin sqs destination spec %s", spec.Name)
	}

	region := assertParam[string](params, "region")
	if region == "" {
		return nil, fmt.Errorf("missing regionin sqs destination spec %s", spec.Name)
	}

	retryMaxAttempts := assertParamWithDefault[int](params, "retryMaxAttempts", DefaultRetryMaxAttempts)
	if retryMaxAttempts <= 0 {
		return nil, fmt.Errorf("invalid retry max attempts in sqs destination spec %s: %d", spec.Name, retryMaxAttempts)
	}
	// Load the configuration from the parameters
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(region),
		config.WithSharedCredentialsFiles([]string{credentialsPath}),
		config.WithRetryMaxAttempts(retryMaxAttempts))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	// Create an SQS client
	c := &sqsConfig{
		queueUrl:         assertParam[string](params, "queueUrl"),
		credentialsPath:  credentialsPath,
		group:            newGroupConfig(assertParam[map[string]any](params, "group")),
		retryMaxAttempts: retryMaxAttempts,
		client:           sqs.NewFromConfig(cfg),
	}
	if c.group.count <= 0 {
		return nil, fmt.Errorf("invalid group count in sqs destination spec %s: %d", spec.Name, c.group.count)
	}
	if c.credentialsPath == "" {
		return nil, fmt.Errorf("missing credentialsPathin sqs destination spec %s", spec.Name)
	}
	if c.queueUrl == "" {
		return nil, fmt.Errorf("missing queueUrlin sqs destination spec %s", spec.Name)
	}
	return c, nil
}

func SendToSqsDestination(spec *DestinationSpec, change *Change) error {
	if change == nil || spec == nil {
		return nil
	}
	s := getSqsConfig(spec)
	if s == nil {
		return nil
	}
	message := spec.Message(change)
	jsonBytes, err := json.Marshal(message)
	if err != nil {
		return err
	}

	deduplicationId := message.Id
	if deduplicationId == "" {
		hash := sha256.New()
		hash.Write(jsonBytes)
		deduplicationId = hex.EncodeToString(hash.Sum(nil))
	}

	// Send the message
	sendMessageInput := &sqs.SendMessageInput{
		QueueUrl:               aws.String(s.queueUrl),
		MessageBody:            aws.String(string(jsonBytes)),
		MessageGroupId:         aws.String(s.group.name(message.Payload)),
		MessageDeduplicationId: aws.String(deduplicationId),
	}

	_, err = s.client.SendMessage(context.TODO(), sendMessageInput)
	if err != nil {
		logger().Errorf("failed to send message, %v", err)
		return err
	}

	return nil
}
