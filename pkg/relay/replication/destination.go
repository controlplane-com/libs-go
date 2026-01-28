package replication

import (
	"github.com/controlplane-com/libs-go/pkg/logging"
	"github.com/controlplane-com/libs-go/pkg/pipeline"
)

type DestinationKind string

const (
	DestinationKindSQS  DestinationKind = "sqs"
	DestinationKindHTTP DestinationKind = "http"
)

func Initialize(replications []*Config) error {
	for _, cfg := range replications {
		if err := initializeSqs(cfg.Destinations); err != nil {
			return err
		}
		if err := initializeHttp(cfg.Destinations); err != nil {
			return err
		}
	}
	return nil
}

type DestinationSpec struct {
	Map        `json:",inline"`
	Kind       DestinationKind `json:"kind"`
	Parameters map[string]any  `json:"parameters"`
}

func (s *DestinationSpec) HandleChange(change *Change) error {
	switch s.Kind {
	case DestinationKindSQS:
		return SendToSqsDestination(s, change)
	case DestinationKindHTTP:
		return SendToHttpDestination(s, change)
	}
	return nil
}

type DestinationList []*DestinationSpec

var mapAll = Map{}

func (l DestinationList) HandleChange(change *Change) error {
	if len(l) == 0 {
		return nil
	}
	for _, spec := range l {
		if !spec.Match(change) {
			continue
		}
		return spec.HandleChange(change)
	}
	logging.Logger().Sugar().Debugf("Change ignore because there is no matching DestinationSpec: %+v", change)
	return nil
}

func (l DestinationList) Tables() []string {
	tableMap := map[string]string{}
	for _, spec := range l {
		tableMap[spec.Table] = spec.Table
	}
	return pipeline.ExtractMapValues(tableMap)
}

func assertParam[T any](params map[string]any, paramName string) T {
	var zero T
	return assertParamWithDefault[T](params, paramName, zero)
}

func assertParamWithDefault[T any](params map[string]any, paramName string, defaultValue T) T {
	if v, ok := params[paramName]; ok {
		if converted, ok := v.(T); ok {
			return converted
		}
	}
	return defaultValue
}
