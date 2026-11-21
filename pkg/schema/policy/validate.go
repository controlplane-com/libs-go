package policy

// validate.go attaches validation to the canonical policy types via the Validator
// interface — the go-libs home of what used to be pkg/handlers/schema_policy.go plus
// the policy entries in pkg/handlers/schema_reflect_iam.go (RegisterType(Permission),
// RegisterType(TargetLink), RegisterType(PrincipalLink), RegisterReflect("policy",
// Policy)).
//
// policy.ts reflects from Policy; there is NO root object rule on the reflected schema
// (data-service's createSchema requireds and .or() and the per-targetKind permission
// checks live in the data-service PolicyHandler).
//
// Ported here (policy.ts):
//   - Permission / TargetLink: Joi.string().trim() element leaf.
//   - PrincipalLink: LocalLink.custom(toKind('user','serviceaccount','group',
//     '!identity')).
//   - Permissions / PrincipalLinks / TargetLinks: the array-level rules
//     (items → [min] → [max] → sort → makeUnique, in Joi order) kept in the nominal
//     slice types' rules (not field tags) so the exact Joi rule ORDER is preserved.
//   - bindings: max(50) only — makeUnique() over distinct object refs is a JS no-op
//     (_.uniq by reference), so no dedupe (matches the builder; see models.go).
// targetQuery resolves to query.Query's own Validator (foundation).

import (
	"github.com/controlplane-com/libs-go/pkg/schema/base"
	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// ValidationRule reflects Policy (policy.ts: Base.keys({...})).
func (Policy) ValidationRule() *validate.Rule { return validate.Reflect(Policy{}) }

// ValidationRule ports `Joi.string().trim()` — a binding permission string.
func (Permission) ValidationRule() *validate.Rule { return validate.String().Trim() }

// ValidationRule ports `Joi.string().trim()` — a policy targetLinks entry.
func (TargetLink) ValidationRule() *validate.Rule { return validate.String().Trim() }

// ValidationRule ports `LocalLink.custom(toKind('user','serviceaccount','group',
// '!identity'))` (policy.ts): a local link that must resolve to a user, serviceaccount
// or group in the same org (and NOT an identity).
func (PrincipalLink) ValidationRule() *validate.Rule {
	return base.LocalLinkRule().Custom(base.ToKind("user", "serviceaccount", "group", "!identity"))
}

// ValidationRule ports `Joi.array().items(Joi.string().trim()).sort().custom(
// makeUnique())` (policy.ts Binding.permissions) with the same operation order
// (sort → makeUnique); the element rule comes from Permission and the array-level
// normalization lives here. `.required()` stays a field-presence tag on the Binding
// field.
func (Permissions) ValidationRule() *validate.Rule {
	return validate.Array(Permission("").ValidationRule()).
		Sort().
		Dedupe()
}

// ValidationRule ports `Joi.array().items(LocalLink.custom(toKind(...))).min(1).max(
// 200).sort().required().custom(makeUnique())` (policy.ts Binding.principalLinks) with
// the same operation order (min → max → sort → makeUnique); the element kind-check
// lives on PrincipalLink. `.required()` stays a field-presence tag on the Binding
// field.
func (PrincipalLinks) ValidationRule() *validate.Rule {
	return validate.Array(PrincipalLink("").ValidationRule()).
		Min(1).
		Max(200).
		Sort().
		Dedupe()
}

// ValidationRule ports `Joi.array().items(Joi.string().trim()).max(200).sort().custom(
// makeUnique())` (policy.ts Policy.targetLinks) with the same operation order
// (max → sort → makeUnique); the element rule comes from TargetLink.
func (TargetLinks) ValidationRule() *validate.Rule {
	return validate.Array(TargetLink("").ValidationRule()).
		Max(200).
		Sort().
		Dedupe()
}
