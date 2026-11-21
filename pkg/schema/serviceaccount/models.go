/* hand-maintained — canonical source of truth for the `serviceaccount` resource.
 * Was joi-to-go-generated from nodelibs/schema/src/serviceaccount.ts; `validate:`
 * tags carry the rules. See serviceaccount.ts for the Joi source.
 */

package serviceaccount

import "github.com/controlplane-com/libs-go/pkg/schema/base"

// ServiceAccountTags aliases base.Tags for the reflection base-type registry.
type ServiceAccountTags = base.Tags

type ServiceAccountOrigin string

const (
	ServiceAccountOriginDefault ServiceAccountOrigin = "default"
	ServiceAccountOriginBuiltin ServiceAccountOrigin = "builtin"
)

func (ServiceAccountOrigin) EnumValues() []string {
	return []string{string(ServiceAccountOriginDefault), string(ServiceAccountOriginBuiltin)}
}

type ServiceAccount struct {
	Id           string               `json:"id,omitzero" validate:"uuid,readOnly"`
	Name         base.Name            `json:"name,omitzero"`
	Kind         base.Kind            `json:"kind,omitzero"`
	Version      *float64             `json:"version,omitzero" validate:"positive,integer,readOnly"`
	Description  string               `json:"description,omitzero" validate:"trim,max=250,allow="`
	Tags         ServiceAccountTags   `json:"tags,omitzero" validate:"default={}"`
	Created      string               `json:"created,omitzero" validate:"date,readOnly"`
	LastModified string               `json:"lastModified,omitzero" validate:"date,readOnly"`
	Links        base.Links           `json:"links,omitzero" validate:"readOnly"`
	Keys         []ServiceAccountKey  `json:"keys,omitzero" validate:"readOnly"`
	Origin       ServiceAccountOrigin `json:"origin,omitzero" validate:"default=default,optional,readOnly"`
}

type ServiceAccountKey struct {
	Name        string `json:"name,omitzero" validate:"readOnly"`
	Created     string `json:"created,omitzero" validate:"date,readOnly"`
	Key         string `json:"key,omitzero" validate:"readOnly"`
	Description string `json:"description" validate:"required,trim,max=250,allow="`
}
