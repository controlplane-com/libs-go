package workloadOptions

// validate.go attaches validation to the canonical workloadOptions types via the
// Validator interface — replacing the old registry-backed struct tags with
// strongly-typed nominal field types. TS reference:
// nodelibs/schema/src/workloadOptions.ts.
//
//   - localLink=location -> LocationLink (LocalLink.custom(toKind('location')))

import (
	"github.com/controlplane-com/libs-go/pkg/schema/base"
	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// distributionMetrics is DISTRIBUTION_METRICS (workloadOptions.ts:12) — the metrics
// whose target is a percentile of a distribution rather than a mean.
var distributionMetrics = map[OptionsAutoscalingMetric]bool{OptionsAutoscalingMetricLatency: true}

// ValidationRule reflects OptionsAutoscaling and layers on the three object-level
// relations TS declares (workloadOptions.ts:223-238):
//
//	.nand('metric','multi').nand('target','multi').custom(metricPercentile default)
//
// plus ResourceTarget's parent-aware default and its `.when('metric')` switch on
// `target`. All of it was missing, so `{metric:'cpu', multi:[…]}`, `{metric:'cpu',
// target:5000}` and `{metric:'keda', target:50}` were all accepted.
func (OptionsAutoscaling) ValidationRule() *validate.Rule {
	return validate.RefineT(
		validate.Reflect(OptionsAutoscaling{}).
			Fork("target", resourceTargetRule).
			Nand("metric", "multi").
			Nand("target", "multi"),
		autoscalingRefine)
}

// ValidationRule gives multi[] items the same ResourceTarget treatment as the scalar
// target — ResourceTarget is the SAME Joi symbol in both places (workloadOptions.ts:175),
// so the default and the metric switch apply per item.
func (OptionsAutoscalingMulti) ValidationRule() *validate.Rule {
	return validate.Reflect(OptionsAutoscalingMulti{}).
		Fork("target", func(r *validate.Rule) *validate.Rule {
			return resourceTargetRule(r).Default(resourceTargetDefault)
		})
}

// ValidationRule ports the KEDA object custom (workloadOptions.ts:109-127): a
// fallback is meaningless without a trigger KEDA can actually fall back on.
func (OptionsAutoscalingKeda) ValidationRule() *validate.Rule {
	return validate.RefineT(validate.Reflect(OptionsAutoscalingKeda{}), kedaFallbackRefine)
}

// ValidationRule ports the DefaultOptions object custom (workloadOptions.ts:281-292)
// and spot's deleteActiveFieldValidator (:267-275).
func (DefaultOptions) ValidationRule() *validate.Rule {
	return validate.RefineT(validate.Reflect(DefaultOptions{}), func(d *DefaultOptions, h *validate.Helper) error {
		return optionsRefine(d.Autoscaling, &d.CapacityAI, &d.Spot, h)
	})
}

// ValidationRule: LocalOptionsItem is `DefaultOptions.keys({location})`
// (workloadOptions.ts:296-298), so it inherits the same object custom.
func (LocalOptionsItem) ValidationRule() *validate.Rule {
	return validate.RefineT(validate.Reflect(LocalOptionsItem{}), func(o *LocalOptionsItem, h *validate.Helper) error {
		return optionsRefine(o.Autoscaling, &o.CapacityAI, &o.Spot, h)
	})
}

// resourceTargetRule layers ResourceTarget's parent-aware default and its
// `.when('metric', {switch: […]})` onto whichever `target` field it is forked onto
// (workloadOptions.ts:134-157).
func resourceTargetRule(r *validate.Rule) *validate.Rule {
	// `is: 'cpu'` in Joi compiles to `Joi.valid('cpu').required()` — the Required()
	// is what makes an ABSENT metric fail the condition instead of matching it.
	metricIs := func(m string) *validate.Rule { return validate.Any().Valid(m).Required() }
	return r.
		When("metric", validate.When{Is: metricIs("cpu"), Then: validate.Number().Custom(targetMax100("cpu"))}).
		When("metric", validate.When{Is: metricIs("memory"), Then: validate.Number().Custom(targetMax100("memory"))}).
		When("metric", validate.When{Is: metricIs("keda"), Then: validate.Number().Custom(targetForbiddenWithKeda)})
}

// targetMax100 is `Joi.number().max(100).message("target can not be greater than 100
// when metric is '<m>'")` — a plain Max() would emit Joi's own wording.
func targetMax100(metric string) validate.CustomFn {
	return func(value any, h *validate.Helper) (any, error) {
		if f, ok := validate.ToFloat(value); ok && f > 100 {
			return nil, h.WithCtx(h.Errorf("number.max", "target can not be greater than 100 when metric is '%s'", metric), "limit", 100, "value", f)
		}
		return value, nil
	}
}

// targetForbiddenWithKeda is the `is:'keda'` branch, a Joi.custom that always throws.
func targetForbiddenWithKeda(value any, h *validate.Helper) (any, error) {
	return nil, h.Errorf("any.custom", "target is not allowed when metric is 'keda'")
}

// resourceTargetDefault returns nil (JS undefined → key stays absent) for the two
// shapes that opt out of a scalar target.
func resourceTargetDefault(parent map[string]any) any {
	if parent == nil {
		return float64(95)
	}
	if m, ok := parent["multi"]; ok && m != nil {
		return nil
	}
	if metric, _ := parent["metric"].(string); metric == "keda" {
		return nil
	}
	return float64(95)
}

// autoscalingRefine ports the OptionsAutoscaling custom: a distribution metric with
// no explicit percentile targets p50 (workloadOptions.ts:226-238).
func autoscalingRefine(a *OptionsAutoscaling, h *validate.Helper) error {
	// ResourceTarget's parent-aware default (workloadOptions.ts:134-138) is applied
	// here rather than as a key .Default(), because this object also carries
	// .nand('target','multi'): the engine counts a key with a default as PRESENT for
	// a presence relation even when the default resolves to undefined, so a plain
	// .Default() turned `{multi:[…]}` — the very shape the nand exists to permit —
	// into an object.nand. Applying it after the relation check is equivalent for
	// every input, since 95 is only ever chosen when `multi` is absent.
	if a.Target == nil && a.Multi == nil && a.Metric != OptionsAutoscalingMetricKeda {
		target := float64(95)
		a.Target = &target
	}
	if a.MetricPercentile == "" && distributionMetrics[a.Metric] {
		a.MetricPercentile = OptionsAutoscalingMetricPercentileP50
	}
	return nil
}

// kedaFallbackRefine ports the Keda object custom (workloadOptions.ts:109-127).
func kedaFallbackRefine(k *OptionsAutoscalingKeda, h *validate.Helper) error {
	if k.Fallback == nil {
		return nil
	}
	if len(k.Triggers) == 0 {
		return h.Errorf("any.custom", "KEDA fallback requires at least one trigger to be defined")
	}
	for _, t := range k.Triggers {
		metricType := t.MetricType
		if metricType == "" {
			metricType = KedaTriggerMetricTypeAverageValue
		}
		if metricType == KedaTriggerMetricTypeAverageValue && t.Type != "cpu" && t.Type != "memory" {
			return nil
		}
	}
	return h.Errorf("any.custom", "KEDA fallback requires at least one trigger with metricType 'AverageValue' (excluding 'cpu' and 'memory' types)")
}

// optionsRefine is the body shared by DefaultOptions and LocalOptionsItem: strip
// `spot` (TS deletes the key outright, so it never reaches the stored document), then
// either downgrade an unset capacityAI or refuse the cpu combination outright.
func optionsRefine(as *OptionsAutoscaling, capacityAI **bool, spot **bool, h *validate.Helper) error {
	*spot = nil
	metricIsCpu := as != nil && as.Metric == OptionsAutoscalingMetricCpu
	multiSet := as != nil && as.Multi != nil
	switch {
	case *capacityAI == nil && (metricIsCpu || multiSet):
		f := false
		*capacityAI = &f
	case *capacityAI != nil && **capacityAI && metricIsCpu:
		return h.Errorf("any.custom", "CapacityAI may not be enabled when the autoscaling metric is 'cpu'")
	}
	return nil
}

func (LocationLink) ValidationRule() *validate.Rule {
	return base.LocalLinkRule().Custom(base.ToKind("location"))
}
