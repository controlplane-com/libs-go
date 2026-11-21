package mk8s

import (
	"testing"

	"github.com/controlplane-com/libs-go/pkg/schema/mk8sCommon"
	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// TestFirewallSourceCIDRRequiresMask pins mk8s.ts:67-74's `cidr:'required'`. The port
// passed "optional", so a maskless IP was accepted — and the firewall reconciler parses
// these with net.ParseCIDR and silently DROPS a maskless entry, leaving the operator
// with an allow-list rule that does not exist.
func TestFirewallSourceCIDRRequiresMask(t *testing.T) {
	r := FirewallSourceCIDR("").ValidationRule()
	for _, ok := range []string{"1.2.3.4/32", "10.0.0.0/8", "2001:db8::/64"} {
		if _, err := r.ValidateAt([]string{"spec", "firewall", "0", "sourceCIDR"}, ok, nil, validate.Prefs{}); err != nil {
			t.Errorf("%q rejected: %v", ok, err)
		}
	}
	for _, bad := range []string{"1.2.3.4", "2001:db8::1", "nope", "1.2.3.4/33"} {
		_, err := r.ValidateAt([]string{"sourceCIDR"}, bad, nil, validate.Prefs{})
		if err == nil {
			t.Errorf("%q accepted without a mask", bad)
			continue
		}
		if d := err.(*validate.Error).Details()[0]; d.Type != "string.ipVersion" {
			t.Errorf("%q: type = %q, want string.ipVersion", bad, d.Type)
		}
	}
}

// TestMaxPodsBounds pins MaxPods (mk8sCommon.ts:68, master ec9ca5659): an optional
// per-nodepool integer in [16,250]. The rule is shared by every provider pool.
func TestMaxPodsBounds(t *testing.T) {
	r := mk8sCommon.MaxPodsRule()
	for _, ok := range []float64{16, 110, 250} {
		if _, err := r.ValidateAt([]string{"maxPods"}, ok, nil, validate.Prefs{}); err != nil {
			t.Errorf("%v rejected: %v", ok, err)
		}
	}
	for _, bad := range []float64{15, 251, 16.5} {
		if _, err := r.ValidateAt([]string{"maxPods"}, bad, nil, validate.Prefs{}); err == nil {
			t.Errorf("%v accepted", bad)
		}
	}
}
