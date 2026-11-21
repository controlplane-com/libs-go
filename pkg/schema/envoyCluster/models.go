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
	Cluster_name string `json:"cluster_name" validate:"required"`
	Endpoints    any    `json:"endpoints,omitzero"`
	Policy       any    `json:"policy,omitzero"`
}

type Cluster struct {
	Name                                      string                 `json:"name" validate:"required"`
	ExcludedWorkloads                         []string               `json:"excludedWorkloads,omitzero"`
	Transport_socket_matches                  any                    `json:"transport_socket_matches,omitzero"`
	Alt_stat_name                             any                    `json:"alt_stat_name,omitzero"`
	Type                                      ClusterType            `json:"type,omitzero"`
	Cluster_type                              any                    `json:"cluster_type,omitzero"`
	Eds_cluster_config                        any                    `json:"eds_cluster_config,omitzero"`
	Connect_timeout                           any                    `json:"connect_timeout,omitzero"`
	Per_connection_buffer_limit_bytes         any                    `json:"per_connection_buffer_limit_bytes,omitzero"`
	Lb_policy                                 any                    `json:"lb_policy,omitzero"`
	Load_assignment                           *ClusterLoadAssignment `json:"load_assignment,omitzero"`
	Health_checks                             any                    `json:"health_checks,omitzero"`
	Max_requests_per_connection               any                    `json:"max_requests_per_connection,omitzero"`
	Circuit_breakers                          any                    `json:"circuit_breakers,omitzero"`
	Upstream_http_protocol_options            any                    `json:"upstream_http_protocol_options,omitzero"`
	Common_http_protocol_options              any                    `json:"common_http_protocol_options,omitzero"`
	Http_protocol_options                     any                    `json:"http_protocol_options,omitzero"`
	Http2_protocol_options                    any                    `json:"http2_protocol_options,omitzero"`
	Typed_extension_protocol_options          any                    `json:"typed_extension_protocol_options,omitzero"`
	Dns_refresh_rate                          any                    `json:"dns_refresh_rate,omitzero"`
	Dns_failure_refresh_rate                  any                    `json:"dns_failure_refresh_rate,omitzero"`
	Respect_dns_ttl                           any                    `json:"respect_dns_ttl,omitzero"`
	Dns_lookup_family                         any                    `json:"dns_lookup_family,omitzero"`
	Dns_resolvers                             any                    `json:"dns_resolvers,omitzero"`
	Use_tcp_for_dns_lookups                   any                    `json:"use_tcp_for_dns_lookups,omitzero"`
	Dns_resolution_config                     any                    `json:"dns_resolution_config,omitzero"`
	Typed_dns_resolver_config                 any                    `json:"typed_dns_resolver_config,omitzero"`
	Wait_for_warm_on_init                     any                    `json:"wait_for_warm_on_init,omitzero"`
	Outlier_detection                         any                    `json:"outlier_detection,omitzero"`
	Cleanup_interval                          any                    `json:"cleanup_interval,omitzero"`
	Upstream_bind_config                      any                    `json:"upstream_bind_config,omitzero"`
	Lb_subset_config                          any                    `json:"lb_subset_config,omitzero"`
	Ring_hash_lb_config                       any                    `json:"ring_hash_lb_config,omitzero"`
	Maglev_lb_config                          any                    `json:"maglev_lb_config,omitzero"`
	Least_request_lb_config                   any                    `json:"least_request_lb_config,omitzero"`
	Common_lb_config                          any                    `json:"common_lb_config,omitzero"`
	Transport_socket                          any                    `json:"transport_socket,omitzero"`
	Metadata                                  any                    `json:"metadata,omitzero"`
	Protocol_selection                        any                    `json:"protocol_selection,omitzero"`
	Upstream_connection_options               any                    `json:"upstream_connection_options,omitzero"`
	Close_connections_on_host_health_failure  any                    `json:"close_connections_on_host_health_failure,omitzero"`
	Ignore_health_on_host_removal             any                    `json:"ignore_health_on_host_removal,omitzero"`
	Filters                                   any                    `json:"filters,omitzero"`
	Load_balancing_policy                     any                    `json:"load_balancing_policy,omitzero"`
	Track_timeout_budgets                     any                    `json:"track_timeout_budgets,omitzero"`
	Upstream_config                           any                    `json:"upstream_config,omitzero"`
	Track_cluster_stats                       any                    `json:"track_cluster_stats,omitzero"`
	Preconnect_policy                         any                    `json:"preconnect_policy,omitzero"`
	Connection_pool_per_downstream_connection any                    `json:"connection_pool_per_downstream_connection,omitzero"`
}

type MetadataFilterMetadata map[string]envoyCommon.Struct

type MetadataTypedFilterMetadata map[string]any

type Metadata struct {
	Filter_metadata       MetadataFilterMetadata      `json:"filter_metadata,omitzero"`
	Typed_filter_metadata MetadataTypedFilterMetadata `json:"typed_filter_metadata,omitzero"`
}
