/* auto-generated */

package gvc

import "github.com/controlplane-com/libs-go/pkg/schema/base"
import "github.com/controlplane-com/libs-go/pkg/schema/workload"
import "github.com/controlplane-com/libs-go/pkg/schema/tracing"
import "github.com/controlplane-com/libs-go/pkg/schema/envoy"
import "github.com/controlplane-com/libs-go/pkg/schema/env"
import "github.com/controlplane-com/libs-go/pkg/schema/query"

type GvcTags map[string]any

type Gvc struct {
	Id           string     `json:"id,omitempty"`
	Name         base.Name  `json:"name,omitempty"`
	Kind         base.Kind  `json:"kind,omitempty"`
	Version      *float32   `json:"version,omitempty"`
	Description  string     `json:"description,omitempty"`
	Tags         GvcTags    `json:"tags,omitempty"`
	Created      string     `json:"created,omitempty"`
	LastModified string     `json:"lastModified,omitempty"`
	Links        base.Links `json:"links,omitempty"`
	Alias        string     `json:"alias,omitempty"`
	Spec         *GvcSpec   `json:"spec,omitempty"`
	Status       *GvcStatus `json:"status,omitempty"`
}

type GvcConfigClusters map[string]GvcConfigClusterEntry

type GvcConfigPreferredClusters map[string]GvcConfigPreferredClusterEntry

type GvcConfigProxy struct {
	MinCpu *float32 `json:"minCpu,omitempty"`
}

type GvcConfigLoadBalancerReadinessProbe struct {
	TimeoutSeconds   *float32 `json:"timeoutSeconds,omitempty"`
	FailureThreshold *float32 `json:"failureThreshold,omitempty"`
	SuccessThreshold *float32 `json:"successThreshold,omitempty"`
}

type GvcConfigLoadBalancerLivenessProbe struct {
	TimeoutSeconds   *float32 `json:"timeoutSeconds,omitempty"`
	FailureThreshold *float32 `json:"failureThreshold,omitempty"`
	SuccessThreshold *float32 `json:"successThreshold,omitempty"`
}

type GvcConfigLoadBalancer struct {
	DistributedIngress bool                                 `json:"distributedIngress,omitempty"`
	MinScale           *float32                             `json:"minScale,omitempty"`
	MaxScale           *float32                             `json:"maxScale,omitempty"`
	MinCpu             string                               `json:"minCpu,omitempty"`
	MinMemory          string                               `json:"minMemory,omitempty"`
	ReadinessProbe     *GvcConfigLoadBalancerReadinessProbe `json:"readinessProbe,omitempty"`
	LivenessProbe      *GvcConfigLoadBalancerLivenessProbe  `json:"livenessProbe,omitempty"`
}

type GvcConfigCapacityAi struct {
	MemToCpuRatio *float32 `json:"memToCpuRatio,omitempty"`
}

type GvcConfig struct {
	Clusters          GvcConfigClusters          `json:"clusters,omitempty"`
	PreferredClusters GvcConfigPreferredClusters `json:"preferredClusters,omitempty"`
	Proxy             *GvcConfigProxy            `json:"proxy,omitempty"`
	LoadBalancer      *GvcConfigLoadBalancer     `json:"loadBalancer,omitempty"`
	ThinProvision     *float32                   `json:"thinProvision,omitempty"`
	LargeDiskSize     workload.Memory            `json:"largeDiskSize,omitempty"`
	CapacityAI        *GvcConfigCapacityAi       `json:"capacityAI,omitempty"`
}

type GvcConfigClusterEntry struct {
	ClusterId string `json:"clusterId,omitempty"`
	Since     string `json:"since,omitempty"`
}

type GvcConfigPreferredClusterEntry struct {
	ClusterId string `json:"clusterId,omitempty"`
}

type GvcLoadBalancerConfigReadinessProbe struct {
	TimeoutSeconds   *float32 `json:"timeoutSeconds,omitempty"`
	FailureThreshold *float32 `json:"failureThreshold,omitempty"`
	SuccessThreshold *float32 `json:"successThreshold,omitempty"`
}

type GvcLoadBalancerConfigLivenessProbe struct {
	TimeoutSeconds   *float32 `json:"timeoutSeconds,omitempty"`
	FailureThreshold *float32 `json:"failureThreshold,omitempty"`
	SuccessThreshold *float32 `json:"successThreshold,omitempty"`
}

type GvcLoadBalancerConfig struct {
	DistributedIngress bool                                 `json:"distributedIngress,omitempty"`
	MinScale           *float32                             `json:"minScale,omitempty"`
	MaxScale           *float32                             `json:"maxScale,omitempty"`
	MinCpu             string                               `json:"minCpu,omitempty"`
	MinMemory          string                               `json:"minMemory,omitempty"`
	ReadinessProbe     *GvcLoadBalancerConfigReadinessProbe `json:"readinessProbe,omitempty"`
	LivenessProbe      *GvcLoadBalancerConfigLivenessProbe  `json:"livenessProbe,omitempty"`
}

type GvcSpecEndpointNamingFormat string

const (
	GvcSpecEndpointNamingFormatDefault GvcSpecEndpointNamingFormat = "default"
	GvcSpecEndpointNamingFormatLegacy  GvcSpecEndpointNamingFormat = "legacy"
	GvcSpecEndpointNamingFormatOrg     GvcSpecEndpointNamingFormat = "org"
)

type GvcSpecTracingLightstep struct {
	Endpoint    string `json:"endpoint"`
	Credentials string `json:"credentials,omitempty"`
}

type GvcSpecTracingCustomTags map[string]tracing.TracingCustomTag

type GvcSpecTracingProviderOtel struct {
	Endpoint string `json:"endpoint"`
}

type GvcSpecTracingProviderLightstep struct {
	Endpoint    string `json:"endpoint"`
	Credentials string `json:"credentials,omitempty"`
}

type GvcSpecTracingProviderControlplane struct {
}

type GvcSpecTracingProvider struct {
	Otel         *GvcSpecTracingProviderOtel         `json:"otel,omitempty"`
	Lightstep    *GvcSpecTracingProviderLightstep    `json:"lightstep,omitempty"`
	Controlplane *GvcSpecTracingProviderControlplane `json:"controlplane,omitempty"`
}

type GvcSpecTracing struct {
	Sampling   float32                   `json:"sampling"`
	Lightstep  *GvcSpecTracingLightstep  `json:"lightstep,omitempty"`
	CustomTags *GvcSpecTracingCustomTags `json:"customTags,omitempty"`
	Provider   *GvcSpecTracingProvider   `json:"provider,omitempty"`
}

type GvcSpecSidecar struct {
	Envoy envoy.EnvoyFilters `json:"envoy,omitempty"`
}

type GvcSpecLoadBalancerMultiZone struct {
	Enabled bool `json:"enabled,omitempty"`
}

type GvcSpecLoadBalancerRedirectClass struct {
	Status5xx string `json:"status5xx,omitempty"`
	Status401 string `json:"status401,omitempty"`
}

type GvcSpecLoadBalancerRedirect struct {
	Class *GvcSpecLoadBalancerRedirectClass `json:"class,omitempty"`
}

type GvcSpecLoadBalancer struct {
	Dedicated      bool                          `json:"dedicated,omitempty"`
	MultiZone      *GvcSpecLoadBalancerMultiZone `json:"multiZone,omitempty"`
	TrustedProxies *float32                      `json:"trustedProxies,omitempty"`
	Redirect       *GvcSpecLoadBalancerRedirect  `json:"redirect,omitempty"`
	IpSet          string                        `json:"ipSet,omitempty"`
}

type GvcSpecKeda struct {
	Enabled      bool     `json:"enabled,omitempty"`
	IdentityLink string   `json:"identityLink,omitempty"`
	Secrets      []string `json:"secrets,omitempty"`
}

type GvcSpec struct {
	StaticPlacement      *StaticPlacement            `json:"staticPlacement,omitempty"`
	PullSecretLinks      []string                    `json:"pullSecretLinks,omitempty"`
	Domain               string                      `json:"domain,omitempty"`
	EndpointNamingFormat GvcSpecEndpointNamingFormat `json:"endpointNamingFormat,omitempty"`
	AliasWorkloadLink    string                      `json:"aliasWorkloadLink,omitempty"`
	Tracing              *GvcSpecTracing             `json:"tracing,omitempty"`
	Sidecar              *GvcSpecSidecar             `json:"sidecar,omitempty"`
	Env                  []env.EnvVar                `json:"env,omitempty"`
	LoadBalancer         *GvcSpecLoadBalancer        `json:"loadBalancer,omitempty"`
	Keda                 *GvcSpecKeda                `json:"keda,omitempty"`
}

type GvcStatus map[string]any

type LocationOptions []LocationOptionsItem

type LocationOptionsItem struct {
	LocationLink       string   `json:"locationLink"`
	RoutingTier        *float32 `json:"routingTier,omitempty"`
	LatencyOffsetMs    *float32 `json:"latencyOffsetMs,omitempty"`
	LatencyToleranceMs *float32 `json:"latencyToleranceMs,omitempty"`
}

type StaticPlacementLocationQueryContext map[string]any

type StaticPlacementLocationQueryFetch string

const (
	StaticPlacementLocationQueryFetchLinks StaticPlacementLocationQueryFetch = "links"
	StaticPlacementLocationQueryFetchItems StaticPlacementLocationQueryFetch = "items"
)

type StaticPlacementLocationQuerySpecMatch string

const (
	StaticPlacementLocationQuerySpecMatchAll  StaticPlacementLocationQuerySpecMatch = "all"
	StaticPlacementLocationQuerySpecMatchAny  StaticPlacementLocationQuerySpecMatch = "any"
	StaticPlacementLocationQuerySpecMatchNone StaticPlacementLocationQuerySpecMatch = "none"
)

type StaticPlacementLocationQuerySpecSortOrder string

const (
	StaticPlacementLocationQuerySpecSortOrderAsc  StaticPlacementLocationQuerySpecSortOrder = "asc"
	StaticPlacementLocationQuerySpecSortOrderDesc StaticPlacementLocationQuerySpecSortOrder = "desc"
)

type StaticPlacementLocationQuerySpecSort struct {
	By    string                                    `json:"by"`
	Order StaticPlacementLocationQuerySpecSortOrder `json:"order,omitempty"`
}

type StaticPlacementLocationQuerySpec struct {
	Match StaticPlacementLocationQuerySpecMatch `json:"match,omitempty"`
	Terms []query.Term                          `json:"terms,omitempty"`
	Sort  *StaticPlacementLocationQuerySpecSort `json:"sort,omitempty"`
}

type StaticPlacementLocationQuery struct {
	Kind    base.Kind                            `json:"kind,omitempty"`
	Context *StaticPlacementLocationQueryContext `json:"context,omitempty"`
	Fetch   StaticPlacementLocationQueryFetch    `json:"fetch,omitempty"`
	Spec    *StaticPlacementLocationQuerySpec    `json:"spec,omitempty"`
}

type StaticPlacement struct {
	LocationLinks   []string                      `json:"locationLinks,omitempty"`
	LocationQuery   *StaticPlacementLocationQuery `json:"locationQuery,omitempty"`
	LocationOptions *[]LocationOptionsItem        `json:"locationOptions,omitempty"`
}
