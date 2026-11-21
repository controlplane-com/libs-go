/* hand-maintained — canonical source of truth for the `policy` resource schema.
 *
 * `validate:` struct tags drive pkg/validate.Reflect. See
 * nodelibs/schema/src/policy.ts for the Joi source.
 *
 * Type-attached validation (validate.Validator) lives in validate.go:
 *   - Policy.ValidationRule reflects this struct (the "policy" kind schema);
 *   - Permission / TargetLink: Joi.string().trim() element leaves;
 *   - PrincipalLink: LocalLink.custom(toKind('user','serviceaccount','group',
 *     '!identity'));
 *   - Permissions / PrincipalLinks / TargetLinks: the array-level max/min/sort/
 *     makeUnique normalization (nominal slice types).
 * targetQuery resolves to query.Query's own Validator (foundation).
 */

package policy

import "github.com/controlplane-com/libs-go/pkg/schema/base"
import "github.com/controlplane-com/libs-go/pkg/schema/query"

// Permission is a binding permission string (`Joi.string().trim()`).
type Permission string

// PrincipalLink ports `LocalLink.custom(toKind('user','serviceaccount','group','!identity'))`.
type PrincipalLink string

// TargetLink ports `Joi.string().trim()` (a policy targetLinks entry).
type TargetLink string

// Permissions ports the Binding.permissions array rule
// `Joi.array().items(Joi.string().trim()).sort().custom(makeUnique())` (policy.ts).
// Nominal slice type so the array-level sort/makeUnique carry as its own Validator
// (validate.go); element trim lives on Permission. `.required()` is the field tag.
type Permissions []Permission

// PrincipalLinks ports the Binding.principalLinks array rule
// `Joi.array().items(LocalLink.custom(toKind(...))).min(1).max(200).sort().custom(
// makeUnique())` (policy.ts). Nominal slice type; element kind-check lives on
// PrincipalLink. `.required()` is the field tag.
type PrincipalLinks []PrincipalLink

// TargetLinks ports the Policy.targetLinks array rule
// `Joi.array().items(Joi.string().trim()).max(200).sort().custom(makeUnique())`
// (policy.ts). Nominal slice type; element trim lives on TargetLink.
type TargetLinks []TargetLink

type Binding struct {
	Permissions    Permissions    `json:"permissions" validate:"required"`
	PrincipalLinks PrincipalLinks `json:"principalLinks" validate:"required"`
}

// PolicyTags aliases base.Tags for the reflection base-type registry.
type PolicyTags = base.Tags

type PolicyTarget string

const (
	PolicyTargetAll PolicyTarget = "all"
)

func (PolicyTarget) EnumValues() []string { return []string{string(PolicyTargetAll)} }

type PolicyOrigin string

const (
	PolicyOriginDefault PolicyOrigin = "default"
	PolicyOriginBuiltin PolicyOrigin = "builtin"
)

func (PolicyOrigin) EnumValues() []string {
	return []string{string(PolicyOriginDefault), string(PolicyOriginBuiltin)}
}

type Policy struct {
	Id           string     `json:"id,omitzero" validate:"uuid,readOnly"`
	Name         base.Name  `json:"name,omitzero"`
	Kind         base.Kind  `json:"kind,omitzero"`
	Version      *float64   `json:"version,omitzero" validate:"positive,integer,readOnly"`
	Description  string     `json:"description,omitzero" validate:"trim,max=250,allow="`
	Tags         PolicyTags `json:"tags,omitzero" validate:"default={}"`
	Created      string     `json:"created,omitzero" validate:"date,readOnly"`
	LastModified string     `json:"lastModified,omitzero" validate:"date,readOnly"`
	Links        base.Links `json:"links,omitzero" validate:"readOnly"`
	// targetKind + origin are set on create but immutable after (policy.ts retainProps
	// ['targetKind','origin']). `unmodifiable` is enforced generically at merge (pinned
	// to the stored value) — see CRUD.retainProps.
	TargetKind  base.Kind    `json:"targetKind,omitzero" validate:"unmodifiable"`
	TargetLinks TargetLinks  `json:"targetLinks,omitzero"`
	TargetQuery *query.Query `json:"targetQuery,omitzero"`
	Target      PolicyTarget `json:"target,omitzero"`
	Origin      PolicyOrigin `json:"origin,omitzero" validate:"default=default,optional,unmodifiable"`
	// makeUnique() with no `by` is _.uniq — a no-op on distinct object refs, so no
	// dedupe custom on bindings (matches the builder).
	Bindings []Binding `json:"bindings,omitzero" validate:"max=50"`
}
