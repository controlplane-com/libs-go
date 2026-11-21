/* hand-maintained — canonical source of truth for the `tenant` resource schema.
 * `validate:` tags drive pkg/validate.FromType. See nodelibs/schema/src/tenant.ts.
 * Tenant = Base.keys({spec: TenantSpec.required()}). Registered as `tenant`.
 */

package tenant

import "github.com/controlplane-com/libs-go/pkg/schema/base"

// TenantTags aliases base.Tags for the reflection base-type registry.
type TenantTags = base.Tags

// NameList ports TenantSpec.orgs/clusters `Joi.array().items(Name).sort()
// .custom(makeUnique())`: an array of base.Name leaves, sorted ascending, with
// duplicate elements removed in place — the strongly-typed replacement for the
// `sort,custom=dedupeEmpty` field tag.
type NameList []base.Name

type TenantSpec struct {
	Orgs     NameList `json:"orgs,omitzero"`
	Clusters NameList `json:"clusters,omitzero"`
}

type Tenant struct {
	Id           string     `json:"id,omitzero" validate:"uuid,readOnly"`
	Name         base.Name  `json:"name,omitzero"`
	Kind         base.Kind  `json:"kind,omitzero"`
	Version      *float64   `json:"version,omitzero" validate:"positive,integer,readOnly"`
	Description  string     `json:"description,omitzero" validate:"trim,max=250,allow="`
	Tags         TenantTags `json:"tags,omitzero" validate:"default={}"`
	Created      string     `json:"created,omitzero" validate:"date,readOnly"`
	LastModified string     `json:"lastModified,omitzero" validate:"date,readOnly"`
	Links        base.Links `json:"links,omitzero" validate:"readOnly"`
	Spec         TenantSpec `json:"spec" validate:"required"`
}
