package workloadOptions

import (
	"testing"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// TestHpaBehaviorMatchesJoi pins keda advanced.horizontalPodAutoscalerConfig (master
// fd601b5f6). Verdicts captured from the built workloadOptions.ts.
func TestHpaBehaviorMatchesJoi(t *testing.T) {
	for _, c := range []struct {
		body string
		want string // "" = accepted, else the Joi error type
	}{
		{`{}`, ""},
		{`{"behavior":{}}`, ""},
		{`{"behavior":{"scaleUp":{"stabilizationWindowSeconds":0,"selectPolicy":"Max","policies":[{"type":"Pods","value":4,"periodSeconds":15}]}}}`, ""},
		{`{"behavior":{"scaleDown":{"stabilizationWindowSeconds":3601}}}`, "number.max"},
		{`{"behavior":{"scaleDown":{"selectPolicy":"Sometimes"}}}`, "any.only"},
		{`{"behavior":{"scaleUp":{"policies":[]}}}`, "array.min"},
		{`{"behavior":{"scaleUp":{"policies":[{"type":"Nodes","value":1,"periodSeconds":1}]}}}`, "any.only"},
		{`{"behavior":{"scaleUp":{"policies":[{"type":"Pods","value":0,"periodSeconds":1}]}}}`, "number.min"},
		{`{"behavior":{"scaleUp":{"policies":[{"type":"Pods","value":1,"periodSeconds":1801}]}}}`, "number.max"},
		{`{"behavior":{"scaleUp":{"policies":[{"type":"Pods","value":1}]}}}`, "any.required"},
		{`{"behavior":{"bogus":1}}`, "object.unknown"},
	} {
		_, verr := validate.Decode[OptionsAutoscalingKedaAdvancedHorizontalPodAutoscalerConfig](validate.Bytes(c.body), nil, validate.Prefs{})
		got := ""
		if verr != nil {
			got = verr.Details()[0].Type
		}
		if got != c.want {
			t.Errorf("%s: got %q, want %q", c.body, got, c.want)
		}
	}
}
