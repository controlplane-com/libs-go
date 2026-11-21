package mk8s

// validate.go attaches validation to the canonical mk8s types via the Validator
// interface — the go-libs home of what used to be pkg/handlers/schema_mk8s.go's
// Mk8sSchema (FromType(Mk8sCluster)) plus the object-level hooks previously
// registered from pkg/handlers/schema_reflect_mk8s.go:
//   - the spec.provider .xor discrimination (exactly one of the twelve providers) —
//     an .Xor on the reflected Mk8SClusterSpecProvider;
//   - the allow-all firewall default (a content-ful array default a field tag cannot
//     carry) — a Refine on Mk8SClusterSpec;
//   - the firewall sourceCIDR ip custom — the strongly-typed FirewallSourceCIDR field.
//
// The twelve provider bodies, spec.addOns and status resolve to their own
// ValidationRule methods (in the sibling mk8s* packages) when reflection walks the
// struct — the strongly-typed replacement for the old base-type registry.
//
// NOTE: the `virtual` column (mk8sCluster.ts computeVirtual) is a server-managed
// query-only field handled entirely by the handler (ValidationStrip); it is not a
// field of Mk8sCluster and is deliberately absent from this schema. The version
// upgrade-path transition guard (mk8sUpgradePathSupported / patchAcceptable) is the
// handler's OnPatch/OnReplace, not part of the create schema — the schema only pins
// spec.version to the enum (Mk8SClusterSpecVersion.EnumValues).

import (
	"github.com/controlplane-com/libs-go/pkg/schema/mk8sCommon"
	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// ValidationRule ports mk8s.ts:67-74 `ip({version:['ipv4','ipv6'], cidr:'required'})`.
// The mask is REQUIRED: the firewall reconciler parses these with net.ParseCIDR and
// silently drops a maskless entry, so accepting one hands the operator an allow-list
// rule that does not exist.
func (FirewallSourceCIDR) ValidationRule() *validate.Rule {
	return mk8sCommon.IpStringRule([]string{"ipv4", "ipv6"}, "required")
}

// ValidationRule reflects Mk8sCluster.
func (Mk8sCluster) ValidationRule() *validate.Rule {
	return validate.Reflect(Mk8sCluster{})
}

// ValidationRule reflects Mk8SClusterSpec and injects the allow-all firewall default
// (mk8s.ts:80) when spec.firewall is absent — a content-ful array default a field tag
// cannot carry.
func (Mk8SClusterSpec) ValidationRule() *validate.Rule {
	return validate.RefineT(validate.Reflect(Mk8SClusterSpec{}), firewallDefaultRefine)
}

// ValidationRule reflects the twelve provider bodies and enforces the Mk8sSpec.provider
// .xor: exactly one provider key is present (mk8s.ts:92).
func (Mk8SClusterSpecProvider) ValidationRule() *validate.Rule {
	return validate.Reflect(Mk8SClusterSpecProvider{}).Xor(
		"generic", "hetzner", "aws", "linode", "oblivus", "lambdalabs",
		"paperspace", "ephemeral", "triton", "azure", "digitalocean", "gcp",
	)
}

// ValidationRule reflects spec.addOns (each add-on config resolves to its own
// ValidationRule in mk8sAddons).
func (Mk8SClusterSpecAddOns) ValidationRule() *validate.Rule {
	return validate.Reflect(Mk8SClusterSpecAddOns{})
}

// ValidationRule reflects status (mk8s.ts:143 Mk8sStatus); the status add-on bodies
// resolve to their own ValidationRule in mk8sAddons.
func (Mk8sStatus) ValidationRule() *validate.Rule {
	return validate.Reflect(Mk8sStatus{})
}

// firewallDefaultRefine injects the allow-all firewall default (mk8s.ts:80) when
// spec.firewall is absent.
func firewallDefaultRefine(spec *Mk8SClusterSpec, h *validate.Helper) error {
	if h.Present("firewall") {
		return nil
	}
	spec.Firewall = []Mk8SClusterSpecFirewall{{
		SourceCIDR:  "0.0.0.0/0",
		Description: "Default allow-all rule",
	}}
	return nil
}
