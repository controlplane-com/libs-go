/* hand-maintained — canonical source of truth for the `accessreport` resource.
 * Was joi-to-go-generated from nodelibs/schema/src/accessreport.ts. `validate:` tags
 * carry the rules; pkg/validate.FromType reflects over these types. See
 * accessreport.ts for the Joi source.
 */

package accessreport

import "github.com/controlplane-com/libs-go/pkg/schema/base"

type AccessReportKind string

const (
	AccessReportKindAccessreport AccessReportKind = "accessreport"
)

func (AccessReportKind) EnumValues() []string {
	return []string{string(AccessReportKindAccessreport)}
}

type AccessReport struct {
	Kind        AccessReportKind    `json:"kind,omitzero"`
	Permissions []GrantedPermission `json:"permissions,omitzero"`
	Created     string              `json:"created,omitzero" validate:"date"`
	Links       base.Links          `json:"links,omitzero"`
}

type GrantedPermission struct {
	Name        string             `json:"name,omitzero"`
	Description string             `json:"description,omitzero"`
	Bindings    []PrincipalBinding `json:"bindings,omitzero"`
}

type PrincipalBindingMatch string

const (
	PrincipalBindingMatchLink  PrincipalBindingMatch = "link"
	PrincipalBindingMatchQuery PrincipalBindingMatch = "query"
	PrincipalBindingMatchAll   PrincipalBindingMatch = "all"
)

func (PrincipalBindingMatch) EnumValues() []string {
	return []string{
		string(PrincipalBindingMatchLink),
		string(PrincipalBindingMatchQuery),
		string(PrincipalBindingMatchAll),
	}
}

type PrincipalBinding struct {
	PrincipalLink      string                `json:"principalLink,omitzero"`
	GrantingPolicyLink string                `json:"grantingPolicyLink,omitzero"`
	GrantedPermissions []string              `json:"grantedPermissions,omitzero"`
	Match              PrincipalBindingMatch `json:"match,omitzero"`
}
