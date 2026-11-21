/* hand-maintained — canonical source of truth for the `spicedbcluster` schema.
 * `validate:` tags drive pkg/validate.FromType reflection (replaces the
 * schema_spicedb.go builder). See nodelibs/schema/src/spicedb.ts. locations use
 * the SpicedbLocationLink leaf type (LocalLink+toKind('location')); ClusterStatus
 * .unknown(true) is a RegisterObjectRules (schema_reflect_datanet.go). */

package spicedb

import "github.com/controlplane-com/libs-go/pkg/schema/base"

// SpicedbLocationLink is the constrained element type for spec.locations.
type SpicedbLocationLink string

// SpicedbClusterTags aliases base.Tags for the reflection base-type registry.
type SpicedbClusterTags = base.Tags

type ClusterSpecVersion string

const (
	ClusterSpecVersion1141 ClusterSpecVersion = "1.14.1"
)

type ClusterSpec struct {
	Version   ClusterSpecVersion    `json:"version,omitzero" validate:"valid=1.14.1"`
	Locations []SpicedbLocationLink `json:"locations,omitzero" validate:"sort,min=1,dedupe"`
}

// ClusterStatus ports `Joi.object({externalEndpoint}).unknown(true)`.
type ClusterStatus struct {
	ExternalEndpoint string `json:"externalEndpoint,omitzero"`
}

type SpicedbClusterSpecVersion string

const (
	SpicedbClusterSpecVersion1141 SpicedbClusterSpecVersion = "1.14.1"
)

type SpicedbClusterSpec struct {
	Version   SpicedbClusterSpecVersion `json:"version,omitzero" validate:"valid=1.14.1"`
	Locations []SpicedbLocationLink     `json:"locations,omitzero" validate:"sort,min=1,dedupe"`
}

// SpicedbCluster ports `Base.keys({spec: required, alias: optional readOnly, status})`.
type SpicedbCluster struct {
	Id           string             `json:"id,omitzero" validate:"uuid,readOnly"`
	Name         base.Name          `json:"name,omitzero"`
	Kind         base.Kind          `json:"kind,omitzero"`
	Version      *float64           `json:"version,omitzero" validate:"positive,integer,readOnly"`
	Description  string             `json:"description,omitzero" validate:"trim,max=250,allow="`
	Tags         SpicedbClusterTags `json:"tags,omitzero" validate:"default={}"`
	Created      string             `json:"created,omitzero" validate:"date,readOnly"`
	LastModified string             `json:"lastModified,omitzero" validate:"date,readOnly"`
	Links        base.Links         `json:"links,omitzero" validate:"readOnly"`
	Spec         SpicedbClusterSpec `json:"spec" validate:"required"`
	Alias        string             `json:"alias,omitzero" validate:"readOnly"`
	Status       *ClusterStatus     `json:"status,omitzero"`
}
