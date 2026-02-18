/* auto-generated */

package tenant

import "github.com/controlplane-com/libs-go/pkg/schema/base"

type TenantTags map[string]any

type TenantSpec struct {
	Orgs     []base.Name `json:"orgs,omitempty"`
	Clusters []base.Name `json:"clusters,omitempty"`
}

type Tenant struct {
	Id           string     `json:"id,omitempty"`
	Name         base.Name  `json:"name,omitempty"`
	Kind         base.Kind  `json:"kind,omitempty"`
	Version      float32    `json:"version"`
	Description  string     `json:"description,omitempty"`
	Tags         TenantTags `json:"tags,omitempty"`
	Created      string     `json:"created,omitempty"`
	LastModified string     `json:"lastModified,omitempty"`
	Links        base.Links `json:"links,omitempty"`
	Spec         TenantSpec `json:"spec"`
}
