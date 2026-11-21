/* hand-maintained — canonical source of truth (nodelibs/schema/src/config.ts).
 * `validate:` struct tags drive pkg/validate.FromType; DSL on go-libs/schema/ipSet/models.go. */

package config

import "github.com/controlplane-com/libs-go/pkg/schema/base"

type ConfigKind string

const (
	ConfigKindConfig ConfigKind = "config"
)

// EnumValues ports config.ts:6 `kind: Joi.string().valid('config')`.
func (ConfigKind) EnumValues() []string { return []string{"config"} }

type Config[D any] struct {
	Kind         ConfigKind `json:"kind,omitzero"`
	Id           string     `json:"id,omitzero"`
	LastModified string     `json:"lastModified,omitzero" validate:"date"`
	Version      *float64   `json:"version,omitzero"`
	Data         D          `json:"data,omitzero"`
	Links        base.Links `json:"links,omitzero"`
}
