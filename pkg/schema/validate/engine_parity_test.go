package validate

import (
	"encoding/json"
	"math"
	"regexp"
	"strings"
	"testing"
)

// Joi coerces number strings with its own grammar (lib/types/number.js numberRx), not
// the platform parser. Go's strconv additionally accepts NaN/Inf/Infinity and digit
// separators, which validated and then failed to marshal — a 500 on any number field.
func TestNumberStringCoercionMatchesJoi(t *testing.T) {
	rejected := []string{"NaN", "Inf", "+Inf", "-Inf", "Infinity", "1_0", "0x10", "", "e5", "5e"}
	for _, in := range rejected {
		if _, err := Validate(Number(), in, nil, Prefs{}); err == nil {
			t.Errorf("%q accepted; Joi rejects it", in)
		}
	}
	accepted := map[string]float64{" 5 ": 5, "5": 5, "-2.5": -2.5, ".5": 0.5, "1e3": 1000, "5.": 5}
	for in, want := range accepted {
		got, err := Validate(Number(), in, nil, Prefs{})
		if err != nil {
			t.Errorf("%q rejected: %v", in, err)
			continue
		}
		if f, _ := got.(float64); f != want {
			t.Errorf("%q -> %v, want %v", in, got, want)
		}
	}
}

// required beats default: Joi raises any.required for an absent required key even when
// it carries a .default(). Both evaluators must agree.
func TestRequiredBeatsDefault(t *testing.T) {
	rule := Object(map[string]*Rule{"a": Number().Default(5).Required()})
	if _, err := Validate(rule, map[string]any{}, nil, Prefs{}); err == nil {
		t.Fatal("map evaluator applied the default where Joi requires the key")
	}
	type row struct {
		A *float64 `json:"a"`
	}
	var r row
	d, derr := DecodeInto(&r, Bytes("{}"))
	if derr != nil {
		t.Fatal(derr)
	}
	if err := d.Validate(rule, &Context{}, Prefs{}); err == nil {
		t.Fatal("typed walk applied the default where Joi requires the key")
	}
}

// `Joi.object()` (no key schema) accepts unknown keys; `Joi.object({})` rejects them.
// Object(nil) is the first form.
func TestBareObjectAcceptsUnknownKeys(t *testing.T) {
	if _, err := Validate(Object(nil), map[string]any{"x": 1}, nil, Prefs{}); err != nil {
		t.Errorf("Object(nil) rejected an unknown key: %v", err)
	}
	if _, err := Validate(Object(map[string]*Rule{}), map[string]any{"x": 1}, nil, Prefs{}); err == nil {
		t.Error("Object(empty) accepted an unknown key; Joi.object({}) rejects it")
	}
	// ...but a .pattern() IS a key schema, so keys must still match it (a dictionary
	// secret is Object(nil).Pattern(keyRe, string)).
	dict := Object(nil).Pattern(String().Regex(regexp.MustCompile(`^[-._a-zA-Z0-9]+$`)), String())
	if _, err := Validate(dict, map[string]any{"ok": "v"}, nil, Prefs{}); err != nil {
		t.Errorf("pattern-matching key rejected: %v", err)
	}
	if _, err := Validate(dict, map[string]any{"bad key!": "v"}, nil, Prefs{}); err == nil {
		t.Error("a key the pattern does not match was accepted")
	}
}

// A forbidden key must be rejected on the MAP path of the typed walk too — that path
// serves every map-backed object on the write path, and it used to store the key.
func TestTypedMapWalkEnforcesForbidden(t *testing.T) {
	type row struct {
		M map[string]any `json:"m"`
	}
	rule := Object(map[string]*Rule{
		"m": Object(map[string]*Rule{"bad": Any().Forbidden()}),
	})
	var r row
	d, derr := DecodeInto(&r, Bytes(`{"m":{"bad":1}}`))
	if derr != nil {
		t.Fatal(derr)
	}
	if err := d.Validate(rule, &Context{}, Prefs{}); err == nil {
		t.Fatal("typed map walk accepted a forbidden key")
	}
}

// Every object dependency error carries Joi's peer context; a client renders its own
// message from it. nand additionally names the first peer and is NOT the oxor message.
func TestDependencyErrorContexts(t *testing.T) {
	obj := func(fn func(*Rule) *Rule) *Rule {
		return fn(Object(map[string]*Rule{"a": Any(), "b": Any()}))
	}
	both := map[string]any{"a": 1, "b": 2}
	cases := []struct {
		name     string
		rule     *Rule
		in       map[string]any
		typ, msg string
		ctx      []string
	}{
		{"xor missing", obj(func(r *Rule) *Rule { return r.Xor("a", "b") }), map[string]any{},
			"object.missing", "must contain at least one of [a, b]", []string{"peers", "peersWithLabels"}},
		{"xor conflict", obj(func(r *Rule) *Rule { return r.Xor("a", "b") }), both,
			"object.xor", "conflict between exclusive peers", []string{"peers", "present", "presentWithLabels"}},
		{"oxor conflict", obj(func(r *Rule) *Rule { return r.Oxor("a", "b") }), both,
			"object.oxor", "conflict between optional exclusive peers", []string{"peers", "present"}},
		{"nand", obj(func(r *Rule) *Rule { return r.Nand("a", "b") }), both,
			"object.nand", `"a" must not exist simultaneously with [b]`, []string{"peers", "main", "mainWithLabel"}},
		{"or missing", obj(func(r *Rule) *Rule { return r.Or("a", "b") }), map[string]any{},
			"object.missing", "must contain at least one of [a, b]", []string{"peers"}},
	}
	for _, c := range cases {
		_, err := Validate(c.rule, c.in, nil, Prefs{})
		if err == nil {
			t.Errorf("%s: accepted", c.name)
			continue
		}
		if err.Type != c.typ || !strings.Contains(err.Message, c.msg) {
			t.Errorf("%s: type=%s msg=%q", c.name, err.Type, err.Message)
		}
		for _, k := range c.ctx {
			if _, ok := err.Ctx[k]; !ok {
				t.Errorf("%s: context is missing %q (have %v)", c.name, k, err.Ctx)
			}
		}
	}
}

// A .valid() set makes the rule an enum: Joi checks membership BEFORE the base type,
// so a wrong-typed value reports any.only listing the valid values, not <type>.base.
func TestEnumReportsAnyOnlyForWrongTypes(t *testing.T) {
	rule := String().Valid("a", "b")
	for _, in := range []any{5, true, map[string]any{}, []any{"a"}, nil} {
		_, err := Validate(Object(map[string]*Rule{"t": rule}), map[string]any{"t": in}, nil, Prefs{})
		if err == nil {
			t.Errorf("%v accepted", in)
			continue
		}
		if err.Type != "any.only" {
			t.Errorf("%v -> %s (%q), want any.only", in, err.Type, err.Message)
		}
	}
	// a valid member still passes, and a wrong STRING still reports any.only
	if _, err := Validate(rule, "a", nil, Prefs{}); err != nil {
		t.Errorf("valid member rejected: %v", err)
	}
	if _, err := Validate(rule, "z", nil, Prefs{}); err == nil || err.Type != "any.only" {
		t.Errorf("wrong member: %v", err)
	}
}

// --- engine-vs-Joi differential fixes (vp-engine.md) ---

// bothAgree runs the same rule through the generic evaluator over the map tree and
// through the typed walk over a struct, and reports each verdict. Every E-class finding
// was the two halves of the service disagreeing on one rule.
func bothAgree(t *testing.T, rule *Rule, body string, dst any, prefs Prefs) (*Error, *Error) {
	t.Helper()
	var tree map[string]any
	if err := json.Unmarshal([]byte(body), &tree); err != nil {
		t.Fatal(err)
	}
	_, gerr := Validate(rule, tree, nil, prefs)
	d, derr := DecodeInto(dst, Bytes(body))
	if derr != nil {
		t.Fatal(derr)
	}
	terr := d.Validate(rule, &Context{}, prefs)
	return gerr, terr
}

func errType(e *Error) string {
	if e == nil {
		return ""
	}
	return e.Type
}

type mapRow struct {
	M map[string]string `json:"m"`
}

// E2: Joi drains the DECLARED keys before it looks at the leftovers, so an absent
// required key is reported before an unknown one. The typed map walk ran one pass over
// the value's own keys and reported the unknown first.
func TestMapDeclaredKeysBeforeUnknown(t *testing.T) {
	rule := Object(map[string]*Rule{"m": Object(map[string]*Rule{"zz": String().Required()})})
	g, tr := bothAgree(t, rule, `{"m":{"aa":"q"}}`, &mapRow{}, Prefs{})
	if errType(g) != "any.required" || errType(tr) != "any.required" {
		t.Fatalf("map=%v typed=%v; Joi reports any.required on m.zz", g, tr)
	}
}

// E1: a forbidden key of a map-backed object is rejected on both paths.
func TestMapForbiddenKey(t *testing.T) {
	rule := Object(map[string]*Rule{"m": Object(map[string]*Rule{"k": String().Forbidden()}).Unknown(true)})
	g, tr := bothAgree(t, rule, `{"m":{"k":"q"}}`, &mapRow{}, Prefs{})
	if errType(g) != "any.unknown" || errType(tr) != "any.unknown" {
		t.Fatalf("map=%v typed=%v; Joi rejects a forbidden key", g, tr)
	}
}

// E3: a .pattern() object under .unknown(true) accepts a key the pattern misses — the
// pattern is not itself a whitelist (keys.js computes forbidUnknown from the flag).
func TestMapPatternMissFollowsUnknownPolicy(t *testing.T) {
	pat := Object(nil).Pattern(String().Regex(regexp.MustCompile(`^x`)), String())
	rule := Object(map[string]*Rule{"m": pat.Unknown(true)})
	g, tr := bothAgree(t, rule, `{"m":{"ya":"q"}}`, &mapRow{}, Prefs{})
	if g != nil || tr != nil {
		t.Fatalf("map=%v typed=%v; .unknown(true) accepts a pattern miss", g, tr)
	}
	// with no unknown flag the same miss is rejected, on both paths
	strict := Object(map[string]*Rule{"m": pat})
	g, tr = bothAgree(t, strict, `{"m":{"ya":"q"}}`, &mapRow{}, Prefs{})
	if errType(g) != "object.unknown" || errType(tr) != "object.unknown" {
		t.Fatalf("map=%v typed=%v; a bare pattern rejects a miss", g, tr)
	}
	// and stripped under the pref
	row := &mapRow{}
	d, derr := DecodeInto(row, Bytes(`{"m":{"ya":"q"}}`))
	if derr != nil {
		t.Fatal(derr)
	}
	if err := d.Validate(strict, &Context{}, Prefs{StripUnknown: true}); err != nil {
		t.Fatalf("stripUnknown should drop a pattern miss: %v", err)
	}
	if _, ok := row.M["ya"]; ok {
		t.Fatalf("key not stripped: %v", row.M)
	}
}

// B4: an explicit .unknown(true) outranks the stripUnknown PREFERENCE
// (`prefs.stripUnknown && !schema._flags.unknown`), on both paths.
func TestSchemaUnknownTrueBeatsStripPref(t *testing.T) {
	rule := Object(map[string]*Rule{"m": Object(map[string]*Rule{}).Unknown(true)})
	row := &mapRow{}
	d, derr := DecodeInto(row, Bytes(`{"m":{"b":"1"}}`))
	if derr != nil {
		t.Fatal(derr)
	}
	if err := d.Validate(rule, &Context{}, Prefs{StripUnknown: true}); err != nil {
		t.Fatal(err)
	}
	if _, ok := row.M["b"]; !ok {
		t.Fatalf("typed walk stripped a key .unknown(true) keeps: %v", row.M)
	}
	out, err := Validate(rule, map[string]any{"m": map[string]any{"b": "1"}}, nil, Prefs{StripUnknown: true})
	if err != nil {
		t.Fatal(err)
	}
	if m, _ := out.(map[string]any)["m"].(map[string]any); m["b"] == nil {
		t.Fatalf("map evaluator stripped a key .unknown(true) keeps: %v", out)
	}
}

// B3: `.unknown(false)` clears the flag rather than pinning a reject, so an
// allowUnknown preference still applies (Joi drops a flag equal to its default).
func TestUnknownFalseDoesNotBlockAllowUnknownPref(t *testing.T) {
	rule := Object(map[string]*Rule{"a": String()}).Unknown(false)
	if _, err := Validate(rule, map[string]any{"a": "x", "b": 1.0}, nil, Prefs{AllowUnknown: true}); err != nil {
		t.Fatalf(".unknown(false) blocked the allowUnknown pref: %v", err)
	}
	if _, err := Validate(rule, map[string]any{"a": "x", "b": 1.0}, nil, Prefs{}); err == nil {
		t.Fatal("unknown key accepted with default prefs")
	}
}

// E7: a default that resolves to nothing leaves the key UNSET; writing the nil stored a
// JSON null where TS omits the key.
func TestUnresolvedDefaultLeavesKeyUnset(t *testing.T) {
	rule := Object(map[string]*Rule{"b": String().Default(RefTo("a"))})
	out, err := Validate(rule, map[string]any{}, nil, Prefs{})
	if err != nil {
		t.Fatal(err)
	}
	if _, has := out.(map[string]any)["b"]; has {
		t.Fatalf("unresolved default wrote a key: %v", out)
	}
}

// E6: a rule whose type the field cannot hold answers <type>.base, as the generic
// evaluator does — it used to panic, turning a 400 into a 500 on the write path.
func TestRuleFieldTypeMismatchIsAnError(t *testing.T) {
	type row struct {
		Items []string `json:"items"`
	}
	rule := Object(map[string]*Rule{"items": String()})
	g, tr := bothAgree(t, rule, `{"items":["a"]}`, &row{}, Prefs{})
	if errType(g) != "string.base" || errType(tr) != "string.base" {
		t.Fatalf("map=%v typed=%v; want string.base from both", g, tr)
	}
}

// E5 + C8: several item rules over object elements validate (Joi accepts an element
// that matches ANY of them) and report array.includes when none match.
func TestMultiItemRulesOverObjects(t *testing.T) {
	type el struct {
		A string `json:"a"`
		B string `json:"b"`
	}
	type row struct {
		Items []el `json:"items"`
	}
	rule := Object(map[string]*Rule{"items": Array(
		Object(map[string]*Rule{"a": String().Required()}),
		Object(map[string]*Rule{"b": String().Required()}),
	)})
	g, tr := bothAgree(t, rule, `{"items":[{"a":"x"}]}`, &row{}, Prefs{})
	if g != nil || tr != nil {
		t.Fatalf("map=%v typed=%v; an element matching one item rule is accepted", g, tr)
	}
	g, tr = bothAgree(t, rule, `{"items":[{}]}`, &row{}, Prefs{})
	if errType(g) != "array.includes" || errType(tr) != "array.includes" {
		t.Fatalf("map=%v typed=%v; want array.includes", g, tr)
	}
}

// C8: a SINGLE item rule still passes its own error through.
func TestSingleItemRuleErrorPassesThrough(t *testing.T) {
	rule := Array(String())
	_, err := Validate(rule, []any{true}, nil, Prefs{})
	if errType(err) != "string.base" {
		t.Fatalf("got %v; a single inclusion reports the branch error", err)
	}
}

// --- numbers ---

// A2/C15/A17: number.unsafe, -0 normalisation, and valids checked BEFORE .precision()
// rounds. Verdicts taken from joi@17.6.0 directly.
func TestNumberJoiParity(t *testing.T) {
	cases := []struct {
		in   any
		rule *Rule
		want string // error type, or the JSON of the accepted value
	}{
		{9007199254740993.0, Number(), "number.unsafe"},
		{9007199254740992.0, Number(), "number.unsafe"},
		{9007199254740991.0, Number(), "9007199254740991"},
		{1e30, Number(), "number.unsafe"},
		{"9007199254740993", Number(), "number.unsafe"},
		{"1.0000000000000001", Number(), "number.unsafe"},
		{"1e400", Number(), "number.unsafe"},
		{"1e-400", Number(), "number.unsafe"},
		{math.Copysign(0, -1), Number(), "0"},
		{"5.0", Number(), "5"},
		{"005", Number(), "5"},
		{0.6, Number().Valid(1.0).Precision(0), "any.only"},
		{0.6, Number().Integer().Precision(0), "1"},
		{"1e3", Number(), "1000"},
	}
	for _, c := range cases {
		out, err := Validate(c.rule, c.in, nil, Prefs{})
		got := ""
		if err != nil {
			got = err.Type
		} else {
			b, _ := json.Marshal(out)
			got = string(b)
		}
		if got != c.want {
			t.Errorf("%v -> %s, want %s", c.in, got, c.want)
		}
	}
}

// --- dates ---

// A3/A4/C16: .iso() is a real ISO-8601 gate, a digit string is an epoch TIMESTAMP before
// it is a year, and a slash date is LOCAL time.
func TestDateJoiParity(t *testing.T) {
	iso := map[string]bool{
		"2020": true, "2020-01": true, "2020-01-02": true, "2020-01-02T03:04:05Z": true,
		"2020-01-02 03:04:05": true, "2020/01/02": false, "Jan 2 2020": false,
		"1577934245000": false, "20200102": true, "2020-W01": true, "2020-01-02T": false,
		"20200102T030405Z": false, "202001": false, "2020-13-01": false,
	}
	for in, want := range iso {
		if got := isISODate(in); got != want {
			t.Errorf("isISODate(%q) = %v, want %v", in, got, want)
		}
	}
	if _, err := Validate(DateRule().Iso(), "2020/01/02", nil, Prefs{}); err == nil || err.Type != "date.format" {
		t.Errorf(".iso() accepted a slash date: %v", err)
	}
	if _, err := Validate(DateRule().Iso(), 1577934245000.0, nil, Prefs{}); err == nil || err.Type != "date.base" {
		t.Errorf(".iso() on a number: %v, want date.base", err)
	}
	// "2020" is 2020 MILLISECONDS after the epoch to a bare Joi.date()
	out, err := Validate(DateRule(), "2020", nil, Prefs{})
	if err != nil || out.(string) != "1970-01-01T00:00:02.020Z" {
		t.Errorf(`date("2020") = %v, %v; Joi reads a digit string as epoch millis`, out, err)
	}
}

// --- refs ---

// A5/C10/A11: a limit ref that does not resolve to a number is any.ref, a resolved one
// renders as `ref:<path>`, and a dotted ref walks into the nested object.
func TestRefBounds(t *testing.T) {
	rule := Object(map[string]*Rule{"a": Number(), "b": Number().MinRef("a")})
	_, err := Validate(rule, map[string]any{"b": 3.0}, nil, Prefs{})
	if err == nil || err.Type != "any.ref" {
		t.Fatalf("absent ref: %v, want any.ref", err)
	}
	if !strings.Contains(err.Message, `references "ref:a"`) {
		t.Errorf("any.ref message: %q", err.Message)
	}
	_, err = Validate(rule, map[string]any{"a": 5.0, "b": 3.0}, nil, Prefs{})
	if err == nil || !strings.HasSuffix(err.Message, "greater than or equal to ref:a") {
		t.Fatalf("resolved ref message: %v", err)
	}
	dotted := Object(map[string]*Rule{
		"o": Object(map[string]*Rule{"f": Number()}),
		"b": Number().MinRef("o.f"),
	})
	_, err = Validate(dotted, map[string]any{"o": map[string]any{"f": 5.0}, "b": 3.0}, nil, Prefs{})
	if err == nil || err.Type != "number.min" {
		t.Fatalf("dotted ref: %v, want number.min", err)
	}
}

// --- when ---

// A14: a .when() branch is merged ONCE. Resolving twice ran a then-branch custom twice,
// which for a transforming custom means the transform applies twice.
func TestWhenBranchMergedOnce(t *testing.T) {
	runs := 0
	rule := Object(map[string]*Rule{
		"a": String(),
		"b": String().When("a", When{IsEq: "x", Then: String().Custom(func(v any, h *Helper) (any, error) {
			runs++
			return v.(string) + "!", nil
		})}),
	})
	out, err := Validate(rule, map[string]any{"a": "x", "b": "q"}, nil, Prefs{})
	if err != nil {
		t.Fatal(err)
	}
	if runs != 1 {
		t.Errorf("branch custom ran %d times, want 1", runs)
	}
	if got := out.(map[string]any)["b"]; got != "q!" {
		t.Errorf("b = %v, want q!", got)
	}
}

// A9 + the IsEq/Is split: Joi compiles a LITERAL `is:` to valid(x).required(), so an
// absent sibling misses; a SCHEMA `is:` gets no .required(), so an absent sibling
// matches. Both shapes are live in the TS schemas, so both must be expressible.
func TestWhenConditionPresence(t *testing.T) {
	lit := Object(map[string]*Rule{
		"a": String(),
		"b": String().When("a", When{IsEq: "x", Then: String().Required()}),
	})
	if _, err := Validate(lit, map[string]any{}, nil, Prefs{}); err != nil {
		t.Errorf("literal condition fired on an absent sibling: %v", err)
	}
	if _, err := Validate(lit, map[string]any{"a": "x"}, nil, Prefs{}); err == nil {
		t.Error("literal condition did not fire on a matching sibling")
	}
	sch := Object(map[string]*Rule{
		"a": String(),
		"b": String().When("a", When{Is: String(), Then: String().Required()}),
	})
	if _, err := Validate(sch, map[string]any{}, nil, Prefs{}); err == nil {
		t.Error("schema condition missed an absent sibling; Joi validates undefined against it")
	}
}

// A10: a .when() branch's rules all survive the merge — the hand-written field list used
// to drop .length()/.unique()/.sort()/.pattern()/.xor().
func TestWhenBranchMergesEveryRule(t *testing.T) {
	rule := Object(map[string]*Rule{
		"a": String(),
		"b": Array(String()).When("a", When{IsEq: "x", Then: Array(String()).Length(2).Unique()}),
	})
	if _, err := Validate(rule, map[string]any{"a": "x", "b": []any{"q"}}, nil, Prefs{}); err == nil {
		t.Error("branch .Length(2) was dropped by the merge")
	}
	if _, err := Validate(rule, map[string]any{"a": "x", "b": []any{"q", "q"}}, nil, Prefs{}); err == nil {
		t.Error("branch .Unique() was dropped by the merge")
	}
}

// --- allow / valid ---

// A15/A16/B2: allow()/valid() are matched on the COERCED value and short-circuit the
// base validate, and a string never equals a number.
func TestAllowMatchesCoercedValue(t *testing.T) {
	cases := []struct {
		rule *Rule
		in   any
		want string
	}{
		{String().Trim().Allow(""), "   ", `""`},
		{String().Trim().Valid(""), "   ", `""`},
		{Number().Allow(0.0), "0", `0`},
		{String().Lowercase().Valid("ab"), "AB", `"ab"`},
		{Any().Valid(1.0), "1", "any.only"}, // Joi: the string "1" is not the number 1
		{String().Valid("1"), 1.0, "any.only"},
	}
	for _, c := range cases {
		out, err := Validate(c.rule, c.in, nil, Prefs{})
		got := ""
		if err != nil {
			got = err.Type
		} else {
			b, _ := json.Marshal(out)
			got = string(b)
		}
		if got != c.want {
			t.Errorf("%v -> %s, want %s", c.in, got, c.want)
		}
	}
}

// --- alternatives ---

// C1/C2/C3/C4: the reported error is chosen from the branch ERRORS, Joi's
// internals.errors, not from the shape of the branch schemas.
func TestAlternativesErrorSelection(t *testing.T) {
	cases := []struct {
		rule *Rule
		in   any
		want string
	}{
		{Alternatives(String().Min(5)), "ab", "string.min"},
		{Alternatives(String().Min(5), Number().NumMin(10)), 3.0, "number.min"},
		{Alternatives(String(), Object(map[string]*Rule{"a": String().Required()})), map[string]any{}, "any.required"},
		{Alternatives(String().Valid("a"), String().Valid("b")), "z", "alternatives.types"},
		{Alternatives(String(), Number()), nil, "alternatives.types"},
		{Alternatives(String().Valid("a"), Number()), "b", "alternatives.types"},
		{Alternatives(
			Object(map[string]*Rule{"a": String().Required()}),
			Object(map[string]*Rule{"b": String().Required()}),
		), map[string]any{}, "alternatives.match"},
		{Alternatives(), "x", "alternatives.any"},
	}
	for _, c := range cases {
		_, err := Validate(c.rule, c.in, nil, Prefs{})
		if errType(err) != c.want {
			t.Errorf("%v -> %v, want %s", c.in, err, c.want)
		}
	}
	// alternatives.types lists the valid VALUES alongside the type names
	_, err := Validate(Alternatives(String().Valid("a"), Number()), "b", nil, Prefs{})
	if err.Message != `"value" must be one of [a, number]` {
		t.Errorf("types message: %q", err.Message)
	}
	// alternatives.match carries every branch's details
	_, err = Validate(Alternatives(
		Object(map[string]*Rule{"a": String().Required()}),
		Object(map[string]*Rule{"b": String().Required()}),
	), map[string]any{}, nil, Prefs{})
	if d, _ := err.Ctx["details"].([]Detail); len(d) != 2 {
		t.Errorf("alternatives.match details = %v, want 2", err.Ctx["details"])
	}
	// Joi's Errors.details collapses IDENTICAL branch messages before the ". " join
	// (details[] keeps every one). Three branches failing the same way say it once.
	_, err = Validate(Alternatives(
		Object(map[string]*Rule{"a": String().Required()}),
		Object(map[string]*Rule{"a": String().Required(), "b": Number()}),
		Object(map[string]*Rule{"a": String().Required(), "c": Number()}),
	), map[string]any{}, nil, Prefs{})
	if got := err.Ctx["message"]; got != `"a" is required` {
		t.Errorf("alternatives.match message = %v, want deduped single message", got)
	}
	if d, _ := err.Ctx["details"].([]Detail); len(d) != 3 {
		t.Errorf("alternatives.match details = %d, want 3 (details are NOT deduped)", len(d))
	}
}

// --- arrays ---

// C7: array.unique blames the DUPLICATE ELEMENT and carries pos/dupePos/dupeValue, on
// both evaluators.
func TestArrayUniqueReportsTheDuplicate(t *testing.T) {
	type row struct {
		A []string `json:"a"`
	}
	rule := Object(map[string]*Rule{"a": Array(String()).Unique()})
	g, tr := bothAgree(t, rule, `{"a":["x","y","x"]}`, &row{}, Prefs{})
	for _, e := range []*Error{g, tr} {
		if e == nil || e.Type != "array.unique" {
			t.Fatalf("want array.unique, got %v", e)
		}
		if e.Message != `"a[2]" contains a duplicate value` {
			t.Errorf("message %q", e.Message)
		}
		if e.Ctx["pos"] != 2 || e.Ctx["dupePos"] != 0 {
			t.Errorf("ctx %v", e.Ctx)
		}
	}
}

// A18: a mixed-type array is array.sort.mismatching, not silently reordered.
func TestArraySortMismatchingTypes(t *testing.T) {
	_, err := Validate(Array().Sort(), []any{"c", 1.0}, nil, Prefs{})
	if errType(err) != "array.sort.mismatching" {
		t.Fatalf("got %v, want array.sort.mismatching", err)
	}
	out, err := Validate(Array().Sort(), []any{"c", "a"}, nil, Prefs{})
	if err != nil || out.([]any)[0] != "a" {
		t.Fatalf("homogeneous sort broke: %v %v", out, err)
	}
}

// --- labels ---

// C11: .Label() was stored and never read, so every override in the schema set was a
// no-op. It applies to leaf errors and to the object loop's own presence errors.
func TestLabelOverride(t *testing.T) {
	type row struct {
		A *string `json:"a"`
		B *string `json:"b"`
	}
	rule := Object(map[string]*Rule{
		"a": String().Label("Alpha"),
		"b": String().Label("Beta").Required(),
	})
	g, tr := bothAgree(t, rule, `{"a":5,"b":"x"}`, &row{}, Prefs{})
	for _, e := range []*Error{g, tr} {
		if e == nil || e.Message != `"Alpha" must be a string` || e.Ctx["label"] != "Alpha" {
			t.Errorf("leaf label: %v", e)
		}
	}
	g, tr = bothAgree(t, rule, `{}`, &row{}, Prefs{})
	for _, e := range []*Error{g, tr} {
		if e == nil || e.Message != `"Beta" is required` {
			t.Errorf("required label: %v", e)
		}
	}
}

// --- string formats ---

// A6/A7/B5/B6: .email()/.uri()/.guid()/.domain() follow @sideway/address rather than a
// one-line approximation. Verdicts taken from joi@17.6.0.
func TestStringFormatsMatchJoi(t *testing.T) {
	email := map[string]bool{
		"a@b.com": true, "a+b@c.com": true, "user.name+tag@example.com": true, "A@B.COM": true,
		"a@sub.domain.co.uk": true, "a@b.museum": true,
		"a@b.c": false, "a..b@c.com": false, ".a@b.co": false, "a.@b.co": false, "a@b..co": false,
		"a@-b.co": false, "a@b-.co": false, "a@b.co.": false, "a@127.0.0.1": false,
		"a@[127.0.0.1]": false, "a@b_c.co": false, "<a@b.co>": false, "a@b.zzzzz": false,
	}
	for in, want := range email {
		if got := isEmail(in); got != want {
			t.Errorf("isEmail(%q) = %v, want %v", in, got, want)
		}
	}
	uri := map[string]bool{
		"http://x.com": true, "https://x.com/a/b?c=d#e": true, "mailto:a@b.com": true,
		"urn:isbn:1234": true, "file:///tmp/x": true, "http://[::1]:80/": true,
		"zz:": false, "zz://": false, "http://": false, "http://x.com/a b": false,
		"http://x.com/\\": false, "x.com": false, "http:/x.com": false, "relative/path": false,
	}
	for in, want := range uri {
		if got := isURI(in); got != want {
			t.Errorf("isURI(%q) = %v, want %v", in, got, want)
		}
	}
	guid := map[string]bool{
		"2c9f6a3e-1b2c-4d3e-8f4a-5b6c7d8e9f01":   true,
		"{2c9f6a3e-1b2c-4d3e-8f4a-5b6c7d8e9f01}": true,
		"(2c9f6a3e-1b2c-4d3e-8f4a-5b6c7d8e9f01)": true,
		"2c9f6a3e1b2c4d3e8f4a5b6c7d8e9f01":       true,
		"2c9f6a3e:1b2c:4d3e:8f4a:5b6c7d8e9f01":   true,
		"{2c9f6a3e-1b2c-4d3e-8f4a-5b6c7d8e9f01":  false,
		"2c9f6a3e-1b2c:4d3e-8f4a-5b6c7d8e9f01":   false,
		"zzzzzzzz-1b2c-4d3e-8f4a-5b6c7d8e9f01":   false,
	}
	for in, want := range guid {
		if got := isGUID(in); got != want {
			t.Errorf("isGUID(%q) = %v, want %v", in, got, want)
		}
	}
	// .domain() carries NO tld list (Joi passes no options), so a one-char tld is fine
	for in, want := range map[string]bool{"a.b": true, "a.c": true, "example.com": true, "a": false, "a..com": false, "-a.com": false} {
		if got := isDomain(in, false); got != want {
			t.Errorf("isDomain(%q) = %v, want %v", in, got, want)
		}
	}
}

// --- key order ---

// C14: .Keys() overriding an existing key moves it to the END, which decides which of
// several missing required keys abortEarly reports.
func TestKeysOverrideMovesToTail(t *testing.T) {
	base := ObjectOrdered([]string{"a", "b", "cc"}, map[string]*Rule{
		"a": String().Required(), "b": String().Required(), "cc": String().Required(),
	})
	_, err := Validate(base.Keys(map[string]*Rule{"a": String().Required()}), map[string]any{}, nil, Prefs{})
	if err == nil || err.Path[0] != "b" {
		t.Fatalf("got %v, want the error on b", err)
	}
	_, err = Validate(base, map[string]any{}, nil, Prefs{})
	if err == nil || err.Path[0] != "a" {
		t.Fatalf("got %v, want the error on a", err)
	}
}

// --- presence / defaults ---

// A default that resolves to nothing must not make an xor/nand peer look present, on
// either evaluator.
func TestNilDefaultIsNotPresent(t *testing.T) {
	type row struct {
		T *float64 `json:"t"`
		M []string `json:"m"`
	}
	rule := Object(map[string]*Rule{
		"t": Number().Default(func(p map[string]any) any { return nil }),
		"m": Array(String()),
	}).Nand("t", "m")
	g, tr := bothAgree(t, rule, `{"m":["x"]}`, &row{}, Prefs{})
	if g != nil || tr != nil {
		t.Fatalf("map=%v typed=%v; an unresolved default is not a present peer", g, tr)
	}
}

// Joi resolves a key's refs against the object it is BUILDING, so an earlier key's
// default is visible — which is how query.ts's `value.when('op')` sees op's default '='.
func TestRefsSeeEarlierDefaults(t *testing.T) {
	rule := ObjectOrdered([]string{"op", "value"}, map[string]*Rule{
		"op":    String().Valid("=", "exists").Default("="),
		"value": String().When("op", When{IsEq: "exists", Then: String().Optional(), Otherwise: String().Required()}),
	})
	if _, err := Validate(rule, map[string]any{}, nil, Prefs{}); errType(err) != "any.required" {
		t.Errorf("got %v; the defaulted op should make value required", err)
	}
	if _, err := Validate(rule, map[string]any{"op": "exists"}, nil, Prefs{}); err != nil {
		t.Errorf("got %v; op=exists makes value optional", err)
	}
}

// The typed container walks stamp Joi's context.value the way the validateValue seam
// does for the generic evaluator.
func TestTypedContainerErrorsCarryValue(t *testing.T) {
	type row struct {
		A []string `json:"a"`
	}
	rule := Object(map[string]*Rule{"a": Array(String()).Min(2)})
	g, tr := bothAgree(t, rule, `{"a":["x"]}`, &row{}, Prefs{})
	for _, e := range []*Error{g, tr} {
		if e == nil || e.Ctx["value"] == nil {
			t.Errorf("array.min without context.value: %v", e)
		}
	}
}

// Helper.Ancestor walks the object chain a cross-field check needs when the data it
// must read lives above its immediate parent.
func TestHelperAncestors(t *testing.T) {
	var got []string
	rule := Object(map[string]*Rule{
		"tag": String(),
		"spec": Object(map[string]*Rule{
			"inner": Object(map[string]*Rule{
				"leaf": String().Custom(func(v any, h *Helper) (any, error) {
					for n := 0; n < 4; n++ {
						a := h.Ancestor(n)
						if a == nil {
							got = append(got, "-")
							continue
						}
						got = append(got, sortedKeysOf(a))
					}
					return v, nil
				}),
			}),
		}),
	})
	_, err := Validate(rule, map[string]any{"tag": "t", "spec": map[string]any{"inner": map[string]any{"leaf": "x"}}}, nil, Prefs{})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"leaf", "inner", "spec,tag", "-"}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("ancestors = %v, want %v", got, want)
	}
}

func sortedKeysOf(m map[string]any) string { return strings.Join(sortedKeys(m), ",") }

// ftoa must render a number the way JavaScript's String(n) does, because Joi
// interpolates it into every limit and valid-value message. Go's 'g' verb switches to
// exponent form far earlier and zero-pads the exponent.
func TestNumberRenderingMatchesJS(t *testing.T) {
	cases := map[float64]string{
		0: "0", 1: "1", 1.5: "1.5", 1e6: "1000000", 4294967295: "4294967295",
		18446744073709551615: "18446744073709552000", 1e20: "100000000000000000000",
		1e21: "1e+21", 1e-6: "0.000001", 1e-7: "1e-7", 2.5e-10: "2.5e-10",
		1e300: "1e+300", 1e-300: "1e-300", 9007199254740991: "9007199254740991",
	}
	for in, want := range cases {
		if got := ftoa(in); got != want {
			t.Errorf("ftoa(%v) = %s, want %s", in, got, want)
		}
	}
	_, err := Validate(Number().MaxNum(4294967295), 5e9, nil, Prefs{})
	if err == nil || !strings.HasSuffix(err.Message, "less than or equal to 4294967295") {
		t.Errorf("limit message: %v", err)
	}
	_, err = Validate(Number().Valid(1e6), 5.0, nil, Prefs{})
	if err == nil || err.Message != `"value" must be [1000000]` {
		t.Errorf("valids message: %v", err)
	}
}

// An alternatives rule's own .custom() runs after a branch matches — returning the
// branch's value straight out dropped it silently, which is an ACCEPT gap.
func TestAlternativesRunsItsOwnCustoms(t *testing.T) {
	seen := 0
	reject := func(v any, h *Helper) (any, error) {
		seen++
		if v == "0s" {
			return nil, h.Errorf("any.custom", "zero not allowed")
		}
		return v, nil
	}
	rule := Alternatives(String(), Number()).Custom(reject)
	if _, err := Validate(rule, "0s", nil, Prefs{}); err == nil {
		t.Error("a custom on an alternatives rule was skipped after a branch matched")
	}
	if seen == 0 {
		t.Error("the custom never ran")
	}
	cond := Alternatives().Conditional("k", AltCase{IsEq: "x", Then: String(), Otherwise: String()}).Custom(reject)
	if _, err := Validate(Object(map[string]*Rule{"k": String(), "v": cond}),
		map[string]any{"k": "x", "v": "0s"}, nil, Prefs{}); err == nil {
		t.Error("a custom on an alternatives().conditional() was skipped")
	}
}

// mergeRules must not write into the BASE rule: clone() shares the keys map, and schema
// rules are built once and reused for every request, so a .when() branch carrying keys
// would permanently rewrite the cached rule.
func TestWhenBranchDoesNotMutateBase(t *testing.T) {
	base := Object(map[string]*Rule{"a": String()})
	rule := Object(map[string]*Rule{
		"flag": String(),
		"o":    base.When("flag", When{IsEq: "x", Then: Object(map[string]*Rule{"b": String()})}),
	})
	// the branch applies: b is accepted
	if _, err := Validate(rule, map[string]any{"flag": "x", "o": map[string]any{"b": "1"}}, nil, Prefs{}); err != nil {
		t.Fatalf("branch key rejected: %v", err)
	}
	// the branch does NOT apply: b must still be unknown, and the base rule untouched
	if _, err := Validate(rule, map[string]any{"o": map[string]any{"b": "1"}}, nil, Prefs{}); err == nil {
		t.Error("the branch's keys leaked into the base rule for every later request")
	}
	if _, ok := base.keys["b"]; ok {
		t.Error("mergeRules wrote into the base rule's shared keys map")
	}
}

// An object may carry SEVERAL .pattern() clauses — envoyHttp declares two on one object.
// Joi tries them in order and the first matching key schema wins; keeping a single
// pair meant the second silently replaced the first.
func TestMultiplePatterns(t *testing.T) {
	rule := Object(nil).
		Pattern(String().RegexInvert(regexp.MustCompile(`^cpln_`)), Object(map[string]*Rule{"a": String()})).
		Pattern(String().Regex(regexp.MustCompile(`^cpln_`)), Object(map[string]*Rule{"b": String()}))
	ok := map[string]any{
		"x":      map[string]any{"a": "1"},
		"cpln_y": map[string]any{"b": "2"},
	}
	if _, err := Validate(rule, ok, nil, Prefs{}); err != nil {
		t.Fatalf("both patterns should apply: %v", err)
	}
	// a cpln_ key is judged by the SECOND pattern, which has no `a`
	bad := map[string]any{"cpln_y": map[string]any{"a": "2"}}
	if _, err := Validate(rule, bad, nil, Prefs{}); err == nil {
		t.Error("the second pattern was dropped; a cpln_ key took the first pattern's rule")
	}
}
