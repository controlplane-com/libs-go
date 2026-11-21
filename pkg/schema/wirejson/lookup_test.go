package wirejson

import (
	"bufio"
	"encoding/json"
	"math/rand"
	"os"
	"reflect"
	"strings"
	"testing"
)

// mapWalk is the implementation Lookup replaces: parse the whole document, then index
// maps segment by segment. Every answer Lookup gives must be this answer.
func mapWalk(data []byte, path []string) (any, bool) {
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, false
	}
	if len(path) == 0 {
		return v, true
	}
	cur := v
	for _, seg := range path {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil, false
		}
		x, present := m[seg]
		if !present {
			return nil, false
		}
		cur = x
	}
	return cur, true
}

// everyPath enumerates every dot-path in a decoded document, plus a few that are not
// there, so absence is compared as carefully as presence.
func everyPath(v any, prefix []string, out *[][]string) {
	*out = append(*out, append([]string(nil), prefix...))
	m, ok := v.(map[string]any)
	if !ok {
		return
	}
	for k, x := range m {
		everyPath(x, append(prefix, k), out)
	}
}

func compareAllPaths(t *testing.T, label string, raw []byte) {
	t.Helper()
	var doc any
	if err := json.Unmarshal(raw, &doc); err != nil {
		return // not our document to judge
	}
	var paths [][]string
	everyPath(doc, nil, &paths)
	paths = append(paths, []string{"nosuchkey"}, []string{"spec", "nosuchkey"}, []string{"a", "b", "c"})
	for _, p := range paths {
		gotV, gotOK := Lookup(raw, p)
		wantV, wantOK := mapWalk(raw, p)
		if gotOK != wantOK {
			t.Fatalf("%s: path %v found=%v, map walk found=%v\n  doc: %s", label, p, gotOK, wantOK, truncateDoc(raw))
		}
		if gotOK && !reflect.DeepEqual(gotV, wantV) {
			t.Fatalf("%s: path %v\n  lookup: %#v\n  walk:   %#v", label, p, gotV, wantV)
		}
	}
}

func truncateDoc(b []byte) string {
	if len(b) > 200 {
		return string(b[:200]) + "…"
	}
	return string(b)
}

// The stored corpus is the real thing: every distinct spec/status/tags shape in the
// diff and IT databases.
func TestLookupMatchesTheMapWalkOnStoredJson(t *testing.T) {
	f, err := os.Open("../handlers/testdata/stored_jsonb.tsv")
	if err != nil {
		t.Skipf("corpus unavailable: %v", err)
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 1<<20), 1<<22)
	rows := 0
	for sc.Scan() {
		parts := strings.SplitN(sc.Text(), "\t", 3)
		if len(parts) != 3 {
			continue
		}
		rows++
		compareAllPaths(t, parts[0]+"."+parts[1], []byte(parts[2]))
	}
	if rows < 100 {
		t.Fatalf("corpus shrank to %d rows", rows)
	}
}

func TestLookupMatchesTheMapWalkOnRandomJson(t *testing.T) {
	r := rand.New(rand.NewSource(20260902))
	for i := 0; i < 4000; i++ {
		v := randValue(r, 0)
		b, err := json.Marshal(v)
		if err != nil {
			continue
		}
		compareAllPaths(t, "random", b)
	}
}

// Duplicate keys are last-wins in encoding/json, and Lookup has to agree — it scans the
// whole object rather than stopping at the first hit precisely for this.
func TestLookupTakesTheLastDuplicateKey(t *testing.T) {
	raw := []byte(`{"a":1,"b":{"c":2},"a":{"c":9},"a":3}`)
	got, ok := Lookup(raw, []string{"a"})
	want, wantOK := mapWalk(raw, []string{"a"})
	if !ok || !wantOK || !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v (%v), want %#v (%v)", got, ok, want, wantOK)
	}
}

// A brace or bracket inside a string must not unbalance the skipper.
func TestLookupSkipsStringsWithBraces(t *testing.T) {
	raw := []byte(`{"a":"}{[]\"","b":{"c":"]}"},"target":42}`)
	got, ok := Lookup(raw, []string{"target"})
	if !ok || got != float64(42) {
		t.Fatalf("got %#v (%v), want 42", got, ok)
	}
}

func TestLookupRejectsMalformedSiblings(t *testing.T) {
	if _, ok := Lookup([]byte(`{"a":01,"target":1}`), []string{"target"}); ok {
		t.Fatal("a malformed sibling number must not be skipped over silently")
	}
}
