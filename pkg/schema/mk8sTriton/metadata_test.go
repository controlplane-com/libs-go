package mk8sTriton

import (
	"strings"
	"testing"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// TestLoadBalancerMetadataHasNoMax pins mk8sTriton.ts:33 — the TritonMetadata VALUE is
// `Joi.string().allow(”)` with NO max (it is TritonTags, a different type, that caps
// values at 255). The port carried a Go-only .Max(255) that rejected a metadata value
// TS stores.
func TestLoadBalancerMetadataHasNoMax(t *testing.T) {
	r := tritonMetadataRule()
	long := map[string]any{"k": strings.Repeat("z", 256)}
	if _, err := r.ValidateAt([]string{"metadata"}, long, nil, validate.Prefs{}); err != nil {
		t.Errorf("a 256-char metadata value rejected: %v", err)
	}
	if _, err := r.ValidateAt([]string{"metadata"}, map[string]any{"k": ""}, nil, validate.Prefs{}); err != nil {
		t.Errorf("an empty metadata value rejected: %v", err)
	}
	// The tags value DOES have the max (and a min of 1).
	if _, err := tritonTagsRule().ValidateAt([]string{"tags"}, long, nil, validate.Prefs{}); err == nil {
		t.Error("a 256-char TAG value accepted")
	}
}
