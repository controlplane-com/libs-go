/* auto-generated */

package quota

import "github.com/controlplane-com/libs-go/pkg/schema/base"

type QuotaDimensions map[string]any

type QuotaOrigin string

const (
	QuotaOriginDefault QuotaOrigin = "default"
	QuotaOriginBuiltin QuotaOrigin = "builtin"
)

type Quota struct {
	Id           string          `json:"id,omitempty"`
	Kind         base.Kind       `json:"kind,omitempty"`
	Version      float32         `json:"version"`
	Description  string          `json:"description,omitempty"`
	Name         base.Name       `json:"name,omitempty"`
	Unit         string          `json:"unit,omitempty"`
	Dimensions   QuotaDimensions `json:"dimensions,omitempty"`
	Max          float32         `json:"max"`
	Created      string          `json:"created,omitempty"`
	LastModified string          `json:"lastModified,omitempty"`
	Current      float32         `json:"current"`
	Origin       QuotaOrigin     `json:"origin,omitempty"`
	Links        base.Links      `json:"links,omitempty"`
}
