package resourcePolicy

// validate.go attaches validation to the canonical resourcePolicy types via the
// Validator interface — the go-libs home of what used to be
// pkg/handlers/schema_resourcepolicy.go plus the resourcepolicy entries in
// pkg/handlers/schema_reflect_iam.go (RegisterType(Permission),
// RegisterReflect("resourcepolicy", ResourcePolicy)).
//
// resourcePolicy.ts reflects from ResourcePolicy; there is NO root object rule /
// refine on the reflected schema. Ported here:
//   - Permission: Joi.string().trim() (resourcePolicy.ts).
//   - Permissions: the bindings permissions array rule
//     .sort().custom(makeUnique()).required() — kept in the nominal slice type's rule
//     (not field tags) so the exact Joi op ORDER is preserved.
//   - PrincipalLinks: the bindings principalLinks array rule
//     .min(1).max(50).sort().custom(makeUnique()) — plain strings, NO LocalLink/toKind
//     (unlike policy's principalLinks).

import (
	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// ValidationRule reflects ResourcePolicy — the schema the "resourcepolicy" kind
// validates against (resourcePolicy.ts:22). No root refine.
func (ResourcePolicy) ValidationRule() *validate.Rule { return validate.Reflect(ResourcePolicy{}) }

// ValidationRule ports the Permission element rule `Joi.string().trim()`.
func (Permission) ValidationRule() *validate.Rule { return validate.String().Trim() }

// ValidationRule ports the bindings `permissions` array rule
// `Joi.array().items(Permission).sort().custom(makeUnique()).required()` with the same
// operation order (sort -> makeUnique -> required); the element trim comes from
// Permission and the array-level normalization lives here.
func (Permissions) ValidationRule() *validate.Rule {
	return validate.Array(Permission("").ValidationRule()).
		Sort().Dedupe().
		Required()
}

// ValidationRule ports the bindings `principalLinks` array rule
// `Joi.array().min(1).max(50).items(Joi.string()).sort().custom(makeUnique())` with the
// same operation order (min -> max -> sort -> makeUnique). Plain strings — resource
// policy principal links are NOT LocalLink/toKind checked (resourcePolicy.ts).
func (PrincipalLinks) ValidationRule() *validate.Rule {
	return validate.Array(validate.String()).
		Min(1).
		Max(50).
		Sort().Dedupe()
}
