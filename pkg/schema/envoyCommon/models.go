/* auto-generated */

package envoyCommon

type AdditionalAddress struct {
	Address *Address `json:"address,omitzero"`
}

type Address struct {
	Socket_address         *SocketAddress        `json:"socket_address,omitzero"`
	Pipe                   *Pipe                 `json:"pipe,omitzero"`
	Envoy_internal_address *EnvoyInternalAddress `json:"envoy_internal_address,omitzero"`
}

type ApiConfigSource struct {
	Api_type                       ApiType                `json:"api_type,omitzero"`
	Transport_api_version          ApiVersion             `json:"transport_api_version,omitzero"`
	Cluster_names                  []string               `json:"cluster_names,omitzero"`
	Grpc_services                  []GrpcService          `json:"grpc_services,omitzero"`
	Refresh_delay                  Duration               `json:"refresh_delay,omitzero"`
	Request_timeout                Duration               `json:"request_timeout,omitzero"`
	Rate_limit_settings            *RateLimitSettings     `json:"rate_limit_settings,omitzero"`
	Set_node_on_first_message_only *bool                  `json:"set_node_on_first_message_only,omitzero"`
	Config_validators              []TypedExtensionConfig `json:"config_validators,omitzero"`
}

type ApiType string

const (
	ApiTypeDeprecatedAndUnavailableDoNotUse ApiType = "DEPRECATED_AND_UNAVAILABLE_DO_NOT_USE"
	ApiTypeRest                             ApiType = "REST"
	ApiTypeGrpc                             ApiType = "GRPC"
	ApiTypeDeltaGrpc                        ApiType = "DELTA_GRPC"
)

type ApiVersion string

const (
	ApiVersionAuto ApiVersion = "AUTO"
	ApiVersionV2   ApiVersion = "V2"
	ApiVersionV3   ApiVersion = "V3"
)

// Joi.binary() fields serialize as untyped `any` in the generated TS (the
// Joi→TS converter never expressed a binary type); kept `any` here for parity.

type BindConfig struct {
	Source_address              *SocketAddress        `json:"source_address,omitzero"`
	Freebind                    *bool                 `json:"freebind,omitzero"`
	Socket_options              []SocketOption        `json:"socket_options,omitzero"`
	Extra_source_addresses      []ExtraSourceAddress  `json:"extra_source_addresses,omitzero"`
	Additional_source_addresses []SocketAddress       `json:"additional_source_addresses,omitzero"`
	Local_address_selector      *TypedExtensionConfig `json:"local_address_selector,omitzero"`
}

type CodecClientType string

const (
	CodecClientTypeHttp1 CodecClientType = "HTTP1"
	CodecClientTypeHttp2 CodecClientType = "HTTP2"
)

type ConfigSource struct {
	Path                  string            `json:"path,omitzero"`
	Path_config_source    *PathConfigSource `json:"path_config_source,omitzero"`
	Api_config_source     *ApiConfigSource  `json:"api_config_source,omitzero"`
	Initial_fetch_timeout Duration          `json:"initial_fetch_timeout,omitzero"`
	Resource_api_version  ApiVersion        `json:"resource_api_version,omitzero"`
}

type CustomHealthCheck struct {
	Name         string `json:"name" validate:"required"`
	Typed_config any    `json:"typed_config,omitzero"`
}

type DataSource struct {
	Filename             string `json:"filename,omitzero"`
	Inline_bytes         Binary `json:"inline_bytes,omitzero"`
	Inline_string        string `json:"inline_string,omitzero"`
	Environment_variable string `json:"environment_variable,omitzero"`
}

type DnsResolutionConfig struct {
	Resolvers            []Address           `json:"resolvers,omitzero"`
	Dns_resolver_options *DnsResolverOptions `json:"dns_resolver_options,omitzero"`
}

type DnsResolverOptions struct {
	Use_tcp_for_dns_lookups  *bool `json:"use_tcp_for_dns_lookups,omitzero"`
	No_default_search_domain *bool `json:"no_default_search_domain,omitzero"`
}

type DoubleMatcher struct {
	Range *DoubleRange `json:"range,omitzero"`
	Exact *float64     `json:"exact,omitzero"`
}

type DoubleRange struct {
	Start *float64 `json:"start,omitzero"`
	End   *float64 `json:"end,omitzero"`
}

// Duration is a proto duration: either a string like "5s"/"15.5s" or a DurationObject{seconds,nanos}. Go has no union type, so it stays any-backed but is referenced by name.
type Duration any //ts:string | DurationObject

// DurationObject is the object form of a proto Duration (seconds 0..315576000000, nanos 0..999999999).
type DurationObject struct {
	Seconds *int `json:"seconds,omitzero"`
	Nanos   *int `json:"nanos,omitzero"`
}

type DurationRestricted string

type Empty struct {
}

type EnvoyInternalAddress struct {
	Server_listener_name string `json:"server_listener_name" validate:"required"`
	Endpoint_id          string `json:"endpoint_id,omitzero"`
}

type ExtensionConfigSource struct {
	Config_source                        *ConfigSource `json:"config_source,omitzero"`
	Default_config                       any           `json:"default_config,omitzero"`
	Apply_default_config_without_warming *bool         `json:"apply_default_config_without_warming,omitzero"`
	Type_urls                            []string      `json:"type_urls,omitzero"`
}

type ExtraSourceAddress struct {
	Address        *Address               `json:"address,omitzero"`
	Socket_options *SocketOptionsOverride `json:"socket_options,omitzero"`
}

type FractionalPercent struct {
	Numerator   *float64 `json:"numerator,omitzero" validate:"integer"`
	Denominator *float64 `json:"denominator,omitzero" validate:"integer,valid=100|10000|1000000"`
}

type GoogleCallCredentialsServiceAccountJwtAccess struct {
	Json_key               string   `json:"json_key,omitzero"`
	Token_lifetime_seconds *float64 `json:"token_lifetime_seconds,omitzero" validate:"integer"`
}

type GoogleCallCredentialsGoogleIam struct {
	Authorization_token string `json:"authorization_token,omitzero"`
	Authority_selector  string `json:"authority_selector,omitzero"`
}

type GoogleCallCredentialsFromPlugin struct {
	Name         string `json:"name,omitzero"`
	Typed_config any    `json:"typed_config,omitzero"`
}

type GoogleCallCredentialsStsService struct {
	Token_exchange_service_uri string `json:"token_exchange_service_uri,omitzero"`
	Resource                   string `json:"resource,omitzero"`
	Audience                   string `json:"audience,omitzero"`
	Scope                      string `json:"scope,omitzero"`
	Requested_token_type       string `json:"requested_token_type,omitzero"`
	Subject_token_path         string `json:"subject_token_path" validate:"required"`
	Subject_token_type         string `json:"subject_token_type" validate:"required"`
	Actor_token_path           string `json:"actor_token_path,omitzero"`
	Actor_token_type           string `json:"actor_token_type,omitzero"`
}

type GoogleCallCredentials struct {
	Access_token               string                                        `json:"access_token,omitzero"`
	Google_compute_engine      *Empty                                        `json:"google_compute_engine,omitzero"`
	Google_refresh_token       string                                        `json:"google_refresh_token,omitzero"`
	Service_account_jwt_access *GoogleCallCredentialsServiceAccountJwtAccess `json:"service_account_jwt_access,omitzero"`
	Google_iam                 *GoogleCallCredentialsGoogleIam               `json:"google_iam,omitzero"`
	From_plugin                *GoogleCallCredentialsFromPlugin              `json:"from_plugin,omitzero"`
	Sts_service                *GoogleCallCredentialsStsService              `json:"sts_service,omitzero"`
}

type GrpcHealthCheck struct {
	Service_name     string              `json:"service_name,omitzero"`
	Authority        string              `json:"authority,omitzero"`
	Initial_metadata []HeaderValueOption `json:"initial_metadata,omitzero"`
}

type GrpcServiceEnvoyGrpc struct {
	Cluster_name string       `json:"cluster_name" validate:"required"`
	Authority    string       `json:"authority,omitzero"`
	Retry_policy *RetryPolicy `json:"retry_policy,omitzero"`
}

type GrpcServiceGoogleGrpcChannelCredentialsSslCredentials struct {
	Root_certs  *DataSource `json:"root_certs,omitzero"`
	Private_key *DataSource `json:"private_key,omitzero"`
	Cert_chain  *DataSource `json:"cert_chain,omitzero"`
}

type GrpcServiceGoogleGrpcChannelCredentials struct {
	Ssl_credentials *GrpcServiceGoogleGrpcChannelCredentialsSslCredentials `json:"ssl_credentials,omitzero"`
	Google_default  *Empty                                                 `json:"google_default,omitzero"`
}

type GrpcServiceGoogleGrpc struct {
	Target_uri               string                                   `json:"target_uri" validate:"required"`
	Channel_credentials      *GrpcServiceGoogleGrpcChannelCredentials `json:"channel_credentials,omitzero"`
	Call_credentials         []GoogleCallCredentials                  `json:"call_credentials,omitzero"`
	Stat_prefix              string                                   `json:"stat_prefix" validate:"required"`
	Credentials_factory_name string                                   `json:"credentials_factory_name,omitzero"`
	Config                   Struct                                   `json:"config,omitzero"`
}

type GrpcService struct {
	Envoy_grpc       *GrpcServiceEnvoyGrpc  `json:"envoy_grpc,omitzero"`
	Google_grpc      *GrpcServiceGoogleGrpc `json:"google_grpc,omitzero"`
	Timeout          Duration               `json:"timeout,omitzero"`
	Initial_metadata []HeaderValue          `json:"initial_metadata,omitzero"`
}

type HeaderAppendAction string

const (
	HeaderAppendActionAppendIfExistsOrAdd    HeaderAppendAction = "APPEND_IF_EXISTS_OR_ADD"
	HeaderAppendActionAddIfAbsent            HeaderAppendAction = "ADD_IF_ABSENT"
	HeaderAppendActionOverwriteIfExistsOrAdd HeaderAppendAction = "OVERWRITE_IF_EXISTS_OR_ADD"
	HeaderAppendActionOverwriteIfExists      HeaderAppendAction = "OVERWRITE_IF_EXISTS"
)

type HeaderMatcher struct {
	Name                          string         `json:"name" validate:"required"`
	Safe_regex_match              *RegexMatcher  `json:"safe_regex_match,omitzero"`
	Range_match                   *IntRange      `json:"range_match,omitzero"`
	Present_match                 *bool          `json:"present_match,omitzero"`
	String_match                  *StringMatcher `json:"string_match,omitzero"`
	Invert_match                  *bool          `json:"invert_match,omitzero"`
	Treat_missing_header_as_empty *bool          `json:"treat_missing_header_as_empty,omitzero"`
}

type HeaderValue struct {
	Key       string `json:"key" validate:"required"`
	Value     string `json:"value,omitzero"`
	Raw_value Binary `json:"raw_value,omitzero"`
}

type HeaderValueOption struct {
	Header           *HeaderValue       `json:"header,omitzero"`
	Append           *bool              `json:"append,omitzero"`
	Append_action    HeaderAppendAction `json:"append_action,omitzero"`
	Keep_empty_value *bool              `json:"keep_empty_value,omitzero"`
}

type HealthCheck struct {
	Timeout                          Duration               `json:"timeout,omitzero"`
	Interval                         Duration               `json:"interval,omitzero"`
	Initial_jitter                   Duration               `json:"initial_jitter,omitzero"`
	Interval_jitter                  Duration               `json:"interval_jitter,omitzero"`
	Interval_jitter_percent          *UInt32                `json:"interval_jitter_percent,omitzero"`
	Unhealthy_threshold              *UInt32                `json:"unhealthy_threshold,omitzero"`
	Healthy_threshold                *UInt32                `json:"healthy_threshold,omitzero"`
	Reuse_connection                 *bool                  `json:"reuse_connection,omitzero"`
	Http_health_check                *HttpHealthCheck       `json:"http_health_check,omitzero"`
	Tcp_health_check                 *TcpHealthCheck        `json:"tcp_health_check,omitzero"`
	Grpc_health_check                *GrpcHealthCheck       `json:"grpc_health_check,omitzero"`
	Custom_health_check              *CustomHealthCheck     `json:"custom_health_check,omitzero"`
	No_traffic_interval              Duration               `json:"no_traffic_interval,omitzero"`
	No_traffic_healthy_interval      Duration               `json:"no_traffic_healthy_interval,omitzero"`
	Unhealthy_interval               Duration               `json:"unhealthy_interval,omitzero"`
	Healthy_edge_interval            Duration               `json:"healthy_edge_interval,omitzero"`
	Event_log_path                   string                 `json:"event_log_path,omitzero"`
	Event_logger                     []TypedExtensionConfig `json:"event_logger,omitzero"`
	Always_log_health_check_failures *bool                  `json:"always_log_health_check_failures,omitzero"`
	Tls_options                      *HealthCheckTlsOptions `json:"tls_options,omitzero"`
	Transport_socket_match_criteria  Struct                 `json:"transport_socket_match_criteria,omitzero"`
}

type HealthCheckPayload struct {
	Text   string `json:"text,omitzero"`
	Binary Binary `json:"binary,omitzero"`
}

type HealthCheckTlsOptions struct {
	Alpn_protocols []string `json:"alpn_protocols,omitzero"`
} //ts:export

type HealthStatus string

const (
	HealthStatusUnknown   HealthStatus = "UNKNOWN"
	HealthStatusHealthy   HealthStatus = "HEALTHY"
	HealthStatusUnhealthy HealthStatus = "UNHEALTHY"
	HealthStatusDraining  HealthStatus = "DRAINING"
	HealthStatusTimeout   HealthStatus = "TIMEOUT"
	HealthStatusDegraded  HealthStatus = "DEGRADED"
)

type HealthStatusSet struct {
	Statuses []HealthStatus `json:"statuses,omitzero"`
}

type Http1ProtocolOptionsHeaderKeyFormatProperCaseWords struct {
	Interval                 Duration `json:"interval,omitzero"`
	Timeout                  Duration `json:"timeout,omitzero"`
	Interval_jitter          *Percent `json:"interval_jitter,omitzero"`
	Connection_idle_interval Duration `json:"connection_idle_interval,omitzero"`
}

type Http1ProtocolOptionsHeaderKeyFormat struct {
	Proper_case_words  *Http1ProtocolOptionsHeaderKeyFormatProperCaseWords `json:"proper_case_words,omitzero"`
	Stateful_formatter *TypedExtensionConfig                               `json:"stateful_formatter,omitzero"`
}

type Http1ProtocolOptions struct {
	Allow_absolute_url                            *bool                                `json:"allow_absolute_url,omitzero"`
	Accept_http_10                                *bool                                `json:"accept_http_10,omitzero"`
	Default_host_for_http_10                      string                               `json:"default_host_for_http_10,omitzero"`
	Header_key_format                             *Http1ProtocolOptionsHeaderKeyFormat `json:"header_key_format,omitzero"`
	Enable_trailers                               *bool                                `json:"enable_trailers,omitzero"`
	Allow_chunked_length                          *bool                                `json:"allow_chunked_length,omitzero"`
	Override_stream_error_on_invalid_http_message *bool                                `json:"override_stream_error_on_invalid_http_message,omitzero"`
	Send_fully_qualified_url                      *bool                                `json:"send_fully_qualified_url,omitzero"`
}

type Http2ProtocolOptions struct {
	Hpack_table_size                                     *UInt32            `json:"hpack_table_size,omitzero"`
	Max_concurrent_streams                               *UInt32            `json:"max_concurrent_streams,omitzero"`
	Initial_stream_window_size                           *UInt32            `json:"initial_stream_window_size,omitzero"`
	Initial_connection_window_size                       *UInt32            `json:"initial_connection_window_size,omitzero"`
	Allow_connect                                        *bool              `json:"allow_connect,omitzero"`
	Max_outbound_frames                                  *UInt32            `json:"max_outbound_frames,omitzero"`
	Max_outbound_control_frames                          *UInt32            `json:"max_outbound_control_frames,omitzero"`
	Max_consecutive_inbound_frames_with_empty_payload    *UInt32            `json:"max_consecutive_inbound_frames_with_empty_payload,omitzero"`
	Max_inbound_priority_frames_per_stream               *UInt32            `json:"max_inbound_priority_frames_per_stream,omitzero"`
	Max_inbound_window_update_frames_per_data_frame_sent *UInt32            `json:"max_inbound_window_update_frames_per_data_frame_sent,omitzero"`
	Stream_error_on_invalid_http_messaging               *bool              `json:"stream_error_on_invalid_http_messaging,omitzero"`
	Override_stream_error_on_invalid_http_message        *bool              `json:"override_stream_error_on_invalid_http_message,omitzero"`
	Connection_keepalive                                 *KeepaliveSettings `json:"connection_keepalive,omitzero"`
}

type HttpHealthCheck struct {
	Host                      string               `json:"host,omitzero"`
	Path                      string               `json:"path" validate:"required"`
	Receive                   []HealthCheckPayload `json:"receive,omitzero"`
	Response_buffer_size      *UInt64              `json:"response_buffer_size,omitzero"`
	Request_headers_to_add    []HeaderValueOption  `json:"request_headers_to_add,omitzero"`
	Request_headers_to_remove []string             `json:"request_headers_to_remove,omitzero"`
	Expected_statuses         []IntRange           `json:"expected_statuses,omitzero"`
	Retriable_statuses        []IntRange           `json:"retriable_statuses,omitzero"`
	Codec_client_type         CodecClientType      `json:"codec_client_type,omitzero"`
	Service_name_matcher      *StringMatcher       `json:"service_name_matcher,omitzero"`
	Method                    RequestMethod        `json:"method,omitzero"`
}

type HttpProtocolOptionsHeadersWithUnderscoresAction string

const (
	HttpProtocolOptionsHeadersWithUnderscoresActionAllow         HttpProtocolOptionsHeadersWithUnderscoresAction = "ALLOW"
	HttpProtocolOptionsHeadersWithUnderscoresActionRejectRequest HttpProtocolOptionsHeadersWithUnderscoresAction = "REJECT_REQUEST"
	HttpProtocolOptionsHeadersWithUnderscoresActionDropHeader    HttpProtocolOptionsHeadersWithUnderscoresAction = "DROP_HEADER"
)

type HttpProtocolOptions struct {
	Idle_timeout                    Duration                                        `json:"idle_timeout,omitzero"`
	Max_connection_duration         Duration                                        `json:"max_connection_duration,omitzero"`
	Max_headers_count               *UInt32                                         `json:"max_headers_count,omitzero"`
	Max_stream_duration             Duration                                        `json:"max_stream_duration,omitzero"`
	Headers_with_underscores_action HttpProtocolOptionsHeadersWithUnderscoresAction `json:"headers_with_underscores_action,omitzero"`
}

type HttpStatus struct {
	Code StatusCode `json:"code,omitzero"`
}

type IntRange struct {
	Start *float64 `json:"start,omitzero"`
	End   *float64 `json:"end,omitzero"`
}

type KeepaliveSettings struct {
	Interval                 Duration `json:"interval,omitzero"`
	Timeout                  Duration `json:"timeout" validate:"required"`
	Interval_jitter          *Percent `json:"interval_jitter,omitzero"`
	Connection_idle_interval Duration `json:"connection_idle_interval,omitzero"`
}

type ListStringMatcher struct {
	Patterns []StringMatcher `json:"patterns,omitzero"`
}

type ListValue []any

type Locality struct {
	Region   string `json:"region,omitzero"`
	Zone     string `json:"zone,omitzero"`
	Sub_zone string `json:"sub_zone,omitzero"`
}

type MetadataMatcherPath struct {
	Key string `json:"key" validate:"required"`
}

type MetadataMatcherValue struct {
	Null_match    any            `json:"null_match,omitzero"`
	Double_match  *DoubleMatcher `json:"double_match,omitzero"`
	String_match  *StringMatcher `json:"string_match,omitzero"`
	Bool_match    *bool          `json:"bool_match,omitzero"`
	Present_match *bool          `json:"present_match,omitzero"`
}

type MetadataMatcher struct {
	Filter string                `json:"filter" validate:"required"`
	Path   []MetadataMatcherPath `json:"path,omitzero"`
	Value  MetadataMatcherValue  `json:"value" validate:"required"`
	Invert *bool                 `json:"invert,omitzero"`
}

type PathConfigSource struct {
	Path              string            `json:"path" validate:"required"`
	Watched_directory *WatchedDirectory `json:"watched_directory,omitzero"`
}

type Percent float64

type Pipe struct {
	Path string   `json:"path" validate:"required"`
	Mode *float64 `json:"mode,omitzero" validate:"integer"`
}

type Port float64

type Priority float64

type QueryParameterMatcher struct {
	Name          string         `json:"name" validate:"required"`
	String_match  *StringMatcher `json:"string_match,omitzero"`
	Present_match *bool          `json:"present_match,omitzero"`
}

type RateLimitSettings struct {
	Max_tokens *UInt32  `json:"max_tokens,omitzero"`
	Fill_rate  *float64 `json:"fill_rate,omitzero"`
}

type RegexMatcherGoogleRe2 struct {
	Max_program_size *float64 `json:"max_program_size,omitzero" validate:"integer"`
}

type RegexMatcher struct {
	Google_re2 *RegexMatcherGoogleRe2 `json:"google_re2,omitzero"`
	Regex      string                 `json:"regex" validate:"required"`
}

type RequestMethod string

const (
	RequestMethodMethodUnspecified RequestMethod = "METHOD_UNSPECIFIED"
	RequestMethodGet               RequestMethod = "GET"
	RequestMethodHead              RequestMethod = "HEAD"
	RequestMethodPost              RequestMethod = "POST"
	RequestMethodPut               RequestMethod = "PUT"
	RequestMethodDelete            RequestMethod = "DELETE"
	RequestMethodConnect           RequestMethod = "CONNECT"
	RequestMethodOptions           RequestMethod = "OPTIONS"
	RequestMethodTrace             RequestMethod = "TRACE"
	RequestMethodPatch             RequestMethod = "PATCH"
)

type RetryPolicy struct {
	Retry_back_off *TimeInterval `json:"retry_back_off,omitzero"`
	Num_retries    *float64      `json:"num_retries,omitzero" validate:"default=1"`
}

type RouteMatchGrpc struct {
	Presented *bool `json:"presented,omitzero"`
	Validated *bool `json:"validated,omitzero"`
}

type RouteMatchTlsContext struct {
	Presented *bool `json:"presented,omitzero"`
	Validated *bool `json:"validated,omitzero"`
}

type RouteMatch struct {
	Prefix                string                    `json:"prefix,omitzero"`
	Path                  string                    `json:"path,omitzero"`
	Safe_regex            *RegexMatcher             `json:"safe_regex,omitzero"`
	Path_separated_prefix string                    `json:"path_separated_prefix,omitzero"`
	Path_match_policy     *TypedExtensionConfig     `json:"path_match_policy,omitzero"`
	Case_sensitive        *bool                     `json:"case_sensitive,omitzero"`
	Runtime_fraction      *RuntimeFractionalPercent `json:"runtime_fraction,omitzero"`
	Headers               []HeaderMatcher           `json:"headers,omitzero"`
	Query_parameters      []QueryParameterMatcher   `json:"query_parameters,omitzero"`
	Grpc                  *RouteMatchGrpc           `json:"grpc,omitzero"`
	Tls_context           *RouteMatchTlsContext     `json:"tls_context,omitzero"`
	Dynamic_metadata      []MetadataMatcher         `json:"dynamic_metadata,omitzero"`
}

type RouteMatchRestricted struct {
	Prefix  string          `json:"prefix,omitzero"`
	Headers []HeaderMatcher `json:"headers,omitzero"`
}

type RoutingPriority string

const (
	RoutingPriorityDefault RoutingPriority = "DEFAULT"
	RoutingPriorityHigh    RoutingPriority = "HIGH"
)

type RuntimeDouble struct {
	Default_value *float64 `json:"default_value,omitzero"`
	Runtime_key   string   `json:"runtime_key" validate:"required"`
}

type RuntimeFeatureFlag struct {
	Default_value bool   `json:"default_value" validate:"required"`
	Runtime_key   string `json:"runtime_key" validate:"required"`
}

type RuntimeFractionalPercent struct {
	Default_value *FractionalPercent `json:"default_value,omitzero"`
	Runtime_key   string             `json:"runtime_key" validate:"required"`
}

type SocketAddressProtocol string

const (
	SocketAddressProtocolTcp SocketAddressProtocol = "TCP"
	SocketAddressProtocolUdp SocketAddressProtocol = "UDP"
)

type SocketAddress struct {
	Protocol      SocketAddressProtocol `json:"protocol,omitzero"`
	Address       string                `json:"address" validate:"required"`
	Port_value    *float64              `json:"port_value,omitzero" validate:"integer"`
	Named_port    string                `json:"named_port,omitzero"`
	Resolver_name any                   `json:"resolver_name,omitzero"`
	Ipv4_compat   *bool                 `json:"ipv4_compat,omitzero"`
}

type SocketOptionState string

const (
	SocketOptionStateStatePrebind   SocketOptionState = "STATE_PREBIND"
	SocketOptionStateStateBound     SocketOptionState = "STATE_BOUND"
	SocketOptionStateStateListening SocketOptionState = "STATE_LISTENING"
)

type SocketOption struct {
	Description string            `json:"description,omitzero"`
	Level       *float64          `json:"level,omitzero" validate:"integer"`
	Name        *float64          `json:"name,omitzero" validate:"integer"`
	Int_value   *float64          `json:"int_value,omitzero" validate:"integer"`
	Buf_value   Binary            `json:"buf_value,omitzero"`
	State       SocketOptionState `json:"state,omitzero"`
}

type SocketOptionsOverride struct {
	Socket_options []SocketOption `json:"socket_options,omitzero"`
}

type StatusCode string

const (
	StatusCodeEmpty                         StatusCode = "Empty"
	StatusCodeContinue                      StatusCode = "Continue"
	StatusCodeOk                            StatusCode = "OK"
	StatusCodeCreated                       StatusCode = "Created"
	StatusCodeAccepted                      StatusCode = "Accepted"
	StatusCodeNonAuthoritativeInformation   StatusCode = "NonAuthoritativeInformation"
	StatusCodeNoContent                     StatusCode = "NoContent"
	StatusCodeResetContent                  StatusCode = "ResetContent"
	StatusCodePartialContent                StatusCode = "PartialContent"
	StatusCodeMultiStatus                   StatusCode = "MultiStatus"
	StatusCodeAlradyReported                StatusCode = "AlradyReported"
	StatusCodeImUsed                        StatusCode = "IMUsed"
	StatusCodeMultipleChoices               StatusCode = "MultipleChoices"
	StatusCodeMovedPermanently              StatusCode = "MovedPermanently"
	StatusCodeFound                         StatusCode = "Found"
	StatusCodeSeeOther                      StatusCode = "SeeOther"
	StatusCodeNotModified                   StatusCode = "NotModified"
	StatusCodeUseProxy                      StatusCode = "UseProxy"
	StatusCodeTemporaryRedirect             StatusCode = "TemporaryRedirect"
	StatusCodePermanentRedirect             StatusCode = "PermanentRedirect"
	StatusCodeBadRequest                    StatusCode = "BadRequest"
	StatusCodeUnauthroized                  StatusCode = "Unauthroized"
	StatusCodePaymentRequired               StatusCode = "PaymentRequired"
	StatusCodeForbidden                     StatusCode = "Forbidden"
	StatusCodeNotFound                      StatusCode = "NotFound"
	StatusCodeMethodNotAllowed              StatusCode = "MethodNotAllowed"
	StatusCodeNotAcceptable                 StatusCode = "NotAcceptable"
	StatusCodeProxyAuthenticationRequired   StatusCode = "ProxyAuthenticationRequired"
	StatusCodeRequestTimeout                StatusCode = "RequestTimeout"
	StatusCodeConflict                      StatusCode = "Conflict"
	StatusCodeGone                          StatusCode = "Gone"
	StatusCodeLengthRequired                StatusCode = "LengthRequired"
	StatusCodePreconditionFailed            StatusCode = "PreconditionFailed"
	StatusCodePayloadTooLarge               StatusCode = "PayloadTooLarge"
	StatusCodeUriTooLong                    StatusCode = "URITooLong"
	StatusCodeUnsupportedMediaType          StatusCode = "UnsupportedMediaType"
	StatusCodeRangeNotSatisfiable           StatusCode = "RangeNotSatisfiable"
	StatusCodeExpectationFailed             StatusCode = "ExpectationFailed"
	StatusCodeMisdirectedRequest            StatusCode = "MisdirectedRequest"
	StatusCodeUnprocessableEntity           StatusCode = "UnprocessableEntity"
	StatusCodeLocked                        StatusCode = "Locked"
	StatusCodeFailedDependency              StatusCode = "FailedDependency"
	StatusCodeUpgradeRequired               StatusCode = "UpgradeRequired"
	StatusCodePreconditionRequired          StatusCode = "PreconditionRequired"
	StatusCodeTooManyRequests               StatusCode = "TooManyRequests"
	StatusCodeRequestHeaderFieldsTooLarge   StatusCode = "RequestHeaderFieldsTooLarge"
	StatusCodeInternalServerError           StatusCode = "InternalServerError"
	StatusCodeNotImplemented                StatusCode = "NotImplemented"
	StatusCodeBadGateway                    StatusCode = "BadGateway"
	StatusCodeServiceUnavailable            StatusCode = "ServiceUnavailable"
	StatusCodeGatewayTimeout                StatusCode = "GatewayTimeout"
	StatusCodeHttpVersionNotSupported       StatusCode = "HTTPVersionNotSupported"
	StatusCodeVariantAlsoNegotiates         StatusCode = "VariantAlsoNegotiates"
	StatusCodeInsufficientStorage           StatusCode = "InsufficientStorage"
	StatusCodeLoopDetected                  StatusCode = "LoopDetected"
	StatusCodeNotExtended                   StatusCode = "NotExtended"
	StatusCodeNetworkAuthenticationRequired StatusCode = "NetworkAuthenticationRequired"
)

type StringMatcher struct {
	Exact       string        `json:"exact,omitzero"`
	Prefix      string        `json:"prefix,omitzero"`
	Suffix      string        `json:"suffix,omitzero"`
	Safe_regex  *RegexMatcher `json:"safe_regex,omitzero"`
	Contains    string        `json:"contains,omitzero"`
	Ignore_case *bool         `json:"ignore_case,omitzero"`
}

type Struct map[string]any

type TcpHealthCheck struct {
	Send    *HealthCheckPayload  `json:"send,omitzero"`
	Receive []HealthCheckPayload `json:"receive,omitzero"`
}

type TcpKeepalive struct {
	Keepalive_probes   *UInt32 `json:"keepalive_probes,omitzero"`
	Keepalive_time     *UInt32 `json:"keepalive_time,omitzero"`
	Keepalive_interval *UInt32 `json:"keepalive_interval,omitzero"`
}

type TimeInterval struct {
	Base_interval Duration `json:"base_interval" validate:"required"`
	Max_interval  Duration `json:"max_interval,omitzero"`
}

type TransportSocketTypedConfigType string

const (
	TransportSocketTypedConfigTypeTypeGoogleapisComEnvoyExtensionsTransportSocketsTlsV3UpstreamTlsContext TransportSocketTypedConfigType = "type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.UpstreamTlsContext"
)

type TransportSocketTypedConfig struct {
	Type TransportSocketTypedConfigType `json:"@type,omitzero"`
	Sni  string                         `json:"sni,omitzero"`
}

type TransportSocket struct {
	Name         string                      `json:"name" validate:"required"`
	Typed_config *TransportSocketTypedConfig `json:"typed_config,omitzero"`
}

type TypedExtensionConfig struct {
	Name         string `json:"name" validate:"required"`
	Typed_config any    `json:"typed_config,omitzero"`
}

type UInt32 float64

type UInt64 float64

type UpstreamHttpProtocolOptions struct {
	Auto_sni                 *bool  `json:"auto_sni,omitzero"`
	Auto_san_validation      *bool  `json:"auto_san_validation,omitzero"`
	Override_auto_sni_header string `json:"override_auto_sni_header,omitzero"`
}

type ValueMatcher struct {
	Null_match    any            `json:"null_match,omitzero"`
	Double_match  *DoubleMatcher `json:"double_match,omitzero"`
	String_match  *StringMatcher `json:"string_match,omitzero"`
	Bool_match    *bool          `json:"bool_match,omitzero"`
	Present_match *bool          `json:"present_match,omitzero"`
}

type WatchedDirectory struct {
	Path string `json:"path" validate:"required"`
}
