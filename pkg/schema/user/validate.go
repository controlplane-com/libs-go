package user

// validate.go attaches validation to the canonical user types via the Validator
// interface — the go-libs home of what used to be pkg/handlers/schema_user.go +
// the RegisterReflect("user", ...) wiring in schema_reflect_iam.go. user.ts is a
// standalone Joi.object() (a base-field subset, NOT Base.keys) with every key
// optional; the createSchema fork (id/name/idp/email required) lives in the user
// handler. User has no cross-field refines, customs, or patterns. InviteRequest/InviteResponse keep their hand-written
// pkg/handlers rules (auxiliary, not a registered write kind).

import "github.com/controlplane-com/libs-go/pkg/schema/validate"

// ValidationRule reflects User — the schema the "user" kind validates against.
func (User) ValidationRule() *validate.Rule {
	return validate.Reflect(User{})
}
