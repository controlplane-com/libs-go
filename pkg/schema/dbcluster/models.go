/* hand-maintained — canonical source of truth for the `dbcluster` schema.
 * `validate:` tags drive pkg/validate.FromType reflection (replaces the
 * schema_dbcluster.go builder). See nodelibs/schema/src/dbcluster.ts. The
 * PostgresConfig.locations single-writable + makeUnique('name') checks are the
 * `uniqueBy=name` modifier + the dbWritable array custom (dbWritable registered in
 * schema_reflect_datanet.go); ClusterStatus.unknown(true) is a RegisterObjectRules. */

package dbcluster

import "github.com/controlplane-com/libs-go/pkg/schema/base"

// DbClusterTags aliases base.Tags for the reflection base-type registry.
type DbClusterTags = base.Tags

type ClusterSpecPostgresVersion string

const (
	ClusterSpecPostgresVersion128 ClusterSpecPostgresVersion = "12.8"
	ClusterSpecPostgresVersion134 ClusterSpecPostgresVersion = "13.4"
)

// EnumValues resolves version → String().Valid('12.8','13.4') (string identity,
// avoiding the float coercion a `valid=12.8|13.4` tag would suffer).
func (ClusterSpecPostgresVersion) EnumValues() []string { return []string{"12.8", "13.4"} }

// dbSizeValues ports `Size` (dbcluster.ts:13).
var dbSizeValues = []string{"developer-2c-4g", "small-2c-8g", "standard-4c-16g", "large-8c-32g", "xlarge-16c-64g"}

type ClusterSpecPostgresDefaultsSize string

const (
	ClusterSpecPostgresDefaultsSizeDeveloper2C4G ClusterSpecPostgresDefaultsSize = "developer-2c-4g"
	ClusterSpecPostgresDefaultsSizeSmall2C8G     ClusterSpecPostgresDefaultsSize = "small-2c-8g"
	ClusterSpecPostgresDefaultsSizeStandard4C16G ClusterSpecPostgresDefaultsSize = "standard-4c-16g"
	ClusterSpecPostgresDefaultsSizeLarge8C32G    ClusterSpecPostgresDefaultsSize = "large-8c-32g"
	ClusterSpecPostgresDefaultsSizeXlarge16C64G  ClusterSpecPostgresDefaultsSize = "xlarge-16c-64g"
)

func (ClusterSpecPostgresDefaultsSize) EnumValues() []string { return dbSizeValues }

type ClusterSpecPostgresDefaults struct {
	Size      ClusterSpecPostgresDefaultsSize `json:"size,omitzero" validate:"required"`
	Instances float64                         `json:"instances" validate:"integer,required,min=1,max=3"`
}

type ClusterSpecPostgresLocationsSize string

const (
	ClusterSpecPostgresLocationsSizeDeveloper2C4G ClusterSpecPostgresLocationsSize = "developer-2c-4g"
	ClusterSpecPostgresLocationsSizeSmall2C8G     ClusterSpecPostgresLocationsSize = "small-2c-8g"
	ClusterSpecPostgresLocationsSizeStandard4C16G ClusterSpecPostgresLocationsSize = "standard-4c-16g"
	ClusterSpecPostgresLocationsSizeLarge8C32G    ClusterSpecPostgresLocationsSize = "large-8c-32g"
	ClusterSpecPostgresLocationsSizeXlarge16C64G  ClusterSpecPostgresLocationsSize = "xlarge-16c-64g"
)

func (ClusterSpecPostgresLocationsSize) EnumValues() []string { return dbSizeValues }

type ClusterSpecPostgresLocations struct {
	Name      string                           `json:"name,omitzero"`
	Writable  *bool                            `json:"writable,omitzero"`
	Size      ClusterSpecPostgresLocationsSize `json:"size,omitzero"`
	Instances *float64                         `json:"instances,omitzero" validate:"integer,min=1,max=3"`
}

// ClusterSpecPostgresLocationsList is the strongly-typed replacement for the
// dbWritable array custom tag on ClusterSpecPostgres.Locations. The declarative
// `min=1,uniqueBy=name` modifiers stay on the field tag; this type contributes the
// irreducible single-writable custom over the collection.
type ClusterSpecPostgresLocationsList []ClusterSpecPostgresLocations

type ClusterSpecPostgres struct {
	Version   ClusterSpecPostgresVersion       `json:"version,omitzero"`
	StorageGB float64                          `json:"storageGB" validate:"integer,required,min=1,max=10000"`
	Defaults  *ClusterSpecPostgresDefaults     `json:"defaults,omitzero"`
	Locations ClusterSpecPostgresLocationsList `json:"locations,omitzero" validate:"min=1,dedupeBy=name"`
}

type ClusterSpec struct {
	ExternalAccessEnabled *bool               `json:"externalAccessEnabled,omitzero" validate:"default=false"`
	Postgres              ClusterSpecPostgres `json:"postgres" validate:"required"`
}

// ClusterStatus ports `Joi.object({parentId, externalEndpoint}).unknown(true)`.
// unknown(true) is a RegisterObjectRules (schema_reflect_datanet.go).
type ClusterStatus struct {
	ParentId         string `json:"parentId,omitzero" validate:"uuid"`
	ExternalEndpoint string `json:"externalEndpoint,omitzero"`
}

type DbCluster struct {
	Id           string         `json:"id,omitzero" validate:"uuid,readOnly"`
	Name         base.Name      `json:"name,omitzero"`
	Kind         base.Kind      `json:"kind,omitzero"`
	Version      *float64       `json:"version,omitzero" validate:"positive,integer,readOnly"`
	Description  string         `json:"description,omitzero" validate:"trim,max=250,allow="`
	Tags         DbClusterTags  `json:"tags,omitzero" validate:"default={}"`
	Created      string         `json:"created,omitzero" validate:"date,readOnly"`
	LastModified string         `json:"lastModified,omitzero" validate:"date,readOnly"`
	Links        base.Links     `json:"links,omitzero" validate:"readOnly"`
	Spec         *ClusterSpec   `json:"spec,omitzero"`
	Status       *ClusterStatus `json:"status,omitzero"`
	Alias        string         `json:"alias,omitzero"`
}

// DbClusterStatus ports dbcluster.ts:86-92 — lastModified is a Joi.date() and version
// is number().positive().integer(); both rules were dropped in the port.
type DbClusterStatus struct {
	Data         *ClusterStatus `json:"data,omitzero"`
	LastModified string         `json:"lastModified,omitzero" validate:"date"`
	Version      *float64       `json:"version,omitzero" validate:"positive,integer"`
}
