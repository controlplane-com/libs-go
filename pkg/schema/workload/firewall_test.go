package workload

import "testing"

// the expected hashes are what nodelibs/schema workload.ts firewallFingerprint computes for the same configs
func TestFirewallFingerprintMatchesTypeScript(t *testing.T) {
	if got := FirewallFingerprint(SandboxDefaultFirewall(), "", ""); got != "bzncc0mfyy0j" {
		t.Fatalf("sandbox default = %q", got)
	}
	session := map[string]any{"external": map[string]any{
		"outboundAllowHostname": []any{"api.example.com"},
		"inboundAllowCIDR":      []any{"0.0.0.0/0"},
		"inboundBlockedCIDR":    []any{},
	}}
	if got := FirewallFingerprint(session, "", ""); got != "4xz3vrn3hja2" {
		t.Fatalf("session = %q", got)
	}
	mixed := map[string]any{
		"external": map[string]any{
			"outboundAllowCIDR":     []any{"10.0.0.1", "1.2.3.0/24", "10.0.0.1/32"},
			"outboundAllowPort":     []any{map[string]any{"protocol": "https", "number": 443}, map[string]any{"protocol": "tcp", "number": 8080}},
			"outboundAllowHostname": []any{"b.com", "a.com"},
		},
		"internal": map[string]any{"inboundAllowType": "same-gvc"},
	}
	if got := FirewallFingerprint(mixed, "", ""); got != "hmgzjrh3dhyx" {
		t.Fatalf("mixed = %q", got)
	}
}

func TestFirewallFingerprintOfNoneOrInvalidIsEmpty(t *testing.T) {
	if got := FirewallFingerprint(nil, "", ""); got != "" {
		t.Fatalf("nil = %q", got)
	}
	bad := map[string]any{"external": map[string]any{"outboundAllowCIDR": []any{"nope"}}}
	if got := FirewallFingerprint(bad, "", ""); got != "" {
		t.Fatalf("invalid = %q", got)
	}
}

func TestSessionFirewallRefusedNamesTheFirstPlatformKey(t *testing.T) {
	cases := map[string]map[string]any{
		"external.inboundAllowCIDR": {"external": map[string]any{"inboundAllowCIDR": []any{}}},
		"external.http":             {"external": map[string]any{"outboundAllowHostname": []any{}, "http": map[string]any{}}},
		"internal":                  {"internal": map[string]any{}},
		"":                          {"external": map[string]any{"outboundAllowCIDR": []any{"1.2.3.4"}}},
	}
	for want, fw := range cases {
		if got := SessionFirewallRefused(fw); got != want {
			t.Fatalf("SessionFirewallRefused(%v) = %q, want %q", fw, got, want)
		}
	}
}
