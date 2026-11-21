package base

import (
	"math"
	"testing"
)

// TestMemoryToMiB pins convertMemoryStringToNumber, captured from the TS function.
// The decimal units floor (master fd5da6590): 1G is 953 MiB, not 954, so a 1G request
// against a 1G limit cannot come out above it.
func TestMemoryToMiB(t *testing.T) {
	for _, c := range []struct {
		in   string
		want float64
	}{
		{"", 0},
		{"128Mi", 128},
		{"1Gi", 1024},
		{"0.5Gi", 512},
		{"1G", 953},
		{"2G", 1907},
		{"512M", 488},
		{"1000k", 0},
		{"2048k", 1},
		{"1024Ki", 1},
		{"1048576", 1},
		{"1.5Mi", 1}, // JS parseInt truncates
		{"1.5Gi", 1536},
		{"100Zi", 0}, // a unit the switch does not name
	} {
		if got := MemoryToMiB(c.in); got != c.want {
			t.Errorf("MemoryToMiB(%q) = %v, want %v", c.in, got, c.want)
		}
	}
	if got := MemoryToMiB("Gi"); !math.IsNaN(got) {
		t.Errorf("MemoryToMiB(\"Gi\") = %v, want NaN (JS parseInt)", got)
	}
}
