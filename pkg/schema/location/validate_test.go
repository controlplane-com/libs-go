package location

import (
	"testing"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// TestIpRangeRule pins status.ipRanges[] to Joi's ip({cidr:'required'}) — including
// the error type, which was a nonexistent "string.ipVersion" quoting the value.
func TestIpRangeRule(t *testing.T) {
	r := IpRange("").ValidationRule()
	if _, err := r.ValidateAt([]string{"status", "ipRanges", "0"}, "10.0.0.0/8", nil, validate.Prefs{}); err != nil {
		t.Fatalf("valid CIDR rejected: %v", err)
	}
	_, err := r.ValidateAt([]string{"status", "ipRanges"}, "1.2.3.4", nil, validate.Prefs{})
	if err == nil {
		t.Fatal("maskless IP accepted")
	}
	d := err.(*validate.Error).Details()[0]
	if d.Type != "string.ip" {
		t.Errorf("type = %q, want string.ip", d.Type)
	}
	if d.Message != `"status.ipRanges" must be a valid ip address with a required CIDR` {
		t.Errorf("message = %q", d.Message)
	}
	if d.Context["cidr"] != "required" {
		t.Errorf("context.cidr = %v", d.Context["cidr"])
	}
}
