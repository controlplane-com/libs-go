package identity

// validate.go attaches validation to the canonical identity types via the
// Validator interface — the go-libs home of what used to be
// data-service/pkg/handlers/schema_identity.go + the identity block of
// schema_reflect_iam.go. The top Identity reflects from its tagged struct and
// layers the per-array Name/FQDN uniqueness custom; the nested aws/gcp/network
// object types carry their xor/oxor presence relations (and gcp its scopes
// content-default) on their own ValidationRule. The old registry leaf types
// (AwsRoleName/PolicyRef/GcpRoleName/NetworkName/Ipv4Addr/NetworkPort/NatsSubject)
// and the pattern=/custom= tag sites are now strongly-typed nominal types that
// carry their own rule. See nodelibs/schema/src/identity.ts for the Joi source.

import (
	"regexp"
	"strings"

	"github.com/controlplane-com/libs-go/pkg/schema/base"
	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// --- regexes (identity.ts) ---

var (
	awsRoleNameRe       = regexp.MustCompile(`^([a-zA-Z0-9/+=,.@_-])+$`)
	policyRefRe         = regexp.MustCompile(`^(aws::)?([a-zA-Z0-9/+=,.@_-])+$`)
	gcpRoleNameRe       = regexp.MustCompile(`^roles/([a-zA-Z0-9])+(\.([a-zA-Z0-9])+)?$`)
	gserviceaccountRe   = regexp.MustCompile(`^.*\.gserviceaccount\.com$`)
	ngsTTLRe            = regexp.MustCompile(`^[0-9]+(ms|s|m|h)$`)
	gcpServiceConnectRe = regexp.MustCompile(`^/?projects/(.+)/regions/(.+)/serviceAttachments/(.+)/?$`)
)

// --- named leaf types (element-level rules; formerly RegisterType leaves) ---

// ValidationRule ports `AwsRoleName` (identity.ts:6).
func (AwsRoleName) ValidationRule() *validate.Rule {
	return validate.String().Regex(awsRoleNameRe).Trim().Max(64)
}

// ValidationRule ports `PolicyRef` (identity.ts:11).
func (PolicyRef) ValidationRule() *validate.Rule {
	return validate.String().Regex(policyRefRe).Trim()
}

// ValidationRule ports `GcpRoleName` (identity.ts:15).
func (GcpRoleName) ValidationRule() *validate.Rule {
	return validate.String().Regex(gcpRoleNameRe)
}

// ValidationRule ports `NetworkName = Joi.alternatives(Name, Joi.string().domain())`.
func (NetworkName) ValidationRule() *validate.Rule {
	return validate.Alternatives(base.Name("").ValidationRule(), validate.String().Domain())
}

// ValidationRule ports `Ipv4Addr = Joi.string().ip({ version: 'ipv4' })` — an IPv4
// address with an optional CIDR suffix (Joi's default cidr is 'optional').
func (Ipv4Addr) ValidationRule() *validate.Rule {
	return base.IPRule([]string{"ipv4"}, "optional")
}

// ValidationRule ports `NetworkPort = Joi.number().integer().min(0).max(65535)`.
func (NetworkPort) ValidationRule() *validate.Rule {
	return validate.Number().Integer().Min(0).Max(65535)
}

// ValidationRule ports `NatsSubject = Joi.string().trim().custom(validSubject)`.
func (NatsSubject) ValidationRule() *validate.Rule {
	return validate.String().Trim().Custom(validSubjectCustom)
}

// --- pattern=/custom= tag sites, now strongly-typed nominal types ---

func (GserviceAccount) ValidationRule() *validate.Rule {
	return validate.String().Regex(gserviceaccountRe)
}

func (NgsTtl) ValidationRule() *validate.Rule {
	return validate.String().Regex(ngsTTLRe)
}

func (GcpTargetService) ValidationRule() *validate.Rule {
	return validate.String().Regex(gcpServiceConnectRe).Custom(func(value any, _ *validate.Helper) (any, error) {
		s, _ := value.(string)
		return strings.Trim(s, "/"), nil
	})
}

// --- nominal link types (the `localLink=<kind>` tag sites) ---
// LocalLink expanded then kind-checked (Joi LocalLink.custom(toKind('<kind>'))),
// formerly the localLink= tag resolved via the RegisterLocalLinkFactory.

func (CloudAccountLink) ValidationRule() *validate.Rule {
	return base.LocalLinkRule().Custom(base.ToKind("cloudaccount"))
}

func (AgentLink) ValidationRule() *validate.Rule {
	return base.LocalLinkRule().Custom(base.ToKind("agent"))
}

func (MemcacheClusterLink) ValidationRule() *validate.Rule {
	return base.LocalLinkRule().Custom(base.ToKind("memcachecluster"))
}

func (SpicedbClusterLink) ValidationRule() *validate.Rule {
	return base.LocalLinkRule().Custom(base.ToKind("spicedbcluster"))
}

// --- collection nominal slice types (the `custom=dedupeEmpty` array sites) ---
// dedupeEmpty = makeUnique() with no `by`: remove duplicate elements (mutating,
// NOT an error like Joi array.unique).

func (NatsSubjectList) ValidationRule() *validate.Rule {
	return validate.Array(NatsSubject("").ValidationRule()).Dedupe()
}

func (Ipv4AddrList) ValidationRule() *validate.Rule {
	return validate.Array(Ipv4Addr("").ValidationRule()).Dedupe()
}

func (NetworkPortList) ValidationRule() *validate.Rule {
	return validate.Array(NetworkPort(0).ValidationRule()).Dedupe()
}

// --- resource / nested object rules ---

// ValidationRule reflects Identity and layers the per-array Name/FQDN uniqueness
// custom (identity.ts:213).
func (Identity) ValidationRule() *validate.Rule {
	return validate.Reflect(Identity{}).CustomTyped(identityUniqueRefine)
}

// ValidationRule reflects AwsIdentity and layers its presence relations
// (identity.ts:24): exactly one of roleName/policyRefs, at most one of
// roleName/trustPolicy.
func (AwsIdentity) ValidationRule() *validate.Rule {
	return validate.Reflect(AwsIdentity{}).
		Xor("roleName", "policyRefs").
		Oxor("roleName", "trustPolicy")
}

// ValidationRule reflects GcpIdentity and layers its presence relation (exactly one
// of serviceAccount/bindings) plus the scopes content-default (identity.ts:43,46).
func (GcpIdentity) ValidationRule() *validate.Rule {
	return validate.RefineT(validate.Reflect(GcpIdentity{}).
		Xor("serviceAccount", "bindings"), gcpScopesRefine)
}

// ValidationRule reflects NetworkResource and layers its presence relation: exactly
// one of FQDN/IPs (identity.ts, NetworkResource).
func (NetworkResource) ValidationRule() *validate.Rule {
	return validate.Reflect(NetworkResource{}).Xor("FQDN", "IPs")
}

// ValidationRule reflects NativeNetworkResource and layers its presence relation:
// exactly one of awsPrivateLink/gcpServiceConnect (identity.ts:166).
func (NativeNetworkResource) ValidationRule() *validate.Rule {
	return validate.Reflect(NativeNetworkResource{}).Xor("awsPrivateLink", "gcpServiceConnect")
}

// --- irreducible customs / refines (the "complex parts of Joi") ---

// isValidSubject ports isValidSubject (identity.ts:55) — the NATS subject validity
// check borrowed from nats-server sublist.go.
func isValidSubject(subject string) bool {
	const fwc = '>'
	if subject == "" {
		return false
	}
	sfwc := false
	tokens := strings.Split(subject, ".")
	for _, t := range tokens {
		length := len(t)
		if length == 0 || sfwc {
			return false
		}
		if length > 1 {
			if strings.ContainsAny(t, "\t\n\f\r ") {
				return false
			}
			continue
		}
		switch t[0] {
		case fwc:
			sfwc = true
		case ' ', '\t', '\n', '\r', '\f':
			return false
		}
	}
	return true
}

// validSubjectCustom ports validSubject (identity.ts:96); backs the NatsSubject leaf.
func validSubjectCustom(value any, h *validate.Helper) (any, error) {
	s, _ := value.(string)
	if !isValidSubject(s) {
		return nil, h.Errorf("any.custom", "'%s' in not a valid NATS subject", s)
	}
	return s, nil
}

// identityUniqueRefine ports the Identity object .custom (identity.ts:213): no two
// (native)networkResources may share a Name or FQDN across the combined set.
func identityUniqueRefine(id any, h *validate.Helper) (any, error) {
	var nr []NetworkResource
	var nnr []NativeNetworkResource
	if p := validate.Sub[[]NetworkResource](id, "networkResources"); p != nil {
		nr = *p
	}
	if p := validate.Sub[[]NativeNetworkResource](id, "nativeNetworkResources"); p != nil {
		nnr = *p
	}

	identifier := "networkResources"
	if len(nnr) > 0 {
		identifier = "nativeNetworkResources"
		if len(nr) > 0 {
			identifier = "networkResources and nativeNetworkResources"
		}
	}

	known := map[string]bool{}
	type named struct{ name, fqdn string }
	all := make([]named, 0, len(nr)+len(nnr))
	for _, r := range nr {
		all = append(all, named{string(r.Name), r.FQDN})
	}
	for _, r := range nnr {
		all = append(all, named{string(r.Name), r.FQDN})
	}

	for _, n := range all {
		if fqdn := n.fqdn; fqdn != "" {
			if known[fqdn] {
				return nil, h.Errorf("any.custom", "%s use the same Name/FQDN: '%s'", identifier, fqdn)
			}
			known[fqdn] = true
		}
		if name := n.name; name != "" {
			if known[name] {
				return nil, h.Errorf("any.custom", "%s use the same Name/FQDN: '%s'", identifier, name)
			}
			known[name] = true
		}
	}
	return id, nil
}

// gcpScopesRefine ports `scopes: Joi.array().items(Joi.string().trim())
// .default(['https://www.googleapis.com/auth/cloud-platform']).optional()`
// (identity.ts:46): inject the content-ful default when absent, else trim each item.
func gcpScopesRefine(g *GcpIdentity, _ *validate.Helper) error {
	if g.Scopes == nil {
		g.Scopes = []string{"https://www.googleapis.com/auth/cloud-platform"}
		return nil
	}
	for i, s := range g.Scopes {
		g.Scopes[i] = strings.TrimSpace(s)
	}
	return nil
}
