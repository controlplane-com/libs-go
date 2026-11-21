/* hand-maintained — canonical source of truth for the `auditctx` resource schema.
 * `validate:` tags drive pkg/validate.FromType. See nodelibs/schema/src/auditctx.ts.
 * AuditContext = Base.keys({status, origin}). The createSchema fork (name required,
 * description default-to-name) is applied on POST by auditctxCreateRefine
 * (schema_reflect_orggvc.go).
 */

package auditctx

import "github.com/controlplane-com/libs-go/pkg/schema/base"

// AuditContextTags aliases base.Tags for the reflection base-type registry.
type AuditContextTags = base.Tags

type AuditContextOrigin string

const (
	AuditContextOriginDefault AuditContextOrigin = "default"
	AuditContextOriginBuiltin AuditContextOrigin = "builtin"
)

// AuditContextStatus = Joi.object({}).unknown(true) (Unknown registered on this type).
type AuditContextStatus map[string]any

type AuditContext struct {
	Id           string             `json:"id,omitzero" validate:"uuid,readOnly"`
	Name         base.Name          `json:"name,omitzero"`
	Kind         base.Kind          `json:"kind,omitzero"`
	Version      *float64           `json:"version,omitzero" validate:"positive,integer,readOnly"`
	Description  string             `json:"description,omitzero" validate:"trim,max=250,allow="`
	Tags         AuditContextTags   `json:"tags,omitzero" validate:"default={}"`
	Created      string             `json:"created,omitzero" validate:"date,readOnly"`
	LastModified string             `json:"lastModified,omitzero" validate:"date,readOnly"`
	Links        base.Links         `json:"links,omitzero" validate:"readOnly"`
	Status       AuditContextStatus `json:"status,omitzero"`
	Origin       AuditContextOrigin `json:"origin,omitzero" validate:"valid=default|builtin,default=default"`
}
