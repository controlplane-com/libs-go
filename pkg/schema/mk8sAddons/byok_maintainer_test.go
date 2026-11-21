package mk8sAddons

import (
	"testing"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// TestByokConfigMaintainerEnv pins master c68aa1e0e: byok config.maintainer is an
// `{env: {[string]:string}}` map like the sibling actuator/middlebox env maps. Its key
// was missing from byokConfigRule, so a valid body got "...maintainer is not allowed".
func TestByokConfigMaintainerEnv(t *testing.T) {
	r := byokConfigRule()
	ok := map[string]any{"maintainer": map[string]any{"env": map[string]any{"FOO": "bar"}}}
	if _, err := r.ValidateAt([]string{"config"}, ok, nil, validate.Prefs{}); err != nil {
		t.Fatalf("maintainer.env rejected: %v", err)
	}
	bad := map[string]any{"maintainer": map[string]any{"env": map[string]any{"FOO": 1}}}
	if _, err := r.ValidateAt([]string{"config"}, bad, nil, validate.Prefs{}); err == nil {
		t.Fatal("maintainer.env with a non-string value was accepted")
	}
}
