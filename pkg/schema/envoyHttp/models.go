/* auto-generated */

package envoyHttp

import "github.com/controlplane-com/libs-go/pkg/schema/envoyCommon"

type BufferSettings struct {
	Max_request_bytes     *float64 `json:"max_request_bytes,omitzero" validate:"integer,min=0"`
	Allow_partial_message *bool    `json:"allow_partial_message,omitzero"`
	Pack_as_bytes         *bool    `json:"pack_as_bytes,omitzero"`
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
	Type ConnectRpcTypedConfigType `json:"@type,omitzero" validate:"required"`
}

type ConnectRpc struct {
	Priority          *envoyCommon.Priority  `json:"priority,omitzero"`
	ExcludedWorkloads []string               `json:"excludedWorkloads,omitzero"`
	Name              ConnectRpcName         `json:"name,omitzero" validate:"required"`
	Typed_config      *ConnectRpcTypedConfig `json:"typed_config,omitzero"`
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
	Allow_origin_string_match    []envoyCommon.StringMatcher           `json:"allow_origin_string_match,omitzero"`
	Allow_methods                string                                `json:"allow_methods,omitzero"`
	Allow_headers                string                                `json:"allow_headers,omitzero"`
	Expose_headers               string                                `json:"expose_headers,omitzero"`
	Max_age                      string                                `json:"max_age,omitzero"`
	Allow_credentials            *bool                                 `json:"allow_credentials,omitzero"`
	Filter_enabled               *envoyCommon.RuntimeFractionalPercent `json:"filter_enabled,omitzero"`
	Shadow_enabled               *envoyCommon.RuntimeFractionalPercent `json:"shadow_enabled,omitzero"`
	Allow_private_network_access *bool                                 `json:"allow_private_network_access,omitzero"`
	Type                         CorsTypedConfigType                   `json:"@type,omitzero" validate:"required"`
}

type Cors struct {
	Priority          *envoyCommon.Priority `json:"priority,omitzero"`
	ExcludedWorkloads []string              `json:"excludedWorkloads,omitzero"`
	Name              CorsName              `json:"name,omitzero" validate:"required"`
	Typed_config      *CorsTypedConfig      `json:"typed_config,omitzero"`
}

type DescriptorRateLimitReplaces struct {
	Name string `json:"name,omitzero"`
}

type DescriptorRateLimitUnit string

const (
	DescriptorRateLimitUnitSecond DescriptorRateLimitUnit = "second"
	DescriptorRateLimitUnitMinute DescriptorRateLimitUnit = "minute"
	DescriptorRateLimitUnitHour   DescriptorRateLimitUnit = "hour"
	DescriptorRateLimitUnitDay    DescriptorRateLimitUnit = "day"
)

type DescriptorRateLimit struct {
	Name              string                        `json:"name,omitzero"`
	Replaces          []DescriptorRateLimitReplaces `json:"replaces,omitzero"`
	Unit              DescriptorRateLimitUnit       `json:"unit,omitzero"`
	Requests_per_unit *envoyCommon.UInt32           `json:"requests_per_unit,omitzero"`
}

type Descriptor struct {
	Key             string               `json:"key" validate:"required"`
	Value           string               `json:"value,omitzero"`
	Rate_limit      *DescriptorRateLimit `json:"rate_limit,omitzero"`
	Shadow_mode     string               `json:"shadow_mode,omitzero"`
	Detailed_metric string               `json:"detailed_metric,omitzero"`
}

type ExtAuthzName string

const (
	ExtAuthzNameEnvoyFiltersHttpExtAuthz ExtAuthzName = "envoy.filters.http.ext_authz"
)

type ExtAuthzTypedConfigHttpServiceAuthorizationRequest struct {
	Allowed_headers *envoyCommon.ListStringMatcher `json:"allowed_headers,omitzero"`
	Headers_to_add  []envoyCommon.HeaderValue      `json:"headers_to_add,omitzero"`
}

type ExtAuthzTypedConfigHttpServiceAuthorizationResponse struct {
	Allowed_upstream_headers           *envoyCommon.ListStringMatcher `json:"allowed_upstream_headers,omitzero"`
	Allowed_upstream_headers_to_append *envoyCommon.ListStringMatcher `json:"allowed_upstream_headers_to_append,omitzero"`
	Allowed_client_headers             *envoyCommon.ListStringMatcher `json:"allowed_client_headers,omitzero"`
	Allowed_client_headers_on_success  *envoyCommon.ListStringMatcher `json:"allowed_client_headers_on_success,omitzero"`
	Dynamic_metadata_from_headers      *envoyCommon.ListStringMatcher `json:"dynamic_metadata_from_headers,omitzero"`
}

type ExtAuthzTypedConfigHttpService struct {
	Server_uri             *HttpUri                                             `json:"server_uri,omitzero"`
	Path_prefix            string                                               `json:"path_prefix,omitzero"`
	Authorization_request  *ExtAuthzTypedConfigHttpServiceAuthorizationRequest  `json:"authorization_request,omitzero"`
	Authorization_response *ExtAuthzTypedConfigHttpServiceAuthorizationResponse `json:"authorization_response,omitzero"`
}

type ExtAuthzTypedConfigType string

const (
	ExtAuthzTypedConfigTypeTypeGoogleapisComEnvoyExtensionsFiltersHttpExtAuthzV3ExtAuthz ExtAuthzTypedConfigType = "type.googleapis.com/envoy.extensions.filters.http.ext_authz.v3.ExtAuthz"
)

type ExtAuthzTypedConfig struct {
	Grpc_service                            *envoyCommon.GrpcService        `json:"grpc_service,omitzero"`
	Http_service                            *ExtAuthzTypedConfigHttpService `json:"http_service,omitzero"`
	Failure_mode_allow                      *bool                           `json:"failure_mode_allow,omitzero"`
	Failure_mode_allow_header_add           *bool                           `json:"failure_mode_allow_header_add,omitzero"`
	With_request_body                       *BufferSettings                 `json:"with_request_body,omitzero"`
	Clear_route_cache                       *bool                           `json:"clear_route_cache,omitzero"`
	Status_on_error                         *envoyCommon.HttpStatus         `json:"status_on_error,omitzero"`
	Metadata_context_namespaces             []string                        `json:"metadata_context_namespaces,omitzero"`
	Typed_metadata_context_namespaces       []string                        `json:"typed_metadata_context_namespaces,omitzero"`
	Route_metadata_context_namespaces       []string                        `json:"route_metadata_context_namespaces,omitzero"`
	Route_typed_metadata_context_namespaces []string                        `json:"route_typed_metadata_context_namespaces,omitzero"`
	Filter_enabled                          *envoyCommon.FractionalPercent  `json:"filter_enabled,omitzero"`
	Deny_at_disable                         *envoyCommon.RuntimeFeatureFlag `json:"deny_at_disable,omitzero"`
	Include_peer_certificate                *bool                           `json:"include_peer_certificate,omitzero"`
	Stat_prefix                             string                          `json:"stat_prefix,omitzero"`
	Bootstrap_metadata_labels_key           string                          `json:"bootstrap_metadata_labels_key,omitzero"`
	Allowed_headers                         *envoyCommon.ListStringMatcher  `json:"allowed_headers,omitzero"`
	Include_tls_session                     *bool                           `json:"include_tls_session,omitzero"`
	Charge_cluster_response_stats           *bool                           `json:"charge_cluster_response_stats,omitzero"`
	Transport_api_version                   envoyCommon.ApiVersion          `json:"transport_api_version,omitzero"`
	Type                                    ExtAuthzTypedConfigType         `json:"@type,omitzero" validate:"required"`
}

type ExtAuthz struct {
	Priority          *envoyCommon.Priority `json:"priority,omitzero"`
	ExcludedWorkloads []string              `json:"excludedWorkloads,omitzero"`
	Name              ExtAuthzName          `json:"name,omitzero" validate:"required"`
	Typed_config      *ExtAuthzTypedConfig  `json:"typed_config,omitzero"`
}

type GrpcJsonTranscoderName string

const (
	GrpcJsonTranscoderNameEnvoyFiltersHttpGrpcJsonTranscoder GrpcJsonTranscoderName = "envoy.filters.http.grpc_json_transcoder"
)

type GrpcJsonTranscoderTypedConfigPrintOptions struct {
	Add_whitespace                *bool `json:"add_whitespace,omitzero"`
	Always_print_primitive_fields *bool `json:"always_print_primitive_fields,omitzero"`
	Always_print_enums_as_ints    *bool `json:"always_print_enums_as_ints,omitzero"`
	Preserve_proto_field_names    *bool `json:"preserve_proto_field_names,omitzero"`
	Stream_newline_delimited      *bool `json:"stream_newline_delimited,omitzero"`
}

type GrpcJsonTranscoderTypedConfigUrlUnescapeSpec string

const (
	GrpcJsonTranscoderTypedConfigUrlUnescapeSpecAllCharactersExceptReserved GrpcJsonTranscoderTypedConfigUrlUnescapeSpec = "ALL_CHARACTERS_EXCEPT_RESERVED"
	GrpcJsonTranscoderTypedConfigUrlUnescapeSpecAllCharactersExceptSlash    GrpcJsonTranscoderTypedConfigUrlUnescapeSpec = "ALL_CHARACTERS_EXCEPT_SLASH"
	GrpcJsonTranscoderTypedConfigUrlUnescapeSpecAllCharacters               GrpcJsonTranscoderTypedConfigUrlUnescapeSpec = "ALL_CHARACTERS"
)

type GrpcJsonTranscoderTypedConfigRequestValidationOptions struct {
	Reject_unknown_method                *bool `json:"reject_unknown_method,omitzero"`
	Reject_unknown_query_parameters      *bool `json:"reject_unknown_query_parameters,omitzero"`
	Reject_binding_body_field_collisions *bool `json:"reject_binding_body_field_collisions,omitzero"`
}

type GrpcJsonTranscoderTypedConfigType string

const (
	GrpcJsonTranscoderTypedConfigTypeTypeGoogleapisComEnvoyExtensionsFiltersHttpGrpcJsonTranscoderV3GrpcJsonTranscoder GrpcJsonTranscoderTypedConfigType = "type.googleapis.com/envoy.extensions.filters.http.grpc_json_transcoder.v3.GrpcJsonTranscoder"
)

type GrpcJsonTranscoderTypedConfig struct {
	Proto_descriptor                string                                                 `json:"proto_descriptor,omitzero"`
	Proto_descriptor_bin            envoyCommon.Binary                                     `json:"proto_descriptor_bin,omitzero"`
	Services                        []string                                               `json:"services,omitzero"`
	Print_options                   *GrpcJsonTranscoderTypedConfigPrintOptions             `json:"print_options,omitzero"`
	Match_incoming_request_route    *bool                                                  `json:"match_incoming_request_route,omitzero"`
	Ignored_query_parameters        []string                                               `json:"ignored_query_parameters,omitzero"`
	Auto_mapping                    *bool                                                  `json:"auto_mapping,omitzero"`
	Ignore_unknown_query_parameters *bool                                                  `json:"ignore_unknown_query_parameters,omitzero"`
	Convert_grpc_status             *bool                                                  `json:"convert_grpc_status,omitzero"`
	Url_unescape_spec               GrpcJsonTranscoderTypedConfigUrlUnescapeSpec           `json:"url_unescape_spec,omitzero"`
	Query_param_unescape_plus       *bool                                                  `json:"query_param_unescape_plus,omitzero"`
	Match_unregistered_custom_verb  *bool                                                  `json:"match_unregistered_custom_verb,omitzero"`
	Request_validation_options      *GrpcJsonTranscoderTypedConfigRequestValidationOptions `json:"request_validation_options,omitzero"`
	Case_insensitive_enum_parsing   *bool                                                  `json:"case_insensitive_enum_parsing,omitzero"`
	Max_request_body_size           *envoyCommon.UInt32                                    `json:"max_request_body_size,omitzero"`
	Max_response_body_size          *envoyCommon.UInt32                                    `json:"max_response_body_size,omitzero"`
	Type                            GrpcJsonTranscoderTypedConfigType                      `json:"@type,omitzero" validate:"required"`
}

type GrpcJsonTranscoder struct {
	Priority          *envoyCommon.Priority          `json:"priority,omitzero"`
	ExcludedWorkloads []string                       `json:"excludedWorkloads,omitzero"`
	Name              GrpcJsonTranscoderName         `json:"name,omitzero" validate:"required"`
	Typed_config      *GrpcJsonTranscoderTypedConfig `json:"typed_config,omitzero"`
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
	Type GrpcWebTypedConfigType `json:"@type,omitzero" validate:"required"`
}

type GrpcWeb struct {
	Priority          *envoyCommon.Priority `json:"priority,omitzero"`
	ExcludedWorkloads []string              `json:"excludedWorkloads,omitzero"`
	Name              GrpcWebName           `json:"name,omitzero" validate:"required"`
	Typed_config      *GrpcWebTypedConfig   `json:"typed_config,omitzero"`
}

// --- header_mutation (envoyHttp.ts header_mutation, master 894c1be3c) ---

type HeaderMutationName string

const HeaderMutationNameEnvoyFiltersHttpHeaderMutation HeaderMutationName = "envoy.filters.http.header_mutation"

type HeaderMutationTypedConfigType string

const HeaderMutationTypedConfigTypeHeaderMutation HeaderMutationTypedConfigType = "type.googleapis.com/envoy.extensions.filters.http.header_mutation.v3.HeaderMutation"

// HeaderMutationRule is the Joi alternatives {remove} | {append}; the union rule is
// registered against this named type (validate.go), as HttpFilter does.
type HeaderMutationRule any

type HeaderMutationMutations struct {
	Request_mutations  []HeaderMutationRule `json:"request_mutations,omitzero"`
	Response_mutations []HeaderMutationRule `json:"response_mutations,omitzero"`
}

type HeaderMutationTypedConfig struct {
	Type      HeaderMutationTypedConfigType `json:"@type,omitzero" validate:"required"`
	Mutations *HeaderMutationMutations      `json:"mutations,omitzero"`
}

type HeaderMutation struct {
	Priority          *envoyCommon.Priority      `json:"priority,omitzero"`
	ExcludedWorkloads []string                   `json:"excludedWorkloads,omitzero"`
	Name              HeaderMutationName         `json:"name,omitzero" validate:"required"`
	Typed_config      *HeaderMutationTypedConfig `json:"typed_config,omitzero"`
}

// HttpFilter is a union of the concrete filter types (ExtAuthz, JwtAuthn, GrpcWeb, ConnectRpc, GrpcJsonTranscoder, Cors, RateLimit, HeaderMutation). Go has no union type; kept any-backed, referenced by name.
type HttpFilter any //ts:ExtAuthz | JwtAuthn | GrpcWeb | ConnectRpc | GrpcJsonTranscoder | Cors | RateLimit | HeaderMutation

type HttpUri struct {
	Uri     string               `json:"uri" validate:"required"`
	Cluster string               `json:"cluster" validate:"required"`
	Timeout envoyCommon.Duration `json:"timeout" validate:"required"`
}

type HttpUriRestricted struct {
	Uri     string                         `json:"uri" validate:"required"`
	Cluster string                         `json:"cluster" validate:"required"`
	Timeout envoyCommon.DurationRestricted `json:"timeout,omitzero"`
}

type JwtAuthnName string

const (
	JwtAuthnNameEnvoyFiltersHttpJwtAuthn JwtAuthnName = "envoy.filters.http.jwt_authn"
)

type JwtAuthnTypedConfigProviders map[string]JwtProvider

type JwtAuthnTypedConfigFilterStateRules struct {
	Name     string            `json:"name" validate:"required"`
	Requires JwtRequirementMap `json:"requires,omitzero"`
}

type JwtAuthnTypedConfigRequirementMap map[string]JwtRequirement

type JwtAuthnTypedConfigType string

const (
	JwtAuthnTypedConfigTypeTypeGoogleapisComEnvoyExtensionsFiltersHttpJwtAuthnV3JwtAuthentication JwtAuthnTypedConfigType = "type.googleapis.com/envoy.extensions.filters.http.jwt_authn.v3.JwtAuthentication"
)

type JwtAuthnTypedConfig struct {
	Providers             JwtAuthnTypedConfigProviders         `json:"providers,omitzero"`
	Rules                 []JwtRequirementRuleRestricted       `json:"rules,omitzero"`
	Filter_state_rules    *JwtAuthnTypedConfigFilterStateRules `json:"filter_state_rules,omitzero"`
	Bypass_cors_preflight *bool                                `json:"bypass_cors_preflight,omitzero"`
	Requirement_map       JwtAuthnTypedConfigRequirementMap    `json:"requirement_map,omitzero"`
	Type                  JwtAuthnTypedConfigType              `json:"@type,omitzero" validate:"required"`
}

type JwtAuthn struct {
	Priority          *envoyCommon.Priority `json:"priority,omitzero"`
	ExcludedWorkloads []string              `json:"excludedWorkloads,omitzero"`
	Name              JwtAuthnName          `json:"name,omitzero" validate:"required"`
	Typed_config      *JwtAuthnTypedConfig  `json:"typed_config,omitzero"`
}

type JwtProviderClaimToHeaders struct {
	Header_name string `json:"header_name" validate:"required"`
	Claim_name  string `json:"claim_name" validate:"required"`
}

type JwtProviderRemoteJwksAsyncFetch struct {
	Fast_listener           *bool                `json:"fast_listener,omitzero"`
	Failed_refetch_duration envoyCommon.Duration `json:"failed_refetch_duration,omitzero"`
}

type JwtProviderRemoteJwks struct {
	Http_uri       *HttpUri                         `json:"http_uri,omitzero"`
	Cache_duration envoyCommon.Duration             `json:"cache_duration,omitzero"`
	Async_fetch    *JwtProviderRemoteJwksAsyncFetch `json:"async_fetch,omitzero"`
	Retry_policy   *envoyCommon.RetryPolicy         `json:"retry_policy,omitzero"`
}

type JwtProvider struct {
	Issuer           string                      `json:"issuer,omitzero"`
	Audiences        []string                    `json:"audiences,omitzero"`
	Claim_to_headers []JwtProviderClaimToHeaders `json:"claim_to_headers,omitzero"`
	Remote_jwks      *JwtProviderRemoteJwks      `json:"remote_jwks,omitzero"`
}

type JwtProviderUiRestrictedClaimToHeaders struct {
	Header_name string `json:"header_name" validate:"required"`
	Claim_name  string `json:"claim_name" validate:"required"`
}

type JwtProviderUiRestrictedRemoteJwks struct {
	Http_uri       *HttpUriRestricted             `json:"http_uri,omitzero"`
	Cache_duration envoyCommon.DurationRestricted `json:"cache_duration,omitzero"`
}

type JwtProviderUIRestricted struct {
	Issuer           string                                  `json:"issuer,omitzero"`
	Audiences        []string                                `json:"audiences,omitzero"`
	Claim_to_headers []JwtProviderUiRestrictedClaimToHeaders `json:"claim_to_headers,omitzero"`
	Remote_jwks      *JwtProviderUiRestrictedRemoteJwks      `json:"remote_jwks,omitzero"`
}

type JwtRequirementProviderAndAudiences struct {
	Provider_name string   `json:"provider_name,omitzero"`
	Audiences     []string `json:"audiences,omitzero"`
}

type JwtRequirementRequiresAny struct {
	Requirements []any `json:"requirements,omitzero"`
}

type JwtRequirementRequiresAll struct {
	Requirements []any `json:"requirements,omitzero"`
}

type JwtRequirement struct {
	Provider_name           string                              `json:"provider_name,omitzero"`
	Provider_and_audiences  *JwtRequirementProviderAndAudiences `json:"provider_and_audiences,omitzero"`
	Requires_any            *JwtRequirementRequiresAny          `json:"requires_any,omitzero"`
	Requires_all            *JwtRequirementRequiresAll          `json:"requires_all,omitzero"`
	Allow_missing_or_failed *envoyCommon.Empty                  `json:"allow_missing_or_failed,omitzero"`
	Allow_missing           *envoyCommon.Empty                  `json:"allow_missing,omitzero"`
}

type JwtRequirementMap map[string]JwtRequirement

type JwtRequirementRestricted struct {
	Provider_name string `json:"provider_name,omitzero"`
}

type JwtRequirementRule struct {
	Match            *envoyCommon.RouteMatch `json:"match,omitzero"`
	Requires         *JwtRequirement         `json:"requires,omitzero"`
	Requirement_name string                  `json:"requirement_name,omitzero"`
}

type JwtRequirementRuleRestricted struct {
	Match    *envoyCommon.RouteMatchRestricted `json:"match,omitzero"`
	Requires *JwtRequirementRestricted         `json:"requires,omitzero"`
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
	Domain                             string                                      `json:"domain" validate:"required"`
	Stage                              *float64                                    `json:"stage,omitzero" validate:"integer,min=0,max=10"`
	Request_type                       RateLimitTypedConfigRequestType             `json:"request_type,omitzero"`
	Timeout                            envoyCommon.Duration                        `json:"timeout,omitzero"`
	Failure_mode_deny                  *bool                                       `json:"failure_mode_deny,omitzero"`
	Rate_limited_as_resource_exhausted *bool                                       `json:"rate_limited_as_resource_exhausted,omitzero"`
	Rate_limit_service                 *RateLimitService                           `json:"rate_limit_service,omitzero"`
	Enable_x_ratelimit_headers         RateLimitTypedConfigEnableXRatelimitHeaders `json:"enable_x_ratelimit_headers,omitzero"`
	Disable_x_envoy_ratelimited_header *bool                                       `json:"disable_x_envoy_ratelimited_header,omitzero"`
	Rate_limited_status                *envoyCommon.HttpStatus                     `json:"rate_limited_status,omitzero"`
	Response_headers_to_add            []envoyCommon.HeaderValueOption             `json:"response_headers_to_add,omitzero"`
	Status_on_error                    *envoyCommon.HttpStatus                     `json:"status_on_error,omitzero"`
	Stat_prefix                        string                                      `json:"stat_prefix,omitzero"`
	Type                               RateLimitTypedConfigType                    `json:"@type,omitzero" validate:"required"`
}

type RateLimit struct {
	Priority          *envoyCommon.Priority `json:"priority,omitzero"`
	ExcludedWorkloads []string              `json:"excludedWorkloads,omitzero"`
	Name              RateLimitName         `json:"name,omitzero" validate:"required"`
	Typed_config      *RateLimitTypedConfig `json:"typed_config,omitzero"`
}

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
	Envoy_grpc       *RateLimitServiceGrpcServiceEnvoyGrpc  `json:"envoy_grpc,omitzero"`
	Google_grpc      *RateLimitServiceGrpcServiceGoogleGrpc `json:"google_grpc,omitzero"`
	Timeout          envoyCommon.Duration                   `json:"timeout,omitzero"`
	Initial_metadata []envoyCommon.HeaderValue              `json:"initial_metadata,omitzero"`
}

type RateLimitService struct {
	Grpc_service          RateLimitServiceGrpcService `json:"grpc_service" validate:"required"`
	Transport_api_version envoyCommon.ApiVersion      `json:"transport_api_version,omitzero"`
}
