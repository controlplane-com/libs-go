/* hand-maintained — canonical source of truth for the `permissions` catalog schema.
 * Was joi-to-go-generated from nodelibs/schema/src/permissions.ts. `validate:` tags
 * carry the rules; pkg/validate.FromType reflects over these types. permissions is a
 * read-only served resource (not a registered write kind). The name regex (was
 * pattern=permissionName) and the items sort({by:'name'}) custom (was permSortByName)
 * are carried by the PermissionName / PermissionItems nominal types in validate.go.
 * See permissions.ts for the Joi source.
 */

package permissions

import "github.com/controlplane-com/libs-go/pkg/schema/base"

// PermissionName is a permission name string; its ValidationRule carries the name
// regex (was pattern=permissionName). See validate.go.
type PermissionName string

type Permission struct {
	Name        PermissionName `json:"name" validate:"required,trim"`
	Description string         `json:"description,omitzero"`
}

type PermissionsKind string

const (
	PermissionsKindPermissions PermissionsKind = "permissions"
)

func (PermissionsKind) EnumValues() []string {
	return []string{string(PermissionsKindPermissions)}
}

type PermissionsTargetKind string

const (
	PermissionsTargetKindOrg                  PermissionsTargetKind = "org"
	PermissionsTargetKindCloudaccount         PermissionsTargetKind = "cloudaccount"
	PermissionsTargetKindPolicy               PermissionsTargetKind = "policy"
	PermissionsTargetKindUser                 PermissionsTargetKind = "user"
	PermissionsTargetKindGroup                PermissionsTargetKind = "group"
	PermissionsTargetKindResource             PermissionsTargetKind = "resource"
	PermissionsTargetKindTask                 PermissionsTargetKind = "task"
	PermissionsTargetKindPermissions          PermissionsTargetKind = "permissions"
	PermissionsTargetKindServiceaccount       PermissionsTargetKind = "serviceaccount"
	PermissionsTargetKindSecret               PermissionsTargetKind = "secret"
	PermissionsTargetKindLocation             PermissionsTargetKind = "location"
	PermissionsTargetKindGvc                  PermissionsTargetKind = "gvc"
	PermissionsTargetKindWorkload             PermissionsTargetKind = "workload"
	PermissionsTargetKindQuota                PermissionsTargetKind = "quota"
	PermissionsTargetKindIdentity             PermissionsTargetKind = "identity"
	PermissionsTargetKindDeployment           PermissionsTargetKind = "deployment"
	PermissionsTargetKindEvent                PermissionsTargetKind = "event"
	PermissionsTargetKindDomain               PermissionsTargetKind = "domain"
	PermissionsTargetKindImage                PermissionsTargetKind = "image"
	PermissionsTargetKindIpset                PermissionsTargetKind = "ipset"
	PermissionsTargetKindResourcepolicy       PermissionsTargetKind = "resourcepolicy"
	PermissionsTargetKindAgent                PermissionsTargetKind = "agent"
	PermissionsTargetKindAccessreport         PermissionsTargetKind = "accessreport"
	PermissionsTargetKindPolicymembership     PermissionsTargetKind = "policymembership"
	PermissionsTargetKindDbcluster            PermissionsTargetKind = "dbcluster"
	PermissionsTargetKindAuditctx             PermissionsTargetKind = "auditctx"
	PermissionsTargetKindMemcachecluster      PermissionsTargetKind = "memcachecluster"
	PermissionsTargetKindSpicedbcluster       PermissionsTargetKind = "spicedbcluster"
	PermissionsTargetKindTenant               PermissionsTargetKind = "tenant"
	PermissionsTargetKindMk8S                 PermissionsTargetKind = "mk8s"
	PermissionsTargetKindCommand              PermissionsTargetKind = "command"
	PermissionsTargetKindImagesummary         PermissionsTargetKind = "imagesummary"
	PermissionsTargetKindVolumeset            PermissionsTargetKind = "volumeset"
	PermissionsTargetKindSandboximage         PermissionsTargetKind = "sandboximage"
	PermissionsTargetKindCertificategenerator PermissionsTargetKind = "certificategenerator"
)

// EnumValues ports `targetKind: Kind.required()` (permissions.ts:19) — the shared
// Kind valid-list.
func (PermissionsTargetKind) EnumValues() []string {
	return []string{
		string(PermissionsTargetKindOrg), string(PermissionsTargetKindCloudaccount),
		string(PermissionsTargetKindPolicy), string(PermissionsTargetKindUser),
		string(PermissionsTargetKindGroup), string(PermissionsTargetKindResource),
		string(PermissionsTargetKindTask), string(PermissionsTargetKindPermissions),
		string(PermissionsTargetKindServiceaccount), string(PermissionsTargetKindSecret),
		string(PermissionsTargetKindLocation), string(PermissionsTargetKindGvc),
		string(PermissionsTargetKindWorkload), string(PermissionsTargetKindQuota),
		string(PermissionsTargetKindIdentity), string(PermissionsTargetKindDeployment),
		string(PermissionsTargetKindEvent), string(PermissionsTargetKindDomain),
		string(PermissionsTargetKindImage), string(PermissionsTargetKindIpset),
		string(PermissionsTargetKindResourcepolicy), string(PermissionsTargetKindAgent),
		string(PermissionsTargetKindAccessreport), string(PermissionsTargetKindPolicymembership),
		string(PermissionsTargetKindDbcluster), string(PermissionsTargetKindAuditctx),
		string(PermissionsTargetKindMemcachecluster), string(PermissionsTargetKindSpicedbcluster),
		string(PermissionsTargetKindTenant), string(PermissionsTargetKindMk8S),
		string(PermissionsTargetKindCommand), string(PermissionsTargetKindImagesummary),
		string(PermissionsTargetKindVolumeset),
	}
}

type PermissionsImplied map[string]any

// PermissionItems is the permission-list collection; its ValidationRule carries the
// sort({by:'name'}) custom (was custom=permSortByName). See validate.go.
type PermissionItems []Permission

type Permissions struct {
	Links      base.Links            `json:"links,omitzero"`
	Kind       PermissionsKind       `json:"kind,omitzero" validate:"required"`
	TargetKind PermissionsTargetKind `json:"targetKind,omitzero" validate:"required"`
	Items      PermissionItems       `json:"items" validate:"min=1,required,uniqueBy=name"`
	Implied    PermissionsImplied    `json:"implied,omitzero"`
}
