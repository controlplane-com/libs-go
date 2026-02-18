/* auto-generated */

package base

type ApiError struct {
	Status  float32 `json:"status"`
	Message string  `json:"message,omitempty"`
	Code    string  `json:"code,omitempty"`
	Details any     `json:"details,omitempty"`
	Id      string  `json:"id,omitempty"`
}

type BaseTags map[string]any

type Base struct {
	Id           string   `json:"id,omitempty"`
	Name         Name     `json:"name,omitempty"`
	Kind         Kind     `json:"kind,omitempty"`
	Version      float32  `json:"version"`
	Description  string   `json:"description,omitempty"`
	Tags         BaseTags `json:"tags,omitempty"`
	Created      string   `json:"created,omitempty"`
	LastModified string   `json:"lastModified,omitempty"`
	Links        Links    `json:"links,omitempty"`
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
	KindOrg              Kind = "org"
	KindCloudaccount     Kind = "cloudaccount"
	KindPolicy           Kind = "policy"
	KindUser             Kind = "user"
	KindGroup            Kind = "group"
	KindResource         Kind = "resource"
	KindTask             Kind = "task"
	KindPermissions      Kind = "permissions"
	KindServiceaccount   Kind = "serviceaccount"
	KindSecret           Kind = "secret"
	KindLocation         Kind = "location"
	KindGvc              Kind = "gvc"
	KindWorkload         Kind = "workload"
	KindQuota            Kind = "quota"
	KindIdentity         Kind = "identity"
	KindDeployment       Kind = "deployment"
	KindEvent            Kind = "event"
	KindDomain           Kind = "domain"
	KindImage            Kind = "image"
	KindIpset            Kind = "ipset"
	KindResourcepolicy   Kind = "resourcepolicy"
	KindAgent            Kind = "agent"
	KindAccessreport     Kind = "accessreport"
	KindPolicymembership Kind = "policymembership"
	KindDbcluster        Kind = "dbcluster"
	KindAuditctx         Kind = "auditctx"
	KindMemcachecluster  Kind = "memcachecluster"
	KindSpicedbcluster   Kind = "spicedbcluster"
	KindTenant           Kind = "tenant"
	KindMk8S             Kind = "mk8s"
	KindCommand          Kind = "command"
	KindImagesummary     Kind = "imagesummary"
	KindVolumeset        Kind = "volumeset"
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
	ListItemKindOrg              ListItemKind = "org"
	ListItemKindCloudaccount     ListItemKind = "cloudaccount"
	ListItemKindPolicy           ListItemKind = "policy"
	ListItemKindUser             ListItemKind = "user"
	ListItemKindGroup            ListItemKind = "group"
	ListItemKindResource         ListItemKind = "resource"
	ListItemKindTask             ListItemKind = "task"
	ListItemKindPermissions      ListItemKind = "permissions"
	ListItemKindServiceaccount   ListItemKind = "serviceaccount"
	ListItemKindSecret           ListItemKind = "secret"
	ListItemKindLocation         ListItemKind = "location"
	ListItemKindGvc              ListItemKind = "gvc"
	ListItemKindWorkload         ListItemKind = "workload"
	ListItemKindQuota            ListItemKind = "quota"
	ListItemKindIdentity         ListItemKind = "identity"
	ListItemKindDeployment       ListItemKind = "deployment"
	ListItemKindEvent            ListItemKind = "event"
	ListItemKindAccount          ListItemKind = "account"
	ListItemKindDomain           ListItemKind = "domain"
	ListItemKindImage            ListItemKind = "image"
	ListItemKindIpset            ListItemKind = "ipset"
	ListItemKindResourcepolicy   ListItemKind = "resourcepolicy"
	ListItemKindAccessreport     ListItemKind = "accessreport"
	ListItemKindAgent            ListItemKind = "agent"
	ListItemKindAuditctx         ListItemKind = "auditctx"
	ListItemKindDbcluster        ListItemKind = "dbcluster"
	ListItemKindMemcachecluster  ListItemKind = "memcachecluster"
	ListItemKindSpicedbcluster   ListItemKind = "spicedbcluster"
	ListItemKindTenant           ListItemKind = "tenant"
	ListItemKindMk8S             ListItemKind = "mk8s"
	ListItemKindCommand          ListItemKind = "command"
	ListItemKindPolicymembership ListItemKind = "policymembership"
	ListItemKindPrimitive        ListItemKind = "primitive"
	ListItemKindForeign          ListItemKind = "foreign"
	ListItemKindLink             ListItemKind = "link"
	ListItemKindImagesummary     ListItemKind = "imagesummary"
	ListItemKindVolumeset        ListItemKind = "volumeset"
)

type List struct {
	Kind     ListKind     `json:"kind,omitempty"`
	ItemKind ListItemKind `json:"itemKind,omitempty"`
	Items    []any        `json:"items"`
	Links    []Link       `json:"links"`
}

type LocalLink string

type MultiZoneOptions struct {
	Enabled bool `json:"enabled,omitempty"`
}

type Name string

type Regex string

type Tags map[string]any
