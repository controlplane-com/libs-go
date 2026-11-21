package config

import (
	"testing"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// TestConfigRules pins config.ts:6,8 — `kind: valid('config')` and
// `lastModified: Joi.date()`. The port declared the kind constant but no EnumValues
// and no date tag, so both rules were absent.
func TestConfigRules(t *testing.T) {
	r := validate.For[Config[any]]()
	if err := r.ApplyCtx(map[string]any{"kind": "config"}, nil, validate.Prefs{}); err != nil {
		t.Fatalf("kind=config rejected: %v", err)
	}
	if err := r.ApplyCtx(map[string]any{"kind": "bogus"}, nil, validate.Prefs{}); err == nil {
		t.Error("kind=bogus accepted")
	}
	if err := r.ApplyCtx(map[string]any{"lastModified": "nope"}, nil, validate.Prefs{}); err == nil {
		t.Error("a non-date lastModified accepted")
	}
	if err := r.ApplyCtx(map[string]any{"lastModified": "2020-01-02T03:04:05.000Z"}, nil, validate.Prefs{}); err != nil {
		t.Errorf("a valid lastModified rejected: %v", err)
	}
}
