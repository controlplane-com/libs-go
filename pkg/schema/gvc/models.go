/* hand-maintained — canonical source of truth for the `gvc` resource schema.
 *
 * `validate:` struct tags drive pkg/validate.FromType. See
 * nodelibs/schema/src/gvc.ts and the tag DSL on go-libs/schema/ipSet/models.go.
 *
 * Cross-partition embeds resolve via the base-type registry: env.EnvCollection
 * (workload slice), envoy.EnvoyFilters (registered here -> EnvoyRule), query.Query
 * (registered here -> QueryRule), and tracing.Tracing (Partition D, reflected).
 *
 * Object-level rules (validate.go):
 *   - GvcSpec: endpointNamingFormat POST default (gvcSpecRefine);
 *   - GvcLoadBalancerConfig: minScale<=maxScale (gvcLbScaleRefine).
 */

package gvc

import (
	"github.com/controlplane-com/libs-go/pkg/schema/base"
	"github.com/controlplane-com/libs-go/pkg/schema/env"
	"github.com/controlplane-com/libs-go/pkg/schema/envoy"
	"github.com/controlplane-com/libs-go/pkg/schema/query"
	"github.com/controlplane-com/libs-go/pkg/schema/tracing"
)

// LocationLink / SecretLink are localLink leaf types (validate.go)
// so they can appear as slice elements (which cannot carry field tags).
type LocationLink string
type SecretLink string

// IpSetLink / IdentityLink / WorkloadLink are localLink leaf types carrying their
// kind check via the Validator interface (were localLink field tags).
type IpSetLink string
type IdentityLink string
type WorkloadLink string

// LocationLinkSet is a slice of location links deduplicated by value (was
// custom=dedupeEmpty). The `sort` tag on the field applies on top.
type LocationLinkSet []LocationLink

// SecretLinkSet is a slice of secret links deduplicated by value (was
// custom=dedupeEmpty). The `sort` tag on the field applies on top.
type SecretLinkSet []SecretLink

// MemFmt is a memory-quantity string (pattern=memFmt) whose unit casing is
// normalised (custom=memNormalize). The field's max/default tags apply on top.
type MemFmt string

// CpuFmt is a cpu-quantity string (pattern=cpuFmt). The field's max/default tags
// apply on top.
type CpuFmt string

// GvcTags aliases base.Tags for the reflection base-type registry.
type GvcTags = base.Tags

type GvcStatus map[string]any

type Gvc struct {
	Id           string     `json:"id,omitzero" validate:"uuid,readOnly"`
	Name         base.Name  `json:"name,omitzero"`
	Kind         base.Kind  `json:"kind,omitzero"`
	Version      *float64   `json:"version,omitzero" validate:"positive,integer,readOnly"`
	Description  string     `json:"description,omitzero" validate:"trim,max=250,allow="`
	Tags         GvcTags    `json:"tags,omitzero" validate:"default={}"`
	Created      string     `json:"created,omitzero" validate:"date,readOnly"`
	LastModified string     `json:"lastModified,omitzero" validate:"date,readOnly"`
	Links        base.Links `json:"links,omitzero" validate:"readOnly"`
	Alias        string     `json:"alias,omitzero" validate:"optional,readOnly"`
	Spec         *GvcSpec   `json:"spec,omitzero"`
	Status       *GvcStatus `json:"status,omitzero"`
}

type LocationOptionsItem struct {
	LocationLink       LocationLink `json:"locationLink" validate:"required"`
	RoutingTier        *float64     `json:"routingTier,omitzero" validate:"integer,min=0,optional"`
	LatencyOffsetMs    *float64     `json:"latencyOffsetMs,omitzero" validate:"integer,optional"`
	LatencyToleranceMs *float64     `json:"latencyToleranceMs,omitzero" validate:"integer,min=0,optional"`
}

type LocationOptions []LocationOptionsItem //ts:export

type StaticPlacement struct {
	LocationLinks   LocationLinkSet `json:"locationLinks,omitzero" validate:"sort"`
	LocationQuery   *query.Query    `json:"locationQuery,omitzero"`
	LocationOptions LocationOptions `json:"locationOptions,omitzero" validate:"optional"`
}

type GvcSpecEndpointNamingFormat string

const (
	GvcSpecEndpointNamingFormatDefault GvcSpecEndpointNamingFormat = "default"
	GvcSpecEndpointNamingFormatLegacy  GvcSpecEndpointNamingFormat = "legacy"
	GvcSpecEndpointNamingFormatOrg     GvcSpecEndpointNamingFormat = "org"
)

type GvcSpecSidecar struct {
	Envoy *envoy.EnvoyFilters `json:"envoy,omitzero"`
}

type GvcSpecLoadBalancerMultiZone struct {
	Enabled *bool `json:"enabled,omitzero" validate:"default=false,optional"`
}

type GvcSpecLoadBalancerRedirectClass struct {
	Status5xx string `json:"status5xx,omitzero" validate:"uri,optional"`
	Status401 string `json:"status401,omitzero" validate:"optional"`
}

type GvcSpecLoadBalancerRedirect struct {
	Class *GvcSpecLoadBalancerRedirectClass `json:"class,omitzero" validate:"optional"`
}

type GvcSpecLoadBalancer struct {
	Dedicated      *bool                         `json:"dedicated,omitzero"`
	MultiZone      *GvcSpecLoadBalancerMultiZone `json:"multiZone,omitzero" validate:"optional"`
	TrustedProxies *float64                      `json:"trustedProxies,omitzero" validate:"integer,min=0,max=2,default=0"`
	Redirect       *GvcSpecLoadBalancerRedirect  `json:"redirect,omitzero" validate:"optional"`
	IpSet          IpSetLink                     `json:"ipSet,omitzero" validate:"optional"`
}

type GvcSpecKeda struct {
	Enabled      *bool         `json:"enabled,omitzero" validate:"default=false"`
	IdentityLink IdentityLink  `json:"identityLink,omitzero" validate:"optional"`
	Secrets      SecretLinkSet `json:"secrets,omitzero" validate:"sort"`
}

type GvcSpec struct {
	StaticPlacement      *StaticPlacement            `json:"staticPlacement,omitzero"`
	PullSecretLinks      SecretLinkSet               `json:"pullSecretLinks,omitzero" validate:"sort"`
	Domain               string                      `json:"domain,omitzero" validate:"domaintlds,optional"`
	EndpointNamingFormat GvcSpecEndpointNamingFormat `json:"endpointNamingFormat,omitzero" validate:"valid=default|legacy|org"`
	AliasWorkloadLink    WorkloadLink                `json:"aliasWorkloadLink,omitzero" validate:"optional"`
	Tracing              *tracing.Tracing            `json:"tracing,omitzero"`
	Sidecar              *GvcSpecSidecar             `json:"sidecar,omitzero"`
	Env                  env.EnvCollection           `json:"env,omitzero" validate:"optional"`
	LoadBalancer         *GvcSpecLoadBalancer        `json:"loadBalancer,omitzero"`
	Keda                 *GvcSpecKeda                `json:"keda,omitzero"`
}

// --- config sub-schemas (gvc.ts:190+; not a registered kind) ---

type GvcConfigClusters map[string]GvcConfigClusterEntry

type GvcConfigPreferredClusters map[string]GvcConfigPreferredClusterEntry

type GvcConfigProxy struct {
	MinCpu *float64 `json:"minCpu,omitzero" validate:"min=0,max=1000,optional"`
}

type GvcConfigCapacityAi struct {
	MemToCpuRatio *float64 `json:"memToCpuRatio,omitzero" validate:"integer,min=1,max=10"`
}

type GvcConfig struct {
	Clusters          GvcConfigClusters          `json:"clusters,omitzero"`
	PreferredClusters GvcConfigPreferredClusters `json:"preferredClusters,omitzero"`
	Proxy             *GvcConfigProxy            `json:"proxy,omitzero" validate:"optional"`
	LoadBalancer      *GvcLoadBalancerConfig     `json:"loadBalancer,omitzero" validate:"optional"`
	ThinProvision     *float64                   `json:"thinProvision,omitzero" validate:"min=0.01,max=1.0"`
	LargeDiskSize     MemFmt                     `json:"largeDiskSize,omitzero" validate:"max=20"`
	CapacityAI        *GvcConfigCapacityAi       `json:"capacityAI,omitzero"`
	Spire             *bool                      `json:"spire,omitzero"`
}

type GvcConfigClusterEntry struct {
	ClusterId string `json:"clusterId,omitzero"`
	Since     string `json:"since,omitzero" validate:"date"`
}

type GvcConfigPreferredClusterEntry struct {
	ClusterId string `json:"clusterId,omitzero"`
}

type GvcLoadBalancerProbe struct {
	TimeoutSeconds   *float64 `json:"timeoutSeconds,omitzero" validate:"integer,min=1,max=60"`
	FailureThreshold *float64 `json:"failureThreshold,omitzero" validate:"integer,min=1,max=10"`
	SuccessThreshold *float64 `json:"successThreshold,omitzero" validate:"integer,min=1,max=10"`
}

// GvcLoadBalancerConfig ports gvc.ts:190. gvcLbScaleRefine enforces minScale<=maxScale.
type GvcLoadBalancerConfig struct {
	// Retired (master 54353d516): still validated, then stripped, so stored configs keep
	// validating and the key is never written back or echoed.
	DistributedIngress *bool                 `json:"distributedIngress,omitzero" validate:"strip"`
	DeploymentCount    *float64              `json:"deploymentCount,omitzero" validate:"integer,min=1,max=10,optional"`
	MinScale           *float64              `json:"minScale,omitzero" validate:"integer,min=2,default=2"`
	MaxScale           *float64              `json:"maxScale,omitzero" validate:"integer,min=2,default=10"`
	MinCpu             CpuFmt                `json:"minCpu,omitzero" validate:"max=20,default=100m"`
	MinMemory          MemFmt                `json:"minMemory,omitzero" validate:"max=20,default=500Mi"`
	ReadinessProbe     *GvcLoadBalancerProbe `json:"readinessProbe,omitzero" validate:"optional"`
	LivenessProbe      *GvcLoadBalancerProbe `json:"livenessProbe,omitzero" validate:"optional"`
}
