package policyMembership

// validate.go attaches validation to the canonical policyMembership types via the
// Validator interface — the go-libs home of what used to be
// pkg/handlers/schema_policymembership.go (validate.FromType) plus the
// RegisterReflect("policymembership", ...) wiring in schema_reflect_iam.go.
//
// policyMembership.ts reflects PolicyMembership directly with no createSchema fork,
// no custom/pattern tokens, and no object-level refines: PolicySummary is a plain
// nested struct, and base.Kind / base.Links carry their own attached Validators. So
// the rule is exactly validate.Reflect over the canonical struct.

import "github.com/controlplane-com/libs-go/pkg/schema/validate"

// ValidationRule reflects PolicyMembership. This is the schema the "policymembership"
// kind validates against.
func (PolicyMembership) ValidationRule() *validate.Rule {
	return validate.Reflect(PolicyMembership{})
}
