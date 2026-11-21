package base

import "testing"

// TestJSString pins JavaScript's `${v}` coercion, which is what every ported message
// interpolating a value has to produce: Go's %v renders a map as map[b:1] and a slice
// as [1 2] where JS renders [object Object] and 1,2.
func TestJSString(t *testing.T) {
	cases := []struct {
		in   any
		want string
	}{
		{map[string]any{"b": 1}, "[object Object]"},
		{[]any{1, 2}, "1,2"},
		{[]any{1}, "1"},
		{[]any{nil, 1}, ",1"},
		{[]any{[]any{1, 2}, 3}, "1,2,3"},
		{"x", "x"},
		{true, "true"},
		{float64(1), "1"},
		{float64(1.5), "1.5"},
		{nil, "null"},
	}
	for _, c := range cases {
		if got := JSString(c.in); got != c.want {
			t.Errorf("JSString(%#v) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestContainsJSWhitespace pins the ImageLink whitespace check to JavaScript's \s,
// which covers NBSP/BOM/the U+2000 block (Go's ASCII set did not) and excludes NEL
// (Go's unicode.IsSpace includes it, JS does not).
func TestContainsJSWhitespace(t *testing.T) {
	yes := []string{"a b", "a\tb", "a\u00a0b", "a\ufeffb", "a\u2003b", "a\u3000b", "a\nb"}
	no := []string{"ab", "a\u0085b", "a-b"}
	for _, s := range yes {
		if !ContainsJSWhitespace(s) {
			t.Errorf("ContainsJSWhitespace(%q) = false, want true", s)
		}
	}
	for _, s := range no {
		if ContainsJSWhitespace(s) {
			t.Errorf("ContainsJSWhitespace(%q) = true, want false", s)
		}
	}
}
