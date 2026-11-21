package gvc

import (
	"encoding/json"
	"testing"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// TestLoadBalancerMatchesJoi pins gvc.ts GvcLoadBalancerConfig after master 54353d516:
// distributedIngress is retired but still validated, then stripped; deploymentCount is
// an integer 1..10. Verdicts captured from the built gvc.ts.
func TestLoadBalancerMatchesJoi(t *testing.T) {
	for _, c := range []struct {
		body string
		want string // "" = accepted, else the Joi error type
	}{
		{`{}`, ""},
		{`{"distributedIngress":true}`, ""},
		{`{"distributedIngress":"x"}`, "boolean.base"},
		{`{"deploymentCount":3}`, ""},
		{`{"deploymentCount":0}`, "number.min"},
		{`{"deploymentCount":11}`, "number.max"},
		{`{"deploymentCount":1.5}`, "number.integer"},
	} {
		row, verr := validate.Decode[GvcLoadBalancerConfig](validate.Bytes(c.body), nil, validate.Prefs{})
		got := ""
		if verr != nil {
			got = verr.Details()[0].Type
		}
		if got != c.want {
			t.Errorf("%s: got %q, want %q", c.body, got, c.want)
			continue
		}
		if verr == nil {
			// Joi drops the stripped key from the result, so it is never stored or echoed.
			b, _ := json.Marshal(row)
			var m map[string]any
			_ = json.Unmarshal(b, &m)
			if _, ok := m["distributedIngress"]; ok {
				t.Errorf("%s: distributedIngress survived validation: %s", c.body, b)
			}
		}
	}
}
