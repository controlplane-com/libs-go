/* auto-generated */

package envoyHttp

import "github.com/controlplane-com/libs-go/pkg/schema/envoyCommon"

type BufferSettings struct {
	Max_request_bytes     *float32 `json:"max_request_bytes,omitempty"`
	Allow_partial_message bool     `json:"allow_partial_message,omitempty"`
	Pack_as_bytes         bool     `json:"pack_as_bytes,omitempty"`
}

type ConnectRpcName string

const (
	ConnectRpcNameEnvoyFiltersHttpConnectGrpcBridge ConnectRpcName = "envoy.filters.http.connect_grpc_bridge"
)

type ConnectRpcTypedConfigType string

const (
	ConnectRpcTypedConfigTypeTypeGoogleapisComEnvoyExtensionsFiltersHttpConnectGrpcBridgeV3FilterConfig ConnectRpcTypedConfigType = "type.googleapis.com/envoy.extensions.filters.http.connect_grpc_bridge.v3.FilterConfig"
)

type ConnectRpcTypedConfig struct {
	Type ConnectRpcTypedConfigType `json:"@type,omitempty"`
}

type ConnectRpc struct {
	Priority          envoyCommon.Priority  `json:"priority,omitempty"`
	ExcludedWorkloads []string              `json:"excludedWorkloads,omitempty"`
	Name              ConnectRpcName        `json:"name,omitempty"`
	Typed_config      ConnectRpcTypedConfig `json:"typed_config,omitempty"`
}

type CorsName string

const (
	CorsNameEnvoyFiltersHttpCors CorsName = "envoy.filters.http.cors"
)

type CorsTypedConfigType string

const (
	CorsTypedConfigTypeTypeGoogleapisComEnvoyExtensionsFiltersHttpCorsV3Cors CorsTypedConfigType = "type.googleapis.com/envoy.extensions.filters.http.cors.v3.Cors"
)

type CorsTypedConfig struct {
	Allow_origin_string_match    []envoyCommon.StringMatcher          `json:"allow_origin_string_match,omitempty"`
	Allow_methods                string                               `json:"allow_methods,omitempty"`
	Allow_headers                string                               `json:"allow_headers,omitempty"`
	Expose_headers               string                               `json:"expose_headers,omitempty"`
	Max_age                      string                               `json:"max_age,omitempty"`
	Allow_credentials            bool                                 `json:"allow_credentials,omitempty"`
	Filter_enabled               envoyCommon.RuntimeFractionalPercent `json:"filter_enabled,omitempty"`
	Shadow_enabled               envoyCommon.RuntimeFractionalPercent `json:"shadow_enabled,omitempty"`
	Allow_private_network_access bool                                 `json:"allow_private_network_access,omitempty"`
	Type                         CorsTypedConfigType                  `json:"@type,omitempty"`
}

type Cors struct {
	Priority          envoyCommon.Priority `json:"priority,omitempty"`
	ExcludedWorkloads []string             `json:"excludedWorkloads,omitempty"`
	Name              CorsName             `json:"name,omitempty"`
	Typed_config      CorsTypedConfig      `json:"typed_config,omitempty"`
}

type DescriptorRateLimitReplaces struct {
	Name string `json:"name,omitempty"`
}

type DescriptorRateLimitUnit string

const (
	DescriptorRateLimitUnitSecond DescriptorRateLimitUnit = "second"
	DescriptorRateLimitUnitMinute DescriptorRateLimitUnit = "minute"
	DescriptorRateLimitUnitHour   DescriptorRateLimitUnit = "hour"
	DescriptorRateLimitUnitDay    DescriptorRateLimitUnit = "day"
)

type DescriptorRateLimit struct {
	Name              string                        `json:"name,omitempty"`
	Replaces          []DescriptorRateLimitReplaces `json:"replaces,omitempty"`
	Unit              DescriptorRateLimitUnit       `json:"unit,omitempty"`
	Requests_per_unit envoyCommon.UInt32            `json:"requests_per_unit,omitempty"`
}

type Descriptor struct {
	Key             string              `json:"key"`
	Value           string              `json:"value,omitempty"`
	Rate_limit      DescriptorRateLimit `json:"rate_limit,omitempty"`
	Shadow_mode     string              `json:"shadow_mode,omitempty"`
	Detailed_metric string              `json:"detailed_metric,omitempty"`
}

type ExtAuthzName string

const (
	ExtAuthzNameEnvoyFiltersHttpExtAuthz ExtAuthzName = "envoy.filters.http.ext_authz"
)

type ExtAuthzTypedConfigHttpServiceAuthorizationRequest struct {
	Allowed_headers envoyCommon.ListStringMatcher `json:"allowed_headers,omitempty"`
	Headers_to_add  []envoyCommon.HeaderValue     `json:"headers_to_add,omitempty"`
}

type ExtAuthzTypedConfigHttpServiceAuthorizationResponse struct {
	Allowed_upstream_headers           envoyCommon.ListStringMatcher `json:"allowed_upstream_headers,omitempty"`
	Allowed_upstream_headers_to_append envoyCommon.ListStringMatcher `json:"allowed_upstream_headers_to_append,omitempty"`
	Allowed_client_headers             envoyCommon.ListStringMatcher `json:"allowed_client_headers,omitempty"`
	Allowed_client_headers_on_success  envoyCommon.ListStringMatcher `json:"allowed_client_headers_on_success,omitempty"`
	Dynamic_metadata_from_headers      envoyCommon.ListStringMatcher `json:"dynamic_metadata_from_headers,omitempty"`
}

type ExtAuthzTypedConfigHttpService struct {
	Server_uri             HttpUri                                             `json:"server_uri,omitempty"`
	Path_prefix            string                                              `json:"path_prefix,omitempty"`
	Authorization_request  ExtAuthzTypedConfigHttpServiceAuthorizationRequest  `json:"authorization_request,omitempty"`
	Authorization_response ExtAuthzTypedConfigHttpServiceAuthorizationResponse `json:"authorization_response,omitempty"`
}

type ExtAuthzTypedConfigType string

const (
	ExtAuthzTypedConfigTypeTypeGoogleapisComEnvoyExtensionsFiltersHttpExtAuthzV3ExtAuthz ExtAuthzTypedConfigType = "type.googleapis.com/envoy.extensions.filters.http.ext_authz.v3.ExtAuthz"
)

type ExtAuthzTypedConfig struct {
	Grpc_service                            envoyCommon.GrpcService        `json:"grpc_service,omitempty"`
	Http_service                            ExtAuthzTypedConfigHttpService `json:"http_service,omitempty"`
	Failure_mode_allow                      bool                           `json:"failure_mode_allow,omitempty"`
	Failure_mode_allow_header_add           bool                           `json:"failure_mode_allow_header_add,omitempty"`
	With_request_body                       BufferSettings                 `json:"with_request_body,omitempty"`
	Clear_route_cache                       bool                           `json:"clear_route_cache,omitempty"`
	Status_on_error                         envoyCommon.HttpStatus         `json:"status_on_error,omitempty"`
	Metadata_context_namespaces             []string                       `json:"metadata_context_namespaces,omitempty"`
	Typed_metadata_context_namespaces       []string                       `json:"typed_metadata_context_namespaces,omitempty"`
	Route_metadata_context_namespaces       []string                       `json:"route_metadata_context_namespaces,omitempty"`
	Route_typed_metadata_context_namespaces []string                       `json:"route_typed_metadata_context_namespaces,omitempty"`
	Filter_enabled                          envoyCommon.FractionalPercent  `json:"filter_enabled,omitempty"`
	Deny_at_disable                         envoyCommon.RuntimeFeatureFlag `json:"deny_at_disable,omitempty"`
	Include_peer_certificate                bool                           `json:"include_peer_certificate,omitempty"`
	Stat_prefix                             string                         `json:"stat_prefix,omitempty"`
	Bootstrap_metadata_labels_key           string                         `json:"bootstrap_metadata_labels_key,omitempty"`
	Allowed_headers                         envoyCommon.ListStringMatcher  `json:"allowed_headers,omitempty"`
	Include_tls_session                     bool                           `json:"include_tls_session,omitempty"`
	Charge_cluster_response_stats           bool                           `json:"charge_cluster_response_stats,omitempty"`
	Transport_api_version                   envoyCommon.ApiVersion         `json:"transport_api_version,omitempty"`
	Type                                    ExtAuthzTypedConfigType        `json:"@type,omitempty"`
}

type ExtAuthz struct {
	Priority          envoyCommon.Priority `json:"priority,omitempty"`
	ExcludedWorkloads []string             `json:"excludedWorkloads,omitempty"`
	Name              ExtAuthzName         `json:"name,omitempty"`
	Typed_config      ExtAuthzTypedConfig  `json:"typed_config,omitempty"`
}

type GrpcJsonTranscoderName string

const (
	GrpcJsonTranscoderNameEnvoyFiltersHttpGrpcJsonTranscoder GrpcJsonTranscoderName = "envoy.filters.http.grpc_json_transcoder"
)

type GrpcJsonTranscoderTypedConfigPrintOptions struct {
	Add_whitespace                bool `json:"add_whitespace,omitempty"`
	Always_print_primitive_fields bool `json:"always_print_primitive_fields,omitempty"`
	Always_print_enums_as_ints    bool `json:"always_print_enums_as_ints,omitempty"`
	Preserve_proto_field_names    bool `json:"preserve_proto_field_names,omitempty"`
	Stream_newline_delimited      bool `json:"stream_newline_delimited,omitempty"`
}

type GrpcJsonTranscoderTypedConfigUrlUnescapeSpec string

const (
	GrpcJsonTranscoderTypedConfigUrlUnescapeSpecAllCharactersExceptReserved GrpcJsonTranscoderTypedConfigUrlUnescapeSpec = "ALL_CHARACTERS_EXCEPT_RESERVED"
	GrpcJsonTranscoderTypedConfigUrlUnescapeSpecAllCharactersExceptSlash    GrpcJsonTranscoderTypedConfigUrlUnescapeSpec = "ALL_CHARACTERS_EXCEPT_SLASH"
	GrpcJsonTranscoderTypedConfigUrlUnescapeSpecAllCharacters               GrpcJsonTranscoderTypedConfigUrlUnescapeSpec = "ALL_CHARACTERS"
)

type GrpcJsonTranscoderTypedConfigRequestValidationOptions struct {
	Reject_unknown_method                bool `json:"reject_unknown_method,omitempty"`
	Reject_unknown_query_parameters      bool `json:"reject_unknown_query_parameters,omitempty"`
	Reject_binding_body_field_collisions bool `json:"reject_binding_body_field_collisions,omitempty"`
}

type GrpcJsonTranscoderTypedConfigType string

const (
	GrpcJsonTranscoderTypedConfigTypeTypeGoogleapisComEnvoyExtensionsFiltersHttpGrpcJsonTranscoderV3GrpcJsonTranscoder GrpcJsonTranscoderTypedConfigType = "type.googleapis.com/envoy.extensions.filters.http.grpc_json_transcoder.v3.GrpcJsonTranscoder"
)

type GrpcJsonTranscoderTypedConfig struct {
	Proto_descriptor                string `json:"proto_descriptor,omitempty"`
	Proto_descriptor_bin            any/* TODO: [object Object]*/ `json:"proto_descriptor_bin,omitempty"`
	Services                        []string                                              `json:"services,omitempty"`
	Print_options                   GrpcJsonTranscoderTypedConfigPrintOptions             `json:"print_options,omitempty"`
	Match_incoming_request_route    bool                                                  `json:"match_incoming_request_route,omitempty"`
	Ignored_query_parameters        []string                                              `json:"ignored_query_parameters,omitempty"`
	Auto_mapping                    bool                                                  `json:"auto_mapping,omitempty"`
	Ignore_unknown_query_parameters bool                                                  `json:"ignore_unknown_query_parameters,omitempty"`
	Convert_grpc_status             bool                                                  `json:"convert_grpc_status,omitempty"`
	Url_unescape_spec               GrpcJsonTranscoderTypedConfigUrlUnescapeSpec          `json:"url_unescape_spec,omitempty"`
	Query_param_unescape_plus       bool                                                  `json:"query_param_unescape_plus,omitempty"`
	Match_unregistered_custom_verb  bool                                                  `json:"match_unregistered_custom_verb,omitempty"`
	Request_validation_options      GrpcJsonTranscoderTypedConfigRequestValidationOptions `json:"request_validation_options,omitempty"`
	Case_insensitive_enum_parsing   bool                                                  `json:"case_insensitive_enum_parsing,omitempty"`
	Max_request_body_size           envoyCommon.UInt32                                    `json:"max_request_body_size,omitempty"`
	Max_response_body_size          envoyCommon.UInt32                                    `json:"max_response_body_size,omitempty"`
	Type                            GrpcJsonTranscoderTypedConfigType                     `json:"@type,omitempty"`
}

type GrpcJsonTranscoder struct {
	Priority          envoyCommon.Priority          `json:"priority,omitempty"`
	ExcludedWorkloads []string                      `json:"excludedWorkloads,omitempty"`
	Name              GrpcJsonTranscoderName        `json:"name,omitempty"`
	Typed_config      GrpcJsonTranscoderTypedConfig `json:"typed_config,omitempty"`
}

type GrpcWebName string

const (
	GrpcWebNameEnvoyFiltersHttpGrpcWeb GrpcWebName = "envoy.filters.http.grpc_web"
)

type GrpcWebTypedConfigType string

const (
	GrpcWebTypedConfigTypeTypeGoogleapisComEnvoyExtensionsFiltersHttpGrpcWebV3GrpcWeb GrpcWebTypedConfigType = "type.googleapis.com/envoy.extensions.filters.http.grpc_web.v3.GrpcWeb"
)

type GrpcWebTypedConfig struct {
	Type GrpcWebTypedConfigType `json:"@type,omitempty"`
}

type GrpcWeb struct {
	Priority          envoyCommon.Priority `json:"priority,omitempty"`
	ExcludedWorkloads []string             `json:"excludedWorkloads,omitempty"`
	Name              GrpcWebName          `json:"name,omitempty"`
	Typed_config      GrpcWebTypedConfig   `json:"typed_config,omitempty"`
}

type HttpFilter any /* TODO: [object Object]*/

type HttpUri struct {
	Uri     string `json:"uri"`
	Cluster string `json:"cluster"`
	Timeout any/* TODO: [object Object]*/ `json:"timeout"`
}

type HttpUriRestricted struct {
	Uri     string                         `json:"uri"`
	Cluster string                         `json:"cluster"`
	Timeout envoyCommon.DurationRestricted `json:"timeout,omitempty"`
}

type JwtAuthnName string

const (
	JwtAuthnNameEnvoyFiltersHttpJwtAuthn JwtAuthnName = "envoy.filters.http.jwt_authn"
)

type JwtAuthnTypedConfigProviders map[string]JwtProvider

type JwtAuthnTypedConfigFilterStateRules struct {
	Name     string            `json:"name"`
	Requires JwtRequirementMap `json:"requires,omitempty"`
}

type JwtAuthnTypedConfigRequirementMap map[string]JwtRequirement

type JwtAuthnTypedConfigType string

const (
	JwtAuthnTypedConfigTypeTypeGoogleapisComEnvoyExtensionsFiltersHttpJwtAuthnV3JwtAuthentication JwtAuthnTypedConfigType = "type.googleapis.com/envoy.extensions.filters.http.jwt_authn.v3.JwtAuthentication"
)

type JwtAuthnTypedConfig struct {
	Providers             JwtAuthnTypedConfigProviders        `json:"providers,omitempty"`
	Rules                 []JwtRequirementRuleRestricted      `json:"rules,omitempty"`
	Filter_state_rules    JwtAuthnTypedConfigFilterStateRules `json:"filter_state_rules,omitempty"`
	Bypass_cors_preflight bool                                `json:"bypass_cors_preflight,omitempty"`
	Requirement_map       JwtAuthnTypedConfigRequirementMap   `json:"requirement_map,omitempty"`
	Type                  JwtAuthnTypedConfigType             `json:"@type,omitempty"`
}

type JwtAuthn struct {
	Priority          envoyCommon.Priority `json:"priority,omitempty"`
	ExcludedWorkloads []string             `json:"excludedWorkloads,omitempty"`
	Name              JwtAuthnName         `json:"name,omitempty"`
	Typed_config      JwtAuthnTypedConfig  `json:"typed_config,omitempty"`
}

type JwtProviderClaimToHeaders struct {
	Header_name string `json:"header_name"`
	Claim_name  string `json:"claim_name"`
}

type JwtProviderRemoteJwksAsyncFetch struct {
	Fast_listener           bool                 `json:"fast_listener,omitempty"`
	Failed_refetch_duration envoyCommon.Duration `json:"failed_refetch_duration,omitempty"`
}

type JwtProviderRemoteJwks struct {
	Http_uri       HttpUri                         `json:"http_uri,omitempty"`
	Cache_duration envoyCommon.Duration            `json:"cache_duration,omitempty"`
	Async_fetch    JwtProviderRemoteJwksAsyncFetch `json:"async_fetch,omitempty"`
	Retry_policy   envoyCommon.RetryPolicy         `json:"retry_policy,omitempty"`
}

type JwtProvider struct {
	Issuer           string                      `json:"issuer,omitempty"`
	Audiences        []string                    `json:"audiences,omitempty"`
	Claim_to_headers []JwtProviderClaimToHeaders `json:"claim_to_headers,omitempty"`
	Remote_jwks      JwtProviderRemoteJwks       `json:"remote_jwks,omitempty"`
}

type JwtProviderUiRestrictedClaimToHeaders struct {
	Header_name string `json:"header_name"`
	Claim_name  string `json:"claim_name"`
}

type JwtProviderUiRestrictedRemoteJwks struct {
	Http_uri       HttpUriRestricted              `json:"http_uri,omitempty"`
	Cache_duration envoyCommon.DurationRestricted `json:"cache_duration,omitempty"`
}

type JwtProviderUIRestricted struct {
	Issuer           string                                  `json:"issuer,omitempty"`
	Audiences        []string                                `json:"audiences,omitempty"`
	Claim_to_headers []JwtProviderUiRestrictedClaimToHeaders `json:"claim_to_headers,omitempty"`
	Remote_jwks      JwtProviderUiRestrictedRemoteJwks       `json:"remote_jwks,omitempty"`
}

type JwtRequirementProviderAndAudiences struct {
	Provider_name string   `json:"provider_name,omitempty"`
	Audiences     []string `json:"audiences,omitempty"`
}

type JwtRequirementRequiresAny struct {
	Requirements []any `json:"requirements,omitempty"`
}

type JwtRequirementRequiresAll struct {
	Requirements []any `json:"requirements,omitempty"`
}

type JwtRequirement struct {
	Provider_name           string                             `json:"provider_name,omitempty"`
	Provider_and_audiences  JwtRequirementProviderAndAudiences `json:"provider_and_audiences,omitempty"`
	Requires_any            JwtRequirementRequiresAny          `json:"requires_any,omitempty"`
	Requires_all            JwtRequirementRequiresAll          `json:"requires_all,omitempty"`
	Allow_missing_or_failed envoyCommon.Empty                  `json:"allow_missing_or_failed,omitempty"`
	Allow_missing           envoyCommon.Empty                  `json:"allow_missing,omitempty"`
}

type JwtRequirementMap map[string]JwtRequirement

type JwtRequirementRestricted struct {
	Provider_name string `json:"provider_name,omitempty"`
}

type JwtRequirementRule struct {
	Match            envoyCommon.RouteMatch `json:"match,omitempty"`
	Requires         JwtRequirement         `json:"requires,omitempty"`
	Requirement_name string                 `json:"requirement_name,omitempty"`
}

type JwtRequirementRuleRestricted struct {
	Match    envoyCommon.RouteMatchRestricted `json:"match,omitempty"`
	Requires JwtRequirementRestricted         `json:"requires,omitempty"`
}

type RateLimitName string

const (
	RateLimitNameEnvoyFiltersHttpRatelimit RateLimitName = "envoy.filters.http.ratelimit"
)

type RateLimitTypedConfigRequestType string

const (
	RateLimitTypedConfigRequestTypeInternal RateLimitTypedConfigRequestType = "internal"
	RateLimitTypedConfigRequestTypeExternal RateLimitTypedConfigRequestType = "external"
	RateLimitTypedConfigRequestTypeBoth     RateLimitTypedConfigRequestType = "both"
)

type RateLimitTypedConfigEnableXRatelimitHeaders string

const (
	RateLimitTypedConfigEnableXRatelimitHeadersOff            RateLimitTypedConfigEnableXRatelimitHeaders = "OFF"
	RateLimitTypedConfigEnableXRatelimitHeadersDraftVersion03 RateLimitTypedConfigEnableXRatelimitHeaders = "DRAFT_VERSION_03"
)

type RateLimitTypedConfigType string

const (
	RateLimitTypedConfigTypeTypeGoogleapisComEnvoyExtensionsFiltersHttpRatelimitV3RateLimit RateLimitTypedConfigType = "type.googleapis.com/envoy.extensions.filters.http.ratelimit.v3.RateLimit"
)

type RateLimitTypedConfig struct {
	Domain                             string                                      `json:"domain"`
	Stage                              *float32                                    `json:"stage,omitempty"`
	Request_type                       RateLimitTypedConfigRequestType             `json:"request_type,omitempty"`
	Timeout                            envoyCommon.Duration                        `json:"timeout,omitempty"`
	Failure_mode_deny                  bool                                        `json:"failure_mode_deny,omitempty"`
	Rate_limited_as_resource_exhausted bool                                        `json:"rate_limited_as_resource_exhausted,omitempty"`
	Rate_limit_service                 RateLimitService                            `json:"rate_limit_service,omitempty"`
	Enable_x_ratelimit_headers         RateLimitTypedConfigEnableXRatelimitHeaders `json:"enable_x_ratelimit_headers,omitempty"`
	Disable_x_envoy_ratelimited_header bool                                        `json:"disable_x_envoy_ratelimited_header,omitempty"`
	Rate_limited_status                envoyCommon.HttpStatus                      `json:"rate_limited_status,omitempty"`
	Response_headers_to_add            []envoyCommon.HeaderValueOption             `json:"response_headers_to_add,omitempty"`
	Status_on_error                    envoyCommon.HttpStatus                      `json:"status_on_error,omitempty"`
	Stat_prefix                        string                                      `json:"stat_prefix,omitempty"`
	Type                               RateLimitTypedConfigType                    `json:"@type,omitempty"`
}

type RateLimit struct {
	Priority          envoyCommon.Priority `json:"priority,omitempty"`
	ExcludedWorkloads []string             `json:"excludedWorkloads,omitempty"`
	Name              RateLimitName        `json:"name,omitempty"`
	Typed_config      RateLimitTypedConfig `json:"typed_config,omitempty"`
}

type RateLimitServiceGrpcServiceEnvoyGrpc struct {
	Cluster_name string                  `json:"cluster_name"`
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
	Target_uri               string                                                  `json:"target_uri"`
	Channel_credentials      RateLimitServiceGrpcServiceGoogleGrpcChannelCredentials `json:"channel_credentials,omitempty"`
	Call_credentials         []envoyCommon.GoogleCallCredentials                     `json:"call_credentials,omitempty"`
	Stat_prefix              string                                                  `json:"stat_prefix"`
	Credentials_factory_name string                                                  `json:"credentials_factory_name,omitempty"`
	Config                   envoyCommon.Struct                                      `json:"config,omitempty"`
}

type RateLimitServiceGrpcService struct {
	Envoy_grpc       RateLimitServiceGrpcServiceEnvoyGrpc  `json:"envoy_grpc,omitempty"`
	Google_grpc      RateLimitServiceGrpcServiceGoogleGrpc `json:"google_grpc,omitempty"`
	Timeout          any/* TODO: [object Object]*/ `json:"timeout,omitempty"`
	Initial_metadata []envoyCommon.HeaderValue `json:"initial_metadata,omitempty"`
}

type RateLimitService struct {
	Grpc_service          RateLimitServiceGrpcService `json:"grpc_service"`
	Transport_api_version envoyCommon.ApiVersion      `json:"transport_api_version,omitempty"`
}
