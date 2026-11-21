/* hand-maintained — canonical source of truth for the `memcachecluster` schema.
 * `validate:` tags drive pkg/validate.FromType reflection (replaces the
 * schema_memcache.go builder). See nodelibs/schema/src/memcache.ts for the Joi
 * source and pkg/validate/reflect.go for the tag DSL. */

package memcache

import "github.com/controlplane-com/libs-go/pkg/schema/base"

// MemcacheLocationLink is the constrained element type for spec.locations —
// LocalLink.custom(toKind('location')) (registered → LocalLinkRule+ToKind).
type MemcacheLocationLink string

// MemcacheClusterTags aliases base.Tags so the reflection base-type registry
// resolves it to the shared TagsRule.
type MemcacheClusterTags = base.Tags

type ClusterSpecVersion string

const (
	ClusterSpecVersion1617 ClusterSpecVersion = "1.6.17"
	ClusterSpecVersion1522 ClusterSpecVersion = "1.5.22"
)

type ClusterSpecOptions struct {
	EvictionsDisabled  *bool    `json:"evictionsDisabled,omitzero" validate:"default=true"`
	IdleTimeoutSeconds *float64 `json:"idleTimeoutSeconds,omitzero" validate:"integer,min=0,default=600"`
	MaxItemSizeKiB     *float64 `json:"maxItemSizeKiB,omitzero" validate:"integer,min=512,default=1024"`
	MaxConnections     *float64 `json:"maxConnections,omitzero" validate:"integer,min=10,default=1024"`
}

type ClusterSpec struct {
	NodeCount   float64                `json:"nodeCount" validate:"integer,min=1,max=8,required"`
	NodeSizeGiB float64                `json:"nodeSizeGiB" validate:"min=0.25,max=4,precision=2,required"`
	Version     ClusterSpecVersion     `json:"version,omitzero" validate:"valid=1.6.17|1.5.22"`
	Options     *ClusterSpecOptions    `json:"options,omitzero" validate:"default={}"`
	Locations   []MemcacheLocationLink `json:"locations,omitzero" validate:"sort,min=1,dedupe"`
}

type ClusterStatus map[string]any

type MemcacheClusterSpecVersion string

const (
	MemcacheClusterSpecVersion1617 MemcacheClusterSpecVersion = "1.6.17"
	MemcacheClusterSpecVersion1522 MemcacheClusterSpecVersion = "1.5.22"
)

type MemcacheClusterSpecOptions struct {
	EvictionsDisabled  *bool    `json:"evictionsDisabled,omitzero" validate:"default=true"`
	IdleTimeoutSeconds *float64 `json:"idleTimeoutSeconds,omitzero" validate:"integer,min=0,default=600"`
	MaxItemSizeKiB     *float64 `json:"maxItemSizeKiB,omitzero" validate:"integer,min=512,default=1024"`
	MaxConnections     *float64 `json:"maxConnections,omitzero" validate:"integer,min=10,default=1024"`
}

type MemcacheClusterSpec struct {
	NodeCount   float64                     `json:"nodeCount" validate:"integer,min=1,max=8,required"`
	NodeSizeGiB float64                     `json:"nodeSizeGiB" validate:"min=0.25,max=4,precision=2,required"`
	Version     MemcacheClusterSpecVersion  `json:"version,omitzero" validate:"valid=1.6.17|1.5.22"`
	Options     *MemcacheClusterSpecOptions `json:"options,omitzero" validate:"default={}"`
	Locations   []MemcacheLocationLink      `json:"locations,omitzero" validate:"sort,min=1,dedupe"`
}

// MemcacheCluster ports `MemcacheCluster = Base.keys({spec: required, status})`.
type MemcacheCluster struct {
	Id           string              `json:"id,omitzero" validate:"uuid,readOnly"`
	Name         base.Name           `json:"name,omitzero"`
	Kind         base.Kind           `json:"kind,omitzero"`
	Version      *float64            `json:"version,omitzero" validate:"positive,integer,readOnly"`
	Description  string              `json:"description,omitzero" validate:"trim,max=250,allow="`
	Tags         MemcacheClusterTags `json:"tags,omitzero" validate:"default={}"`
	Created      string              `json:"created,omitzero" validate:"date,readOnly"`
	LastModified string              `json:"lastModified,omitzero" validate:"date,readOnly"`
	Links        base.Links          `json:"links,omitzero" validate:"readOnly"`
	Spec         MemcacheClusterSpec `json:"spec" validate:"required"`
	Status       ClusterStatus       `json:"status,omitzero"`
}

type MemcacheOptions struct {
	EvictionsDisabled  *bool    `json:"evictionsDisabled,omitzero" validate:"default=true"`
	IdleTimeoutSeconds *float64 `json:"idleTimeoutSeconds,omitzero" validate:"integer,min=0,default=600"`
	MaxItemSizeKiB     *float64 `json:"maxItemSizeKiB,omitzero" validate:"integer,min=512,default=1024"`
	MaxConnections     *float64 `json:"maxConnections,omitzero" validate:"integer,min=10,default=1024"`
}
