package util

// helpers.go holds the generic map/slice value helpers the schema ValidationRule
// customs use across kinds — the go-libs home for what used to be unexported
// getStr/numOf/asArray/truthy/... scattered across data-service/pkg/handlers.
// Pure (no request context), so they live beside the canonical structs.

// GetStr returns m[key] as a string, or "" if absent/nil/not a string.
func GetStr(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	s, _ := m[key].(string)
	return s
}

// GetBool returns m[key] as a bool, or false if absent/nil/not a bool.
func GetBool(m map[string]any, key string) bool {
	if m == nil {
		return false
	}
	b, _ := m[key].(bool)
	return b
}

// NumOf coerces a decoded-JSON number (float64 or int) to float64.
func NumOf(v any) (float64, bool) {
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

// AsArray returns v as a []any, or nil.
func AsArray(v any) []any {
	a, _ := v.([]any)
	return a
}

// Truthy mirrors JS truthiness for a decoded-JSON value.
func Truthy(v any) bool {
	switch t := v.(type) {
	case bool:
		return t
	case string:
		return t != "" && t != "false"
	case float64:
		return t != 0
	}
	return v != nil
}

// ContainsStr reports whether ss contains s.
func ContainsStr(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}

// Difference returns the elements of a not present in b.
func Difference(a, b []string) []string {
	set := make(map[string]bool, len(b))
	for _, x := range b {
		set[x] = true
	}
	out := []string{}
	for _, x := range a {
		if !set[x] {
			out = append(out, x)
		}
	}
	return out
}

// Float64 dereferences an optional number (nil → 0).
func Float64(p *float64) float64 {
	if p == nil {
		return 0
	}
	return *p
}

// Bool dereferences an optional boolean (nil → false).
func Bool(p *bool) bool { return p != nil && *p }

// Ptr returns a pointer to v, for filling an optional field.
func Ptr[T any](v T) *T { return &v }
