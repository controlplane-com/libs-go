package wirejson

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"reflect"
	"testing"
	"unicode/utf8"
)

// Transcode replaces "parse the jsonb into maps, then encode the maps". It is only safe
// if it produces the same bytes that pair produced — with one deliberate exception, key
// ORDER, which it preserves from the input instead of sorting (see the doc comment).
//
// The strongest statement of that is an identity: bytes this package already produced
// must survive a Transcode unchanged, since their keys are already in sorted order.

func TestTranscodeIsIdentityOnWireBytes(t *testing.T) {
	rng := rand.New(rand.NewSource(3))
	for i := 0; i < 3000; i++ {
		wire, err := Marshal(randValue(rng, 0))
		if err != nil {
			continue
		}
		got, err := Transcode(nil, wire)
		if err != nil {
			t.Fatalf("Transcode(%s): %v", wire, err)
		}
		if string(got) != string(wire) {
			t.Errorf("not identity on wire bytes\n  in:  %s\n  out: %s", wire, got)
		}
	}
}

// The same, from the other side: whatever Postgres hands back must transcode to what the
// decode+encode pair would have produced, once key order is taken out of the comparison.
func TestTranscodeMatchesDecodeThenEncode(t *testing.T) {
	rng := rand.New(rand.NewSource(5))
	for i := 0; i < 3000; i++ {
		// Postgres rejects invalid UTF-8 on write, so a column's bytes never carry the
		// \ufffd substitution the encoder makes for it. Keep the corpus to what jsonb
		// can actually hold.
		v := randValue(rng, 0)
		wire, err := Marshal(v)
		if !utf8.Valid(wire) || bytes.Contains(wire, []byte(`\ufffd`)) {
			continue
		}
		if err != nil {
			continue
		}
		src := jsonbFlavour(wire)

		got, err := Transcode(nil, src)
		if err != nil {
			t.Fatalf("Transcode(%s): %v", src, err)
		}
		var decoded any
		if err := Decode(src, &decoded); err != nil {
			t.Fatalf("Decode(%s): %v", src, err)
		}
		want, err := Marshal(decoded)
		if err != nil {
			t.Fatal(err)
		}
		// Values must be identical...
		var a, b any
		if err := json.Unmarshal(got, &a); err != nil {
			t.Fatalf("Transcode produced invalid JSON: %s (%v)", got, err)
		}
		if err := json.Unmarshal(want, &b); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(a, b) {
			t.Errorf("transcode and decode+encode disagree\n  src:  %s\n  got:  %s\n  want: %s", src, got, want)
		}
		// ...and so must the bytes, since jsonbFlavour only added whitespace to output
		// whose keys were already sorted.
		if string(got) != string(want) {
			t.Errorf("bytes differ\n  src:  %s\n  got:  %s\n  want: %s", src, got, want)
		}
	}
}

// Numbers are where Postgres and Go genuinely disagree: jsonb keeps `numeric`, so it
// renders exponents as digits. Transcode must put them back into Go's form, because that
// is what the map path emitted and what clients have been receiving.
func TestTranscodeRendersNumbersLikeTheEncoder(t *testing.T) {
	cases := map[string]string{
		// what Postgres emits            what the wire has always carried
		`{"n": 1000000000000000000000}`: `{"n":1e+21}`,
		`{"n": 0.0000001}`:              `{"n":1e-7}`,
		`{"n": 1}`:                      `{"n":1}`,
		`{"n": 1.5}`:                    `{"n":1.5}`,
		`{"n": -0}`:                     `{"n":-0}`,
		`{"n": 12345678901234567890}`:   `{"n":12345678901234567000}`,
		`{"n": 100000000000000000000}`:  `{"n":100000000000000000000}`,
		`{"n": 0.000001}`:               `{"n":0.000001}`,
	}
	for src, want := range cases {
		got, err := Transcode(nil, []byte(src))
		if err != nil {
			t.Fatalf("Transcode(%s): %v", src, err)
		}
		if string(got) != want {
			t.Errorf("Transcode(%s) = %s, want %s", src, got, want)
		}
		// and it agrees with the pair it replaces
		var v any
		if err := Decode([]byte(src), &v); err != nil {
			t.Fatal(err)
		}
		viaMap, _ := Marshal(v)
		if string(viaMap) != want {
			t.Errorf("decode+encode(%s) = %s, want %s — the expectation itself is wrong", src, viaMap, want)
		}
	}
}

// Key order is preserved, not sorted. This is the one intentional difference from the
// map path, and it is what the TS service emits.
func TestTranscodePreservesKeyOrder(t *testing.T) {
	src := []byte(`{"A": 3, "a": 4, "b": 1, "aa": 2}`) // Postgres's length-then-bytes order
	got, err := Transcode(nil, src)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `{"A":3,"a":4,"b":1,"aa":2}` {
		t.Errorf("key order not preserved: %s", got)
	}
	// the map path would have sorted them
	var v any
	if err := Decode(src, &v); err != nil {
		t.Fatal(err)
	}
	sorted, _ := Marshal(v)
	if string(sorted) == string(got) {
		t.Error("this input does not distinguish the two orders — pick another")
	}
}

// U+2028/U+2029 come out of Postgres raw and must not reach the wire that way.
func TestTranscodeEscapesLineSeparators(t *testing.T) {
	src := []byte("{\"s\": \"a\u2028b\u2029c\"}")
	got, err := Transcode(nil, src)
	if err != nil {
		t.Fatal(err)
	}
	const want = `{"s":"a\u2028b\u2029c"}`
	if string(got) != want {
		t.Errorf("Transcode = %s, want %s", got, want)
	}
}

func TestTranscodeRejectsWhatDecodeRejects(t *testing.T) {
	for _, in := range []string{
		``, ` `, `{`, `}`, `[1,]`, `{"a":}`, `{a:1}`, `tru`, `"unterminated`, `01`, `1.`,
		`{"a":1}x`, `[1] [2]`, `'a'`, `"\q"`, `NaN`,
	} {
		_, tErr := Transcode(nil, []byte(in))
		var v any
		dErr := Decode([]byte(in), &v)
		if (tErr == nil) != (dErr == nil) {
			t.Errorf("input %q: Transcode err=%v, Decode err=%v", in, tErr, dErr)
		}
	}
}

// jsonbFlavour reproduces Postgres's spacing: a space after every structural ':' and ','.
func jsonbFlavour(b []byte) []byte {
	out := make([]byte, 0, len(b)*2)
	inStr := false
	for i := 0; i < len(b); i++ {
		c := b[i]
		out = append(out, c)
		switch {
		case inStr && c == '\\':
			if i+1 < len(b) {
				i++
				out = append(out, b[i])
			}
		case c == '"':
			inStr = !inStr
		case !inStr && (c == ':' || c == ','):
			out = append(out, ' ')
		}
	}
	return out
}

func BenchmarkTranscodeSpec(b *testing.B) {
	spec, err := Marshal(benchPage()["items"].([]any)[0].(map[string]any)["spec"])
	if err != nil {
		b.Fatal(err)
	}
	src := jsonbFlavour(spec)
	b.Run("transcode", func(b *testing.B) {
		b.ReportAllocs()
		buf := make([]byte, 0, 4096)
		for i := 0; i < b.N; i++ {
			if buf, err = Transcode(buf[:0], src); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("decode_then_encode", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			var v any
			if err := Decode(src, &v); err != nil {
				b.Fatal(err)
			}
			if _, err := Marshal(v); err != nil {
				b.Fatal(err)
			}
		}
	})
}
