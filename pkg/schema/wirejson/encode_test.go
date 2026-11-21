package wirejson

import (
	"bytes"
	"encoding/json"
	"math"
	"math/rand"
	"testing"
)

// The whole contract: Marshal must produce exactly what encoding/json.Marshal produces.
// It is a hand-written encoder standing in for the reflect-based one on the hot path,
// so "close enough" is a silent wire change — a mis-escaped ampersand in a description,
// a float that rounds differently, a key order that drifts.

func same(t *testing.T, v any) {
	t.Helper()
	got, err := Marshal(v)
	if err != nil {
		t.Fatalf("wirejson.Marshal(%#v): %v", v, err)
	}
	want, err := stdlib(v)
	if err != nil {
		t.Fatalf("stdlib(%#v): %v", v, err)
	}
	if string(got) != string(want) {
		t.Errorf("value %#v\n  wirejson:      %s\n  encoding/json: %s", v, got, want)
	}
}

// stdlib is the encoder httpx.WriteJSON used before this package existed — the oracle
// this must match. SetEscapeHTML(false), and the trailing newline Encode appends
// removed, exactly as WriteJSON does.
func stdlib(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return bytes.TrimSuffix(buf.Bytes(), []byte("\n")), nil
}

func TestMatchesStdlibOnScalars(t *testing.T) {
	for _, v := range []any{
		nil, true, false, "", "plain",
		// every escape encoding/json applies, including the HTML set it escapes by
		// default and the JS line separators
		"quote\" backslash\\ slash/ tab\t nl\n cr\r", "<script>&amp;</script>",
		"  ", "\x00\x01\x1f", "unicode: é 日本 \U0001f642",
		// invalid UTF-8 is pinned by TestInvalidUTF8 below, not compared to stdlib:
		// encoding/json's substitution changed between Go 1.26 (� escape) and 1.27
		// (raw U+FFFD); it never reaches the wire (Postgres rejects it on write).
		float64(0), float64(1), float64(-1), 0.5, -0.5,
		float64(1e20), float64(1e21), float64(1e-6), float64(1e-7),
		float64(1 << 53), float64(-(1 << 53)), math.SmallestNonzeroFloat64, math.MaxFloat64,
		float64(3), float64(120), 1.5e-8, 2.5e22,
		int(0), int(-7), int64(1 << 40),
		[]any{}, map[string]any{}, []any(nil), map[string]any(nil),
		json.RawMessage(`{"a":  1}`), json.RawMessage(`[1,2 ,3]`),
	} {
		same(t, v)
	}
}

// Invalid UTF-8 never reaches the wire (Postgres rejects it on write; see
// transcode_test.go), so stdlib is not the oracle for it here — its substitution
// changed between Go 1.26 (� escape) and 1.27 (raw U+FFFD). The encoder still has
// to produce something stable across Go versions: each invalid byte becomes a raw
// U+FFFD, matching TS JSON.stringify.
func TestInvalidUTF8(t *testing.T) {
	got, err := Marshal("\xff\xfe x")
	if err != nil {
		t.Fatal(err)
	}
	if want := "\"�� x\""; string(got) != want {
		t.Errorf("Marshal(invalid utf8) = %s, want %s", got, want)
	}
}

func TestMatchesStdlibOnNesting(t *testing.T) {
	same(t, map[string]any{
		"kind": "workload", "version": float64(1), "tags": map[string]any{},
		"spec": map[string]any{
			"containers": []any{map[string]any{
				"name": "main", "image": "nginx:1.25", "ports": []any{map[string]any{"number": float64(8080)}},
			}},
			"defaultOptions": map[string]any{"capacityAI": false, "timeoutSeconds": float64(5)},
		},
		"links": []any{map[string]any{"rel": "self", "href": "/org/o/gvc/g?a=1&b=2"}},
		"null":  nil,
	})
	same(t, []map[string]any{{"a": float64(1)}, {"b": "two"}})
	same(t, []string{"a", "b<c"})
}

// Key ORDER is part of the bytes. encoding/json sorts map keys; so must this, or two
// runs of the same service disagree with each other.
func TestKeyOrder(t *testing.T) {
	m := map[string]any{}
	for _, k := range []string{"z", "a", "M", "0", "_", "aa", "A"} {
		m[k] = float64(1)
	}
	for i := 0; i < 50; i++ { // map iteration order varies per run
		same(t, m)
	}
}

// A type the encoder does not model must go to encoding/json, not be guessed at.
func TestUnmodeledTypesFallBack(t *testing.T) {
	type row struct {
		A string `json:"a"`
		B []int  `json:"b"`
	}
	for _, v := range []any{
		row{"x", []int{1, 2}},
		map[string]any{"row": row{"y", nil}, "ptr": &row{"z", []int{3}}},
		[]any{row{"q", nil}},
		map[string]any{"i8": int8(3), "u": uint(4), "f32": float32(1.5)},
	} {
		same(t, v)
	}
}

func TestRandomTrees(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	for i := 0; i < 3000; i++ {
		same(t, randValue(rng, 0))
	}
}

var nastyStrings = []string{
	"", "a", "&", "<", ">", "\"", "\\", "/", "\b", "\f", "\n", "\r", "\t",
	"\x00", "\x1f", " ", " ", "é", "日本語", "\U0001f642",
	"mixed <b>&\"'</b>\n\ttail", "/org/o/gvc/g?a=1&b=2#frag",
}

func randValue(rng *rand.Rand, depth int) any {
	kind := rng.Intn(9)
	if depth > 3 && kind > 5 {
		kind = rng.Intn(6)
	}
	switch kind {
	case 0:
		return nil
	case 1:
		return rng.Intn(2) == 0
	case 2:
		return nastyStrings[rng.Intn(len(nastyStrings))]
	case 3:
		return float64(rng.Intn(1000))
	case 4:
		return rng.NormFloat64() * math.Pow(10, float64(rng.Intn(46)-23))
	case 5:
		return float64(rng.Int63())
	case 6:
		n := rng.Intn(4)
		out := make([]any, n)
		for i := range out {
			out[i] = randValue(rng, depth+1)
		}
		return out
	case 7:
		n := rng.Intn(5)
		out := make(map[string]any, n)
		for i := 0; i < n; i++ {
			out[nastyStrings[rng.Intn(len(nastyStrings))]] = randValue(rng, depth+1)
		}
		return out
	default:
		return json.RawMessage(`{"raw": [1, 2, {"a": "b"} ]}`)
	}
}

func BenchmarkPage(b *testing.B) {
	page := benchPage()
	b.Run("wirejson", func(b *testing.B) {
		b.ReportAllocs()
		buf := make([]byte, 0, 1<<16)
		for i := 0; i < b.N; i++ {
			var err error
			if buf, err = Append(buf[:0], page); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("encoding_json", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			if _, err := stdlib(page); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func benchPage() map[string]any {
	items := make([]any, 50)
	for i := range items {
		items[i] = map[string]any{
			"kind": "workload", "id": "5cfc580b-20e9-4363-aaca-0f6a811c5860", "name": "w1",
			"description": "bench workload", "version": float64(1),
			"created": "2026-08-21T15:25:26.626Z", "lastModified": "2026-08-21T15:25:26.626Z",
			"tags": map[string]any{"pool": "bench", "even": "yes"},
			"links": []any{
				map[string]any{"rel": "self", "href": "/org/o/gvc/g/workload/w1"},
				map[string]any{"rel": "org", "href": "/org/o"},
			},
			"spec": map[string]any{
				"type": "standard",
				"containers": []any{map[string]any{
					"name": "main", "image": "nginx:1.25", "cpu": "50m", "memory": "128Mi",
					"inheritEnv": false,
					"ports":      []any{map[string]any{"number": float64(8080), "protocol": "http"}},
				}},
				"defaultOptions": map[string]any{
					"capacityAI": false, "debug": false, "suspend": false,
					"timeoutSeconds": float64(5),
					"autoscaling":    map[string]any{"metric": "cpu"},
				},
				"supportDynamicTags": false,
			},
			"status": map[string]any{"parentId": "938d72e6", "suspendedStatus": "notSuspended"},
		}
	}
	return map[string]any{"kind": "workloadList", "itemKind": "workload", "items": items}
}
