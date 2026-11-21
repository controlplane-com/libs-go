package validate

import "testing"

func run(t *testing.T, rule *Rule, cases []Case) {
	t.Helper()
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			if err := c.Run(rule); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestNumberCoercionAndBounds(t *testing.T) {
	rule := Number().Integer().Min(80).Max(65535)
	run(t, rule, []Case{
		{Name: "numeric string coerced", Input: "8080", Want: float64(8080)},
		{Name: "float ok", Input: float64(80), Want: float64(80)},
		{Name: "below min", Input: float64(79), WantErr: "greater than or equal to 80"},
		{Name: "above max", Input: float64(70000), WantErr: "less than or equal to 65535"},
		{Name: "non-integer", Input: float64(80.5), WantErr: "integer"},
		{Name: "not a number", Input: "abc", WantErr: "must be a number"},
	})
}

func TestNumberRefBound(t *testing.T) {
	// maxSize: Number().min(Joi.ref('minSize')) style
	rule := Object(map[string]*Rule{
		"minSize": Number().Integer().Required(),
		"maxSize": Number().Integer().Required().MinRef("minSize"),
	})
	run(t, rule, []Case{
		{Name: "max>=min ok", Input: map[string]any{"minSize": float64(2), "maxSize": float64(5)}},
		{Name: "max<min fails", Input: map[string]any{"minSize": float64(5), "maxSize": float64(2)}, WantErr: "greater than or equal to ref:minSize"},
	})
}

func TestPrecision(t *testing.T) {
	run(t, Number().Precision(2), []Case{
		{Name: "rounds", Input: float64(1.239), Want: float64(1.24)},
	})
}

func TestStringModes(t *testing.T) {
	run(t, String().Trim().Lowercase(), []Case{
		{Name: "trim+lower", Input: "  HeLLo  ", Want: "hello"},
	})
	run(t, String().Allow(""), []Case{
		{Name: "empty allowed", Input: "", Want: ""},
	})
	run(t, String(), []Case{
		{Name: "empty rejected", Input: "", WantErr: "empty"},
	})
	run(t, String().Email(), []Case{
		{Name: "valid email", Input: "a@b.co"},
		{Name: "invalid email", Input: "nope", WantErr: "valid email"},
	})
	run(t, String().UUID(), []Case{
		{Name: "valid uuid", Input: "123e4567-e89b-12d3-a456-426614174000"},
		{Name: "invalid uuid", Input: "nope", WantErr: "GUID"},
	})
}

func TestBooleanCoercion(t *testing.T) {
	run(t, Boolean(), []Case{
		{Name: "bool", Input: true, Want: true},
		{Name: "string true", Input: "true", Want: true},
		{Name: "string FALSE", Input: "FALSE", Want: false},
		{Name: "bad", Input: "yes", WantErr: "must be a boolean"},
	})
}

func TestDateNormalization(t *testing.T) {
	run(t, DateRule(), []Case{
		{Name: "iso string", Input: "2020-01-02T03:04:05Z", Want: "2020-01-02T03:04:05.000Z"},
		{Name: "iso with offset", Input: "2020-01-02T03:04:05+02:00", Want: "2020-01-02T01:04:05.000Z"},
		{Name: "bad date", Input: "not-a-date", WantErr: "valid date"},
	})
}

func TestArrayFeatures(t *testing.T) {
	// ports: Joi.array().items(number.integer.min0.max65535).min(1).max(10).sort().unique()
	rule := Array(Number().Integer().Min(0).Max(65535)).Min(1).Max(3).Sort().Unique()
	run(t, rule, []Case{
		{Name: "sorted output", Input: []any{float64(3), float64(1), float64(2)}, Want: []any{float64(1), float64(2), float64(3)}},
		{Name: "empty fails min", Input: []any{}, WantErr: "at least 1"},
		{Name: "too many", Input: []any{float64(1), float64(2), float64(3), float64(4)}, WantErr: "less than or equal to 3"},
		{Name: "duplicate", Input: []any{float64(1), float64(1)}, WantErr: "duplicate"},
		{Name: "bad item", Input: []any{float64(70000)}, WantErr: "less than or equal to 65535"},
	})
}

func TestArrayUniqueBy(t *testing.T) {
	rule := Array(Object(map[string]*Rule{"name": String().Required()})).Unique("name")
	run(t, rule, []Case{
		{Name: "unique names", Input: []any{map[string]any{"name": "a"}, map[string]any{"name": "b"}}},
		{Name: "dup names", Input: []any{map[string]any{"name": "a"}, map[string]any{"name": "a"}}, WantErr: "duplicate"},
	})
}

// TestArrayDedupe pins the makeUnique semantics: duplicates are SILENTLY DROPPED (keep
// first, after sort), never rejected — unlike Unique. This is memcache/spicedb/dbcluster
// `.custom(makeUnique())`.
func TestArrayDedupe(t *testing.T) {
	rule := Array(String()).Min(1).Sort().Dedupe()
	run(t, rule, []Case{
		{Name: "dup collapsed + sorted", Input: []any{"b", "a", "b"}, Want: []any{"a", "b"}},
		{Name: "no dup unchanged", Input: []any{"a", "b"}, Want: []any{"a", "b"}},
	})
	byName := Array(Object(map[string]*Rule{"name": String().Required()})).Dedupe("name")
	run(t, byName, []Case{
		{Name: "dedupe by name keeps first", Input: []any{map[string]any{"name": "a"}, map[string]any{"name": "a"}}, Want: []any{map[string]any{"name": "a"}}},
	})
}

func TestUnknownKeyPolicy(t *testing.T) {
	rule := Object(map[string]*Rule{"a": String()})
	// default: reject
	if err := (Case{Input: map[string]any{"a": "x", "z": "y"}, WantErr: "is not allowed"}).Run(rule); err != nil {
		t.Fatal(err)
	}
	// allow via pref
	out, verr := Validate(rule, map[string]any{"a": "x", "z": "y"}, &Context{}, Prefs{AllowUnknown: true})
	if verr != nil {
		t.Fatalf("allowUnknown: %v", verr)
	}
	if out.(map[string]any)["z"] != "y" {
		t.Fatal("allowUnknown should keep z")
	}
	// strip via pref (wins over allow)
	out, verr = Validate(rule, map[string]any{"a": "x", "z": "y"}, &Context{}, Prefs{AllowUnknown: true, StripUnknown: true})
	if verr != nil {
		t.Fatalf("stripUnknown: %v", verr)
	}
	if _, ok := out.(map[string]any)["z"]; ok {
		t.Fatal("stripUnknown should drop z")
	}
}

func TestPresenceAndDefaults(t *testing.T) {
	rule := Object(map[string]*Rule{
		"req": String().Required(),
		"def": String().Default("x"),
	})
	run(t, rule, []Case{
		{Name: "missing required", Input: map[string]any{}, WantErr: "is required"},
		{Name: "default applied", Input: map[string]any{"req": "y"}, Want: map[string]any{"req": "y", "def": "x"}},
	})
	// presence:'optional' override relaxes required
}

func TestPresenceRelations(t *testing.T) {
	xor := Object(map[string]*Rule{"a": String(), "b": String()}).Xor("a", "b")
	run(t, xor, []Case{
		{Name: "exactly one", Input: map[string]any{"a": "x"}},
		{Name: "none", Input: map[string]any{}, WantErr: "must contain at least one of"},
		{Name: "both", Input: map[string]any{"a": "x", "b": "y"}, WantErr: "conflict between exclusive peers"},
	})
	nand := Object(map[string]*Rule{"a": String(), "b": String()}).Nand("a", "b")
	run(t, nand, []Case{
		{Name: "nand none", Input: map[string]any{}},
		{Name: "nand one", Input: map[string]any{"a": "x"}},
		{Name: "nand both", Input: map[string]any{"a": "x", "b": "y"}, WantErr: "must not exist simultaneously with"},
	})
	oxor := Object(map[string]*Rule{"a": String(), "b": String()}).Oxor("a", "b")
	run(t, oxor, []Case{
		{Name: "oxor none ok", Input: map[string]any{}},
		{Name: "oxor both fails", Input: map[string]any{"a": "x", "b": "y"}, WantErr: "optional exclusive peers"},
	})
}

func TestPatternMap(t *testing.T) {
	// Joi.object().pattern(Joi.string(), Joi.string())
	rule := Object(nil).Pattern(String(), String())
	run(t, rule, []Case{
		{Name: "all string values", Input: map[string]any{"x": "1", "y": "2"}, Want: map[string]any{"x": "1", "y": "2"}},
		{Name: "non-string value fails", Input: map[string]any{"x": float64(1)}, WantErr: "must be a string"},
	})
}

func TestWhenConditional(t *testing.T) {
	// resolver: Joi.when('mode', { is: 'a', then: required, otherwise: forbidden })
	rule := Object(map[string]*Rule{
		"mode": String(),
		"val":  String().When("mode", When{Is: String().Valid("a"), Then: String().Required(), Otherwise: String().Forbidden()}),
	})
	run(t, rule, []Case{
		{Name: "mode a requires val", Input: map[string]any{"mode": "a"}, WantErr: "is required"},
		{Name: "mode a with val ok", Input: map[string]any{"mode": "a", "val": "x"}},
		{Name: "mode b forbids val", Input: map[string]any{"mode": "b", "val": "x"}, WantErr: "is not allowed"},
		{Name: "mode b no val ok", Input: map[string]any{"mode": "b"}},
	})
}

func TestWhenNotExist(t *testing.T) {
	// pathType: Joi.when('regex', { not: Joi.exist(), then: Joi.string().default('/') })
	rule := Object(map[string]*Rule{
		"regex": String(),
		"path":  String().When("regex", When{Not: Any().Required(), Then: String().Default("/")}),
	})
	run(t, rule, []Case{
		{Name: "no regex → default path", Input: map[string]any{}, Want: map[string]any{"path": "/"}},
		{Name: "regex present → no default", Input: map[string]any{"regex": "x"}, Want: map[string]any{"regex": "x"}},
	})
}

func TestAlternativesTry(t *testing.T) {
	rule := Alternatives(String(), Number())
	run(t, rule, []Case{
		{Name: "string ok", Input: "x", Want: "x"},
		{Name: "number ok", Input: float64(3), Want: float64(3)},
		// Joi raises alternatives.types (not alternatives.match) when every
		// branch is a bare type schema.
		{Name: "bool fails", Input: true, WantErr: "must be one of [string, number]"},
	})
}

func TestAlternativesConditional(t *testing.T) {
	// data: Joi.alternatives().conditional('provider', { is: 'aws', then: Object(aws), otherwise: Object(gcp) })
	rule := Object(map[string]*Rule{
		"provider": String(),
		"data": Alternatives().Conditional("provider", AltCase{
			Is:        String().Valid("aws"),
			Then:      Object(map[string]*Rule{"role": String().Required()}),
			Otherwise: Object(map[string]*Rule{"project": String().Required()}),
		}),
	})
	run(t, rule, []Case{
		{Name: "aws branch ok", Input: map[string]any{"provider": "aws", "data": map[string]any{"role": "r"}}},
		{Name: "aws branch missing role", Input: map[string]any{"provider": "aws", "data": map[string]any{}}, WantErr: "is required"},
		{Name: "gcp branch ok", Input: map[string]any{"provider": "gcp", "data": map[string]any{"project": "p"}}},
	})
}

// TestApplyCtxInPlace exercises the exact entry point base.go:runValidate calls:
// Schema.ApplyCtx mutates obj in place (coerce + default + strip) and returns a
// *Error whose Details() render the parity body.
func TestApplyCtxInPlace(t *testing.T) {
	s := ObjectSchema(Object(map[string]*Rule{
		"name":  String().Required(),
		"count": Number().Integer().Default(float64(1)),
	}))
	obj := map[string]any{"name": "n", "count": "5"}
	if err := s.ApplyCtx(obj, &Context{Org: "o", Method: "POST"}, Prefs{}); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if obj["count"] != float64(5) {
		t.Fatalf("count not coerced in place: %#v", obj["count"])
	}

	obj2 := map[string]any{"count": float64(2)}
	err := s.ApplyCtx(obj2, &Context{}, Prefs{})
	ve, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	d := ve.Details()
	if len(d) != 1 || d[0].Type != "any.required" || d[0].Message == "" {
		t.Fatalf("bad details: %#v", d)
	}
	// default still applied even on the failing object's processed keys? no — abortEarly.
	_ = obj2
}

func TestNestedObjectPath(t *testing.T) {
	rule := Object(map[string]*Rule{
		"spec": Object(map[string]*Rule{
			"count": Number().Integer().Required(),
		}),
	})
	err := (Case{Input: map[string]any{"spec": map[string]any{"count": "x"}}}).Run(rule)
	if err == nil {
		t.Fatal("expected error")
	}
	// verify the path label is nested
	_, verr := Validate(rule, map[string]any{"spec": map[string]any{"count": "x"}}, &Context{}, Prefs{})
	if verr == nil || label(verr.Path) != "spec.count" {
		t.Fatalf("expected path spec.count, got %v", verr)
	}
}

func TestObjectKeyCountMax(t *testing.T) {
	// Joi object().pattern(...).max(2): key-count cap, not string/array length.
	rule := Object(nil).Pattern(String(), String()).Max(2)

	if err := (Case{Input: map[string]any{"a": "1", "b": "2"}}).Run(rule); err != nil {
		t.Fatalf("2 keys should pass: %v", err)
	}
	_, verr := Validate(rule, map[string]any{"a": "1", "b": "2", "c": "3"}, &Context{}, Prefs{})
	if verr == nil || verr.Type != "object.max" {
		t.Fatalf("expected object.max on 3 keys, got %v", verr)
	}

	// Min mirror
	minRule := Object(nil).Pattern(String(), String()).Min(2)
	_, verr = Validate(minRule, map[string]any{"a": "1"}, &Context{}, Prefs{})
	if verr == nil || verr.Type != "object.min" {
		t.Fatalf("expected object.min on 1 key, got %v", verr)
	}
}

// Joi's `helpers.error` closes over the value as it was on ENTRY to the rule chain
// (validator.js:306 — internals.rules never reassigns that binding), so a custom that
// fails AFTER an earlier custom transformed the value still reports the pre-transform
// value in context.value. base.LocalLink is exactly that shape —
// `.custom(expand).custom(toKind)` — and echoing the EXPANDED link told a client its
// request had said something it never said.
func TestCustomErrorReportsChainEntryValue(t *testing.T) {
	expand := func(v any, h *Helper) (any, error) { return "/org/o/gvc/g/" + v.(string), nil }
	reject := func(v any, h *Helper) (any, error) { return nil, h.Errorf("any.custom", "nope %s", v.(string)) }
	rule := Object(map[string]*Rule{
		"link": String().Custom(expand).Custom(reject),
	})
	err := ObjectSchema(rule).ApplyCtx(map[string]any{"link": "//workload/w"}, &Context{}, Prefs{})
	ve, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected the second custom to reject, got %v", err)
	}
	d := ve.Details()
	if len(d) != 1 || d[0].Type != "any.custom" {
		t.Fatalf("details: %#v", d)
	}
	if got := d[0].Context["value"]; got != "//workload/w" {
		t.Fatalf("context.value = %v, want the chain-entry value \"//workload/w\"", got)
	}
}
