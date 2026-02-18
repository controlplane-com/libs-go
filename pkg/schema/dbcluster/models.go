/* auto-generated */

package dbcluster

import "github.com/controlplane-com/libs-go/pkg/schema/base"

type ClusterSpecPostgresVersion string

const (
	ClusterSpecPostgresVersion128 ClusterSpecPostgresVersion = "12.8"
	ClusterSpecPostgresVersion134 ClusterSpecPostgresVersion = "13.4"
)

type ClusterSpecPostgresDefaultsSize string

const (
	ClusterSpecPostgresDefaultsSizeDeveloper2C4G ClusterSpecPostgresDefaultsSize = "developer-2c-4g"
	ClusterSpecPostgresDefaultsSizeSmall2C8G     ClusterSpecPostgresDefaultsSize = "small-2c-8g"
	ClusterSpecPostgresDefaultsSizeStandard4C16G ClusterSpecPostgresDefaultsSize = "standard-4c-16g"
	ClusterSpecPostgresDefaultsSizeLarge8C32G    ClusterSpecPostgresDefaultsSize = "large-8c-32g"
	ClusterSpecPostgresDefaultsSizeXlarge16C64G  ClusterSpecPostgresDefaultsSize = "xlarge-16c-64g"
)

type ClusterSpecPostgresDefaults struct {
	Size      ClusterSpecPostgresDefaultsSize `json:"size,omitempty"`
	Instances float32                         `json:"instances"`
}

type ClusterSpecPostgresLocationsSize string

const (
	ClusterSpecPostgresLocationsSizeDeveloper2C4G ClusterSpecPostgresLocationsSize = "developer-2c-4g"
	ClusterSpecPostgresLocationsSizeSmall2C8G     ClusterSpecPostgresLocationsSize = "small-2c-8g"
	ClusterSpecPostgresLocationsSizeStandard4C16G ClusterSpecPostgresLocationsSize = "standard-4c-16g"
	ClusterSpecPostgresLocationsSizeLarge8C32G    ClusterSpecPostgresLocationsSize = "large-8c-32g"
	ClusterSpecPostgresLocationsSizeXlarge16C64G  ClusterSpecPostgresLocationsSize = "xlarge-16c-64g"
)

type ClusterSpecPostgresLocations struct {
	Name      string                           `json:"name,omitempty"`
	Writable  bool                             `json:"writable,omitempty"`
	Size      ClusterSpecPostgresLocationsSize `json:"size,omitempty"`
	Instances float32                          `json:"instances"`
}

type ClusterSpecPostgres struct {
	Version   ClusterSpecPostgresVersion     `json:"version,omitempty"`
	StorageGB float32                        `json:"storageGB"`
	Defaults  ClusterSpecPostgresDefaults    `json:"defaults,omitempty"`
	Locations []ClusterSpecPostgresLocations `json:"locations,omitempty"`
}

type ClusterSpec struct {
	ExternalAccessEnabled bool                `json:"externalAccessEnabled,omitempty"`
	Postgres              ClusterSpecPostgres `json:"postgres"`
}

type ClusterStatus struct {
	ParentId         string `json:"parentId,omitempty"`
	ExternalEndpoint string `json:"externalEndpoint,omitempty"`

	/* WARNING!! Arbitrary properties are being ignored! */
}

type DbClusterTags map[string]any

type DbCluster struct {
	Id           string        `json:"id,omitempty"`
	Name         base.Name     `json:"name,omitempty"`
	Kind         base.Kind     `json:"kind,omitempty"`
	Version      float32       `json:"version"`
	Description  string        `json:"description,omitempty"`
	Tags         DbClusterTags `json:"tags,omitempty"`
	Created      string        `json:"created,omitempty"`
	LastModified string        `json:"lastModified,omitempty"`
	Links        base.Links    `json:"links,omitempty"`
	Spec         ClusterSpec   `json:"spec,omitempty"`
	Status       ClusterStatus `json:"status,omitempty"`
	Alias        string        `json:"alias,omitempty"`
}

type DbClusterStatus struct {
	Data         ClusterStatus `json:"data,omitempty"`
	LastModified string        `json:"lastModified,omitempty"`
	Version      float32       `json:"version"`
}
