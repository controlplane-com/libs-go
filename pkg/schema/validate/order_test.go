package validate

import (
	"reflect"
	"testing"

	"github.com/controlplane-com/libs-go/pkg/schema/wirejson"
)

// ordDoc covers the destinations a map-shaped object can land in: a typed map, an
// `any` (the whole subtree stays generic), and an array element. A struct destination
// has no map for a custom to read keys off, so it is only here to keep the document
// mixed.
type ordDoc struct {
	Sub *ordSub           `json:"sub"`
	M   map[string]string `json:"m"`
	Any any               `json:"any"`
	Arr []any             `json:"arr"`
}

type ordSub struct {
	A string `json:"a"`
	B string `json:"b"`
}

// capture is a custom that files the order it was given, so a test can read what a
// rule at that path would have seen.
func capture(into map[string][]string, at string) CustomFn {
	return func(v any, h *Helper) (any, error) {
		into[at] = OrderedKeys(v, h)
		return v, nil
	}
}

func ordRule(seen map[string][]string) *Rule {
	return Object(map[string]*Rule{
		"sub": Object(map[string]*Rule{"a": String(), "b": String()}),
		"m":   Object(nil).CustomTyped(capture(seen, "m")),
		"any": Object(nil).CustomTyped(capture(seen, "any")),
		"arr": Array(Object(nil).CustomTyped(capture(seen, "arr"))),
	})
}

// TestWireOrderReachesEveryDestination pins that the decoder's record survives to a
// custom at each kind of destination, and that it is the BODY's order rather than the
// map's or the schema's.
func TestWireOrderReachesEveryDestination(t *testing.T) {
	body := `{"m":{"z":"1","a":"2"},"any":{"z":1,"a":2},"sub":{"b":"1","a":"2"},"arr":[{"z":1,"a":2}]}`
	var dst ordDoc
	d, err := DecodeInto(&dst, Bytes(body))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	seen := map[string][]string{}
	if verr := d.Validate(ordRule(seen), &Context{}, Prefs{}); verr != nil {
		t.Fatalf("validate: %s", verr.Message)
	}
	want := map[string][]string{
		"m":   {"z", "a"},
		"any": {"z", "a"},
		"arr": {"z", "a"},
	}
	for at, w := range want {
		if got := seen[at]; !reflect.DeepEqual(got, w) {
			t.Errorf("at %q: got %v, want %v", at, got, w)
		}
	}
}

// TestWireOrderIsSharedByBothEvaluators runs one body twice — once into typed
// destinations, once into a single `any` so the whole document stays generic — and
// requires the same answer. The two evaluators reading different orders at the same
// path would be its own class of bug.
func TestWireOrderIsSharedByBothEvaluators(t *testing.T) {
	body := `{"m":{"z":"1","a":"2"},"any":{"z":1,"a":2},"sub":{"b":"1","a":"2"}}`

	typedSeen := map[string][]string{}
	var typed ordDoc
	d, err := DecodeInto(&typed, Bytes(body))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if verr := d.Validate(ordRule(typedSeen), &Context{}, Prefs{}); verr != nil {
		t.Fatalf("typed: %s", verr.Message)
	}

	genericSeen := map[string][]string{}
	var generic struct {
		Doc any `json:"doc"`
	}
	gd, err := DecodeInto(&generic, Bytes(`{"doc":`+body+`}`))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	// the customs run one level down, so file them under the same names
	inner := map[string][]string{}
	if verr := gd.Validate(Object(map[string]*Rule{"doc": ordRule(inner)}), &Context{}, Prefs{}); verr != nil {
		t.Fatalf("generic: %s", verr.Message)
	}
	for k, v := range inner {
		genericSeen[k] = v
	}
	if !reflect.DeepEqual(typedSeen, genericSeen) {
		t.Fatalf("evaluators disagree: typed %v, generic %v", typedSeen, genericSeen)
	}
}

// TestWireOrderAddressesNestedObjects covers the reason the table is keyed by path:
// the custom that needs the order runs several levels ABOVE the object whose keys it
// reads (envoy's uiFilterNamingValidator reads http[i].typed_config.providers).
func TestWireOrderAddressesNestedObjects(t *testing.T) {
	var got []string
	rule := Object(map[string]*Rule{
		"arr": Array(Object(nil)),
	}).CustomTyped(func(v any, h *Helper) (any, error) {
		m, _ := v.(map[string]any)
		arr, _ := m["arr"].([]any)
		got = OrderedKeys(arr[1], h, "arr", 1)
		return v, nil
	})
	var dst struct {
		Doc any `json:"doc"`
	}
	d, err := DecodeInto(&dst, Bytes(`{"doc":{"arr":[{"a":1},{"z":1,"b":2,"a":3}]}}`))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if verr := d.Validate(Object(map[string]*Rule{"doc": rule}), &Context{}, Prefs{}); verr != nil {
		t.Fatalf("validate: %s", verr.Message)
	}
	if want := []string{"z", "b", "a"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

// TestWireOrderUnderABase covers Decoded.ValidateAt: the decode paths are relative to
// the struct, the validation paths carry a base, and the lookup has to bridge that.
func TestWireOrderUnderABase(t *testing.T) {
	var got []string
	var dst struct {
		M map[string]string `json:"m"`
	}
	d, err := DecodeInto(&dst, Bytes(`{"m":{"z":"1","a":"2"}}`))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	rule := Object(map[string]*Rule{"m": Object(nil).CustomTyped(func(v any, h *Helper) (any, error) {
		got = OrderedKeys(v, h)
		return v, nil
	})})
	if verr := d.ValidateAt(rule, []string{"spec", IndexPath(2)}, &Context{}, Prefs{}); verr != nil {
		t.Fatalf("validate: %s", verr.Message)
	}
	if want := []string{"z", "a"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

// TestWireOrderDuplicateKeyKeepsFirstPosition is JS's own answer for a body that
// repeats a key: the LAST value wins, at the FIRST position.
//
//	node -e 'console.log(Object.keys(JSON.parse(`{"b":1,"a":2,"b":3}`)))' -> [ 'b', 'a' ]
func TestWireOrderDuplicateKeyKeepsFirstPosition(t *testing.T) {
	for _, body := range []string{
		`{"m":{"b":"1","a":"2","b":"3"}}`, // map destination
		`{"any":{"b":1,"a":2,"b":3}}`,     // generic subtree
	} {
		var dst ordDoc
		d, err := DecodeInto(&dst, Bytes(body))
		if err != nil {
			t.Fatalf("decode %s: %v", body, err)
		}
		seen := map[string][]string{}
		if verr := d.Validate(ordRule(seen), &Context{}, Prefs{}); verr != nil {
			t.Fatalf("validate %s: %s", body, verr.Message)
		}
		for _, at := range []string{"m", "any"} {
			if got, ok := seen[at]; ok && !reflect.DeepEqual(got, []string{"b", "a"}) {
				t.Errorf("%s at %q: got %v, want [b a]", body, at, got)
			}
		}
	}
}

// TestOrderKeysMatchesObjectKeys pins orderKeys against `Object.keys` for the shapes
// where JS does NOT simply use insertion order. Captured 2026-09-16 with
//
//	node -e 'console.log(Object.keys(JSON.parse(<body>)))'
func TestOrderKeysMatchesObjectKeys(t *testing.T) {
	for _, tc := range []struct {
		name string
		wire []string // the body's order, which is also the key set here
		want []string // what node printed
	}{
		{"plain insertion order", []string{"b", "a"}, []string{"b", "a"}},
		// integer-like keys jump the queue, in numeric order
		{"array indices first", []string{"10", "b", "2"}, []string{"2", "10", "b"}},
		{"non-canonical is not an index", []string{"01", "b", "1"}, []string{"1", "01", "b"}},
		{"negative is not an index", []string{"-1", "b", "1"}, []string{"1", "-1", "b"}},
		{"2^32-1 is not an index", []string{"4294967295", "b", "1"}, []string{"1", "4294967295", "b"}},
		{"2^32-2 is", []string{"4294967294", "b", "1"}, []string{"1", "4294967294", "b"}},
		{"zero is, 00 is not", []string{"0", "b", "00"}, []string{"0", "b", "00"}},
		{"a fraction is not", []string{"1.5", "b", "1"}, []string{"1", "1.5", "b"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			present := append([]string(nil), tc.wire...)
			if got := orderKeys(present, tc.wire); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}

// TestOrderKeysFallsBackToStoredOrder covers the keys the wire never carried — the
// PATCH merge tree, whose order TS gets from the jsonb column it read the row out of:
// shortest key first, then bytewise. A plain lexicographic fallback answered `cpln_aa`
// where TS answers `cpln_z`, which split the PATCH verdict.
func TestOrderKeysFallsBackToStoredOrder(t *testing.T) {
	got := orderKeys([]string{"cpln_aa", "cpln_z", "cpln_ab", "cpln_y", "cpln_bbb"}, nil)
	// captured from the oracle: GET the stored gvc and read the provider key order
	want := []string{"cpln_y", "cpln_z", "cpln_aa", "cpln_ab", "cpln_bbb"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	// keys the wire DID carry still lead, in wire order
	got = orderKeys([]string{"b", "a", "zz"}, []string{"b"})
	if want := []string{"b", "a", "zz"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

// TestTreeSourceHasNoWireOrder: an already-parsed tree has no order left to record, so
// it must degrade to the fallback rather than report a made-up one.
func TestTreeSourceHasNoWireOrder(t *testing.T) {
	var dst ordDoc
	d, err := DecodeInto(&dst, Tree{V: map[string]any{"m": map[string]any{"zz": "1", "a": "2"}}})
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	seen := map[string][]string{}
	if verr := d.Validate(ordRule(seen), &Context{}, Prefs{}); verr != nil {
		t.Fatalf("validate: %s", verr.Message)
	}
	if want := []string{"a", "zz"}; !reflect.DeepEqual(seen["m"], want) {
		t.Fatalf("got %v, want %v", seen["m"], want)
	}
}

// TestOrderingGenericMatchesWirejson: recording the order meant reading generic
// subtrees through the cursor instead of wirejson's one-shot tree builder, so the two
// have to produce the same tree — including float64 for every number, [] vs null for
// an empty array, and last-wins duplicate keys.
func TestOrderingGenericMatchesWirejson(t *testing.T) {
	docs := []string{
		`{}`, `[]`, `null`, `0`, `-1.5e3`, `""`, `"é😀"`, `true`,
		`{"a":1}`, `{"a":{},"b":[]}`, `[[[]]]`, `[{"a":[1,2,{"b":null}]}]`,
		`{"b":1,"a":2,"b":3}`, `{"a":[],"b":{},"c":null,"d":false,"e":"x","f":1e-3}`,
		`{"":1,"0":2,"10":3}`, `[1,[2,[3,[4,[5]]]]]`,
	}
	for _, doc := range docs {
		var dst struct {
			Any any `json:"any"`
		}
		if _, err := DecodeInto(&dst, Bytes(`{"any":`+doc+`}`)); err != nil {
			t.Fatalf("%s: decode: %v", doc, err)
		}
		var want any
		if err := wirejson.Decode([]byte(doc), &want); err != nil {
			t.Fatalf("%s: wirejson: %v", doc, err)
		}
		if !reflect.DeepEqual(dst.Any, want) {
			t.Errorf("%s: got %#v, want %#v", doc, dst.Any, want)
		}
	}
}

// TestPathKeyIsUnambiguous: any byte is legal inside a JSON key, including the
// engine's own array-index marker, so a separator-joined path key would let one body's
// key collide with another's path and hand a custom the wrong object's order.
func TestPathKeyIsUnambiguous(t *testing.T) {
	for _, pair := range [][2][]string{
		{{"a\x00b"}, {"a", "b"}},
		{{"a" + IndexPath(0)}, {"a", IndexPath(0)}},
		{{"a:b"}, {"a", "b"}},
		{{"1:a"}, {"a"}},
	} {
		if pathKey(pair[0]) == pathKey(pair[1]) {
			t.Errorf("%q and %q collide on %q", pair[0], pair[1], pathKey(pair[0]))
		}
	}
}
