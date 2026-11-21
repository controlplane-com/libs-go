package workloadOptions

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// check decodes body into T, validates it against T's rule through the TYPED walk
// (the object rules here carry RefineT customs, which the generic evaluator refuses),
// and returns the error message and the normalized JSON.
func check[T any](t *testing.T, rule *validate.Rule, body string) (string, string) {
	t.Helper()
	var v T
	d, err := validate.DecodeInto(&v, validate.Bytes(body))
	if err != nil {
		t.Fatalf("decode %s: %v", body, err)
	}
	verr := d.Validate(rule, &validate.Context{Org: "o", Gvc: "g", Method: "POST"}, validate.Prefs{})
	out, _ := json.Marshal(v)
	if verr != nil {
		return verr.Message, string(out)
	}
	return "", string(out)
}

func autoscaling(t *testing.T, body string) (string, string) {
	return check[OptionsAutoscaling](t, OptionsAutoscaling{}.ValidationRule(), body)
}

func autoscalingErr(t *testing.T, body string) string {
	t.Helper()
	msg, _ := autoscaling(t, body)
	return msg
}

// assertErr compares a validation message against the expected substring ("" = the
// case must be accepted). A custom's message arrives inside Joi's
// `"<label>" failed custom validation because …` envelope, which is asserted
// end-to-end against the TS oracle rather than re-spelled in every case.
func assertErr(t *testing.T, got, want string) {
	t.Helper()
	switch {
	case want == "" && got != "":
		t.Fatalf("unexpected error: %s", got)
	case want != "" && !strings.Contains(got, want):
		t.Fatalf("got %q want it to contain %q", got, want)
	}
}

// TestAutoscalingNand pins `.nand('metric','multi').nand('target','multi')`
// (workloadOptions.ts:223-225), which the port dropped entirely.
func TestAutoscalingNand(t *testing.T) {
	for _, c := range []struct{ name, body, wantErr string }{
		{"metric+multi", `{"metric":"cpu","multi":[{"metric":"cpu","target":50}]}`, `"metric" must not exist simultaneously with [multi]`},
		{"target+multi", `{"target":50,"multi":[{"metric":"cpu","target":50}]}`, `"target" must not exist simultaneously with [multi]`},
		// A DEFAULTED target must not count as present, or the nand would refuse the
		// very shape it exists to allow.
		{"multi alone", `{"multi":[{"metric":"cpu","target":50}]}`, ""},
		{"metric alone", `{"metric":"cpu","target":50}`, ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			assertErr(t, autoscalingErr(t, c.body), c.wantErr)
		})
	}
}

// TestResourceTargetWhenMetric pins ResourceTarget's `.when('metric', {switch})`
// (workloadOptions.ts:141-157) on both the scalar target and multi[].target.
func TestResourceTargetWhenMetric(t *testing.T) {
	for _, c := range []struct{ name, body, wantErr string }{
		{"cpu over 100", `{"metric":"cpu","target":5000}`, "target can not be greater than 100 when metric is 'cpu'"},
		{"cpu at 100", `{"metric":"cpu","target":100}`, ""},
		{"memory over 100", `{"metric":"memory","target":101}`, "target can not be greater than 100 when metric is 'memory'"},
		{"rps unbounded", `{"metric":"rps","target":5000}`, ""},
		{"keda forbids target", `{"metric":"keda","target":50}`, "target is not allowed when metric is 'keda'"},
		// The metric is ABSENT: Joi's `is:` is required-by-compilation, so no branch
		// applies. A port whose condition matched an absent sibling rejected every
		// bare `{target: N}`.
		{"no metric", `{"target":5000}`, ""},
		{"multi item cpu", `{"multi":[{"metric":"cpu","target":5000}]}`, "target can not be greater than 100 when metric is 'cpu'"},
		{"multi item rps", `{"multi":[{"metric":"rps","target":5000}]}`, ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			assertErr(t, autoscalingErr(t, c.body), c.wantErr)
		})
	}
}

// TestResourceTargetDefault pins the parent-aware default (workloadOptions.ts:134-138)
// on both target sites — 95 unless `multi` is set or the metric is keda.
func TestResourceTargetDefault(t *testing.T) {
	for _, c := range []struct{ name, body, want string }{
		{"scalar defaults", `{"metric":"cpu"}`, `"target":95`},
		{"keda opts out", `{"metric":"keda"}`, `"metric":"keda"`},
		{"multi opts out of scalar", `{"multi":[{"metric":"cpu","target":10}]}`, `"multi":[{"metric":"cpu","target":10}]`},
		{"multi item defaults", `{"multi":[{"metric":"cpu"}]}`, `"multi":[{"metric":"cpu","target":95}]`},
	} {
		t.Run(c.name, func(t *testing.T) {
			errMsg, out := autoscaling(t, c.body)
			if errMsg != "" {
				t.Fatalf("unexpected error: %s", errMsg)
			}
			if !strings.Contains(out, c.want) {
				t.Fatalf("got %s want it to contain %s", out, c.want)
			}
		})
	}
	if _, out := autoscaling(t, `{"metric":"keda"}`); strings.Contains(out, `"target"`) {
		t.Fatalf("keda must not get a target default: %s", out)
	}
	if _, out := autoscaling(t, `{"multi":[{"metric":"cpu","target":10}]}`); strings.Contains(out, `"target":95`) {
		t.Fatalf("multi must not get a scalar target default: %s", out)
	}
}

// TestMetricPercentileDefault pins the distribution-metric percentile default
// (workloadOptions.ts:226-238).
func TestMetricPercentileDefault(t *testing.T) {
	if _, out := autoscaling(t, `{"metric":"latency","target":50}`); !strings.Contains(out, `"metricPercentile":"p50"`) {
		t.Fatalf("latency should default metricPercentile to p50: %s", out)
	}
	if _, out := autoscaling(t, `{"metric":"latency","target":50,"metricPercentile":"p99"}`); !strings.Contains(out, `"metricPercentile":"p99"`) {
		t.Fatalf("explicit percentile must win: %s", out)
	}
	if _, out := autoscaling(t, `{"metric":"cpu","target":50}`); strings.Contains(out, "metricPercentile") {
		t.Fatalf("non-distribution metric should get no percentile: %s", out)
	}
}

// TestKedaFallbackRequiresTrigger pins the Keda object custom (workloadOptions.ts:109-127).
func TestKedaFallbackRequiresTrigger(t *testing.T) {
	rule := OptionsAutoscalingKeda{}.ValidationRule()
	for _, c := range []struct{ name, body, wantErr string }{
		{"no triggers", `{"fallback":{"failureThreshold":3,"replicas":2}}`,
			"KEDA fallback requires at least one trigger to be defined"},
		{"cpu trigger only", `{"triggers":[{"type":"cpu"}],"fallback":{"failureThreshold":3,"replicas":2}}`,
			"KEDA fallback requires at least one trigger with metricType 'AverageValue' (excluding 'cpu' and 'memory' types)"},
		{"non-AverageValue only", `{"triggers":[{"type":"prometheus","metricType":"Value"}],"fallback":{"failureThreshold":3,"replicas":2}}`,
			"KEDA fallback requires at least one trigger with metricType 'AverageValue' (excluding 'cpu' and 'memory' types)"},
		{"implicit AverageValue", `{"triggers":[{"type":"prometheus"}],"fallback":{"failureThreshold":3,"replicas":2}}`, ""},
		{"no fallback", `{"triggers":[{"type":"cpu"}]}`, ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			got, _ := check[OptionsAutoscalingKeda](t, rule, c.body)
			assertErr(t, got, c.wantErr)
		})
	}
}

// TestOptionsCapacityAI pins the DefaultOptions object custom (workloadOptions.ts:281-292)
// and spot's deleteActiveFieldValidator (:267-275), for BOTH options shapes —
// LocalOptionsItem is `DefaultOptions.keys({location})` and inherits them.
func TestOptionsCapacityAI(t *testing.T) {
	cases := []struct{ name, body, wantErr, wantOut string }{
		{"cpu + capacityAI throws", `{"capacityAI":true,"autoscaling":{"metric":"cpu","target":50}}`,
			"CapacityAI may not be enabled when the autoscaling metric is 'cpu'", ""},
		{"cpu defaults capacityAI off", `{"autoscaling":{"metric":"cpu","target":50}}`, "", `"capacityAI":false`},
		{"multi defaults capacityAI off", `{"autoscaling":{"multi":[{"metric":"cpu","target":50}]}}`, "", `"capacityAI":false`},
		{"explicit false with cpu", `{"capacityAI":false,"autoscaling":{"metric":"cpu","target":50}}`, "", `"capacityAI":false`},
		{"capacityAI with rps", `{"capacityAI":true,"autoscaling":{"metric":"rps","target":50}}`, "", `"capacityAI":true`},
	}
	for _, c := range cases {
		t.Run("default/"+c.name, func(t *testing.T) {
			got, out := check[DefaultOptions](t, DefaultOptions{}.ValidationRule(), c.body)
			assertErr(t, got, c.wantErr)
			if c.wantOut != "" && !strings.Contains(out, c.wantOut) {
				t.Fatalf("got %s want it to contain %s", out, c.wantOut)
			}
		})
		t.Run("local/"+c.name, func(t *testing.T) {
			body := `{"location":"//location/l",` + c.body[1:]
			got, out := check[LocalOptionsItem](t, LocalOptionsItem{}.ValidationRule(), body)
			assertErr(t, got, c.wantErr)
			if c.wantOut != "" && !strings.Contains(out, c.wantOut) {
				t.Fatalf("got %s want it to contain %s", out, c.wantOut)
			}
		})
	}
}

// TestSpotIsStripped: TS deletes `spot` from the parent outright, so it never reaches
// the stored document. Go persisted it, which is a stored-spec divergence on an
// otherwise identical accepted request.
func TestSpotIsStripped(t *testing.T) {
	if _, out := check[DefaultOptions](t, DefaultOptions{}.ValidationRule(), `{"spot":true}`); strings.Contains(out, "spot") {
		t.Fatalf("spot must be stripped: %s", out)
	}
	if _, out := check[LocalOptionsItem](t, LocalOptionsItem{}.ValidationRule(), `{"location":"//location/l","spot":false}`); strings.Contains(out, "spot") {
		t.Fatalf("spot must be stripped from localOptions too: %s", out)
	}
}

// TestMultiUniqueComparator pins `.unique((m1, m2) => m1.metric === m2.metric)`
// (workloadOptions.ts:186). Joi's FUNCTION comparator carries no `path` in the error
// context (the string-path spelling does), and `dupeValue` is the EARLIER entry — the
// by-path spelling this replaced reported both wrong, telling a caller the two
// colliding entries were identical when their targets differed.
func TestMultiUniqueComparator(t *testing.T) {
	var v OptionsAutoscaling
	d, err := validate.DecodeInto(&v, validate.Bytes(`{"multi":[{"metric":"cpu","target":50},{"metric":"cpu","target":60}]}`))
	if err != nil {
		t.Fatal(err)
	}
	verr := d.Validate(OptionsAutoscaling{}.ValidationRule(), &validate.Context{Org: "o", Gvc: "g", Method: "POST"}, validate.Prefs{})
	if verr == nil {
		t.Fatal("expected a duplicate-metric rejection")
	}
	det := verr.Details()[0]
	if det.Type != "array.unique" {
		t.Fatalf("type = %q", det.Type)
	}
	if _, has := det.Context["path"]; has {
		t.Fatalf("function comparator must not report a path: %v", det.Context)
	}
	dupe, _ := det.Context["dupeValue"].(map[string]any)
	if dupe["target"] != float64(50) {
		t.Fatalf("dupeValue = %v, want the earlier entry (target 50)", det.Context["dupeValue"])
	}
	cur, _ := det.Context["value"].(map[string]any)
	if cur["target"] != float64(60) {
		t.Fatalf("value = %v, want the repeat (target 60)", det.Context["value"])
	}
	// Distinct metrics stay legal.
	if msg := autoscalingErr(t, `{"multi":[{"metric":"cpu","target":50},{"metric":"memory","target":60}]}`); msg != "" {
		t.Fatalf("unexpected error: %s", msg)
	}
}
