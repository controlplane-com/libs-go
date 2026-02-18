/* auto-generated */

package permissions

import "github.com/controlplane-com/libs-go/pkg/schema/base"

type Permission struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type PermissionsKind string

const (
	PermissionsKindPermissions PermissionsKind = "permissions"
)

type PermissionsTargetKind string

const (
	PermissionsTargetKindOrg              PermissionsTargetKind = "org"
	PermissionsTargetKindCloudaccount     PermissionsTargetKind = "cloudaccount"
	PermissionsTargetKindPolicy           PermissionsTargetKind = "policy"
	PermissionsTargetKindUser             PermissionsTargetKind = "user"
	PermissionsTargetKindGroup            PermissionsTargetKind = "group"
	PermissionsTargetKindResource         PermissionsTargetKind = "resource"
	PermissionsTargetKindTask             PermissionsTargetKind = "task"
	PermissionsTargetKindPermissions      PermissionsTargetKind = "permissions"
	PermissionsTargetKindServiceaccount   PermissionsTargetKind = "serviceaccount"
	PermissionsTargetKindSecret           PermissionsTargetKind = "secret"
	PermissionsTargetKindLocation         PermissionsTargetKind = "location"
	PermissionsTargetKindGvc              PermissionsTargetKind = "gvc"
	PermissionsTargetKindWorkload         PermissionsTargetKind = "workload"
	PermissionsTargetKindQuota            PermissionsTargetKind = "quota"
	PermissionsTargetKindIdentity         PermissionsTargetKind = "identity"
	PermissionsTargetKindDeployment       PermissionsTargetKind = "deployment"
	PermissionsTargetKindEvent            PermissionsTargetKind = "event"
	PermissionsTargetKindDomain           PermissionsTargetKind = "domain"
	PermissionsTargetKindImage            PermissionsTargetKind = "image"
	PermissionsTargetKindIpset            PermissionsTargetKind = "ipset"
	PermissionsTargetKindResourcepolicy   PermissionsTargetKind = "resourcepolicy"
	PermissionsTargetKindAgent            PermissionsTargetKind = "agent"
	PermissionsTargetKindAccessreport     PermissionsTargetKind = "accessreport"
	PermissionsTargetKindPolicymembership PermissionsTargetKind = "policymembership"
	PermissionsTargetKindDbcluster        PermissionsTargetKind = "dbcluster"
	PermissionsTargetKindAuditctx         PermissionsTargetKind = "auditctx"
	PermissionsTargetKindMemcachecluster  PermissionsTargetKind = "memcachecluster"
	PermissionsTargetKindSpicedbcluster   PermissionsTargetKind = "spicedbcluster"
	PermissionsTargetKindTenant           PermissionsTargetKind = "tenant"
	PermissionsTargetKindMk8S             PermissionsTargetKind = "mk8s"
	PermissionsTargetKindCommand          PermissionsTargetKind = "command"
	PermissionsTargetKindImagesummary     PermissionsTargetKind = "imagesummary"
	PermissionsTargetKindVolumeset        PermissionsTargetKind = "volumeset"
)

type PermissionsImplied map[string]any

type Permissions struct {
	Links      base.Links            `json:"links,omitempty"`
	Kind       PermissionsKind       `json:"kind,omitempty"`
	TargetKind PermissionsTargetKind `json:"targetKind,omitempty"`
	Items      []Permission          `json:"items"`
	Implied    PermissionsImplied    `json:"implied,omitempty"`
}
