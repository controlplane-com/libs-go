package validate

import (
	"reflect"
	"regexp"
	"testing"
)

// --- fixtures: a nominal leaf type and a resource type, both via Validator ---

type tCpu string

func (tCpu) ValidationRule() *Rule { return String().Regex(regexp.MustCompile(`^[0-9]+m$`)) }

type tSpec struct {
	Cpu     tCpu   `json:"cpu"`
	Replica int    `json:"replica" validate:"min=1"`
	Name    string `json:"name"`
}

// tRes implements Validator: reflect its fields, then require exactly one of a/b
// analog via a cross-field Refine, proving object-level layering.
type tRes struct {
	Spec tSpec  `json:"spec"`
	Kind string `json:"kind"`
}

func (tRes) ValidationRule() *Rule {
	return Reflect(tRes{}).Refine(func(o map[string]any, h *Helper) error {
		if _, ok := o["kind"]; !ok {
			return h.Errorf("any.required", `"kind" is required`)
		}
		return nil
	})
}

// tPlain implements nothing — must fall back to pure struct-tag reflection.
type tPlain struct {
	N int `json:"n" validate:"min=3"`
}

func TestForValidatorDispatch(t *testing.T) {
	s := For[tRes]()
	if s == nil {
		t.Fatal("For[tRes]() returned nil")
	}
	// leaf nominal type rule is applied (tCpu regex): a bad cpu must fail.
	err := s.ApplyCtx(map[string]any{
		"kind": "x",
		"spec": map[string]any{"cpu": "oops", "replica": 2, "name": "a"},
	}, &Context{}, Prefs{})
	if err == nil {
		t.Fatal("expected tCpu regex to reject cpu=\"oops\"")
	}
	// the object-level Refine is applied: missing kind must fail.
	err = s.ApplyCtx(map[string]any{
		"spec": map[string]any{"cpu": "100m", "replica": 2, "name": "a"},
	}, &Context{}, Prefs{})
	if err == nil {
		t.Fatal("expected Refine to reject missing kind")
	}
	// a fully valid object passes.
	if err := s.ApplyCtx(map[string]any{
		"kind": "x",
		"spec": map[string]any{"cpu": "100m", "replica": 2, "name": "a"},
	}, &Context{}, Prefs{}); err != nil {
		t.Fatalf("valid object rejected: %v", err)
	}
}

func TestForFallbackReflection(t *testing.T) {
	s := For[tPlain]() // implements no interface → pure struct-tag reflection
	if s == nil {
		t.Fatal("For[tPlain]() returned nil")
	}
	if err := s.ApplyCtx(map[string]any{"n": 1}, &Context{}, Prefs{}); err == nil {
		t.Fatal("expected fallback reflection to enforce validate:\"min=3\"")
	}
	if err := s.ApplyCtx(map[string]any{"n": 5}, &Context{}, Prefs{}); err != nil {
		t.Fatalf("valid fallback object rejected: %v", err)
	}
}

func TestOfTypeCachedAndForEquivalent(t *testing.T) {
	a := OfType(reflect.TypeOf(tRes{}))
	b := For[tRes]()
	if a != b {
		t.Fatal("OfType and For[T] must return the same cached *Schema")
	}
}
