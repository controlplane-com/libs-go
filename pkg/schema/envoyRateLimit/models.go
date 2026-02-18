/* auto-generated */

package envoyRateLimit

import "github.com/controlplane-com/libs-go/pkg/schema/envoyCommon"

type RateLimitServiceGrpcServiceEnvoyGrpc struct {
	Cluster_name string                  `json:"cluster_name,omitempty"`
	Authority    string                  `json:"authority,omitempty"`
	Retry_policy envoyCommon.RetryPolicy `json:"retry_policy,omitempty"`
}

type RateLimitServiceGrpcServiceGoogleGrpcChannelCredentialsSslCredentials struct {
	Root_certs  envoyCommon.DataSource `json:"root_certs,omitempty"`
	Private_key envoyCommon.DataSource `json:"private_key,omitempty"`
	Cert_chain  envoyCommon.DataSource `json:"cert_chain,omitempty"`
}

type RateLimitServiceGrpcServiceGoogleGrpcChannelCredentials struct {
	Ssl_credentials RateLimitServiceGrpcServiceGoogleGrpcChannelCredentialsSslCredentials `json:"ssl_credentials,omitempty"`
	Google_default  envoyCommon.Empty                                                     `json:"google_default,omitempty"`
}

type RateLimitServiceGrpcServiceGoogleGrpc struct {
	Target_uri               string                                                  `json:"target_uri,omitempty"`
	Channel_credentials      RateLimitServiceGrpcServiceGoogleGrpcChannelCredentials `json:"channel_credentials,omitempty"`
	Call_credentials         []envoyCommon.GoogleCallCredentials                     `json:"call_credentials,omitempty"`
	Stat_prefix              string                                                  `json:"stat_prefix,omitempty"`
	Credentials_factory_name string                                                  `json:"credentials_factory_name,omitempty"`
	Config                   envoyCommon.Struct                                      `json:"config,omitempty"`
}

type RateLimitServiceGrpcService struct {
	Envoy_grpc  RateLimitServiceGrpcServiceEnvoyGrpc  `json:"envoy_grpc,omitempty"`
	Google_grpc RateLimitServiceGrpcServiceGoogleGrpc `json:"google_grpc,omitempty"`
	Timeout     any/* TODO: [object Object]*/ `json:"timeout,omitempty"`
}

type RateLimitService struct {
	Grpc_service          RateLimitServiceGrpcService `json:"grpc_service,omitempty"`
	Transport_api_version envoyCommon.ApiVersion      `json:"transport_api_version,omitempty"`
}

type RateLimitRequestType string

const (
	RateLimitRequestTypeInternal RateLimitRequestType = "internal"
	RateLimitRequestTypeExternal RateLimitRequestType = "external"
	RateLimitRequestTypeBoth     RateLimitRequestType = "both"
)

type RateLimitEnableXRatelimitHeaders string

const (
	RateLimitEnableXRatelimitHeadersOff            RateLimitEnableXRatelimitHeaders = "OFF"
	RateLimitEnableXRatelimitHeadersDraftVersion03 RateLimitEnableXRatelimitHeaders = "DRAFT_VERSION_03"
)

type RateLimit struct {
	Domain                             string                           `json:"domain,omitempty"`
	Stage                              float32                          `json:"stage"`
	Request_type                       RateLimitRequestType             `json:"request_type,omitempty"`
	Timeout                            envoyCommon.Duration             `json:"timeout,omitempty"`
	Failure_mode_deny                  bool                             `json:"failure_mode_deny,omitempty"`
	Rate_limited_as_resource_exhausted bool                             `json:"rate_limited_as_resource_exhausted,omitempty"`
	Rate_limit_service                 RateLimitService                 `json:"rate_limit_service,omitempty"`
	Enable_x_ratelimit_headers         RateLimitEnableXRatelimitHeaders `json:"enable_x_ratelimit_headers,omitempty"`
	Disable_x_envoy_ratelimited_header bool                             `json:"disable_x_envoy_ratelimited_header,omitempty"`
	Rate_limited_status                envoyCommon.HttpStatus           `json:"rate_limited_status,omitempty"`
	Response_headers                   []envoyCommon.HeaderValueOption  `json:"response_headers,omitempty"`
	Status_on_error                    envoyCommon.HttpStatus           `json:"status_on_error,omitempty"`
	Stat_prefix                        string                           `json:"stat_prefix,omitempty"`
}
