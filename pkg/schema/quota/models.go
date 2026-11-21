/* hand-maintained — canonical source of truth for the `quota` resource schema.
 * `validate:` tags drive pkg/validate.FromType. See nodelibs/schema/src/quota.ts.
 * NOT Base-derived: version is only positive (no integer/readOnly), description is
 * only trim, created/lastModified are plain dates. Registered as the `quota` kind.
 */

package quota

import "github.com/controlplane-com/libs-go/pkg/schema/base"

type QuotaDimensions map[string]any

type QuotaOrigin string

const (
	QuotaOriginDefault QuotaOrigin = "default"
	QuotaOriginBuiltin QuotaOrigin = "builtin"
)

type Quota struct {
	Id           string          `json:"id,omitzero" validate:"uuid"`
	Kind         base.Kind       `json:"kind,omitzero"`
	Version      *float64        `json:"version,omitzero" validate:"positive"`
	Description  string          `json:"description,omitzero" validate:"trim"`
	Name         base.Name       `json:"name,omitzero"`
	Unit         string          `json:"unit,omitzero"`
	Dimensions   QuotaDimensions `json:"dimensions,omitzero"`
	Max          *float64        `json:"max,omitzero"`
	Created      string          `json:"created,omitzero" validate:"date"`
	LastModified string          `json:"lastModified,omitzero" validate:"date"`
	Current      *float64        `json:"current,omitzero"`
	Origin       QuotaOrigin     `json:"origin,omitzero" validate:"valid=default|builtin,default=default"`
	Links        base.Links      `json:"links,omitzero"`
}
