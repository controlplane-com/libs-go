/* auto-generated */

package envoyAccessLog

import "github.com/controlplane-com/libs-go/pkg/schema/envoyCommon"

type AccessLogName string

const (
	AccessLogNameEnvoyHttpGrpcAccessLog AccessLogName = "envoy.http_grpc_access_log"
)

type AccessLogTypedConfigCommonConfig struct {
	Log_name                    string                  `json:"log_name"`
	Grpc_service                envoyCommon.GrpcService `json:"grpc_service,omitempty"`
	Transport_api_version       envoyCommon.ApiVersion  `json:"transport_api_version,omitempty"`
	Buffer_flush_interval       envoyCommon.Duration    `json:"buffer_flush_interval,omitempty"`
	Buffer_size_bytes           envoyCommon.UInt32      `json:"buffer_size_bytes,omitempty"`
	Filter_state_objects_to_log []string                `json:"filter_state_objects_to_log,omitempty"`
	Grpc_stream_retry_policy    envoyCommon.RetryPolicy `json:"grpc_stream_retry_policy,omitempty"`
}

type AccessLogTypedConfig struct {
	Type                                any                              `json:"@type,omitempty"`
	Common_config                       AccessLogTypedConfigCommonConfig `json:"common_config,omitempty"`
	Additional_request_headers_to_log   []string                         `json:"additional_request_headers_to_log,omitempty"`
	Additional_response_headers_to_log  []string                         `json:"additional_response_headers_to_log,omitempty"`
	Additional_response_trailers_to_log []string                         `json:"additional_response_trailers_to_log,omitempty"`
}

type AccessLog struct {
	Priority          envoyCommon.Priority `json:"priority,omitempty"`
	Name              AccessLogName        `json:"name,omitempty"`
	ExcludedWorkloads []string             `json:"excludedWorkloads,omitempty"`
	Typed_config      AccessLogTypedConfig `json:"typed_config,omitempty"`
}
