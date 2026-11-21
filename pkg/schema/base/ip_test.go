package base

import (
	"testing"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// TestIPRuleAcceptSet pins the accept set to Joi's, which is the RFC-3986 grammar
// @hapi/address compiles — NOT net.ParseIP. The two disagree at both edges: ParseIP
// takes an IPv4-mapped IPv6 literal for an IPv4 address and rejects leading-zero
// octets, and both cases were live divergences (ipset status.ipAddresses[].ip).
func TestIPRuleAcceptSet(t *testing.T) {
	cases := []struct {
		versions []string
		cidr     string
		value    string
		ok       bool
	}{
		{[]string{"ipv4"}, "forbidden", "1.2.3.4", true},
		{[]string{"ipv4"}, "forbidden", "01.2.3.4", true},  // leading zeros: Joi accepts
		{[]string{"ipv4"}, "forbidden", "1.2.3.04", true},  // …in any octet
		{[]string{"ipv4"}, "forbidden", "001.2.3.4", true}, // …up to three digits
		{[]string{"ipv4"}, "forbidden", "0001.2.3.4", false},
		{[]string{"ipv4"}, "forbidden", "::ffff:1.2.3.4", false}, // mapped v6 is NOT ipv4
		{[]string{"ipv4"}, "forbidden", "256.1.1.1", false},
		{[]string{"ipv4"}, "forbidden", "1.2.3", false},
		{[]string{"ipv4"}, "forbidden", "1.2.3.4/32", false}, // cidr forbidden
		{[]string{"ipv4"}, "optional", "1.2.3.4/32", true},
		{[]string{"ipv4"}, "optional", "1.2.3.4", true},
		{[]string{"ipv4"}, "optional", "1.2.3.4/33", false},
		{nil, "required", "10.0.0.0/8", true},
		{nil, "required", "1.2.3.4", false},
		{nil, "required", "2001:db8::/64", true},
		{nil, "required", "2001:db8::/129", false},
		{[]string{"ipv4", "ipv6"}, "required", "2001:db8::1", false}, // mask required
		{[]string{"ipv4", "ipv6"}, "required", "1.2.3.4/32", true},
	}
	for _, c := range cases {
		r := IPRule(c.versions, c.cidr)
		_, err := r.ValidateAt(nil, c.value, nil, validate.Prefs{})
		if (err == nil) != c.ok {
			t.Errorf("IPRule(%v,%q) on %q: ok=%v, err=%v", c.versions, c.cidr, c.value, c.ok, err)
		}
	}
}

// TestIPRuleError pins Joi's error: string.ipVersion when a version list is present
// (string.ip otherwise), the LABEL in the message, and the cidr/version context.
func TestIPRuleError(t *testing.T) {
	_, err := IPRule([]string{"ipv4"}, "forbidden").ValidateAt([]string{"status", "ip"}, "1.2.3.4/32", nil, validate.Prefs{})
	ve, ok := err.(*validate.Error)
	if !ok {
		t.Fatalf("want *validate.Error, got %T", err)
	}
	d := ve.Details()[0]
	if d.Type != "string.ipVersion" {
		t.Errorf("type = %q, want string.ipVersion", d.Type)
	}
	want := `"status.ip" must be a valid ip address of one of the following versions [ipv4] with a forbidden CIDR`
	if d.Message != want {
		t.Errorf("message = %q, want %q", d.Message, want)
	}
	if d.Context["cidr"] != "forbidden" {
		t.Errorf("context.cidr = %v", d.Context["cidr"])
	}
	if got, _ := d.Context["version"].([]string); len(got) != 1 || got[0] != "ipv4" {
		t.Errorf("context.version = %v", d.Context["version"])
	}
	if d.Context["value"] != "1.2.3.4/32" {
		t.Errorf("context.value = %v", d.Context["value"])
	}

	_, err = IPRule(nil, "required").ValidateAt([]string{"ipRanges"}, "1.2.3.4", nil, validate.Prefs{})
	d = err.(*validate.Error).Details()[0]
	if d.Type != "string.ip" {
		t.Errorf("type = %q, want string.ip (no version list)", d.Type)
	}
	if d.Message != `"ipRanges" must be a valid ip address with a required CIDR` {
		t.Errorf("message = %q", d.Message)
	}
}
