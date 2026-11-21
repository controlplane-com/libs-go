/* hand-maintained — canonical source of truth for the `policymembership` resource schema.
 *
 * These structs (formerly emitted by joi-to-go from nodelibs/schema/src/policyMembership.ts)
 * are now the CANONICAL definition. `validate:` struct tags carry the validation rules;
 * the type-attached Validator (validate.go) reflects over these types to build the
 * *validate.Rule tree that the engine runs — replacing the reflection registration in
 * pkg/handlers (schema_policymembership.go + schema_reflect_iam.go). See
 * nodelibs/schema/src/policyMembership.ts for the original Joi source, and
 * go-libs/schema/ipSet/models.go for the tag DSL grammar.
 *
 * All fields are optional (the Joi source declares no `.required()`); base.Kind and
 * base.Links resolve to their own attached Validators, and PolicySummary is reflected
 * as a nested struct. No object-level refines or custom tokens apply.
 */

package policyMembership

import "github.com/controlplane-com/libs-go/pkg/schema/base"

type PolicyMembership struct {
	Kind     base.Kind       `json:"kind,omitzero"`
	Policies []PolicySummary `json:"policies,omitzero"`
	Links    base.Links      `json:"links,omitzero"`
}

type PolicySummary struct {
	Link        string    `json:"link,omitzero"`
	Description string    `json:"description,omitzero"`
	TargetKind  base.Kind `json:"targetKind,omitzero"`
}
