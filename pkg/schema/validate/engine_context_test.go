package validate

// engine_context_test.go pins the FULL Joi error envelope — message, path, type and
// every context key — for one instance of each error the engine can raise. The
// expectations are verbatim `error.details[0]` from joi@17.6.0 with
// {convert:true, abortEarly:true}.
//
// It exists because the context keys were being found missing one at a time (peers on
// the dependency errors, invalids on any.invalid, value on the forbidden-key and typed
// container errors). A client renders its own message from these, so a missing key is a
// real difference, and a per-type table is the only way to stop rediscovering it.
//
// Every case also runs through the TYPED walk, because that is the main write path and
// the two evaluators must produce the same envelope.

import (
	"encoding/json"
	"regexp"
	"testing"
)

type ctxAnyRow struct {
	A  any `json:"a,omitzero"`
	B  any `json:"b,omitzero"`
	ZZ any `json:"zz,omitzero"`
}

type ctxSliceRow struct {
	A []any `json:"a,omitzero"`
}

type ctxMapRow struct {
	A map[string]any `json:"a,omitzero"`
}

func TestErrorContextMatchesJoi(t *testing.T) {
	cases := []struct {
		name string
		rule *Rule
		in   string
		dst  func() any
	}{
		{"string.base", Object(map[string]*Rule{"a": String()}), `{"a":5}`, nil},
		{"string.empty", Object(map[string]*Rule{"a": String()}), `{"a":""}`, nil},
		{"string.min", Object(map[string]*Rule{"a": String().Min(3)}), `{"a":"x"}`, nil},
		{"string.max", Object(map[string]*Rule{"a": String().Max(1)}), `{"a":"xy"}`, nil},
		{"string.length", Object(map[string]*Rule{"a": String().Length(3)}), `{"a":"xy"}`, nil},
		{"string.pattern", Object(map[string]*Rule{"a": String().Regex(regexp.MustCompile(`^[a-z]+$`))}), `{"a":"X1"}`, nil},
		{"string.email", Object(map[string]*Rule{"a": String().Email()}), `{"a":"nope"}`, nil},
		{"string.uri", Object(map[string]*Rule{"a": String().URI()}), `{"a":"nope"}`, nil},
		{"string.uriCustomScheme", Object(map[string]*Rule{"a": String().URIScheme("http", "https")}), `{"a":"ftp://x/y"}`, nil},
		{"string.base64", Object(map[string]*Rule{"a": String().Base64()}), `{"a":"!!!!"}`, nil},
		{"string.guid", Object(map[string]*Rule{"a": String().UUID()}), `{"a":"nope"}`, nil},
		{"string.domain", Object(map[string]*Rule{"a": String().Domain()}), `{"a":"nope"}`, nil},
		{"number.base", Object(map[string]*Rule{"a": Number()}), `{"a":"zz"}`, nil},
		{"number.min", Object(map[string]*Rule{"a": Number().Min(5)}), `{"a":1}`, nil},
		{"number.max", Object(map[string]*Rule{"a": Number().Max(5)}), `{"a":9}`, nil},
		{"number.integer", Object(map[string]*Rule{"a": Number().Integer()}), `{"a":1.5}`, nil},
		{"number.positive", Object(map[string]*Rule{"a": Number().Positive()}), `{"a":-1}`, nil},
		{"number.negative", Object(map[string]*Rule{"a": Number().Negative()}), `{"a":1}`, nil},
		{"number.port", Object(map[string]*Rule{"a": Number().Port()}), `{"a":99999}`, nil},
		{"number.unsafe", Object(map[string]*Rule{"a": Number()}), `{"a":1e30}`, nil},
		{"boolean.base", Object(map[string]*Rule{"a": Boolean()}), `{"a":"zz"}`, nil},
		{"date.base", Object(map[string]*Rule{"a": DateRule()}), `{"a":"zz"}`, nil},
		{"date.format", Object(map[string]*Rule{"a": DateRule().Iso()}), `{"a":"2020/01/02"}`, nil},
		{"array.base", Object(map[string]*Rule{"a": Array()}), `{"a":"zz"}`, nil},
		{"array.min", Object(map[string]*Rule{"a": Array().Min(2)}), `{"a":["x"]}`, sliceDst},
		{"array.max", Object(map[string]*Rule{"a": Array().Max(1)}), `{"a":["x","y"]}`, sliceDst},
		{"array.length", Object(map[string]*Rule{"a": Array().Length(2)}), `{"a":["x"]}`, sliceDst},
		{"array.unique", Object(map[string]*Rule{"a": Array(String()).Unique()}), `{"a":["x","x"]}`, sliceDst},
		{"array.unique.path", Object(map[string]*Rule{"a": Array(Object(map[string]*Rule{"k": String(), "v": Number()})).Unique("k")}), `{"a":[{"k":"x","v":1},{"k":"x","v":2}]}`, sliceDst},
		{"array.unique.comparator", Object(map[string]*Rule{"a": Array(Object(map[string]*Rule{"k": String(), "v": Number()})).UniqueComparing("k")}), `{"a":[{"k":"x","v":1},{"k":"x","v":2}]}`, sliceDst},
		{"array.includes", Object(map[string]*Rule{"a": Array(String(), Number())}), `{"a":[true]}`, sliceDst},
		{"array.sort", Object(map[string]*Rule{"a": Array().Sort()}), `{"a":["c",1]}`, sliceDst},
		{"object.base", Object(map[string]*Rule{"a": Object(map[string]*Rule{})}), `{"a":"zz"}`, nil},
		{"object.unknown", Object(map[string]*Rule{"a": String()}), `{"a":"x","zz":1}`, nil},
		{"object.min", Object(map[string]*Rule{"a": Object(nil).Min(2)}), `{"a":{"x":1}}`, mapDst},
		{"object.max", Object(map[string]*Rule{"a": Object(map[string]*Rule{"x": Any(), "y": Any()}).Max(1)}), `{"a":{"x":1,"y":2}}`, mapDst},
		{"object.missing", Object(map[string]*Rule{"a": Any(), "b": Any()}).Or("a", "b"), `{}`, nil},
		{"object.xor", Object(map[string]*Rule{"a": Any(), "b": Any()}).Xor("a", "b"), `{"a":1,"b":2}`, nil},
		{"object.oxor", Object(map[string]*Rule{"a": Any(), "b": Any()}).Oxor("a", "b"), `{"a":1,"b":2}`, nil},
		{"object.nand", Object(map[string]*Rule{"a": Any(), "b": Any()}).Nand("a", "b"), `{"a":1,"b":2}`, nil},
		{"any.required", Object(map[string]*Rule{"a": String().Required()}), `{}`, nil},
		{"any.only", Object(map[string]*Rule{"a": String().Valid("x", "y")}), `{"a":"z"}`, nil},
		{"any.invalid", Object(map[string]*Rule{"a": String().Invalid("no")}), `{"a":"no"}`, nil},
		{"message.override", Object(map[string]*Rule{"a": String().Regex(regexp.MustCompile(`^[a-z]+$`)).Message("only lowercase letters, please")}), `{"a":"X1"}`, nil},
		{"any.unknown", Object(map[string]*Rule{"a": Any().Forbidden()}), `{"a":{"c":3}}`, nil},
		{"any.ref", Object(map[string]*Rule{"a": Number(), "b": Number().MinRef("a")}), `{"b":3}`, nil},
		{"alt.types", Object(map[string]*Rule{"a": Alternatives(String(), Number())}), `{"a":true}`, nil},
		{"alt.match", Object(map[string]*Rule{"a": Alternatives(Object(map[string]*Rule{"x": String().Required()}), Object(map[string]*Rule{"y": String().Required()}))}), `{"a":{}}`, nil},
	}
	for _, c := range cases {
		want, ok := joiDetails[c.name]
		if !ok {
			t.Fatalf("no expectation for %s", c.name)
		}
		var tree map[string]any
		if err := json.Unmarshal([]byte(c.in), &tree); err != nil {
			t.Fatal(err)
		}
		_, gerr := Validate(c.rule, tree, nil, Prefs{})
		if got := detailJSON(t, gerr); got != canonical(t, want) {
			t.Errorf("%s (map evaluator)\n  go:  %s\n  joi: %s", c.name, got, canonical(t, want))
		}
		dst := any(&ctxAnyRow{})
		if c.dst != nil {
			dst = c.dst()
		}
		d, derr := DecodeInto(dst, Bytes(c.in))
		if derr != nil {
			t.Fatalf("%s: %v", c.name, derr)
		}
		terr := d.Validate(c.rule, &Context{}, Prefs{})
		if got := detailJSON(t, terr); got != canonical(t, want) {
			t.Errorf("%s (typed walk)\n  go:  %s\n  joi: %s", c.name, got, canonical(t, want))
		}
	}
}

func sliceDst() any { return &ctxSliceRow{} }
func mapDst() any   { return &ctxMapRow{} }

func detailJSON(t *testing.T, e *Error) string {
	t.Helper()
	if e == nil {
		return "OK"
	}
	return canonical(t, string(mustJSON(t, e.Details()[0])))
}

// canonical re-encodes through a map so key ORDER does not enter the comparison.
func canonical(t *testing.T, js string) string {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal([]byte(js), &m); err != nil {
		t.Fatalf("bad json %q: %v", js, err)
	}
	return string(mustJSON(t, m))
}

func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// joiDetails is `error.details[0]` from joi@17.6.0, verbatim.
var joiDetails = map[string]string{
	"string.base":            `{"message":"\"a\" must be a string","path":["a"],"type":"string.base","context":{"label":"a","value":5,"key":"a"}}`,
	"string.empty":           `{"message":"\"a\" is not allowed to be empty","path":["a"],"type":"string.empty","context":{"label":"a","value":"","key":"a"}}`,
	"string.min":             `{"message":"\"a\" length must be at least 3 characters long","path":["a"],"type":"string.min","context":{"limit":3,"value":"x","label":"a","key":"a"}}`,
	"string.max":             `{"message":"\"a\" length must be less than or equal to 1 characters long","path":["a"],"type":"string.max","context":{"limit":1,"value":"xy","label":"a","key":"a"}}`,
	"string.length":          `{"message":"\"a\" length must be 3 characters long","path":["a"],"type":"string.length","context":{"limit":3,"value":"xy","label":"a","key":"a"}}`,
	"string.pattern":         `{"message":"\"a\" with value \"X1\" fails to match the required pattern: /^[a-z]+$/","path":["a"],"type":"string.pattern.base","context":{"regex":{},"value":"X1","label":"a","key":"a"}}`,
	"string.email":           `{"message":"\"a\" must be a valid email","path":["a"],"type":"string.email","context":{"value":"nope","invalids":["nope"],"label":"a","key":"a"}}`,
	"string.uri":             `{"message":"\"a\" must be a valid uri","path":["a"],"type":"string.uri","context":{"label":"a","value":"nope","key":"a"}}`,
	"string.uriCustomScheme": `{"message":"\"a\" must be a valid uri with a scheme matching the http|https pattern","path":["a"],"type":"string.uriCustomScheme","context":{"scheme":"http|https","value":"ftp://x/y","label":"a","key":"a"}}`,
	"string.base64":          `{"message":"\"a\" must be a valid base64 string","path":["a"],"type":"string.base64","context":{"label":"a","value":"!!!!","key":"a"}}`,
	// `.message()` swaps ONLY the wording — the type and the regex/value context that
	// a client renders from stay the rule's own.
	"message.override": `{"message":"only lowercase letters, please","path":["a"],"type":"string.pattern.base","context":{"regex":{},"value":"X1","label":"a","key":"a"}}`,
	// dupeValue is the EARLIER element; `path` appears only for the string-comparator
	// spelling of .unique(), never for Joi's function comparator.
	"array.unique.path":       `{"message":"\"a[1]\" contains a duplicate value","path":["a",1],"type":"array.unique","context":{"pos":1,"value":{"k":"x","v":2},"dupePos":0,"dupeValue":{"k":"x","v":1},"path":"k","label":"a[1]","key":1}}`,
	"array.unique.comparator": `{"message":"\"a[1]\" contains a duplicate value","path":["a",1],"type":"array.unique","context":{"pos":1,"value":{"k":"x","v":2},"dupePos":0,"dupeValue":{"k":"x","v":1},"label":"a[1]","key":1}}`,
	"string.guid":             `{"message":"\"a\" must be a valid GUID","path":["a"],"type":"string.guid","context":{"label":"a","value":"nope","key":"a"}}`,
	"string.domain":           `{"message":"\"a\" must contain a valid domain name","path":["a"],"type":"string.domain","context":{"label":"a","value":"nope","key":"a"}}`,
	"number.base":             `{"message":"\"a\" must be a number","path":["a"],"type":"number.base","context":{"label":"a","value":"zz","key":"a"}}`,
	"number.min":              `{"message":"\"a\" must be greater than or equal to 5","path":["a"],"type":"number.min","context":{"limit":5,"value":1,"label":"a","key":"a"}}`,
	"number.max":              `{"message":"\"a\" must be less than or equal to 5","path":["a"],"type":"number.max","context":{"limit":5,"value":9,"label":"a","key":"a"}}`,
	"number.integer":          `{"message":"\"a\" must be an integer","path":["a"],"type":"number.integer","context":{"label":"a","value":1.5,"key":"a"}}`,
	"number.positive":         `{"message":"\"a\" must be a positive number","path":["a"],"type":"number.positive","context":{"label":"a","value":-1,"key":"a"}}`,
	"number.negative":         `{"message":"\"a\" must be a negative number","path":["a"],"type":"number.negative","context":{"label":"a","value":1,"key":"a"}}`,
	"number.port":             `{"message":"\"a\" must be a valid port","path":["a"],"type":"number.port","context":{"label":"a","value":99999,"key":"a"}}`,
	"number.unsafe":           `{"message":"\"a\" must be a safe number","path":["a"],"type":"number.unsafe","context":{"label":"a","value":1e+30,"key":"a"}}`,
	"boolean.base":            `{"message":"\"a\" must be a boolean","path":["a"],"type":"boolean.base","context":{"label":"a","value":"zz","key":"a"}}`,
	"date.base":               `{"message":"\"a\" must be a valid date","path":["a"],"type":"date.base","context":{"label":"a","value":"zz","key":"a"}}`,
	"date.format":             `{"message":"\"a\" must be in ISO 8601 date format","path":["a"],"type":"date.format","context":{"format":"iso","label":"a","value":"2020/01/02","key":"a"}}`,
	"array.base":              `{"message":"\"a\" must be an array","path":["a"],"type":"array.base","context":{"label":"a","value":"zz","key":"a"}}`,
	"array.min":               `{"message":"\"a\" must contain at least 2 items","path":["a"],"type":"array.min","context":{"limit":2,"value":["x"],"label":"a","key":"a"}}`,
	"array.max":               `{"message":"\"a\" must contain less than or equal to 1 items","path":["a"],"type":"array.max","context":{"limit":1,"value":["x","y"],"label":"a","key":"a"}}`,
	"array.length":            `{"message":"\"a\" must contain 2 items","path":["a"],"type":"array.length","context":{"limit":2,"value":["x"],"label":"a","key":"a"}}`,
	"array.unique":            `{"message":"\"a[1]\" contains a duplicate value","path":["a",1],"type":"array.unique","context":{"pos":1,"value":"x","dupePos":0,"dupeValue":"x","label":"a[1]","key":1}}`,
	"array.includes":          `{"message":"\"a[0]\" does not match any of the allowed types","path":["a",0],"type":"array.includes","context":{"pos":0,"value":true,"label":"a[0]","key":0}}`,
	"array.sort":              `{"message":"\"a\" cannot be sorted due to mismatching types","path":["a"],"type":"array.sort.mismatching","context":{"label":"a","value":["c",1],"key":"a"}}`,
	"object.base":             `{"message":"\"a\" must be of type object","path":["a"],"type":"object.base","context":{"type":"object","label":"a","value":"zz","key":"a"}}`,
	"object.unknown":          `{"message":"\"zz\" is not allowed","path":["zz"],"type":"object.unknown","context":{"child":"zz","label":"zz","value":1,"key":"zz"}}`,
	"object.min":              `{"message":"\"a\" must have at least 2 keys","path":["a"],"type":"object.min","context":{"limit":2,"value":{"x":1},"label":"a","key":"a"}}`,
	"object.max":              `{"message":"\"a\" must have less than or equal to 1 key","path":["a"],"type":"object.max","context":{"limit":1,"value":{"x":1,"y":2},"label":"a","key":"a"}}`,
	"object.missing":          `{"message":"\"value\" must contain at least one of [a, b]","path":[],"type":"object.missing","context":{"peers":["a","b"],"peersWithLabels":["a","b"],"label":"value","value":{}}}`,
	"object.xor":              `{"message":"\"value\" contains a conflict between exclusive peers [a, b]","path":[],"type":"object.xor","context":{"peers":["a","b"],"peersWithLabels":["a","b"],"present":["a","b"],"presentWithLabels":["a","b"],"label":"value","value":{"a":1,"b":2}}}`,
	"object.oxor":             `{"message":"\"value\" contains a conflict between optional exclusive peers [a, b]","path":[],"type":"object.oxor","context":{"peers":["a","b"],"peersWithLabels":["a","b"],"present":["a","b"],"presentWithLabels":["a","b"],"label":"value","value":{"a":1,"b":2}}}`,
	"object.nand":             `{"message":"\"a\" must not exist simultaneously with [b]","path":[],"type":"object.nand","context":{"main":"a","mainWithLabel":"a","peers":["b"],"peersWithLabels":["b"],"label":"value","value":{"a":1,"b":2}}}`,
	"any.required":            `{"message":"\"a\" is required","path":["a"],"type":"any.required","context":{"label":"a","key":"a"}}`,
	"any.only":                `{"message":"\"a\" must be one of [x, y]","path":["a"],"type":"any.only","context":{"valids":["x","y"],"label":"a","value":"z","key":"a"}}`,
	"any.invalid":             `{"message":"\"a\" contains an invalid value","path":["a"],"type":"any.invalid","context":{"invalids":["no"],"label":"a","value":"no","key":"a"}}`,
	"any.unknown":             `{"message":"\"a\" is not allowed","path":["a"],"type":"any.unknown","context":{"label":"a","value":{"c":3},"key":"a"}}`,
	"any.ref":                 `{"message":"\"b\" limit references \"ref:a\" which must be a number","path":["b"],"type":"any.ref","context":{"arg":"limit","ref":{"adjust":null,"in":false,"iterables":null,"map":null,"separator":".","type":"value","ancestor":1,"path":["a"],"depth":1,"key":"a","root":"a","display":"ref:a"},"reason":"must be a number","label":"b","key":"b"}}`,
	"alt.types":               `{"message":"\"a\" must be one of [string, number]","path":["a"],"type":"alternatives.types","context":{"types":["string","number"],"label":"a","value":true,"key":"a"}}`,
	"alt.match":               `{"message":"\"a\" does not match any of the allowed types","path":["a"],"type":"alternatives.match","context":{"message":"\"a.x\" is required. \"a.y\" is required","details":[{"message":"\"a.x\" is required","path":["a","x"],"type":"any.required","context":{"label":"a.x","key":"x"}},{"message":"\"a.y\" is required","path":["a","y"],"type":"any.required","context":{"label":"a.y","key":"y"}}],"label":"a","value":{},"key":"a"}}`}

// TestCaseInsensitiveRegexMessage pins the JS spelling of a case-insensitive pattern.
// Go carries the flag inline as `(?i)` and JS as a trailing `i`, so the message used
// to quote a pattern (`/(?i)^sha…$/`) that is not valid JS and does not match what
// the TS schema reports for the same value.
func TestCaseInsensitiveRegexMessage(t *testing.T) {
	rule := Object(map[string]*Rule{"a": String().Regex(regexp.MustCompile(`(?i)^sha(256|512):[a-f0-9]+$`))})
	const want = `"a" with value "md5:abc" fails to match the required pattern: /^sha(256|512):[a-f0-9]+$/i`

	var tree map[string]any
	if err := json.Unmarshal([]byte(`{"a":"md5:abc"}`), &tree); err != nil {
		t.Fatal(err)
	}
	if _, gerr := Validate(rule, tree, nil, Prefs{}); gerr == nil || gerr.Message != want {
		t.Errorf("map evaluator: got %v want %q", gerr, want)
	}
	d, err := DecodeInto(&ctxAnyRow{}, Bytes(`{"a":"md5:abc"}`))
	if err != nil {
		t.Fatal(err)
	}
	if terr := d.Validate(rule, &Context{}, Prefs{}); terr == nil || terr.Message != want {
		t.Errorf("typed walk: got %v want %q", terr, want)
	}
}
