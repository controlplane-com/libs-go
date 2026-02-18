/* auto-generated */

package envoyCluster

import "github.com/controlplane-com/libs-go/pkg/schema/envoyCommon"

type ClusterType string

const (
	ClusterTypeStatic      ClusterType = "STATIC"
	ClusterTypeStrictDns   ClusterType = "STRICT_DNS"
	ClusterTypeLogicalDns  ClusterType = "LOGICAL_DNS"
	ClusterTypeEds         ClusterType = "EDS"
	ClusterTypeOriginalDst ClusterType = "ORIGINAL_DST"
)

type ClusterLoadAssignment struct {
	Cluster_name string `json:"cluster_name"`
	Endpoints    any    `json:"endpoints,omitempty"`
	Policy       any    `json:"policy,omitempty"`
}

type Cluster struct {
	Name                                      string                `json:"name"`
	ExcludedWorkloads                         []string              `json:"excludedWorkloads,omitempty"`
	Transport_socket_matches                  any                   `json:"transport_socket_matches,omitempty"`
	Alt_stat_name                             any                   `json:"alt_stat_name,omitempty"`
	Type                                      ClusterType           `json:"type,omitempty"`
	Cluster_type                              any                   `json:"cluster_type,omitempty"`
	Eds_cluster_config                        any                   `json:"eds_cluster_config,omitempty"`
	Connect_timeout                           any                   `json:"connect_timeout,omitempty"`
	Per_connection_buffer_limit_bytes         any                   `json:"per_connection_buffer_limit_bytes,omitempty"`
	Lb_policy                                 any                   `json:"lb_policy,omitempty"`
	Load_assignment                           ClusterLoadAssignment `json:"load_assignment,omitempty"`
	Health_checks                             any                   `json:"health_checks,omitempty"`
	Max_requests_per_connection               any                   `json:"max_requests_per_connection,omitempty"`
	Circuit_breakers                          any                   `json:"circuit_breakers,omitempty"`
	Upstream_http_protocol_options            any                   `json:"upstream_http_protocol_options,omitempty"`
	Common_http_protocol_options              any                   `json:"common_http_protocol_options,omitempty"`
	Http_protocol_options                     any                   `json:"http_protocol_options,omitempty"`
	Http2_protocol_options                    any                   `json:"http2_protocol_options,omitempty"`
	Typed_extension_protocol_options          any                   `json:"typed_extension_protocol_options,omitempty"`
	Dns_refresh_rate                          any                   `json:"dns_refresh_rate,omitempty"`
	Dns_failure_refresh_rate                  any                   `json:"dns_failure_refresh_rate,omitempty"`
	Respect_dns_ttl                           any                   `json:"respect_dns_ttl,omitempty"`
	Dns_lookup_family                         any                   `json:"dns_lookup_family,omitempty"`
	Dns_resolvers                             any                   `json:"dns_resolvers,omitempty"`
	Use_tcp_for_dns_lookups                   any                   `json:"use_tcp_for_dns_lookups,omitempty"`
	Dns_resolution_config                     any                   `json:"dns_resolution_config,omitempty"`
	Typed_dns_resolver_config                 any                   `json:"typed_dns_resolver_config,omitempty"`
	Wait_for_warm_on_init                     any                   `json:"wait_for_warm_on_init,omitempty"`
	Outlier_detection                         any                   `json:"outlier_detection,omitempty"`
	Cleanup_interval                          any                   `json:"cleanup_interval,omitempty"`
	Upstream_bind_config                      any                   `json:"upstream_bind_config,omitempty"`
	Lb_subset_config                          any                   `json:"lb_subset_config,omitempty"`
	Ring_hash_lb_config                       any                   `json:"ring_hash_lb_config,omitempty"`
	Maglev_lb_config                          any                   `json:"maglev_lb_config,omitempty"`
	Least_request_lb_config                   any                   `json:"least_request_lb_config,omitempty"`
	Common_lb_config                          any                   `json:"common_lb_config,omitempty"`
	Transport_socket                          any                   `json:"transport_socket,omitempty"`
	Metadata                                  any                   `json:"metadata,omitempty"`
	Protocol_selection                        any                   `json:"protocol_selection,omitempty"`
	Upstream_connection_options               any                   `json:"upstream_connection_options,omitempty"`
	Close_connections_on_host_health_failure  any                   `json:"close_connections_on_host_health_failure,omitempty"`
	Ignore_health_on_host_removal             any                   `json:"ignore_health_on_host_removal,omitempty"`
	Filters                                   any                   `json:"filters,omitempty"`
	Load_balancing_policy                     any                   `json:"load_balancing_policy,omitempty"`
	Track_timeout_budgets                     any                   `json:"track_timeout_budgets,omitempty"`
	Upstream_config                           any                   `json:"upstream_config,omitempty"`
	Track_cluster_stats                       any                   `json:"track_cluster_stats,omitempty"`
	Preconnect_policy                         any                   `json:"preconnect_policy,omitempty"`
	Connection_pool_per_downstream_connection any                   `json:"connection_pool_per_downstream_connection,omitempty"`
}

type MetadataFilterMetadata map[string]envoyCommon.Struct

type MetadataTypedFilterMetadata map[string]any

type Metadata struct {
	Filter_metadata       MetadataFilterMetadata      `json:"filter_metadata,omitempty"`
	Typed_filter_metadata MetadataTypedFilterMetadata `json:"typed_filter_metadata,omitempty"`
}
