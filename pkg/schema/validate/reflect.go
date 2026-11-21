package validate

// reflect.go is the reflection bridge that inverts the schema source of truth:
// instead of hand-written builder funcs (pkg/handlers/schema_<kind>.go) driving a
// *Rule tree, the CANONICAL tagged Go structs (go-libs/schema/*/models.go) are
// walked by reflection to DERIVE the same *Rule tree, which the existing engine
// (eval.go) then runs unchanged against the incoming map[string]any body.
//
// FromType(t) walks a struct type: each field becomes an object key (the `json`
// tag name), its rule derived from the field's Go type + `validate:"..."` tag DSL.
// Nested structs recurse; slices become Array(elemRule); maps become
// Pattern(String, valRule). Well-known named types (base.Name, base.Kind, …) and
// links resolve to the exact prebuilt rules authored in pkg/handlers via the
// registries below — keeping this file free of any pkg/handlers import (no cycle).
//
// The tag DSL grammar is documented on go-libs/schema/ipSet/models.go.

import (
	"reflect"
	"strconv"
	"strings"
	"sync"
)

// --- named-type dispatch ---

// EnumValuer is implemented by a typed-string enum whose allowed values cannot be
// enumerated by reflection; the reflection bridge calls EnumValues() to add
// .Valid(...) for the field's base rule.
type EnumValuer interface{ EnumValues() []string }

// Refiner lets a struct type attach an object-level cross-field custom (the
// "complex parts of Joi" — cross-field validation). obj is the validated object
// (post type-coercion, siblings readable/mutatable); h carries the Context. Return
// an error to fail; mutating obj is honoured. Equivalent to the .Refine() builder,
// applied automatically when a reflected struct type implements it.
type Refiner interface {
	Refine(obj map[string]any, h *Helper) error
}

var (
	enumValuerType = reflect.TypeOf((*EnumValuer)(nil)).Elem()
	refinerType    = reflect.TypeOf((*Refiner)(nil)).Elem()
)

// FromType builds the *Rule for a struct type (following a leading pointer) by
// pure reflection over its fields + `validate:` tags. Field types implementing
// Validator resolve to their own rule; a resource's ValidationRule typically calls
// Reflect(self) (which routes here) then layers object-level .Refine/.Xor.
func FromType(t reflect.Type) *Rule { return baseRuleForType(deref(t), tokens{}) }

func deref(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

// structRule walks a struct type's fields into an Object rule and applies any
// object-level hooks (JoiObjecter / Refiner).
func structRule(t reflect.Type) *Rule {
	keys := map[string]*Rule{}
	order := make([]string, 0, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.PkgPath != "" { // unexported
			continue
		}
		name := jsonName(f)
		if name == "-" || name == "" {
			continue
		}
		keys[name] = fieldRule(f)
		order = append(order, name)
	}
	// Field order, not sorted: it IS the Joi declaration order, which decides
	// which key an abortEarly failure reports. See ObjectOrdered.
	r := ObjectOrdered(order, keys)
	return applyObjectHooks(t, r)
}

// fieldRule derives the rule for one struct field from its type + `validate:` tag.
func fieldRule(f reflect.StructField) *Rule {
	toks := parseJoiTag(fieldTag(f))
	ft := deref(f.Type)
	base := baseRuleForType(ft, toks)
	return applyModifiers(base, toks)
}

// baseRuleForType picks the primitive/base rule for a type, honouring the
// base-determining tags (localLink/imageLink/date), the base-type registry, the
// EnumValuer Valuer registry, then the Go kind. toks may be empty (slice/map
// elements carry no tags, as Go cannot tag them).
func baseRuleForType(ft reflect.Type, toks tokens) *Rule {
	// 0. a type that defines its own rule (Validator) wins over the registries and
	// the Go-kind fallback — the strongly-typed replacement for typeRegistry /
	// custom= / pattern= / localLink=. (The top-level type is dispatched via
	// topRuleForType, and Reflect walks fields via structRule, so this fires only
	// for FIELD/nested types — no recursion into the caller's own ValidationRule.)
	if impl, ok := asValidator(ft); ok {
		return orderRuleByType(impl.ValidationRule(), ft)
	}
	// 1. a registered rule for a type that CANNOT carry a method — a defined interface
	// type standing for a Joi union (envoy's `type Duration any`). Without this such a
	// field has no rule at all and accepts anything.
	if fn, ok := lookupTypeRule(ft); ok {
		return orderRuleByType(fn(), ft)
	}
	// 2. explicit type override (declare a Go string/number as a Joi.date())
	if toks.has("date") {
		return DateRule()
	}
	// 4. typed-string enum (EnumValuer)
	if vals, ok := enumValues(ft); ok {
		return String().Valid(toAnySlice(vals)...)
	}
	// 5. by Go kind
	switch ft.Kind() {
	case reflect.Struct:
		return structRule(ft)
	case reflect.Slice, reflect.Array:
		return Array(baseRuleForType(deref(ft.Elem()), tokens{}))
	case reflect.Map:
		return Object(nil).Pattern(String(), baseRuleForType(deref(ft.Elem()), tokens{}))
	case reflect.String:
		return String()
	case reflect.Bool:
		return Boolean()
	case reflect.Float32, reflect.Float64,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return Number()
	default:
		return Any()
	}
}

// applyModifiers layers the constraint tags onto the base rule. Base-determining
// tags (localLink/imageLink/date) are skipped here — they were consumed by
// baseRuleForType.
func applyModifiers(base *Rule, toks tokens) *Rule {
	r := base
	for _, tk := range toks {
		switch tk.key {
		// base determiners — already handled
		case "localLink", "imageLink", "date":
			// skip
		case "iso":
			r = r.Iso()
		// presence
		case "required":
			r = r.Required()
		case "optional":
			r = r.Optional()
		// Joi .strip(): validated, then removed from the result.
		case "strip":
			r = r.Strip()
		case "forbidden":
			r = r.Forbidden()
		// core
		case "default":
			r = r.Default(parseDefault(tk.val))
		case "allow":
			r = r.Allow(parseAllow(tk.val))
		case "valid":
			r = r.Valid(parseValues(tk.val)...)
		case "invalid":
			r = r.Invalid(parseValues(tk.val)...)
		// string
		case "trim":
			r = r.Trim()
		case "lowercase":
			r = r.Lowercase()
		case "uppercase":
			r = r.Uppercase()
		case "min":
			if f, isFloat := floatTag(tk.val); isFloat {
				r = r.NumMin(f)
			} else {
				r = r.Min(atoi(tk.val))
			}
		case "max":
			if f, isFloat := floatTag(tk.val); isFloat {
				r = r.NumMax(f)
			} else {
				r = r.Max(atoi(tk.val))
			}
		case "length":
			r = r.Length(atoi(tk.val))
		case "email":
			r = r.Email()
		case "uri":
			r = r.URI()
		case "uuid":
			r = r.UUID()
		case "domain":
			r = r.Domain()
		// `.domain(opts)` — ANY options object turns the IANA TLD list on, so this is a
		// different rule from a bare `domain`, not a decoration of it.
		case "domaintlds":
			r = r.DomainTLDs()
		// number
		case "integer":
			r = r.Integer()
		case "positive":
			r = r.Positive()
		case "negative":
			r = r.Negative()
		case "port":
			r = r.Port()
		case "precision":
			r = r.Precision(atoi(tk.val))
		case "minRef":
			r = r.MinRef(tk.val)
		case "maxRef":
			r = r.MaxRef(tk.val)
		// array
		case "sort":
			r = r.Sort()
		case "unique":
			r = r.Unique()
		case "uniqueBy":
			r = r.Unique(tk.val)
		case "uniqueComparing":
			r = r.UniqueComparing(tk.val)
		case "dedupe":
			r = r.Dedupe()
		case "dedupeBy":
			r = r.Dedupe(tk.val)
		case "single":
			r = r.Single()
		// meta
		case "readOnly":
			r = r.ReadOnly()
		case "unmodifiable":
			r = r.Unmodifiable()
		default:
			panic("validate: unknown validate tag token " + tk.key)
		}
	}
	return r
}

func applyObjectHooks(t reflect.Type, r *Rule) *Rule {
	// A reflected struct type may attach an object-level cross-field custom by
	// implementing Refiner (equivalent to the .Refine() builder on its ValidationRule).
	if impl, ok := asRefiner(t); ok {
		r = withRefine(r, impl.Refine)
	}
	return r
}

func withRefine(r *Rule, refine func(map[string]any, *Helper) error) *Rule {
	return r.Custom(func(v any, h *Helper) (any, error) {
		obj, _ := v.(map[string]any)
		if err := refine(obj, h); err != nil {
			return nil, err
		}
		return obj, nil
	})
}

// --- interface probing (value or pointer receiver) ---

func enumValues(t reflect.Type) ([]string, bool) {
	if t.Implements(enumValuerType) {
		return reflect.Zero(t).Interface().(EnumValuer).EnumValues(), true
	}
	if reflect.PtrTo(t).Implements(enumValuerType) {
		return reflect.New(t).Interface().(EnumValuer).EnumValues(), true
	}
	return nil, false
}

func asRefiner(t reflect.Type) (Refiner, bool) {
	if t.Implements(refinerType) {
		return reflect.Zero(t).Interface().(Refiner), true
	}
	if reflect.PtrTo(t).Implements(refinerType) {
		return reflect.New(t).Interface().(Refiner), true
	}
	return nil, false
}

// --- tag parsing ---

type token struct {
	key string
	val string // "" if the token had no '=' (or an explicitly empty value)
}

type tokens []token

func (ts tokens) get(key string) (string, bool) {
	for _, t := range ts {
		if t.key == key {
			return t.val, true
		}
	}
	return "", false
}

func (ts tokens) has(key string) bool { _, ok := ts.get(key); return ok }

// fieldTag returns the `validate:"..."` validation tag for a field.
func fieldTag(f reflect.StructField) string { return f.Tag.Get("validate") }

// parseJoiTag splits a `validate:"a,b=c,allow="` tag into ordered tokens. Values that
// themselves contain a comma are not supported (by design — named-regex/custom
// registries keep complex values out of tags).
func parseJoiTag(tag string) tokens {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return tokens{}
	}
	var out tokens
	for _, part := range strings.Split(tag, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if i := strings.IndexByte(part, '='); i >= 0 {
			out = append(out, token{key: part[:i], val: part[i+1:]})
		} else {
			out = append(out, token{key: part})
		}
	}
	return out
}

func jsonName(f reflect.StructField) string {
	tag, ok := f.Tag.Lookup("json")
	if !ok {
		return f.Name
	}
	name := strings.Split(tag, ",")[0]
	if name == "" {
		return f.Name
	}
	return name
}

// parseValues parses valid=/invalid= pipe-separated literals (numeric → float64,
// else string).
func parseValues(v string) []any {
	parts := strings.Split(v, "|")
	out := make([]any, 0, len(parts))
	for _, p := range parts {
		if f, err := strconv.ParseFloat(p, 64); err == nil {
			out = append(out, f)
		} else {
			out = append(out, p)
		}
	}
	return out
}

// parseAllow parses an allow=<v> value: bare "" → allow empty string; "null" →
// allow nil; otherwise the literal.
func parseAllow(v string) any {
	switch v {
	case "":
		return ""
	case "null":
		return nil
	default:
		return parseScalar(v)
	}
}

// parseDefault parses a default=<lit>: {} → fresh empty object (func), [] → fresh
// empty array (func), true/false → bool, "str" → unquoted string, numeric →
// float64, else the literal string.
func parseDefault(v string) any {
	switch v {
	case "{}":
		return func() any { return map[string]any{} }
	case "[]":
		return func() any { return []any{} }
	default:
		return parseScalar(v)
	}
}

func parseScalar(v string) any {
	switch v {
	case "true":
		return true
	case "false":
		return false
	}
	if len(v) >= 2 && v[0] == '"' && v[len(v)-1] == '"' {
		return v[1 : len(v)-1]
	}
	if f, err := strconv.ParseFloat(v, 64); err == nil {
		return f
	}
	return v
}

func toAnySlice(ss []string) []any {
	out := make([]any, len(ss))
	for i, s := range ss {
		out[i] = s
	}
	return out
}

// floatTag reports whether a min=/max= value is a non-integer bound (contains a
// decimal point, e.g. min=0.1), returning the parsed float. Integer values fall
// through to the type-dispatched .Min/.Max (string length / int bound / array len).
func floatTag(s string) (float64, bool) {
	s = strings.TrimSpace(s)
	if !strings.Contains(s, ".") {
		return 0, false
	}
	f, err := strconv.ParseFloat(s, 64)
	return f, err == nil
}

func atoi(s string) int {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		panic("validate: expected integer in validate tag, got " + strconv.Quote(s))
	}
	return n
}

// --- declaration order recovery ---

// orderRuleByType re-imposes a struct's FIELD order on a hand-built object rule.
//
// A rule authored as Object(map[string]*Rule{...}) loses its source order — a Go
// map literal has none — so Object() sorts the keys. That order is not cosmetic:
// abortEarly reports the FIRST failing key, so a sorted rule blames a different
// field than Joi does (mk8s's provider rules were alphabetical against TS's
// declaration order). The canonical struct for the same shape is generated from
// the Joi source and DOES carry the declaration order, so wherever a rule is
// attached to a type it is reordered to that type here — once, at build time.
//
// Keys with no matching field keep their relative order after the matched ones.
func orderRuleByType(r *Rule, t reflect.Type) *Rule {
	return orderRule(r, t, map[*Rule]bool{})
}

func orderRule(r *Rule, t reflect.Type, seen map[*Rule]bool) *Rule {
	if r == nil || t == nil || seen[r] {
		return r
	}
	seen[r] = true
	defer delete(seen, r)
	t = deref(t)

	switch {
	case r.typ == tArray && (t.Kind() == reflect.Slice || t.Kind() == reflect.Array):
		var items []*Rule
		changed := false
		for _, it := range r.arrItems {
			o := orderRule(it, t.Elem(), seen)
			changed = changed || o != it
			items = append(items, o)
		}
		if !changed {
			return r
		}
		c := r.clone()
		c.arrItems = items
		return c

	case r.typ == tObject && t.Kind() == reflect.Map:
		if !r.hasPatterns() {
			return r
		}
		var out []keyPattern
		changed := false
		for _, p := range r.patterns {
			o := orderRule(p.val, t.Elem(), seen)
			changed = changed || o != p.val
			out = append(out, keyPattern{key: p.key, val: o})
		}
		if !changed {
			return r
		}
		c := r.clone()
		c.patterns = out
		return c

	case r.typ == tObject && t.Kind() == reflect.Struct:
		fields := map[string]reflect.Type{}
		var order []string
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if f.PkgPath != "" {
				continue
			}
			name := jsonName(f)
			if name == "-" || name == "" {
				continue
			}
			fields[name] = f.Type
			if _, ok := r.keys[name]; ok {
				order = append(order, name)
			}
		}
		if len(order) == 0 {
			return r
		}
		in := map[string]bool{}
		for _, k := range order {
			in[k] = true
		}
		for _, k := range r.keyOrder {
			if !in[k] {
				order = append(order, k)
			}
		}
		keys := make(map[string]*Rule, len(r.keys))
		for k, v := range r.keys {
			if ft, ok := fields[k]; ok {
				keys[k] = orderRule(v, ft, seen)
			} else {
				keys[k] = v
			}
		}
		c := r.clone()
		c.keys = keys
		c.keyOrder = order
		return c
	}
	return r
}

// typeRules maps a Go type to its rule, for types that cannot define a
// ValidationRule method: a defined INTERFACE type (Go forbids methods on one), which
// is how the generator spells a Joi alternatives() union. Registered from an init()
// in the package that owns the type.
var (
	typeRules   = map[reflect.Type]func() *Rule{}
	typeRulesMu sync.RWMutex
)

// RegisterTypeRule attaches a rule to a type by identity. Panics on a duplicate
// registration for the same type, which would make the winner load-order dependent.
func RegisterTypeRule(t reflect.Type, fn func() *Rule) {
	typeRulesMu.Lock()
	defer typeRulesMu.Unlock()
	if _, dup := typeRules[t]; dup {
		panic("validate: duplicate RegisterTypeRule for " + t.String())
	}
	typeRules[t] = fn
}

func lookupTypeRule(t reflect.Type) (func() *Rule, bool) {
	typeRulesMu.RLock()
	defer typeRulesMu.RUnlock()
	fn, ok := typeRules[t]
	return fn, ok
}
