/* auto-generated */

package accessreport

import "github.com/controlplane-com/libs-go/pkg/schema/base"

type AccessReportKind string

const (
	AccessReportKindAccessreport AccessReportKind = "accessreport"
)

type AccessReport struct {
	Kind        AccessReportKind    `json:"kind,omitempty"`
	Permissions []GrantedPermission `json:"permissions,omitempty"`
	Created     string              `json:"created,omitempty"`
	Links       base.Links          `json:"links,omitempty"`
}

type GrantedPermission struct {
	Name        string             `json:"name,omitempty"`
	Description string             `json:"description,omitempty"`
	Bindings    []PrincipalBinding `json:"bindings,omitempty"`
}

type PrincipalBindingMatch string

const (
	PrincipalBindingMatchLink  PrincipalBindingMatch = "link"
	PrincipalBindingMatchQuery PrincipalBindingMatch = "query"
	PrincipalBindingMatchAll   PrincipalBindingMatch = "all"
)

type PrincipalBinding struct {
	PrincipalLink      string                `json:"principalLink,omitempty"`
	GrantingPolicyLink string                `json:"grantingPolicyLink,omitempty"`
	GrantedPermissions []string              `json:"grantedPermissions,omitempty"`
	Match              PrincipalBindingMatch `json:"match,omitempty"`
}
