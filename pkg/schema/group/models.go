/* hand-maintained — canonical source of truth for the `group` resource schema.
 *
 * `validate:` struct tags drive pkg/validate.Reflect. See
 * nodelibs/schema/src/group.ts for the Joi source.
 *
 * Type-attached validation (validate.Validator) lives in validate.go:
 *   - Group.ValidationRule reflects this struct (the "group" kind schema);
 *   - MemberLink / MemberLinks: LocalLink.custom(toKind) + the sort/max/makeUnique
 *     member-array normalization;
 *   - GroupIdentityMatcher.ValidationRule reflects + the JMESPath/JavaScript
 *     expression-parse refine (group.ts:20).
 * memberQuery resolves to query.Query's own Validator (foundation).
 */

package group

import "github.com/controlplane-com/libs-go/pkg/schema/base"
import "github.com/controlplane-com/libs-go/pkg/schema/query"

// MemberLink ports `LocalLink.custom(toKind('user','serviceaccount'))` — the element
// of a group's memberLinks array. Self-validating via validate.go.
type MemberLink string

// MemberLinks ports the memberLinks array rule
// `Joi.array().items(LocalLink.custom(toKind(...))).max(200).sort().custom(makeUnique())`
// (group.ts:10-13). Nominal slice type so the array-level max/sort/makeUnique carry as
// its own Validator (validate.go); element kind-check lives on MemberLink.
type MemberLinks []MemberLink

// GroupTags aliases base.Tags for the reflection base-type registry.
type GroupTags = base.Tags

type GroupIdentityMatcherLanguage string

const (
	GroupIdentityMatcherLanguageJmespath   GroupIdentityMatcherLanguage = "jmespath"
	GroupIdentityMatcherLanguageJavascript GroupIdentityMatcherLanguage = "javascript"
)

func (GroupIdentityMatcherLanguage) EnumValues() []string {
	return []string{
		string(GroupIdentityMatcherLanguageJmespath),
		string(GroupIdentityMatcherLanguageJavascript),
	}
}

// GroupIdentityMatcher's object-level expression-parse .custom (group.ts:20) is a
// Refine on its ValidationRule (validate.go).
type GroupIdentityMatcher struct {
	Expression string                       `json:"expression" validate:"required"`
	Language   GroupIdentityMatcherLanguage `json:"language,omitzero" validate:"default=jmespath,required"`
}

type GroupOrigin string

const (
	GroupOriginSynthetic GroupOrigin = "synthetic"
	GroupOriginDefault   GroupOrigin = "default"
	GroupOriginBuiltin   GroupOrigin = "builtin"
)

func (GroupOrigin) EnumValues() []string {
	return []string{
		string(GroupOriginSynthetic), string(GroupOriginDefault), string(GroupOriginBuiltin),
	}
}

type Group struct {
	Id              string                `json:"id,omitzero" validate:"uuid,readOnly"`
	Name            base.Name             `json:"name,omitzero"`
	Kind            base.Kind             `json:"kind,omitzero"`
	Version         *float64              `json:"version,omitzero" validate:"positive,integer,readOnly"`
	Description     string                `json:"description,omitzero" validate:"trim,max=250,allow="`
	Tags            GroupTags             `json:"tags,omitzero" validate:"default={}"`
	Created         string                `json:"created,omitzero" validate:"date,readOnly"`
	LastModified    string                `json:"lastModified,omitzero" validate:"date,readOnly"`
	Links           base.Links            `json:"links,omitzero" validate:"readOnly"`
	MemberLinks     MemberLinks           `json:"memberLinks,omitzero"`
	MemberQuery     *query.Query          `json:"memberQuery,omitzero"`
	IdentityMatcher *GroupIdentityMatcher `json:"identityMatcher,omitzero"`
	Origin          GroupOrigin           `json:"origin,omitzero" validate:"default=default,readOnly"`
}
