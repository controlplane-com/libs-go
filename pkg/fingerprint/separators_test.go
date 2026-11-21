package fingerprint

import "testing"

func TestValueWritesLineSeparatorsRawAsJSONStringify(t *testing.T) {
	cases := []struct {
		v    any
		want string
	}{
		{"a\u2028b", "/ai4lPJ/M3CrELylW7wGKva+nj0"},
		{`x\u2028y`, "e5GtAVac18GwcXKL6SzXzzxVDVY"},
		{map[string]any{"k": "p\u2029q", "a": []any{"\\\u2028"}}, "RztBJyVwpy3aOEjePy5EfIfSDqk"},
	}
	for _, c := range cases {
		if got := Value(c.v); got != c.want {
			t.Fatalf("Value(%q) = %s, want %s (TS fingerprintValue)", c.v, got, c.want)
		}
	}
}
