package env

import (
	"reflect"
	"testing"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// TestEnvFromSourceValidation mirrors nodelibs/schema/src/env.test.ts envFrom cases
// (master 4533c0c10): a dictionary secretRef with a valid name passes; a missing
// secretRef, a missing name, or a name failing the resource-name regex are rejected.
func TestEnvFromSourceValidation(t *testing.T) {
	s := validate.OfType(reflect.TypeOf(EnvFromSource{}))
	cases := []struct {
		name string
		in   map[string]any
		ok   bool
	}{
		{"valid dictionary secret ref", map[string]any{"secretRef": map[string]any{"name": "snowflake-secrets"}}, true},
		{"missing secretRef", map[string]any{}, false},
		{"missing secret name", map[string]any{"secretRef": map[string]any{}}, false},
		{"bad secret name", map[string]any{"secretRef": map[string]any{"name": "-bad-name"}}, false},
	}
	for _, c := range cases {
		err := s.ApplyCtx(c.in, &validate.Context{}, validate.Prefs{})
		if (err == nil) != c.ok {
			t.Errorf("%s: want ok=%v, err=%v", c.name, c.ok, err)
		}
	}
}
