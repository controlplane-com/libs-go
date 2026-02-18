/* auto-generated */

package memcache

import "github.com/controlplane-com/libs-go/pkg/schema/base"

type ClusterSpecVersion string

const (
	ClusterSpecVersion1617 ClusterSpecVersion = "1.6.17"
	ClusterSpecVersion1522 ClusterSpecVersion = "1.5.22"
)

type ClusterSpecOptions struct {
	EvictionsDisabled  bool    `json:"evictionsDisabled,omitempty"`
	IdleTimeoutSeconds float32 `json:"idleTimeoutSeconds"`
	MaxItemSizeKiB     float32 `json:"maxItemSizeKiB"`
	MaxConnections     float32 `json:"maxConnections"`
}

type ClusterSpec struct {
	NodeCount   float32            `json:"nodeCount"`
	NodeSizeGiB float32            `json:"nodeSizeGiB"`
	Version     ClusterSpecVersion `json:"version,omitempty"`
	Options     ClusterSpecOptions `json:"options,omitempty"`
	Locations   []string           `json:"locations,omitempty"`
}

type ClusterStatus map[string]any

type MemcacheClusterTags map[string]any

type MemcacheClusterSpecVersion string

const (
	MemcacheClusterSpecVersion1617 MemcacheClusterSpecVersion = "1.6.17"
	MemcacheClusterSpecVersion1522 MemcacheClusterSpecVersion = "1.5.22"
)

type MemcacheClusterSpecOptions struct {
	EvictionsDisabled  bool    `json:"evictionsDisabled,omitempty"`
	IdleTimeoutSeconds float32 `json:"idleTimeoutSeconds"`
	MaxItemSizeKiB     float32 `json:"maxItemSizeKiB"`
	MaxConnections     float32 `json:"maxConnections"`
}

type MemcacheClusterSpec struct {
	NodeCount   float32                    `json:"nodeCount"`
	NodeSizeGiB float32                    `json:"nodeSizeGiB"`
	Version     MemcacheClusterSpecVersion `json:"version,omitempty"`
	Options     MemcacheClusterSpecOptions `json:"options,omitempty"`
	Locations   []string                   `json:"locations,omitempty"`
}

type MemcacheCluster struct {
	Id           string              `json:"id,omitempty"`
	Name         base.Name           `json:"name,omitempty"`
	Kind         base.Kind           `json:"kind,omitempty"`
	Version      float32             `json:"version"`
	Description  string              `json:"description,omitempty"`
	Tags         MemcacheClusterTags `json:"tags,omitempty"`
	Created      string              `json:"created,omitempty"`
	LastModified string              `json:"lastModified,omitempty"`
	Links        base.Links          `json:"links,omitempty"`
	Spec         MemcacheClusterSpec `json:"spec"`
	Status       ClusterStatus       `json:"status,omitempty"`
}

type MemcacheOptions struct {
	EvictionsDisabled  bool    `json:"evictionsDisabled,omitempty"`
	IdleTimeoutSeconds float32 `json:"idleTimeoutSeconds"`
	MaxItemSizeKiB     float32 `json:"maxItemSizeKiB"`
	MaxConnections     float32 `json:"maxConnections"`
}
