package wirejson

import (
	"encoding/json"
	"math/rand"
	"reflect"
	"testing"
)

// Decode must be indistinguishable from json.Unmarshal into an `any`: same Go shapes,
// same float64 for every number, same U+FFFD for invalid UTF-8, same last-wins on
// duplicate keys — and it must REFUSE exactly what the stdlib refuses, since a decoder
// that accepts more would let malformed jsonb through as if it were data.

func sameDecode(t *testing.T, in string) {
	t.Helper()
	var got, want any
	gErr := Decode([]byte(in), &got)
	wErr := json.Unmarshal([]byte(in), &want)
	if (gErr == nil) != (wErr == nil) {
		t.Errorf("input %q\n  wirejson err: %v\n  encoding/json err: %v", in, gErr, wErr)
		return
	}
	if gErr != nil {
		return
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("input %q\n  wirejson:      %#v\n  encoding/json: %#v", in, got, want)
	}
}

func TestDecodeMatchesStdlib(t *testing.T) {
	for _, in := range []string{
		`null`, `true`, `false`, `0`, `-0`, `1`, `-1`, `1.5`, `-1.5e10`, `1e-7`, `1E+21`,
		`123456789012345678901234567890`, `0.1`, `1e308`, `5e-324`,
		`""`, `"a"`, `"A"`, `"🙂"`, `"\ud83d"`, `"\udc42"`, `"\ud83dx"`,
		`"tab\tnl\ncr\rbs\bff\fquote\"slash\/back\\"`, `"é 日"`,
		`{}`, `[]`, `{"a":1}`, `[1,2,3]`, `{"a":{"b":[1,{"c":null}]}}`,
		`{"a":1,"a":2}`, `{"":1}`,
		"  \t\n {\"a\" : [ 1 , 2 ] } \r\n ",
		`{"spec":{"containers":[{"name":"main","ports":[{"number":8080}]}]}}`,
		// malformed — both sides must refuse
		``, ` `, `{`, `}`, `[`, `]`, `[1,]`, `{"a":}`, `{"a" 1}`, `{a:1}`, `tru`, `nul`,
		`"unterminated`, `01`, `-`, `1.`, `.5`, `+1`, `{"a":1}x`, `[1] [2]`, `'a'`,
		`"\q"`, `"\u00"`, `"raw` + "\x01" + `ctl"`, `NaN`, `Infinity`,
	} {
		sameDecode(t, in)
	}
}

// Invalid UTF-8 inside a string: encoding/json substitutes U+FFFD rather than failing.
func TestDecodeInvalidUTF8(t *testing.T) {
	for _, in := range []string{
		"\"\xff\"", "\"a\xffb\"", "\"\xed\xa0\x80\"", "\"ok \xc3\x28 tail\"",
		"{\"\xff\":\"\xfe\"}",
	} {
		sameDecode(t, in)
	}
}

func TestDecodeRandomDocuments(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	for i := 0; i < 3000; i++ {
		v := randValue(rng, 0)
		b, err := json.Marshal(v)
		if err != nil {
			continue
		}
		sameDecode(t, string(b))
		// and mutated, to exercise the error paths
		if len(b) > 2 {
			m := make([]byte, len(b))
			copy(m, b)
			m[rng.Intn(len(m))] = "{}[]\",:x\\ \x01"[rng.Intn(11)]
			sameDecode(t, string(m))
		}
	}
}

// Round trip: decode then encode must reproduce what the stdlib pair reproduces.
func TestDecodeEncodeRoundTrip(t *testing.T) {
	rng := rand.New(rand.NewSource(11))
	for i := 0; i < 1000; i++ {
		b, err := json.Marshal(randValue(rng, 0))
		if err != nil {
			continue
		}
		var mine, theirs any
		if err := Decode(b, &mine); err != nil {
			t.Fatalf("decode %s: %v", b, err)
		}
		if err := json.Unmarshal(b, &theirs); err != nil {
			t.Fatalf("unmarshal %s: %v", b, err)
		}
		mb, err := Marshal(mine)
		if err != nil {
			t.Fatal(err)
		}
		tb, err := stdlib(theirs)
		if err != nil {
			t.Fatal(err)
		}
		if string(mb) != string(tb) {
			t.Errorf("round trip of %s\n  wirejson:      %s\n  encoding/json: %s", b, mb, tb)
		}
	}
}

func BenchmarkDecodeSpec(b *testing.B) {
	doc, err := json.Marshal(benchPage()["items"].([]any)[0].(map[string]any)["spec"])
	if err != nil {
		b.Fatal(err)
	}
	b.Run("wirejson", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			var v any
			if err := Decode(doc, &v); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("encoding_json", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			var v any
			if err := json.Unmarshal(doc, &v); err != nil {
				b.Fatal(err)
			}
		}
	})
}
