package validate

import "testing"

// C3/C4: Joi picks an alternatives error by the RESULTING errors, not by the schema
// shape (alternatives.js internals.errors): one failure passes through, failures that
// are all *.base / any.only become alternatives.types listing the type names and
// valids, exactly one complex failure passes through, otherwise alternatives.match.
// Verdicts captured from joi@17.6.0, not inferred.
func TestAlternativesErrorSelectionMatchesJoi(t *testing.T) {
	A, O := Alternatives, Object
	for _, c := range []struct {
		name     string
		rule     *Rule
		in       any
		wantType string
		wantMsg  string
	}{
		{"bare-types", A(String(), Number()), true, "alternatives.types", "\"value\" must be one of [string, number]"},
		{"bare-types-null", A(String(), Number()), nil, "alternatives.types", "\"value\" must be one of [string, number]"},
		{"bare-types-obj", A(String(), Number()), map[string]any{"a": 1.0}, "alternatives.types", "\"value\" must be one of [string, number]"},
		{"one-branch-min", A(String().Min(5)), "ab", "string.min", "\"value\" length must be at least 5 characters long"},
		{"two-min", A(String().Min(5), Number().Min(10)), 3.0, "number.min", "\"value\" must be greater than or equal to 10"},
		{"two-min-str", A(String().Min(5), Number().Min(10)), "ab", "string.min", "\"value\" length must be at least 5 characters long"},
		{"str-or-obj", A(String(), O(map[string]*Rule{"a": String().Required()})), map[string]any{}, "any.required", "\"a\" is required"},
		{"valids", A(String().Valid("a"), String().Valid("b")), "z", "alternatives.types", "\"value\" must be one of [a, b]"},
		{"valids-mixed", A(String().Valid("a"), Number()), "b", "alternatives.types", "\"value\" must be one of [a, number]"},
		{"valids-mixed2", A(String().Valid("a"), Number()), true, "alternatives.types", "\"value\" must be one of [a, number]"},
		{"min-or-star", A(String().Min(5), String().Valid("*")), "ab", "string.min", "\"value\" length must be at least 5 characters long"},
		{"two-complex", A(O(map[string]*Rule{"x": String().Required()}), O(map[string]*Rule{"y": String().Required()})), map[string]any{}, "alternatives.match", "\"value\" does not match any of the allowed types"},
		{"three-types", A(String(), Number(), Boolean(), DateRule()), map[string]any{"o": 1.0}, "alternatives.types", "\"value\" must be one of [string, number, boolean, date]"},
		{"three-types-arr", A(String(), Number(), Boolean(), DateRule()), []any{"a"}, "alternatives.types", "\"value\" must be one of [string, number, boolean, date]"},
		{"nested-arr", A(Array(String()), String()), []any{1.0}, "string.base", "\"[0]\" must be a string"},
		{"one-obj", A(O(map[string]*Rule{"a": String().Required()})), map[string]any{"a": 1.0}, "string.base", "\"a\" must be a string"},
		{"one-bare", A(String()), 5.0, "string.base", "\"value\" must be a string"},
		{"req-missing", O(map[string]*Rule{"k": A(String(), Number()).Required()}), map[string]any{}, "any.required", "\"k\" is required"},
		{"in-object", O(map[string]*Rule{"k": A(String().Min(5), Number().Min(10))}), map[string]any{"k": 3.0}, "number.min", "\"k\" must be greater than or equal to 10"},
		{"in-object-complex", O(map[string]*Rule{"k": A(O(map[string]*Rule{"a": String().Required()}), String())}), map[string]any{"k": map[string]any{}}, "any.required", "\"k.a\" is required"},
		{"date-or-num", A(DateRule(), Number()), "zzz", "alternatives.types", "\"value\" must be one of [date, number]"},
		{"nested-alt", A(A(String().Min(5), Number()), Boolean()), "ab", "string.min", "\"value\" length must be at least 5 characters long"},
	} {
		_, err := c.rule.ValidateAt(nil, c.in, nil, Prefs{})
		if err == nil {
			t.Errorf("%s: want %s, got success", c.name, c.wantType)
			continue
		}
		e := err.(*Error)
		if e.Type != c.wantType || e.Message != c.wantMsg {
			t.Errorf("%s: got %s %q, Joi says %s %q", c.name, e.Type, e.Message, c.wantType, c.wantMsg)
		}
	}
}
