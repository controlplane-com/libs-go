/* auto-generated */

package envoyRateLimit

import "github.com/controlplane-com/libs-go/pkg/schema/envoyCommon"

type RateLimitServiceGrpcServiceEnvoyGrpc struct {
	Cluster_name string                   `json:"cluster_name" validate:"required"`
	Authority    string                   `json:"authority,omitzero"`
	Retry_policy *envoyCommon.RetryPolicy `json:"retry_policy,omitzero"`
}

type RateLimitServiceGrpcServiceGoogleGrpcChannelCredentialsSslCredentials struct {
	Root_certs  *envoyCommon.DataSource `json:"root_certs,omitzero"`
	Private_key *envoyCommon.DataSource `json:"private_key,omitzero"`
	Cert_chain  *envoyCommon.DataSource `json:"cert_chain,omitzero"`
}

type RateLimitServiceGrpcServiceGoogleGrpcChannelCredentials struct {
	Ssl_credentials *RateLimitServiceGrpcServiceGoogleGrpcChannelCredentialsSslCredentials `json:"ssl_credentials,omitzero"`
	Google_default  *envoyCommon.Empty                                                     `json:"google_default,omitzero"`
}

type RateLimitServiceGrpcServiceGoogleGrpc struct {
	Target_uri               string                                                   `json:"target_uri" validate:"required"`
	Channel_credentials      *RateLimitServiceGrpcServiceGoogleGrpcChannelCredentials `json:"channel_credentials,omitzero"`
	Call_credentials         []envoyCommon.GoogleCallCredentials                      `json:"call_credentials,omitzero"`
	Stat_prefix              string                                                   `json:"stat_prefix" validate:"required"`
	Credentials_factory_name string                                                   `json:"credentials_factory_name,omitzero"`
	Config                   envoyCommon.Struct                                       `json:"config,omitzero"`
}

type RateLimitServiceGrpcService struct {
	Envoy_grpc  *RateLimitServiceGrpcServiceEnvoyGrpc  `json:"envoy_grpc,omitzero"`
	Google_grpc *RateLimitServiceGrpcServiceGoogleGrpc `json:"google_grpc,omitzero"`
	Timeout     envoyCommon.Duration                   `json:"timeout,omitzero"`
}

type RateLimitService struct {
	Grpc_service          RateLimitServiceGrpcService `json:"grpc_service" validate:"required"`
	Transport_api_version envoyCommon.ApiVersion      `json:"transport_api_version,omitzero"`
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
	Domain                             string                           `json:"domain" validate:"required"`
	Stage                              *float64                         `json:"stage,omitzero" validate:"integer,min=0,max=10"`
	Request_type                       RateLimitRequestType             `json:"request_type,omitzero"`
	Timeout                            envoyCommon.Duration             `json:"timeout,omitzero"`
	Failure_mode_deny                  *bool                            `json:"failure_mode_deny,omitzero"`
	Rate_limited_as_resource_exhausted *bool                            `json:"rate_limited_as_resource_exhausted,omitzero"`
	Rate_limit_service                 *RateLimitService                `json:"rate_limit_service,omitzero"`
	Enable_x_ratelimit_headers         RateLimitEnableXRatelimitHeaders `json:"enable_x_ratelimit_headers,omitzero"`
	Disable_x_envoy_ratelimited_header *bool                            `json:"disable_x_envoy_ratelimited_header,omitzero"`
	Rate_limited_status                *envoyCommon.HttpStatus          `json:"rate_limited_status,omitzero"`
	Response_headers                   []envoyCommon.HeaderValueOption  `json:"response_headers,omitzero"`
	Status_on_error                    *envoyCommon.HttpStatus          `json:"status_on_error,omitzero"`
	Stat_prefix                        string                           `json:"stat_prefix,omitzero"`
}
