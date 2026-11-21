/* hand-maintained — canonical source of truth for the `resourcepolicy` resource.
 * Was joi-to-go-generated from nodelibs/schema/src/resourcePolicy.ts. `validate:`
 * tags carry the rules; pkg/validate.FromType reflects over these types. Unlike
 * policy, principalLinks are plain strings (no LocalLink/toKind) and the base-like
 * fields are NOT readOnly (they mirror the standalone Joi object, not Base). See
 * resourcePolicy.ts for the Joi source.
 */

package resourcePolicy

import "github.com/controlplane-com/libs-go/pkg/schema/base"

// Permission is a binding permission string (`Joi.string().trim()`). Self-validating
// via validate.go (was RegisterType(Permission, String().Trim())).
type Permission string

// Permissions ports the bindings `permissions` array rule
// `Joi.array().items(Permission).sort().custom(makeUnique()).required()` — the array-
// level sort/makeUnique/required live on this nominal type's Validator (validate.go)
// so the exact Joi op ORDER is preserved; the element trim lives on Permission.
type Permissions []Permission

// PrincipalLinks ports the bindings `principalLinks` array rule
// `Joi.array().min(1).max(50).items(Joi.string()).sort().custom(makeUnique())` — plain
// strings (unlike policy, NO LocalLink/toKind). Array-level rules live on this nominal
// type's Validator (validate.go) to preserve op order.
type PrincipalLinks []string

type ResourcePolicyBindings struct {
	Permissions    Permissions    `json:"permissions"`
	PrincipalLinks PrincipalLinks `json:"principalLinks,omitzero"`
}

type ResourcePolicy struct {
	Id           string     `json:"id,omitzero" validate:"uuid"`
	Kind         base.Kind  `json:"kind,omitzero"`
	Version      *float64   `json:"version,omitzero" validate:"positive"`
	Created      string     `json:"created,omitzero" validate:"date"`
	LastModified string     `json:"lastModified,omitzero" validate:"date"`
	Links        base.Links `json:"links,omitzero"`
	// makeUnique() with no `by` is _.uniq — a no-op on distinct object refs.
	Bindings []ResourcePolicyBindings `json:"bindings,omitzero" validate:"max=50"`
}
