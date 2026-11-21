package discovery

import (
	"testing"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// TestEndpointUris pins the five `Joi.string().uri()` endpoints of discovery.ts:5-8,22.
// The port carried no tags at all, so every endpoint was an unconstrained string.
func TestEndpointUris(t *testing.T) {
	r := validate.For[DiscoveryEndpoints]()
	for _, k := range []string{"billing", "console", "browser", "hub", "auth"} {
		if err := r.ApplyCtx(map[string]any{k: "https://x/y"}, nil, validate.Prefs{}); err != nil {
			t.Errorf("%s: valid uri rejected: %v", k, err)
		}
		if err := r.ApplyCtx(map[string]any{k: "notauri"}, nil, validate.Prefs{}); err == nil {
			t.Errorf("%s: non-uri accepted", k)
		}
	}
	// The remaining endpoints are plain strings in discovery.ts — no uri rule.
	if err := r.ApplyCtx(map[string]any{"registry": "notauri"}, nil, validate.Prefs{}); err != nil {
		t.Errorf("registry (a plain string in TS) rejected: %v", err)
	}
}
