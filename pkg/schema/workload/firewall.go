/* Firewall fingerprints and the session firewall (workload.ts firewallNormalized/firewallHash,
 * SessionFirewallSpec; sandbox.ts sandboxInboundDefault/sandboxDefaultFirewall). */

package workload

import (
	"encoding/json"
	"sort"
	"strings"
	"sync"

	"github.com/controlplane-com/libs-go/pkg/fingerprint"
	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

var firewallRule = sync.OnceValue(func() *validate.Rule { return validate.Reflect(WorkloadSpecFirewallConfig{}) })

// SandboxInboundDefault is the inbound part of the sandbox firewall default: open, since the platform reaches a sandbox through its endpoint.
func SandboxInboundDefault() map[string]any {
	return map[string]any{"inboundAllowCIDR": []any{"0.0.0.0/0"}, "inboundBlockedCIDR": []any{}}
}

// SandboxDefaultFirewall is the firewall a sandbox gets when none is given: inbound open to the platform, egress open.
func SandboxDefaultFirewall() map[string]any {
	external := SandboxInboundDefault()
	external["outboundAllowHostname"] = []any{}
	external["outboundAllowCIDR"] = []any{"0.0.0.0/0"}
	external["outboundBlockedCIDR"] = []any{}
	return map[string]any{"external": external}
}

// SessionFirewallForbidden are the keys a session firewall may not set, in the order the schema reports them.
var SessionFirewallForbidden = [][]string{
	{"external", "inboundAllowCIDR"},
	{"external", "inboundBlockedCIDR"},
	{"external", "http"},
	{"internal"},
}

// SessionFirewallRefused is the first forbidden key a session firewall sets, dotted; "" when none.
func SessionFirewallRefused(fw map[string]any) string {
	for _, path := range SessionFirewallForbidden {
		var cur any = fw
		found := true
		for _, seg := range path {
			m, ok := cur.(map[string]any)
			if !ok {
				found = false
				break
			}
			if cur, ok = m[seg]; !ok {
				found = false
				break
			}
		}
		if found {
			return strings.Join(path, ".")
		}
	}
	return ""
}

// FirewallNormalized validates a firewall config as the workload schema does and fingerprints the result; internal links validate in org/gvc.
func FirewallNormalized(cfg any, org, gvc string) (map[string]any, string, error) {
	obj, ok := generic(cfg).(map[string]any)
	if !ok {
		return nil, "", &validate.Error{Type: "object.base", Message: `"value" must be of type object`}
	}
	if err := firewallRule().ApplyAt(nil, obj, &validate.Context{Org: org, Gvc: gvc}, validate.Prefs{}); err != nil {
		return nil, "", err
	}
	value, _ := generic(obj).(map[string]any)
	return value, FirewallHash(value), nil
}

// FirewallFingerprint is the fingerprint of a firewall config; "" for none or one that does not validate.
func FirewallFingerprint(cfg any, org, gvc string) string {
	if cfg == nil {
		return ""
	}
	_, hash, err := FirewallNormalized(cfg, org, gvc)
	if err != nil {
		return ""
	}
	return hash
}

// FirewallHash fingerprints a validated config: maskless IPv4 as /32, lists sorted and deduplicated, empty lists and blocks dropped.
func FirewallHash(validated any) string {
	return fingerprint.ValueURL(withoutEmpties(sortedLists(canonicalCidrs(generic(validated), ""))), 12)
}

// generic is v as plain JSON values (map[string]any, []any, string, float64, bool, nil).
func generic(v any) any {
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	var out any
	if json.Unmarshal(b, &out) != nil {
		return nil
	}
	return out
}

func canonicalCidr(s string) string {
	if strings.Contains(s, "/") || strings.Contains(s, ":") {
		return s
	}
	return s + "/32"
}

func canonicalCidrs(v any, key string) any {
	switch t := v.(type) {
	case []any:
		out := make([]any, len(t))
		for i, x := range t {
			if s, ok := x.(string); ok && strings.HasSuffix(key, "CIDR") {
				out[i] = canonicalCidr(s)
			} else if strings.HasSuffix(key, "CIDR") {
				out[i] = x
			} else {
				out[i] = canonicalCidrs(x, "")
			}
		}
		return out
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, x := range t {
			out[k] = canonicalCidrs(x, k)
		}
		return out
	}
	return v
}

func sortedLists(v any) any {
	switch t := v.(type) {
	case []any:
		type keyed struct {
			key  string
			item any
		}
		items := make([]keyed, len(t))
		for i, x := range t {
			s := sortedLists(x)
			items[i] = keyed{fingerprint.Value(s), s}
		}
		sort.SliceStable(items, func(i, j int) bool { return items[i].key < items[j].key })
		out := make([]any, 0, len(items))
		for i, it := range items {
			if i == 0 || it.key != items[i-1].key {
				out = append(out, it.item)
			}
		}
		return out
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, x := range t {
			out[k] = sortedLists(x)
		}
		return out
	}
	return v
}

func withoutEmpties(v any) any {
	switch t := v.(type) {
	case []any:
		out := make([]any, len(t))
		for i, x := range t {
			out[i] = withoutEmpties(x)
		}
		return out
	case map[string]any:
		out := map[string]any{}
		for k, x := range t {
			kept := withoutEmpties(x)
			if a, ok := kept.([]any); ok && len(a) == 0 {
				continue
			}
			if m, ok := kept.(map[string]any); ok && len(m) == 0 {
				continue
			}
			out[k] = kept
		}
		return out
	}
	return v
}
