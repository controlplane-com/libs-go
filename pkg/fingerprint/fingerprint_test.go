package fingerprint

import "testing"

// the expected values are what nodelibs/misc fingerprint.ts computes for the same value
func TestMatchesTheTypeScriptFingerprints(t *testing.T) {
	v := map[string]any{
		"spec":     map[string]any{"base": map[string]any{"image": "coke.registry.cpln.io/app:1"}, "tag": "latest"},
		"cplnTags": map[string]any{"cpln/restartedAt": "2026-09-17T00:00:00Z"},
		"z":        []any{3, 1.5, "<a&b>"},
		"a":        nil,
	}
	if got := Value(v); got != "YG2EDD610lZbcWtlmVtAxOCCGok" {
		t.Fatalf("Value = %s", got)
	}
	if got := ValueURL(v, 0); got != "c1pr831ypq95cpvhddjsjpt0rkg846m9" {
		t.Fatalf("ValueURL = %s", got)
	}
	if got := ValueURL(v, 12); got != "c1pr831ypq95" {
		t.Fatalf("ValueURL(12) = %s", got)
	}
}

func TestCanonicalSortsKeysAndKeepsHTML(t *testing.T) {
	type s struct {
		B string `json:"b"`
		A int    `json:"a"`
	}
	if got := string(Canonical(s{B: "<&>", A: 1})); got != `{"a":1,"b":"<&>"}` {
		t.Fatalf("Canonical = %s", got)
	}
}
