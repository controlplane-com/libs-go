package serviceMap

import (
	"testing"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// TestServiceMapRules pins serviceMap.ts's `.required()`s and the workload `type` enum.
// The port carried none of them: ServiceWorkloadType declared its constants but had no
// EnumValues, so even a reflected rule would not have enforced the enum.
func TestServiceMapRules(t *testing.T) {
	r := validate.For[ServiceMap]()
	if err := r.ApplyCtx(map[string]any{"lastUpdate": "2020-01-02T03:04:05.000Z"}, nil, validate.Prefs{}); err == nil {
		t.Error("a service map with no outbound accepted")
	}
	if err := r.ApplyCtx(map[string]any{"outbound": []any{}}, nil, validate.Prefs{}); err == nil {
		t.Error("a service map with no lastUpdate accepted")
	}
	wl := func(typ string) map[string]any {
		return map[string]any{
			"outbound": []any{map[string]any{
				"workload": map[string]any{"name": "w", "canonicalEndpoint": "e", "portMap": []any{}, "type": typ},
				"gvc":      map[string]any{"name": "g", "alias": "a"},
			}},
			"lastUpdate": "2020-01-02T03:04:05.000Z",
		}
	}
	if err := r.ApplyCtx(wl("serverless"), nil, validate.Prefs{}); err != nil {
		t.Fatalf("a valid service map rejected: %v", err)
	}
	if err := r.ApplyCtx(wl("bogus"), nil, validate.Prefs{}); err == nil {
		t.Error("an unknown workload type accepted")
	}
}
