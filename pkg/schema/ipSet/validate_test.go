package ipSet

import (
	"testing"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// TestIpv4NoCidr pins ipSet.ts:7 `ip({version:'ipv4', cidr:'forbidden'})` to Joi's
// accept set. The port hand-rolled it as net.ParseIP + a dot count, which accepted the
// IPv4-MAPPED IPv6 literal ::ffff:1.2.3.4 (stored verbatim as an "IPv4" address) and
// rejected the leading-zero octets Joi allows.
func TestIpv4NoCidr(t *testing.T) {
	r := Ipv4NoCidr("").ValidationRule()
	cases := map[string]bool{
		"1.2.3.4":         true,
		"01.2.3.4":        true,
		"1.2.3.04":        true,
		"255.255.255.255": true,
		"::ffff:1.2.3.4":  false,
		"1.2.3.4/32":      false,
		"256.1.1.1":       false,
		"1.2.3":           false,
		"":                false,
	}
	for v, want := range cases {
		_, err := r.ValidateAt([]string{"status", "ipAddresses", validate.IndexPath(0), "ip"}, v, nil, validate.Prefs{})
		if (err == nil) != want {
			t.Errorf("ip %q: want ok=%v, err=%v", v, want, err)
		}
	}
	// The error names the full label, not a hardcoded "ip", and carries Joi's context.
	_, err := r.ValidateAt([]string{"status", "ipAddresses", validate.IndexPath(0), "ip"}, "1.2.3.4/32", nil, validate.Prefs{})
	d := err.(*validate.Error).Details()[0]
	want := `"status.ipAddresses[0].ip" must be a valid ip address of one of the following versions [ipv4] with a forbidden CIDR`
	if d.Message != want {
		t.Errorf("message = %q, want %q", d.Message, want)
	}
	if d.Context["cidr"] != "forbidden" || d.Context["value"] != "1.2.3.4/32" {
		t.Errorf("context = %v", d.Context)
	}
}
