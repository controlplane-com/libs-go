/* hand-maintained — canonical source of truth (nodelibs/schema/src/workloadOptions.ts).
 * `joi:` tags drive pkg/validate.FromType; DSL on go-libs/schema/ipSet/models.go.
 *
 * Object-level relations/customs wired in pkg/handlers (schema_reflect_workload.go):
 *   - OptionsAutoscaling      : Nand(metric,multi) Nand(target,multi) + Refine
 *       (parent-dependent ResourceTarget default(95); target<=100 when metric
 *        cpu/memory; target forbidden when metric keda; metricPercentile default)
 *   - OptionsAutoscalingMulti : Refine (multi[].target<=100 when metric cpu/memory)
 *   - OptionsAutoscalingKeda  : Refine (fallback requires a non-cpu/memory AverageValue trigger)
 *   - DefaultOptions          : Refine (spot deleteActiveField; capacityAI/cpu rules)
 *   - LocalOptionsItem        : Refine (same DefaultOptions rules)
 *
 * The type-driven optionsValidator (metric/timeout/scale-to-zero by workload type)
 * is applied by workload(A) at spec level (runOptionsValidator), NOT here. */

package workloadOptions

// LocationLink ports `LocalLink.custom(toKind('location'))`: a local link that must
// resolve to a resource of kind `location`. Replaces the `localLink=location` tag
// on LocalOptionsItem.Location.
type LocationLink string

type DefaultOptionsMultiZone struct {
	Enabled *bool `json:"enabled,omitzero" validate:"default=false"`
}

type DefaultOptions struct {
	Autoscaling             *OptionsAutoscaling      `json:"autoscaling,omitzero"`
	TimeoutSeconds          *float64                 `json:"timeoutSeconds,omitzero" validate:"min=1,max=3600,default=5"`
	CapacityAI              *bool                    `json:"capacityAI,omitzero"`
	CapacityAIUpdateMinutes *float64                 `json:"capacityAIUpdateMinutes,omitzero" validate:"integer,positive,min=2"`
	Spot                    *bool                    `json:"spot,omitzero"`
	Debug                   *bool                    `json:"debug,omitzero" validate:"default=false"`
	Suspend                 *bool                    `json:"suspend,omitzero" validate:"default=false"`
	MultiZone               *DefaultOptionsMultiZone `json:"multiZone,omitzero"`
}

type KedaTriggerMetadata map[string]any

type KedaTriggerMetricType string

const (
	KedaTriggerMetricTypeAverageValue KedaTriggerMetricType = "AverageValue"
	KedaTriggerMetricTypeValue        KedaTriggerMetricType = "Value"
	KedaTriggerMetricTypeUtilization  KedaTriggerMetricType = "Utilization"
)

func (KedaTriggerMetricType) EnumValues() []string {
	return []string{"AverageValue", "Value", "Utilization"}
}

type KedaTriggerAuthenticationRef struct {
	Name string `json:"name" validate:"required"`
}

type KedaTrigger struct {
	Type              string                        `json:"type" validate:"required"`
	Metadata          KedaTriggerMetadata           `json:"metadata,omitzero"`
	Name              string                        `json:"name,omitzero"`
	UseCachedMetrics  *bool                         `json:"useCachedMetrics,omitzero"`
	MetricType        KedaTriggerMetricType         `json:"metricType,omitzero"`
	AuthenticationRef *KedaTriggerAuthenticationRef `json:"authenticationRef,omitzero"`
}

type LocalOptions []LocalOptionsItem

type LocalOptionsItemMultiZone struct {
	Enabled *bool `json:"enabled,omitzero" validate:"default=false"`
}

type LocalOptionsItem struct {
	Autoscaling             *OptionsAutoscaling        `json:"autoscaling,omitzero"`
	TimeoutSeconds          *float64                   `json:"timeoutSeconds,omitzero" validate:"min=1,max=3600,default=5"`
	CapacityAI              *bool                      `json:"capacityAI,omitzero"`
	CapacityAIUpdateMinutes *float64                   `json:"capacityAIUpdateMinutes,omitzero" validate:"integer,positive,min=2"`
	Spot                    *bool                      `json:"spot,omitzero"`
	Debug                   *bool                      `json:"debug,omitzero" validate:"default=false"`
	Suspend                 *bool                      `json:"suspend,omitzero" validate:"default=false"`
	MultiZone               *LocalOptionsItemMultiZone `json:"multiZone,omitzero"`
	Location                LocationLink               `json:"location" validate:"required"`
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

func (OptionsAutoscalingMetric) EnumValues() []string {
	return []string{"concurrency", "cpu", "memory", "rps", "latency", "keda", "disabled"}
}

type OptionsAutoscalingMulti struct {
	Metric any      `json:"metric,omitzero" validate:"valid=cpu|memory|rps"`
	Target *float64 `json:"target,omitzero" validate:"integer,min=1,max=20000"`
}

type OptionsAutoscalingMetricPercentile string

const (
	OptionsAutoscalingMetricPercentileP50 OptionsAutoscalingMetricPercentile = "p50"
	OptionsAutoscalingMetricPercentileP75 OptionsAutoscalingMetricPercentile = "p75"
	OptionsAutoscalingMetricPercentileP99 OptionsAutoscalingMetricPercentile = "p99"
)

func (OptionsAutoscalingMetricPercentile) EnumValues() []string {
	return []string{"p50", "p75", "p99"}
}

type OptionsAutoscalingKedaAdvancedScalingModifiersMetricType string

const (
	OptionsAutoscalingKedaAdvancedScalingModifiersMetricTypeAverageValue OptionsAutoscalingKedaAdvancedScalingModifiersMetricType = "AverageValue"
	OptionsAutoscalingKedaAdvancedScalingModifiersMetricTypeValue        OptionsAutoscalingKedaAdvancedScalingModifiersMetricType = "Value"
	OptionsAutoscalingKedaAdvancedScalingModifiersMetricTypeUtilization  OptionsAutoscalingKedaAdvancedScalingModifiersMetricType = "Utilization"
)

func (OptionsAutoscalingKedaAdvancedScalingModifiersMetricType) EnumValues() []string {
	return []string{"AverageValue", "Value", "Utilization"}
}

type OptionsAutoscalingKedaAdvancedScalingModifiers struct {
	Target           string                                                   `json:"target,omitzero"`
	ActivationTarget string                                                   `json:"activationTarget,omitzero"`
	MetricType       OptionsAutoscalingKedaAdvancedScalingModifiersMetricType `json:"metricType,omitzero"`
	Formula          string                                                   `json:"formula,omitzero"`
}

// HpaScalingRules ports workloadOptions.ts HpaScalingRules (master fd601b5f6). Joi
// declares it once and uses it for both scaleUp and scaleDown, so it is one type here
// rather than the generator's two identical copies.
type HpaScalingRules struct {
	StabilizationWindowSeconds *float64               `json:"stabilizationWindowSeconds,omitzero" validate:"integer,min=0,max=3600,optional"`
	SelectPolicy               HpaScalingSelectPolicy `json:"selectPolicy,omitzero" validate:"optional"`
	Policies                   []HpaScalingPolicy     `json:"policies,omitzero" validate:"min=1,optional"`
}

type HpaScalingSelectPolicy string

const (
	HpaScalingSelectPolicyMax      HpaScalingSelectPolicy = "Max"
	HpaScalingSelectPolicyMin      HpaScalingSelectPolicy = "Min"
	HpaScalingSelectPolicyDisabled HpaScalingSelectPolicy = "Disabled"
)

func (HpaScalingSelectPolicy) EnumValues() []string {
	return []string{"Max", "Min", "Disabled"}
}

type HpaScalingPolicy struct {
	Type          HpaScalingPolicyType `json:"type" validate:"required"`
	Value         float64              `json:"value" validate:"integer,min=1,required"`
	PeriodSeconds float64              `json:"periodSeconds" validate:"integer,min=1,max=1800,required"`
}

type HpaScalingPolicyType string

const (
	HpaScalingPolicyTypePods    HpaScalingPolicyType = "Pods"
	HpaScalingPolicyTypePercent HpaScalingPolicyType = "Percent"
)

func (HpaScalingPolicyType) EnumValues() []string {
	return []string{"Pods", "Percent"}
}

type OptionsAutoscalingKedaAdvancedHorizontalPodAutoscalerConfigBehavior struct {
	ScaleUp   *HpaScalingRules `json:"scaleUp,omitzero" validate:"optional"`
	ScaleDown *HpaScalingRules `json:"scaleDown,omitzero" validate:"optional"`
}

type OptionsAutoscalingKedaAdvancedHorizontalPodAutoscalerConfig struct {
	Behavior *OptionsAutoscalingKedaAdvancedHorizontalPodAutoscalerConfigBehavior `json:"behavior,omitzero" validate:"optional"`
}

type OptionsAutoscalingKedaAdvanced struct {
	ScalingModifiers              *OptionsAutoscalingKedaAdvancedScalingModifiers              `json:"scalingModifiers,omitzero"`
	HorizontalPodAutoscalerConfig *OptionsAutoscalingKedaAdvancedHorizontalPodAutoscalerConfig `json:"horizontalPodAutoscalerConfig,omitzero"`
}

type OptionsAutoscalingKedaFallbackBehavior string

const (
	OptionsAutoscalingKedaFallbackBehaviorStatic                  OptionsAutoscalingKedaFallbackBehavior = "static"
	OptionsAutoscalingKedaFallbackBehaviorCurrentReplicas         OptionsAutoscalingKedaFallbackBehavior = "currentReplicas"
	OptionsAutoscalingKedaFallbackBehaviorCurrentReplicasIfHigher OptionsAutoscalingKedaFallbackBehavior = "currentReplicasIfHigher"
	OptionsAutoscalingKedaFallbackBehaviorCurrentReplicasIfLower  OptionsAutoscalingKedaFallbackBehavior = "currentReplicasIfLower"
)

func (OptionsAutoscalingKedaFallbackBehavior) EnumValues() []string {
	return []string{"static", "currentReplicas", "currentReplicasIfHigher", "currentReplicasIfLower"}
}

type OptionsAutoscalingKedaFallback struct {
	FailureThreshold float64                                `json:"failureThreshold" validate:"integer,required"`
	Replicas         float64                                `json:"replicas" validate:"integer,required"`
	Behavior         OptionsAutoscalingKedaFallbackBehavior `json:"behavior,omitzero"`
}

type OptionsAutoscalingKeda struct {
	Triggers              []KedaTrigger                   `json:"triggers,omitzero"`
	Advanced              *OptionsAutoscalingKedaAdvanced `json:"advanced,omitzero"`
	Fallback              *OptionsAutoscalingKedaFallback `json:"fallback,omitzero"`
	PollingInterval       *float64                        `json:"pollingInterval,omitzero" validate:"integer,min=1"`
	CooldownPeriod        *float64                        `json:"cooldownPeriod,omitzero" validate:"integer,min=1"`
	InitialCooldownPeriod *float64                        `json:"initialCooldownPeriod,omitzero" validate:"integer,min=1"`
}

type OptionsAutoscaling struct {
	Metric           OptionsAutoscalingMetric           `json:"metric,omitzero"`
	Multi            []OptionsAutoscalingMulti          `json:"multi,omitzero" validate:"min=1,uniqueComparing=metric"`
	MetricPercentile OptionsAutoscalingMetricPercentile `json:"metricPercentile,omitzero"`
	// Target carries ResourceTarget's PARENT-AWARE default (validate.go) — 95
	// unless `multi` is set or the metric is keda (workloadOptions.ts:134-141) —
	// which no struct tag can express, so it is layered in OptionsAutoscaling's rule.
	Target           *float64                `json:"target,omitzero" validate:"integer,min=1,max=20000"`
	MaxScale         *float64                `json:"maxScale,omitzero" validate:"integer,min=0,default=5"`
	MinScale         *float64                `json:"minScale,omitzero" validate:"integer,min=0,maxRef=maxScale,default=1"`
	ScaleToZeroDelay *float64                `json:"scaleToZeroDelay,omitzero" validate:"min=30,max=3600,default=300"`
	MaxConcurrency   *float64                `json:"maxConcurrency,omitzero" validate:"min=0,max=30000,default=0"`
	Keda             *OptionsAutoscalingKeda `json:"keda,omitzero"`
}
