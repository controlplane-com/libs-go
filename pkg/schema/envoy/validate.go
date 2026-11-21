package envoy

// Validation for `spec.sidecar.envoy`. Ports nodelibs/schema/src/envoy.ts.
//
// The only rule at this level is the object-level refine `uiFilterNamingValidator`,
// which is what keeps a `cpln_`-named filter from referring to a cluster / jwt_authn
// provider pair that does not exist. Without it the service accepted an envoy config
// naming a cluster it never declares.
//
// workload's sidecar envoy is a type ALIAS of EnvoyFilters, so this rule covers the
// workload route as well as the gvc one.

import (
	"reflect"
	"sort"
	"strings"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

const cplnPrefix = "cpln_"

// ValidationRule ports `EnvoyFilters = Joi.object({...}).custom(uiFilterNamingValidator)`.
func (EnvoyFilters) ValidationRule() *validate.Rule {
	return validate.Reflect(EnvoyFilters{}).CustomTyped(uiFilterNamingValidator)
}

// uiFilterNamingValidator ports envoy.ts:20. It is reflective so the one function
// serves both evaluators: the typed walk hands it *EnvoyFilters (clusters are typed
// structs), the generic one a map.
func uiFilterNamingValidator(value any, h *validate.Helper) (any, error) {
	clusters := sliceItems(validate.Get(value, "clusters"))
	https := sliceItems(validate.Get(value, "http"))

	cplnName := ""
	// The field name and index are carried because the `providers` choice below is
	// made from the BODY's key order, which the engine addresses by path.
	scan := func(field string, items []any) {
		for i, item := range items {
			if n := str(validate.Get(item, "name")); strings.HasPrefix(n, cplnPrefix) {
				cplnName = n
			}
			if la := validate.Get(item, "load_assignment"); la != nil {
				if n := str(validate.Get(la, "cluster_name")); strings.HasPrefix(n, cplnPrefix) {
					cplnName = n
				}
			}
			tc := validate.Get(item, "typed_config")
			if tc == nil {
				continue
			}
			if n := firstCplnProvider(validate.Get(tc, "providers"), h, field, i); n != "" {
				cplnName = n
			}
			for _, rule := range sliceItems(validate.Get(tc, "rules")) {
				req := validate.Get(rule, "requires")
				if req == nil {
					continue
				}
				if n := str(validate.Get(req, "provider_name")); strings.HasPrefix(n, cplnPrefix) {
					cplnName = n
				}
			}
		}
	}
	scan("clusters", clusters)
	scan("http", https)
	if cplnName == "" {
		return value, nil
	}

	found := false
	for _, c := range clusters {
		la := validate.Get(c, "load_assignment")
		if la != nil && str(validate.Get(c, "name")) == cplnName && str(validate.Get(la, "cluster_name")) == cplnName {
			found = true
			break
		}
	}
	if !found {
		return nil, h.Errorf("any.custom",
			"There does not exist a cluster filter using the name: %s for the 'name' and 'load_assignment.cluster_name'", cplnName)
	}

	for _, f := range https {
		tc := validate.Get(f, "typed_config")
		if tc == nil {
			continue
		}
		if !hasProvider(validate.Get(tc, "providers"), cplnName) {
			continue
		}
		for _, rule := range sliceItems(validate.Get(tc, "rules")) {
			req := validate.Get(rule, "requires")
			if req != nil && str(validate.Get(req, "provider_name")) == cplnName {
				return value, nil
			}
		}
	}
	return nil, h.Errorf("any.custom", "There does not exist a jwt authn filter using %s as a provider", cplnName)
}

// firstCplnProvider is `Object.keys(providers).find(p => p.startsWith('cpln_'))`.
//
// The key it returns decides which cluster/jwt pair the guard demands, so the order
// is load-bearing: picking the first SORTED key rather than the first the body wrote
// made this port store a config TS refuses and refuse one TS stores, in both
// directions. validate.OrderedKeys answers from the wire (see validate/order.go),
// falling back to sorted only where the body's order is genuinely gone.
func firstCplnProvider(providers any, h *validate.Helper, field string, i int) string {
	for _, k := range validate.OrderedKeys(providers, h, field, i, "typed_config", "providers") {
		if strings.HasPrefix(k, cplnPrefix) {
			return k
		}
	}
	return ""
}

func hasProvider(providers any, name string) bool {
	for _, k := range keys(providers) {
		if k == name {
			return true
		}
	}
	return false
}

func keys(v any) []string {
	rv := deref(reflect.ValueOf(v))
	if !rv.IsValid() || rv.Kind() != reflect.Map {
		return nil
	}
	out := make([]string, 0, rv.Len())
	for _, k := range rv.MapKeys() {
		if k.Kind() == reflect.String {
			out = append(out, k.String())
		}
	}
	sort.Strings(out)
	return out
}

// sliceItems yields each element of a slice field in the form validate.Get accepts:
// a struct element as a pointer to it, anything else as its own value.
func sliceItems(v any) []any {
	rv := deref(reflect.ValueOf(v))
	if !rv.IsValid() || rv.Kind() != reflect.Slice {
		return nil
	}
	out := make([]any, 0, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		e := rv.Index(i)
		if e.Kind() == reflect.Interface {
			e = e.Elem()
		}
		if !e.IsValid() {
			continue
		}
		e = deref(e)
		if !e.IsValid() {
			continue
		}
		if e.Kind() == reflect.Struct && e.CanAddr() {
			out = append(out, e.Addr().Interface())
			continue
		}
		out = append(out, e.Interface())
	}
	return out
}

func deref(v reflect.Value) reflect.Value {
	for v.IsValid() && v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return reflect.Value{}
		}
		v = v.Elem()
	}
	return v
}

func str(v any) string { s, _ := v.(string); return s }
