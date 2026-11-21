package ipSet

// validate.go attaches validation to the canonical ipSet types via the Validator
// interface — the go-libs home of what used to be the reflected `ipset` schema in
// data-service/pkg/handlers (schema_reflect.go: FromType(IpSet{}) + the
// ipv4NoCidr custom + the localLink factory). ipSet.ts reflects from IpSet
// (Base.keys({spec, status})). The old tag-based indirections are now strongly-typed
// field types:
//   - custom=ipv4NoCidr        -> Ipv4NoCidr
//   - localLink=location       -> LocationLink   (LocalLink.custom(toKind('location')))
//   - localLink=!workload|gvc  -> WorkloadOrGvcLink
//
// TS reference: nodelibs/schema/src/ipSet.ts.

import (
	"github.com/controlplane-com/libs-go/pkg/schema/base"
	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// ValidationRule reflects IpSet (Base.keys({spec: required, status})).
func (IpSet) ValidationRule() *validate.Rule {
	return validate.Reflect(IpSet{})
}

func (LocationLink) ValidationRule() *validate.Rule {
	return base.LocalLinkRule().Custom(base.ToKind("location"))
}

func (WorkloadOrGvcLink) ValidationRule() *validate.Rule {
	return base.LocalLinkRule().Custom(base.ToKind("!workload", "gvc"))
}

// ValidationRule ports ipSet.ts:7 `ip({version:'ipv4', cidr:'forbidden'})`. It used to
// be a net.ParseIP + dot-count heuristic, which accepted the IPv4-mapped IPv6 literal
// ::ffff:1.2.3.4 and rejected the leading-zero octets Joi's RFC-3986 grammar allows.
func (Ipv4NoCidr) ValidationRule() *validate.Rule {
	return base.IPRule([]string{"ipv4"}, "forbidden")
}
