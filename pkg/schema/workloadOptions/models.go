/* auto-generated */

package workloadOptions

type DefaultOptionsMultiZone struct {
	Enabled bool `json:"enabled,omitempty"`
}

type DefaultOptions struct {
	Autoscaling             *OptionsAutoscaling      `json:"autoscaling,omitempty"`
	TimeoutSeconds          *float32                 `json:"timeoutSeconds,omitempty"`
	CapacityAI              *bool                    `json:"capacityAI,omitempty"`
	CapacityAIUpdateMinutes *float32                 `json:"capacityAIUpdateMinutes,omitempty"`
	Spot                    *bool                    `json:"spot,omitempty"`
	Debug                   bool                     `json:"debug,omitempty"`
	Suspend                 bool                     `json:"suspend,omitempty"`
	MultiZone               *DefaultOptionsMultiZone `json:"multiZone,omitempty"`
}

type KedaTriggerMetadata map[string]any /* TODO: [object Object]*/

type KedaTriggerMetricType string

const (
	KedaTriggerMetricTypeAverageValue KedaTriggerMetricType = "AverageValue"
	KedaTriggerMetricTypeValue        KedaTriggerMetricType = "Value"
	KedaTriggerMetricTypeUtilization  KedaTriggerMetricType = "Utilization"
)

type KedaTriggerAuthenticationRef struct {
	Name string `json:"name"`
}

type KedaTrigger struct {
	Type              string                        `json:"type"`
	Metadata          KedaTriggerMetadata           `json:"metadata,omitempty"`
	Name              string                        `json:"name,omitempty"`
	UseCachedMetrics  *bool                         `json:"useCachedMetrics,omitempty"`
	MetricType        KedaTriggerMetricType         `json:"metricType,omitempty"`
	AuthenticationRef *KedaTriggerAuthenticationRef `json:"authenticationRef,omitempty"`
}

type LocalOptions []LocalOptionsItem

type LocalOptionsItemMultiZone struct {
	Enabled bool `json:"enabled,omitempty"`
}

type LocalOptionsItem struct {
	Autoscaling             *OptionsAutoscaling        `json:"autoscaling,omitempty"`
	TimeoutSeconds          *float32                   `json:"timeoutSeconds,omitempty"`
	CapacityAI              *bool                      `json:"capacityAI,omitempty"`
	CapacityAIUpdateMinutes *float32                   `json:"capacityAIUpdateMinutes,omitempty"`
	Spot                    *bool                      `json:"spot,omitempty"`
	Debug                   bool                       `json:"debug,omitempty"`
	Suspend                 bool                       `json:"suspend,omitempty"`
	MultiZone               *LocalOptionsItemMultiZone `json:"multiZone,omitempty"`
	Location                string                     `json:"location"`
}

type OptionsAutoscalingMetric string

const (
	OptionsAutoscalingMetricConcurrency OptionsAutoscalingMetric = "concurrency"
	OptionsAutoscalingMetricCpu         OptionsAutoscalingMetric = "cpu"
	OptionsAutoscalingMetricMemory      OptionsAutoscalingMetric = "memory"
	OptionsAutoscalingMetricRps         OptionsAutoscalingMetric = "rps"
	OptionsAutoscalingMetricLatency     OptionsAutoscalingMetric = "latency"
	OptionsAutoscalingMetricKeda        OptionsAutoscalingMetric = "keda"
	OptionsAutoscalingMetricDisabled    OptionsAutoscalingMetric = "disabled"
)

type OptionsAutoscalingMulti struct {
	Metric any      `json:"metric,omitempty"`
	Target *float32 `json:"target,omitempty"`
}

type OptionsAutoscalingMetricPercentile string

const (
	OptionsAutoscalingMetricPercentileP50 OptionsAutoscalingMetricPercentile = "p50"
	OptionsAutoscalingMetricPercentileP75 OptionsAutoscalingMetricPercentile = "p75"
	OptionsAutoscalingMetricPercentileP99 OptionsAutoscalingMetricPercentile = "p99"
)

type OptionsAutoscalingKedaAdvancedScalingModifiersMetricType string

const (
	OptionsAutoscalingKedaAdvancedScalingModifiersMetricTypeAverageValue OptionsAutoscalingKedaAdvancedScalingModifiersMetricType = "AverageValue"
	OptionsAutoscalingKedaAdvancedScalingModifiersMetricTypeValue        OptionsAutoscalingKedaAdvancedScalingModifiersMetricType = "Value"
	OptionsAutoscalingKedaAdvancedScalingModifiersMetricTypeUtilization  OptionsAutoscalingKedaAdvancedScalingModifiersMetricType = "Utilization"
)

type OptionsAutoscalingKedaAdvancedScalingModifiers struct {
	Target           string                                                   `json:"target,omitempty"`
	ActivationTarget string                                                   `json:"activationTarget,omitempty"`
	MetricType       OptionsAutoscalingKedaAdvancedScalingModifiersMetricType `json:"metricType,omitempty"`
	Formula          string                                                   `json:"formula,omitempty"`
}

type OptionsAutoscalingKedaAdvanced struct {
	ScalingModifiers *OptionsAutoscalingKedaAdvancedScalingModifiers `json:"scalingModifiers,omitempty"`
}

type OptionsAutoscalingKedaFallbackBehavior string

const (
	OptionsAutoscalingKedaFallbackBehaviorStatic                  OptionsAutoscalingKedaFallbackBehavior = "static"
	OptionsAutoscalingKedaFallbackBehaviorCurrentReplicas         OptionsAutoscalingKedaFallbackBehavior = "currentReplicas"
	OptionsAutoscalingKedaFallbackBehaviorCurrentReplicasIfHigher OptionsAutoscalingKedaFallbackBehavior = "currentReplicasIfHigher"
	OptionsAutoscalingKedaFallbackBehaviorCurrentReplicasIfLower  OptionsAutoscalingKedaFallbackBehavior = "currentReplicasIfLower"
)

type OptionsAutoscalingKedaFallback struct {
	FailureThreshold float32                                `json:"failureThreshold"`
	Replicas         float32                                `json:"replicas"`
	Behavior         OptionsAutoscalingKedaFallbackBehavior `json:"behavior,omitempty"`
}

type OptionsAutoscalingKeda struct {
	Triggers              []KedaTrigger                   `json:"triggers,omitempty"`
	Advanced              *OptionsAutoscalingKedaAdvanced `json:"advanced,omitempty"`
	Fallback              *OptionsAutoscalingKedaFallback `json:"fallback,omitempty"`
	PollingInterval       *float32                        `json:"pollingInterval,omitempty"`
	CooldownPeriod        *float32                        `json:"cooldownPeriod,omitempty"`
	InitialCooldownPeriod *float32                        `json:"initialCooldownPeriod,omitempty"`
}

type OptionsAutoscaling struct {
	Metric           OptionsAutoscalingMetric           `json:"metric,omitempty"`
	Multi            []OptionsAutoscalingMulti          `json:"multi,omitempty"`
	MetricPercentile OptionsAutoscalingMetricPercentile `json:"metricPercentile,omitempty"`
	Target           *float32                           `json:"target,omitempty"`
	MaxScale         float32                            `json:"maxScale"`
	MinScale         float32                            `json:"minScale"`
	ScaleToZeroDelay *float32                           `json:"scaleToZeroDelay,omitempty"`
	MaxConcurrency   float32                            `json:"maxConcurrency"`
	Keda             *OptionsAutoscalingKeda            `json:"keda,omitempty"`
}
