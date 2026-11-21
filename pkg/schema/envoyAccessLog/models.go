/* auto-generated */

package envoyAccessLog

import "github.com/controlplane-com/libs-go/pkg/schema/envoyCommon"

type AccessLogName string

const (
	AccessLogNameEnvoyHttpGrpcAccessLog AccessLogName = "envoy.http_grpc_access_log"
)

// EnumValues makes AccessLogName reflect to Joi.string().valid(...) (envoyAccessLog.ts).
func (AccessLogName) EnumValues() []string {
	return []string{string(AccessLogNameEnvoyHttpGrpcAccessLog)}
}

type AccessLogTypedConfigCommonConfig struct {
	Log_name                    string                   `json:"log_name" validate:"required"`
	Grpc_service                *envoyCommon.GrpcService `json:"grpc_service,omitzero"`
	Transport_api_version       envoyCommon.ApiVersion   `json:"transport_api_version,omitzero"`
	Buffer_flush_interval       envoyCommon.Duration     `json:"buffer_flush_interval,omitzero"`
	Buffer_size_bytes           *envoyCommon.UInt32      `json:"buffer_size_bytes,omitzero"`
	Filter_state_objects_to_log []string                 `json:"filter_state_objects_to_log,omitzero"`
	Grpc_stream_retry_policy    *envoyCommon.RetryPolicy `json:"grpc_stream_retry_policy,omitzero"`
}

type AccessLogTypedConfig struct {
	// `@type` is a BARE literal in envoyAccessLog.ts:21, which Joi compiles to
	// any().valid(lit) — not string().valid(lit). A Go string would answer "must be a
	// string" where TS answers "must be [<lit>]".
	Type                                any                               `json:"@type,omitzero" validate:"valid=type.googleapis.com/envoy.extensions.access_loggers.grpc.v3.HttpGrpcAccessLogConfig"`
	Common_config                       *AccessLogTypedConfigCommonConfig `json:"common_config,omitzero"`
	Additional_request_headers_to_log   []string                          `json:"additional_request_headers_to_log,omitzero"`
	Additional_response_headers_to_log  []string                          `json:"additional_response_headers_to_log,omitzero"`
	Additional_response_trailers_to_log []string                          `json:"additional_response_trailers_to_log,omitzero"`
}

type AccessLog struct {
	Priority          *envoyCommon.Priority `json:"priority,omitzero"`
	Name              AccessLogName         `json:"name,omitzero"`
	ExcludedWorkloads []string              `json:"excludedWorkloads,omitzero"`
	Typed_config      *AccessLogTypedConfig `json:"typed_config,omitzero"`
}
