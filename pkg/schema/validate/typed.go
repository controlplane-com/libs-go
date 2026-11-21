package validate

// typed.go is the type-attached validation dispatch: the strongly-typed
// replacement for the string/type-keyed registries in reflect.go. A type carries
// its own validation by implementing Validator; the engine dispatches by Go type,
// lazily and cached. A type that implements nothing validates by pure struct-tag
// reflection (the fallback). This makes "the schema is wired" a compile-time
// property of the type — you cannot forget to register a kind.
//
// During the registry→interface migration this coexists with the registries in
// reflect.go: Validator (when implemented) wins; otherwise the reflected rule +
// registries apply exactly as before. Once every kind implements Validator the
// registries are deleted.

import (
	"reflect"
	"sync"
)

// Validator is implemented by a type that defines its own validation rule. A
// resource struct returns its complete rule (typically Reflect(self) layered with
// object-level .Refine/.Xor); a leaf nominal type returns its own rule (e.g.
// String().Regex(cpuRe)). It is the single interface that replaces Register /
// RegisterType / RegisterCustom / RegisterPattern / RegisterObjectRules /
// RegisterRefine and the kind-string registry.
type Validator interface{ ValidationRule() *Rule }

var validatorType = reflect.TypeOf((*Validator)(nil)).Elem()

// asValidator probes t for Validator with a value or pointer receiver.
func asValidator(t reflect.Type) (Validator, bool) {
	if t.Implements(validatorType) {
		return reflect.Zero(t).Interface().(Validator), true
	}
	if reflect.PtrTo(t).Implements(validatorType) {
		return reflect.New(t).Interface().(Validator), true
	}
	return nil, false
}

var (
	typeSchemaCache = map[reflect.Type]*Schema{}
	typeSchemaMu    sync.RWMutex
)

// OfType returns the cached *Schema for a Go type (following pointers). If the
// type implements Validator, its ValidationRule() is the schema root; otherwise
// the reflected struct-tag rule is used (no custom validation — the fallback).
// The rule is built once, lazily, on first ApplyCtx.
func OfType(t reflect.Type) *Schema {
	t = deref(t)
	typeSchemaMu.RLock()
	s, ok := typeSchemaCache[t]
	typeSchemaMu.RUnlock()
	if ok {
		return s
	}
	typeSchemaMu.Lock()
	defer typeSchemaMu.Unlock()
	if s, ok := typeSchemaCache[t]; ok { // re-check under write lock
		return s
	}
	tt := t
	s = &Schema{build: func() *Rule { return topRuleForType(tt) }}
	typeSchemaCache[tt] = s
	return s
}

// For is generic sugar: OfType for the type parameter T.
func For[T any]() *Schema { return OfType(reflect.TypeOf((*T)(nil)).Elem()) }

// OfTypeIfValidator returns OfType(t) only when t (or *t) implements Validator,
// else nil. It is the transitional dispatch seam: runValidate prefers a type's
// own Validator schema when the type has been migrated, and falls back to the
// legacy kind-string registry (ForKind) otherwise — so migrating a kind is just
// "make its struct implement Validator", with no engine/handler coordination.
func OfTypeIfValidator(t reflect.Type) *Schema {
	if _, ok := asValidator(deref(t)); !ok {
		return nil
	}
	return OfType(t)
}

var (
	immutableFieldsCache = map[reflect.Type][]string{}
	immutableFieldsMu    sync.RWMutex
)

// ImmutableFields returns the top-level object field names (json keys) of the
// type's schema that are marked `readOnly` (server-owned: id/version/created/
// lastModified/links, alias, keys, …) or `unmodifiable` (client-settable on create
// but fixed after: provider, region, …). These are exactly the fields the merge
// path must pin to the STORED value so a PATCH/PUT cannot change them — letting the
// engine enforce immutability from the schema instead of every handler having to
// remember to list them in RetainProps. Result is cached per type; nil when the
// schema root is not a plain object (e.g. an alternatives), so the caller falls back
// to its declared RetainProps.
func ImmutableFields(t reflect.Type) []string {
	t = deref(t)
	immutableFieldsMu.RLock()
	if v, ok := immutableFieldsCache[t]; ok {
		immutableFieldsMu.RUnlock()
		return v
	}
	immutableFieldsMu.RUnlock()

	var out []string
	if root := topRuleForType(t); root != nil {
		for k, r := range root.keys {
			if r != nil && (r.meta.readOnly || r.meta.unmodifiable) {
				out = append(out, k)
			}
		}
	}

	immutableFieldsMu.Lock()
	immutableFieldsCache[t] = out
	immutableFieldsMu.Unlock()
	return out
}

// topRuleForType builds the root rule for a top-level type: its ValidationRule if
// implemented, else the pure reflected struct rule (fallback).
func topRuleForType(t reflect.Type) *Rule {
	if impl, ok := asValidator(t); ok {
		return orderRuleByType(impl.ValidationRule(), t)
	}
	return baseRuleForType(t, tokens{})
}

// Reflect builds the pure struct-tag rule for v's type (following a pointer),
// WITHOUT invoking that type's own ValidationRule — so a resource's
// ValidationRule can call Reflect(self) then layer object-level .Refine/.Xor
// without recursing. Field types that themselves implement Validator DO resolve
// to their own rule (via baseRuleForType).
func Reflect(v any) *Rule { return structRule(deref(reflect.TypeOf(v))) }
