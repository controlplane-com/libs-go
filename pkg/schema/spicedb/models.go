/* auto-generated */

package spicedb

import "github.com/controlplane-com/libs-go/pkg/schema/base"

type ClusterSpecVersion string

const (
	ClusterSpecVersion1141 ClusterSpecVersion = "1.14.1"
)

type ClusterSpec struct {
	Version   ClusterSpecVersion `json:"version,omitempty"`
	Locations []string           `json:"locations,omitempty"`
}

type ClusterStatus struct {
	ExternalEndpoint string `json:"externalEndpoint,omitempty"`

	/* WARNING!! Arbitrary properties are being ignored! */
}

type SpicedbClusterTags map[string]any

type SpicedbClusterSpecVersion string

const (
	SpicedbClusterSpecVersion1141 SpicedbClusterSpecVersion = "1.14.1"
)

type SpicedbClusterSpec struct {
	Version   SpicedbClusterSpecVersion `json:"version,omitempty"`
	Locations []string                  `json:"locations,omitempty"`
}

type SpicedbCluster struct {
	Id           string             `json:"id,omitempty"`
	Name         base.Name          `json:"name,omitempty"`
	Kind         base.Kind          `json:"kind,omitempty"`
	Version      *float32           `json:"version,omitempty"`
	Description  string             `json:"description,omitempty"`
	Tags         SpicedbClusterTags `json:"tags,omitempty"`
	Created      string             `json:"created,omitempty"`
	LastModified string             `json:"lastModified,omitempty"`
	Links        base.Links         `json:"links,omitempty"`
	Spec         SpicedbClusterSpec `json:"spec"`
	Alias        string             `json:"alias,omitempty"`
	Status       ClusterStatus      `json:"status,omitempty"`
}
