package mk8sAddons

import (
	"testing"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// TestRegularExpressionIsRE2 pins the deliberate divergence recorded as
// INTENTIONAL-DIVERGENCES A11: mk8sAddons.ts compiles these patterns with the JS
// engine, this port compiles them with RE2 (which is what the consumer — Prometheus
// relabelling — uses), so JS-only syntax is refused at the API instead of failing
// silently on the agent.
func TestRegularExpressionIsRE2(t *testing.T) {
	r := RegularExpression("").ValidationRule()
	for _, ok := range []string{"^kube-.*$", "(a|b)+", `\d{1,3}`} {
		if _, err := r.ValidateAt([]string{"includeNamespaces"}, ok, nil, validate.Prefs{}); err != nil {
			t.Errorf("%q rejected: %v", ok, err)
		}
	}
	// Lookbehind: valid in JS, not expressible in RE2.
	if _, err := r.ValidateAt([]string{"includeNamespaces"}, "(?<=a)b", nil, validate.Prefs{}); err == nil {
		t.Error("a lookbehind pattern was accepted; A11 says this port refuses it")
	}
	// A pattern that is broken in BOTH engines is rejected either way.
	if _, err := r.ValidateAt([]string{"includeNamespaces"}, "([unclosed", nil, validate.Prefs{}); err == nil {
		t.Error("an unterminated character class was accepted")
	}
}
