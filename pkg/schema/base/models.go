/* auto-generated */

package base

type ApiError struct {
	Status  *float64 `json:"status,omitzero"`
	Message string   `json:"message,omitzero"`
	Code    string   `json:"code,omitzero"`
	Details any      `json:"details,omitzero"`
	Id      string   `json:"id,omitzero"`
}

type BaseTags map[string]any

type Base struct {
	Id           string   `json:"id,omitzero"`
	Name         Name     `json:"name,omitzero"`
	Kind         Kind     `json:"kind,omitzero"`
	Version      *float64 `json:"version,omitzero"`
	Description  string   `json:"description,omitzero"`
	Tags         BaseTags `json:"tags,omitzero"`
	Created      string   `json:"created,omitzero"`
	LastModified string   `json:"lastModified,omitzero"`
	Links        Links    `json:"links,omitzero"`
}

type CloudProvider string

const (
	CloudProviderAws   CloudProvider = "aws"
	CloudProviderGcp   CloudProvider = "gcp"
	CloudProviderAzure CloudProvider = "azure"
	CloudProviderNgs   CloudProvider = "ngs"
)

type ImageLink string

type Kind string

const (
	KindOrg                  Kind = "org"
	KindCloudaccount         Kind = "cloudaccount"
	KindPolicy               Kind = "policy"
	KindUser                 Kind = "user"
	KindGroup                Kind = "group"
	KindResource             Kind = "resource"
	KindTask                 Kind = "task"
	KindPermissions          Kind = "permissions"
	KindServiceaccount       Kind = "serviceaccount"
	KindSecret               Kind = "secret"
	KindLocation             Kind = "location"
	KindGvc                  Kind = "gvc"
	KindWorkload             Kind = "workload"
	KindQuota                Kind = "quota"
	KindIdentity             Kind = "identity"
	KindDeployment           Kind = "deployment"
	KindEvent                Kind = "event"
	KindDomain               Kind = "domain"
	KindImage                Kind = "image"
	KindIpset                Kind = "ipset"
	KindResourcepolicy       Kind = "resourcepolicy"
	KindAgent                Kind = "agent"
	KindAccessreport         Kind = "accessreport"
	KindPolicymembership     Kind = "policymembership"
	KindDbcluster            Kind = "dbcluster"
	KindAuditctx             Kind = "auditctx"
	KindMemcachecluster      Kind = "memcachecluster"
	KindSpicedbcluster       Kind = "spicedbcluster"
	KindTenant               Kind = "tenant"
	KindMk8S                 Kind = "mk8s"
	KindCommand              Kind = "command"
	KindImagesummary         Kind = "imagesummary"
	KindVolumeset            Kind = "volumeset"
	KindSandboximage         Kind = "sandboximage"
	KindCertificategenerator Kind = "certificategenerator"
)

type Link struct {
	Rel  string `json:"rel"`
	Href string `json:"href"`
}

type Links []Link

type ListKind string

const (
	ListKindList ListKind = "list"
)

type ListItemKind string

const (
	ListItemKindOrg                  ListItemKind = "org"
	ListItemKindCloudaccount         ListItemKind = "cloudaccount"
	ListItemKindPolicy               ListItemKind = "policy"
	ListItemKindUser                 ListItemKind = "user"
	ListItemKindGroup                ListItemKind = "group"
	ListItemKindResource             ListItemKind = "resource"
	ListItemKindTask                 ListItemKind = "task"
	ListItemKindPermissions          ListItemKind = "permissions"
	ListItemKindServiceaccount       ListItemKind = "serviceaccount"
	ListItemKindSecret               ListItemKind = "secret"
	ListItemKindLocation             ListItemKind = "location"
	ListItemKindGvc                  ListItemKind = "gvc"
	ListItemKindWorkload             ListItemKind = "workload"
	ListItemKindQuota                ListItemKind = "quota"
	ListItemKindIdentity             ListItemKind = "identity"
	ListItemKindDeployment           ListItemKind = "deployment"
	ListItemKindEvent                ListItemKind = "event"
	ListItemKindAccount              ListItemKind = "account"
	ListItemKindDomain               ListItemKind = "domain"
	ListItemKindImage                ListItemKind = "image"
	ListItemKindIpset                ListItemKind = "ipset"
	ListItemKindResourcepolicy       ListItemKind = "resourcepolicy"
	ListItemKindAccessreport         ListItemKind = "accessreport"
	ListItemKindAgent                ListItemKind = "agent"
	ListItemKindAuditctx             ListItemKind = "auditctx"
	ListItemKindDbcluster            ListItemKind = "dbcluster"
	ListItemKindMemcachecluster      ListItemKind = "memcachecluster"
	ListItemKindSpicedbcluster       ListItemKind = "spicedbcluster"
	ListItemKindTenant               ListItemKind = "tenant"
	ListItemKindMk8S                 ListItemKind = "mk8s"
	ListItemKindCommand              ListItemKind = "command"
	ListItemKindPolicymembership     ListItemKind = "policymembership"
	ListItemKindPrimitive            ListItemKind = "primitive"
	ListItemKindForeign              ListItemKind = "foreign"
	ListItemKindLink                 ListItemKind = "link"
	ListItemKindImagesummary         ListItemKind = "imagesummary"
	ListItemKindVolumeset            ListItemKind = "volumeset"
	ListItemKindSandboximage         ListItemKind = "sandboximage"
	ListItemKindCertificategenerator ListItemKind = "certificategenerator"
)

type List struct {
	Kind     ListKind     `json:"kind,omitzero"`
	ItemKind ListItemKind `json:"itemKind,omitzero"`
	Items    []any        `json:"items"`
	Links    []Link       `json:"links"`
}

type LocalLink string

type MultiZoneOptions struct {
	Enabled *bool `json:"enabled,omitzero"`
}

type Name string

type Regex string

type Tags map[string]any
