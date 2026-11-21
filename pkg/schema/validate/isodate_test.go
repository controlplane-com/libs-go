package validate

import "testing"

func TestISODateRendersAsJSToISOString(t *testing.T) {
	cases := map[string]string{
		"2026-11-19T10:00:00+01:00": "2026-11-19T09:00:00.000Z",
		"2026-11-19T09:00:00.5Z":    "2026-11-19T09:00:00.500Z",
		"2026-11-19T09:00:00.123Z":  "2026-11-19T09:00:00.123Z",
		"2026-11-19":                "2026-11-19T00:00:00.000Z",
	}
	for in, want := range cases {
		got, ok := ISODate(in)
		if !ok || got != want {
			t.Fatalf("ISODate(%q) = %q, %v; want %q", in, got, ok, want)
		}
	}
	if _, ok := ISODate("not a date"); ok {
		t.Fatal("a non-date is not an ISO date")
	}
}
