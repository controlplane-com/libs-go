package math

import "testing"

func TestCountLeadingZeroDecimals(t *testing.T) {
	c := CountLeadingZeroDecimals(0.00004)
	if c != 4 {
		t.FailNow()
	}

	c = CountLeadingZeroDecimals(1.)
	if c != 0 {
		t.FailNow()
	}

	c = CountLeadingZeroDecimals(1.0000)
	if c != 0 {
		t.FailNow()
	}

	c = CountLeadingZeroDecimals(.10000)
	if c != 0 {
		t.FailNow()
	}
}
