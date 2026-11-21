// Package validate is the runtime validation subsystem — the Go equivalent of
// the Joi validation the TS service runs on create/patch bodies. It implements
// the DECLARATIVE rule set (required, default, valid-enum, min/max, integer,
// pattern, type) that the plan identifies as ~88% mechanically mappable from the
// Joi schemas, plus a Custom hook for the irreducible `.custom()`/`.when()`
// rules. Schemas are registered per resource kind and applied by the CRUD engine.
//
// The declarative rules here are hand-authored for the ported resources; the
// intended end-state is to have nodelibs/schema/joi-to-go emit these Schemas
// directly from the Joi source of truth (see port-to-go.md "Validation").
package validate

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
)

// FieldRule is the set of declarative constraints on one field (dot-path).
type FieldRule struct {
	Path     string // e.g. "spec.initialCapacity"
	Required bool
	// Default is applied when the field is absent/nil. A string of the form
	// "$name" references another top-level field's value (Joi.ref).
	Default any
	Enum    []any          // valid() — allowed values
	Min     *float64       // numeric minimum (inclusive)
	Max     *float64       // numeric maximum (inclusive)
	Integer bool           // must be a whole number
	Pattern *regexp.Regexp // string must match
	Type    string         // "number" | "string" | "object" | "array" | "boolean"
}

// Schema is a resource's validation definition. Two authoring styles coexist:
//   - Root: a combinator *Rule (the Joi-parity engine, rule.go/eval.go) — the
//     target style all resources migrate to.
//   - Fields/Custom: the legacy declarative FieldRule path (retained until each
//     fan-out agent migrates its resource to a Root rule).
//
// When Root != nil it takes precedence.
type Schema struct {
	Root *Rule

	// build, when set, lazily produces Root on first ApplyCtx — used by
	// reflection registration so a kind's rule is compiled AFTER all init()s have
	// populated the type/custom/pattern/hook registries (eliminates init-order
	// fragility across the many schema files in one package).
	build     func() *Rule
	buildOnce sync.Once

	Fields []FieldRule
	// Custom validators run after declarative checks; return an error to fail.
	// They may mutate obj (e.g. Joi.custom coercions like fileSystemType→pc).
	Custom []func(obj map[string]any) error
}

// Rule returns the schema's combinator root, forcing the lazy build. Returns nil
// for a legacy Fields-only schema. Callers that need to derive a variant of a
// schema (e.g. the create fork) go through this rather than touching Root, which
// is unpopulated until the first ApplyCtx.
func (s *Schema) Rule() *Rule { return s.root() }

// root forces the lazy build and returns Root. Every reader goes through it.
//
// The `if s.Root == nil` pre-check that used to guard the Once was a double-checked
// lock without the atomics: a second goroutine could observe a non-nil s.Root published
// by the builder WITHOUT the happens-before that Once.Do provides, skip the Do, and then
// validate against a rule graph whose interior writes it could not see. Under `go build
// -race`, driving concurrent queries produced ~200 reports, nearly all of them a read in
// resolveEffective/validateString racing the clone() that built the rule. Once.Do is a
// single atomic load once the build has run, so the pre-check bought nothing.
func (s *Schema) root() *Rule {
	if s.build != nil {
		s.buildOnce.Do(func() { s.Root = s.build() })
	}
	return s.Root
}

// Apply validates obj with default Joi prefs and an empty context. It is the
// legacy entry point (kept so existing callers compile); prefer ApplyCtx.
func (s *Schema) Apply(obj map[string]any) error {
	return s.ApplyCtx(obj, nil, Prefs{})
}

// ApplyCtx validates obj against the schema, threading the Joi validation
// context (org/gvc/method) and per-call-site Prefs. For a combinator Root it
// runs the engine and writes the normalized (coerced/defaulted/stripped) result
// back into obj in place. For the legacy Fields path it ignores ctx/prefs and
// runs the declarative checks. Returns nil or a *Error (parity error body).
// ValidateAt validates a single (non-object) value against the rule, reporting under
// base. The scalar counterpart of ApplyAt, for a check that happens outside the schema
// walk but must answer as though it were inside it.
func (r *Rule) ValidateAt(base []string, value any, ctx *Context, prefs Prefs) (any, error) {
	if ctx == nil {
		ctx = &Context{}
	}
	st := &evalState{ctx: ctx, prefs: prefs}
	out, verr := r.validateValue(value, true, base, st)
	if verr != nil {
		return nil, verr
	}
	return out, nil
}

// ApplyAt validates obj against the rule with every error reported UNDER base — for
// validating a sub-document with another schema. TS gets this for free by embedding
// the sub-schema in the parent (group.ts:472 forks `memberQuery` to the org_member
// query schema), so its errors already read `memberQuery.spec.terms[0].property`;
// applying the sub-schema separately without a base path reports at the root and names
// a field the caller cannot find in their request.
func (r *Rule) ApplyAt(base []string, obj map[string]any, ctx *Context, prefs Prefs) error {
	if ctx == nil {
		ctx = &Context{}
	}
	st := &evalState{ctx: ctx, prefs: prefs, parent: mapView(obj)}
	out, verr := r.validateValue(obj, true, base, st)
	if verr != nil {
		return verr
	}
	m, _ := out.(map[string]any)
	for k := range obj {
		delete(obj, k)
	}
	for k, v := range m {
		obj[k] = v
	}
	return nil
}

func (s *Schema) ApplyCtx(obj map[string]any, ctx *Context, prefs Prefs) error {
	root := s.root()
	if root != nil {
		if ctx == nil {
			ctx = &Context{}
		}
		st := &evalState{ctx: ctx, prefs: prefs, parent: mapView(obj)}
		out, verr := root.validateValue(obj, true, nil, st)
		if verr != nil {
			return verr
		}
		m, _ := out.(map[string]any)
		for k := range obj {
			delete(obj, k)
		}
		for k, v := range m {
			obj[k] = v
		}
		return nil
	}
	return s.applyLegacy(obj)
}

// ApplyValue validates a top-level value that need not be an OBJECT, returning the
// normalized result. express's json parser accepts an array as a body under strict mode
// and hands it to the route as-is, so a route whose TS twin passes req.body straight to
// Joi must be able to get Joi's own `object.base` for one — which the map-shaped
// ApplyCtx cannot express.
func (s *Schema) ApplyValue(v any, ctx *Context, prefs Prefs) (any, error) {
	root := s.root()
	if root == nil {
		if m, ok := v.(map[string]any); ok {
			return m, s.applyLegacy(m)
		}
		return v, nil
	}
	if ctx == nil {
		ctx = &Context{}
	}
	st := &evalState{ctx: ctx, prefs: prefs}
	out, verr := root.validateValue(v, true, nil, st)
	if verr != nil {
		return nil, verr
	}
	return out, nil
}

// applyLegacy is the original declarative FieldRule validation.
func (s *Schema) applyLegacy(obj map[string]any) error {
	// 1) defaults first, so later rules see populated values
	for _, f := range s.Fields {
		if f.Default == nil {
			continue
		}
		if v, ok := getPath(obj, f.Path); !ok || v == nil {
			setPath(obj, f.Path, resolveDefault(f.Default, obj))
		}
	}
	// 2) declarative checks
	for _, f := range s.Fields {
		v, ok := getPath(obj, f.Path)
		present := ok && v != nil
		if !present {
			if f.Required {
				return fmt.Errorf("%q is required", f.Path)
			}
			continue
		}
		if err := checkField(f, v); err != nil {
			return err
		}
	}
	// 3) custom cross-field checks
	for _, c := range s.Custom {
		if err := c(obj); err != nil {
			return err
		}
	}
	return nil
}

func checkField(f FieldRule, v any) error {
	if f.Type != "" && !typeMatches(f.Type, v) {
		return fmt.Errorf("%q must be a %s", f.Path, f.Type)
	}
	if len(f.Enum) > 0 && !inEnum(v, f.Enum) {
		return fmt.Errorf("%q must be one of %v", f.Path, f.Enum)
	}
	if f.Integer {
		if n, ok := toNum(v); !ok || n != float64(int64(n)) {
			return fmt.Errorf("%q must be an integer", f.Path)
		}
	}
	if f.Min != nil {
		if n, ok := toNum(v); !ok || n < *f.Min {
			return fmt.Errorf("%q must be greater than or equal to %v", f.Path, *f.Min)
		}
	}
	if f.Max != nil {
		if n, ok := toNum(v); !ok || n > *f.Max {
			return fmt.Errorf("%q must be less than or equal to %v", f.Path, *f.Max)
		}
	}
	if f.Pattern != nil {
		if s, ok := v.(string); !ok || !f.Pattern.MatchString(s) {
			return fmt.Errorf("%q fails to match the required pattern", f.Path)
		}
	}
	return nil
}

// Num returns a *float64 for rule bounds.
func Num(f float64) *float64 { return &f }

// --- path + value helpers ---

func getPath(obj map[string]any, path string) (any, bool) {
	cur := any(obj)
	for _, seg := range strings.Split(path, ".") {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil, false
		}
		v, ok := m[seg]
		if !ok {
			return nil, false
		}
		cur = v
	}
	return cur, true
}

func setPath(obj map[string]any, path string, val any) {
	segs := strings.Split(path, ".")
	cur := obj
	for _, seg := range segs[:len(segs)-1] {
		next, ok := cur[seg].(map[string]any)
		if !ok {
			next = map[string]any{}
			cur[seg] = next
		}
		cur = next
	}
	cur[segs[len(segs)-1]] = val
}

func resolveDefault(def any, obj map[string]any) any {
	if s, ok := def.(string); ok && strings.HasPrefix(s, "$") {
		return obj[s[1:]]
	}
	return def
}

func typeMatches(t string, v any) bool {
	switch t {
	case "number":
		_, ok := toNum(v)
		return ok
	case "string":
		_, ok := v.(string)
		return ok
	case "object":
		_, ok := v.(map[string]any)
		return ok
	case "array":
		_, ok := v.([]any)
		return ok
	case "boolean":
		_, ok := v.(bool)
		return ok
	}
	return true
}

func inEnum(v any, enum []any) bool {
	for _, e := range enum {
		if e == v {
			return true
		}
		if toStr(e) == toStr(v) {
			return true
		}
	}
	return false
}

func toNum(v any) (float64, bool) {
	switch t := v.(type) {
	case float64:
		return t, true
	case int:
		return float64(t), true
	case int64:
		return float64(t), true
	}
	return 0, false
}

func toStr(v any) string { return fmt.Sprintf("%v", v) }
