/* auto-generated */

package envoyCommon

type AdditionalAddress struct {
	Address Address `json:"address,omitempty"`
}

type Address struct {
	Socket_address         SocketAddress        `json:"socket_address,omitempty"`
	Pipe                   Pipe                 `json:"pipe,omitempty"`
	Envoy_internal_address EnvoyInternalAddress `json:"envoy_internal_address,omitempty"`
}

type ApiConfigSource struct {
	Api_type                       ApiType                `json:"api_type,omitempty"`
	Transport_api_version          ApiVersion             `json:"transport_api_version,omitempty"`
	Cluster_names                  []string               `json:"cluster_names,omitempty"`
	Grpc_services                  []GrpcService          `json:"grpc_services,omitempty"`
	Refresh_delay                  Duration               `json:"refresh_delay,omitempty"`
	Request_timeout                Duration               `json:"request_timeout,omitempty"`
	Rate_limit_settings            RateLimitSettings      `json:"rate_limit_settings,omitempty"`
	Set_node_on_first_message_only bool                   `json:"set_node_on_first_message_only,omitempty"`
	Config_validators              []TypedExtensionConfig `json:"config_validators,omitempty"`
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

type BindConfig struct {
	Source_address              SocketAddress        `json:"source_address,omitempty"`
	Freebind                    bool                 `json:"freebind,omitempty"`
	Socket_options              []SocketOption       `json:"socket_options,omitempty"`
	Extra_source_addresses      []ExtraSourceAddress `json:"extra_source_addresses,omitempty"`
	Additional_source_addresses []SocketAddress      `json:"additional_source_addresses,omitempty"`
	Local_address_selector      TypedExtensionConfig `json:"local_address_selector,omitempty"`
}

type CodecClientType string

const (
	CodecClientTypeHttp1 CodecClientType = "HTTP1"
	CodecClientTypeHttp2 CodecClientType = "HTTP2"
)

type ConfigSource struct {
	Path                  string           `json:"path,omitempty"`
	Path_config_source    PathConfigSource `json:"path_config_source,omitempty"`
	Api_config_source     ApiConfigSource  `json:"api_config_source,omitempty"`
	Initial_fetch_timeout Duration         `json:"initial_fetch_timeout,omitempty"`
	Resource_api_version  ApiVersion       `json:"resource_api_version,omitempty"`
}

type CustomHealthCheck struct {
	Name         string `json:"name"`
	Typed_config any    `json:"typed_config,omitempty"`
}

type DataSource struct {
	Filename             string `json:"filename,omitempty"`
	Inline_bytes         any/* TODO: [object Object]*/ `json:"inline_bytes,omitempty"`
	Inline_string        string `json:"inline_string,omitempty"`
	Environment_variable string `json:"environment_variable,omitempty"`
}

type DnsResolutionConfig struct {
	Resolvers            []Address          `json:"resolvers,omitempty"`
	Dns_resolver_options DnsResolverOptions `json:"dns_resolver_options,omitempty"`
}

type DnsResolverOptions struct {
	Use_tcp_for_dns_lookups  bool `json:"use_tcp_for_dns_lookups,omitempty"`
	No_default_search_domain bool `json:"no_default_search_domain,omitempty"`
}

type DoubleMatcher struct {
	Range DoubleRange `json:"range,omitempty"`
	Exact float32     `json:"exact"`
}

type DoubleRange struct {
	Start float32 `json:"start"`
	End   float32 `json:"end"`
}

type Duration any /* TODO: [object Object]*/

type DurationRestricted string

type Empty struct {
}

type EnvoyInternalAddress struct {
	Server_listener_name string `json:"server_listener_name"`
	Endpoint_id          string `json:"endpoint_id,omitempty"`
}

type ExtensionConfigSource struct {
	Config_source                        ConfigSource `json:"config_source,omitempty"`
	Default_config                       any          `json:"default_config,omitempty"`
	Apply_default_config_without_warming bool         `json:"apply_default_config_without_warming,omitempty"`
	Type_urls                            []string     `json:"type_urls,omitempty"`
}

type ExtraSourceAddress struct {
	Address        Address               `json:"address,omitempty"`
	Socket_options SocketOptionsOverride `json:"socket_options,omitempty"`
}

type FractionalPercent struct {
	Numerator   float32 `json:"numerator"`
	Denominator float32 `json:"denominator"`
}

type GoogleCallCredentialsServiceAccountJwtAccess struct {
	Json_key               string  `json:"json_key,omitempty"`
	Token_lifetime_seconds float32 `json:"token_lifetime_seconds"`
}

type GoogleCallCredentialsGoogleIam struct {
	Authorization_token string `json:"authorization_token,omitempty"`
	Authority_selector  string `json:"authority_selector,omitempty"`
}

type GoogleCallCredentialsFromPlugin struct {
	Name         string `json:"name,omitempty"`
	Typed_config any    `json:"typed_config,omitempty"`
}

type GoogleCallCredentialsStsService struct {
	Token_exchange_service_uri string `json:"token_exchange_service_uri,omitempty"`
	Resource                   string `json:"resource,omitempty"`
	Audience                   string `json:"audience,omitempty"`
	Scope                      string `json:"scope,omitempty"`
	Requested_token_type       string `json:"requested_token_type,omitempty"`
	Subject_token_path         string `json:"subject_token_path"`
	Subject_token_type         string `json:"subject_token_type"`
	Actor_token_path           string `json:"actor_token_path,omitempty"`
	Actor_token_type           string `json:"actor_token_type,omitempty"`
}

type GoogleCallCredentials struct {
	Access_token               string                                       `json:"access_token,omitempty"`
	Google_compute_engine      Empty                                        `json:"google_compute_engine,omitempty"`
	Google_refresh_token       string                                       `json:"google_refresh_token,omitempty"`
	Service_account_jwt_access GoogleCallCredentialsServiceAccountJwtAccess `json:"service_account_jwt_access,omitempty"`
	Google_iam                 GoogleCallCredentialsGoogleIam               `json:"google_iam,omitempty"`
	From_plugin                GoogleCallCredentialsFromPlugin              `json:"from_plugin,omitempty"`
	Sts_service                GoogleCallCredentialsStsService              `json:"sts_service,omitempty"`
}

type GrpcHealthCheck struct {
	Service_name     string              `json:"service_name,omitempty"`
	Authority        string              `json:"authority,omitempty"`
	Initial_metadata []HeaderValueOption `json:"initial_metadata,omitempty"`
}

type GrpcServiceEnvoyGrpc struct {
	Cluster_name string      `json:"cluster_name"`
	Authority    string      `json:"authority,omitempty"`
	Retry_policy RetryPolicy `json:"retry_policy,omitempty"`
}

type GrpcServiceGoogleGrpcChannelCredentialsSslCredentials struct {
	Root_certs  DataSource `json:"root_certs,omitempty"`
	Private_key DataSource `json:"private_key,omitempty"`
	Cert_chain  DataSource `json:"cert_chain,omitempty"`
}

type GrpcServiceGoogleGrpcChannelCredentials struct {
	Ssl_credentials GrpcServiceGoogleGrpcChannelCredentialsSslCredentials `json:"ssl_credentials,omitempty"`
	Google_default  Empty                                                 `json:"google_default,omitempty"`
}

type GrpcServiceGoogleGrpc struct {
	Target_uri               string                                  `json:"target_uri"`
	Channel_credentials      GrpcServiceGoogleGrpcChannelCredentials `json:"channel_credentials,omitempty"`
	Call_credentials         []GoogleCallCredentials                 `json:"call_credentials,omitempty"`
	Stat_prefix              string                                  `json:"stat_prefix"`
	Credentials_factory_name string                                  `json:"credentials_factory_name,omitempty"`
	Config                   Struct                                  `json:"config,omitempty"`
}

type GrpcService struct {
	Envoy_grpc       GrpcServiceEnvoyGrpc  `json:"envoy_grpc,omitempty"`
	Google_grpc      GrpcServiceGoogleGrpc `json:"google_grpc,omitempty"`
	Timeout          any/* TODO: [object Object]*/ `json:"timeout,omitempty"`
	Initial_metadata []HeaderValue `json:"initial_metadata,omitempty"`
}

type HeaderAppendAction string

const (
	HeaderAppendActionAppendIfExistsOrAdd    HeaderAppendAction = "APPEND_IF_EXISTS_OR_ADD"
	HeaderAppendActionAddIfAbsent            HeaderAppendAction = "ADD_IF_ABSENT"
	HeaderAppendActionOverwriteIfExistsOrAdd HeaderAppendAction = "OVERWRITE_IF_EXISTS_OR_ADD"
	HeaderAppendActionOverwriteIfExists      HeaderAppendAction = "OVERWRITE_IF_EXISTS"
)

type HeaderMatcher struct {
	Name                          string        `json:"name"`
	Safe_regex_match              RegexMatcher  `json:"safe_regex_match,omitempty"`
	Range_match                   IntRange      `json:"range_match,omitempty"`
	Present_match                 bool          `json:"present_match,omitempty"`
	String_match                  StringMatcher `json:"string_match,omitempty"`
	Invert_match                  bool          `json:"invert_match,omitempty"`
	Treat_missing_header_as_empty bool          `json:"treat_missing_header_as_empty,omitempty"`
}

type HeaderValue struct {
	Key       string `json:"key"`
	Value     string `json:"value,omitempty"`
	Raw_value any/* TODO: [object Object]*/ `json:"raw_value,omitempty"`
}

type HeaderValueOption struct {
	Header           HeaderValue        `json:"header,omitempty"`
	Append           bool               `json:"append,omitempty"`
	Append_action    HeaderAppendAction `json:"append_action,omitempty"`
	Keep_empty_value bool               `json:"keep_empty_value,omitempty"`
}

type HealthCheck struct {
	Timeout                          Duration               `json:"timeout,omitempty"`
	Interval                         Duration               `json:"interval,omitempty"`
	Initial_jitter                   Duration               `json:"initial_jitter,omitempty"`
	Interval_jitter                  Duration               `json:"interval_jitter,omitempty"`
	Interval_jitter_percent          UInt32                 `json:"interval_jitter_percent,omitempty"`
	Unhealthy_threshold              UInt32                 `json:"unhealthy_threshold,omitempty"`
	Healthy_threshold                UInt32                 `json:"healthy_threshold,omitempty"`
	Reuse_connection                 bool                   `json:"reuse_connection,omitempty"`
	Http_health_check                HttpHealthCheck        `json:"http_health_check,omitempty"`
	Tcp_health_check                 TcpHealthCheck         `json:"tcp_health_check,omitempty"`
	Grpc_health_check                GrpcHealthCheck        `json:"grpc_health_check,omitempty"`
	Custom_health_check              CustomHealthCheck      `json:"custom_health_check,omitempty"`
	No_traffic_interval              Duration               `json:"no_traffic_interval,omitempty"`
	No_traffic_healthy_interval      Duration               `json:"no_traffic_healthy_interval,omitempty"`
	Unhealthy_interval               Duration               `json:"unhealthy_interval,omitempty"`
	Healthy_edge_interval            Duration               `json:"healthy_edge_interval,omitempty"`
	Event_log_path                   string                 `json:"event_log_path,omitempty"`
	Event_logger                     []TypedExtensionConfig `json:"event_logger,omitempty"`
	Always_log_health_check_failures bool                   `json:"always_log_health_check_failures,omitempty"`
	Tls_options                      HealthCheckTlsOptions  `json:"tls_options,omitempty"`
	Transport_socket_match_criteria  Struct                 `json:"transport_socket_match_criteria,omitempty"`
}

type HealthCheckPayload struct {
	Text   string `json:"text,omitempty"`
	Binary any/* TODO: [object Object]*/ `json:"binary,omitempty"`
}

type HealthCheckTlsOptions struct {
	Alpn_protocols []string `json:"alpn_protocols,omitempty"`
}

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
	Statuses []HealthStatus `json:"statuses,omitempty"`
}

type Http1ProtocolOptionsHeaderKeyFormatProperCaseWords struct {
	Interval                 Duration `json:"interval,omitempty"`
	Timeout                  Duration `json:"timeout,omitempty"`
	Interval_jitter          Percent  `json:"interval_jitter,omitempty"`
	Connection_idle_interval Duration `json:"connection_idle_interval,omitempty"`
}

type Http1ProtocolOptionsHeaderKeyFormat struct {
	Proper_case_words  Http1ProtocolOptionsHeaderKeyFormatProperCaseWords `json:"proper_case_words,omitempty"`
	Stateful_formatter TypedExtensionConfig                               `json:"stateful_formatter,omitempty"`
}

type Http1ProtocolOptions struct {
	Allow_absolute_url                            bool                                `json:"allow_absolute_url,omitempty"`
	Accept_http_10                                bool                                `json:"accept_http_10,omitempty"`
	Default_host_for_http_10                      string                              `json:"default_host_for_http_10,omitempty"`
	Header_key_format                             Http1ProtocolOptionsHeaderKeyFormat `json:"header_key_format,omitempty"`
	Enable_trailers                               bool                                `json:"enable_trailers,omitempty"`
	Allow_chunked_length                          bool                                `json:"allow_chunked_length,omitempty"`
	Override_stream_error_on_invalid_http_message bool                                `json:"override_stream_error_on_invalid_http_message,omitempty"`
	Send_fully_qualified_url                      bool                                `json:"send_fully_qualified_url,omitempty"`
}

type Http2ProtocolOptions struct {
	Hpack_table_size                                     UInt32            `json:"hpack_table_size,omitempty"`
	Max_concurrent_streams                               UInt32            `json:"max_concurrent_streams,omitempty"`
	Initial_stream_window_size                           UInt32            `json:"initial_stream_window_size,omitempty"`
	Initial_connection_window_size                       UInt32            `json:"initial_connection_window_size,omitempty"`
	Allow_connect                                        bool              `json:"allow_connect,omitempty"`
	Max_outbound_frames                                  UInt32            `json:"max_outbound_frames,omitempty"`
	Max_outbound_control_frames                          UInt32            `json:"max_outbound_control_frames,omitempty"`
	Max_consecutive_inbound_frames_with_empty_payload    UInt32            `json:"max_consecutive_inbound_frames_with_empty_payload,omitempty"`
	Max_inbound_priority_frames_per_stream               UInt32            `json:"max_inbound_priority_frames_per_stream,omitempty"`
	Max_inbound_window_update_frames_per_data_frame_sent UInt32            `json:"max_inbound_window_update_frames_per_data_frame_sent,omitempty"`
	Stream_error_on_invalid_http_messaging               bool              `json:"stream_error_on_invalid_http_messaging,omitempty"`
	Override_stream_error_on_invalid_http_message        bool              `json:"override_stream_error_on_invalid_http_message,omitempty"`
	Connection_keepalive                                 KeepaliveSettings `json:"connection_keepalive,omitempty"`
}

type HttpHealthCheck struct {
	Host                      string               `json:"host,omitempty"`
	Path                      string               `json:"path"`
	Receive                   []HealthCheckPayload `json:"receive,omitempty"`
	Response_buffer_size      UInt64               `json:"response_buffer_size,omitempty"`
	Request_headers_to_add    []HeaderValueOption  `json:"request_headers_to_add,omitempty"`
	Request_headers_to_remove []string             `json:"request_headers_to_remove,omitempty"`
	Expected_statuses         []IntRange           `json:"expected_statuses,omitempty"`
	Retriable_statuses        []IntRange           `json:"retriable_statuses,omitempty"`
	Codec_client_type         CodecClientType      `json:"codec_client_type,omitempty"`
	Service_name_matcher      StringMatcher        `json:"service_name_matcher,omitempty"`
	Method                    RequestMethod        `json:"method,omitempty"`
}

type HttpProtocolOptionsHeadersWithUnderscoresAction string

const (
	HttpProtocolOptionsHeadersWithUnderscoresActionAllow         HttpProtocolOptionsHeadersWithUnderscoresAction = "ALLOW"
	HttpProtocolOptionsHeadersWithUnderscoresActionRejectRequest HttpProtocolOptionsHeadersWithUnderscoresAction = "REJECT_REQUEST"
	HttpProtocolOptionsHeadersWithUnderscoresActionDropHeader    HttpProtocolOptionsHeadersWithUnderscoresAction = "DROP_HEADER"
)

type HttpProtocolOptions struct {
	Idle_timeout                    Duration                                        `json:"idle_timeout,omitempty"`
	Max_connection_duration         Duration                                        `json:"max_connection_duration,omitempty"`
	Max_headers_count               UInt32                                          `json:"max_headers_count,omitempty"`
	Max_stream_duration             Duration                                        `json:"max_stream_duration,omitempty"`
	Headers_with_underscores_action HttpProtocolOptionsHeadersWithUnderscoresAction `json:"headers_with_underscores_action,omitempty"`
}

type HttpStatus struct {
	Code StatusCode `json:"code,omitempty"`
}

type IntRange struct {
	Start float32 `json:"start"`
	End   float32 `json:"end"`
}

type KeepaliveSettings struct {
	Interval                 Duration `json:"interval,omitempty"`
	Timeout                  any/* TODO: [object Object]*/ `json:"timeout"`
	Interval_jitter          Percent  `json:"interval_jitter,omitempty"`
	Connection_idle_interval Duration `json:"connection_idle_interval,omitempty"`
}

type ListStringMatcher struct {
	Patterns []StringMatcher `json:"patterns,omitempty"`
}

type ListValue []any

type Locality struct {
	Region   string `json:"region,omitempty"`
	Zone     string `json:"zone,omitempty"`
	Sub_zone string `json:"sub_zone,omitempty"`
}

type MetadataMatcherPath struct {
	Key string `json:"key"`
}

type MetadataMatcherValue struct {
	Null_match    any           `json:"null_match,omitempty"`
	Double_match  DoubleMatcher `json:"double_match,omitempty"`
	String_match  StringMatcher `json:"string_match,omitempty"`
	Bool_match    bool          `json:"bool_match,omitempty"`
	Present_match bool          `json:"present_match,omitempty"`
}

type MetadataMatcher struct {
	Filter string                `json:"filter"`
	Path   []MetadataMatcherPath `json:"path,omitempty"`
	Value  MetadataMatcherValue  `json:"value"`
	Invert bool                  `json:"invert,omitempty"`
}

type PathConfigSource struct {
	Path              string           `json:"path"`
	Watched_directory WatchedDirectory `json:"watched_directory,omitempty"`
}

type Percent float32

type Pipe struct {
	Path string  `json:"path"`
	Mode float32 `json:"mode"`
}

type Port float32

type Priority float32

type QueryParameterMatcher struct {
	Name          string        `json:"name"`
	String_match  StringMatcher `json:"string_match,omitempty"`
	Present_match bool          `json:"present_match,omitempty"`
}

type RateLimitSettings struct {
	Max_tokens UInt32  `json:"max_tokens,omitempty"`
	Fill_rate  float32 `json:"fill_rate"`
}

type RegexMatcherGoogleRe2 struct {
	Max_program_size float32 `json:"max_program_size"`
}

type RegexMatcher struct {
	Google_re2 RegexMatcherGoogleRe2 `json:"google_re2,omitempty"`
	Regex      string                `json:"regex"`
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
	Retry_back_off TimeInterval `json:"retry_back_off,omitempty"`
	Num_retries    float32      `json:"num_retries"`
}

type RouteMatchGrpc struct {
	Presented bool `json:"presented,omitempty"`
	Validated bool `json:"validated,omitempty"`
}

type RouteMatchTlsContext struct {
	Presented bool `json:"presented,omitempty"`
	Validated bool `json:"validated,omitempty"`
}

type RouteMatch struct {
	Prefix                string                   `json:"prefix,omitempty"`
	Path                  string                   `json:"path,omitempty"`
	Safe_regex            RegexMatcher             `json:"safe_regex,omitempty"`
	Path_separated_prefix string                   `json:"path_separated_prefix,omitempty"`
	Path_match_policy     TypedExtensionConfig     `json:"path_match_policy,omitempty"`
	Case_sensitive        bool                     `json:"case_sensitive,omitempty"`
	Runtime_fraction      RuntimeFractionalPercent `json:"runtime_fraction,omitempty"`
	Headers               []HeaderMatcher          `json:"headers,omitempty"`
	Query_parameters      []QueryParameterMatcher  `json:"query_parameters,omitempty"`
	Grpc                  RouteMatchGrpc           `json:"grpc,omitempty"`
	Tls_context           RouteMatchTlsContext     `json:"tls_context,omitempty"`
	Dynamic_metadata      []MetadataMatcher        `json:"dynamic_metadata,omitempty"`
}

type RouteMatchRestricted struct {
	Prefix  string          `json:"prefix,omitempty"`
	Headers []HeaderMatcher `json:"headers,omitempty"`
}

type RoutingPriority string

const (
	RoutingPriorityDefault RoutingPriority = "DEFAULT"
	RoutingPriorityHigh    RoutingPriority = "HIGH"
)

type RuntimeDouble struct {
	Default_value float32 `json:"default_value"`
	Runtime_key   string  `json:"runtime_key"`
}

type RuntimeFeatureFlag struct {
	Default_value bool   `json:"default_value"`
	Runtime_key   string `json:"runtime_key"`
}

type RuntimeFractionalPercent struct {
	Default_value FractionalPercent `json:"default_value,omitempty"`
	Runtime_key   string            `json:"runtime_key"`
}

type SocketAddressProtocol string

const (
	SocketAddressProtocolTcp SocketAddressProtocol = "TCP"
	SocketAddressProtocolUdp SocketAddressProtocol = "UDP"
)

type SocketAddress struct {
	Protocol      SocketAddressProtocol `json:"protocol,omitempty"`
	Address       string                `json:"address"`
	Port_value    float32               `json:"port_value"`
	Named_port    string                `json:"named_port,omitempty"`
	Resolver_name any                   `json:"resolver_name,omitempty"`
	Ipv4_compat   bool                  `json:"ipv4_compat,omitempty"`
}

type SocketOptionState string

const (
	SocketOptionStateStatePrebind   SocketOptionState = "STATE_PREBIND"
	SocketOptionStateStateBound     SocketOptionState = "STATE_BOUND"
	SocketOptionStateStateListening SocketOptionState = "STATE_LISTENING"
)

type SocketOption struct {
	Description string  `json:"description,omitempty"`
	Level       float32 `json:"level"`
	Name        float32 `json:"name"`
	Int_value   float32 `json:"int_value"`
	Buf_value   any/* TODO: [object Object]*/ `json:"buf_value,omitempty"`
	State       SocketOptionState `json:"state,omitempty"`
}

type SocketOptionsOverride struct {
	Socket_options []SocketOption `json:"socket_options,omitempty"`
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
	Exact       string       `json:"exact,omitempty"`
	Prefix      string       `json:"prefix,omitempty"`
	Suffix      string       `json:"suffix,omitempty"`
	Safe_regex  RegexMatcher `json:"safe_regex,omitempty"`
	Contains    string       `json:"contains,omitempty"`
	Ignore_case bool         `json:"ignore_case,omitempty"`
}

type Struct map[string]any

type TcpHealthCheck struct {
	Send    HealthCheckPayload   `json:"send,omitempty"`
	Receive []HealthCheckPayload `json:"receive,omitempty"`
}

type TcpKeepalive struct {
	Keepalive_probes   UInt32 `json:"keepalive_probes,omitempty"`
	Keepalive_time     UInt32 `json:"keepalive_time,omitempty"`
	Keepalive_interval UInt32 `json:"keepalive_interval,omitempty"`
}

type TimeInterval struct {
	Base_interval any/* TODO: [object Object]*/ `json:"base_interval"`
	Max_interval  Duration `json:"max_interval,omitempty"`
}

type TransportSocketTypedConfigType string

const (
	TransportSocketTypedConfigTypeTypeGoogleapisComEnvoyExtensionsTransportSocketsTlsV3UpstreamTlsContext TransportSocketTypedConfigType = "type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.UpstreamTlsContext"
)

type TransportSocketTypedConfig struct {
	Type TransportSocketTypedConfigType `json:"@type,omitempty"`
	Sni  string                         `json:"sni,omitempty"`
}

type TransportSocket struct {
	Name         string                     `json:"name"`
	Typed_config TransportSocketTypedConfig `json:"typed_config,omitempty"`
}

type TypedExtensionConfig struct {
	Name         string `json:"name"`
	Typed_config any    `json:"typed_config,omitempty"`
}

type UInt32 float32

type UInt64 float32

type UpstreamHttpProtocolOptions struct {
	Auto_sni                 bool   `json:"auto_sni,omitempty"`
	Auto_san_validation      bool   `json:"auto_san_validation,omitempty"`
	Override_auto_sni_header string `json:"override_auto_sni_header,omitempty"`
}

type ValueMatcher struct {
	Null_match    any           `json:"null_match,omitempty"`
	Double_match  DoubleMatcher `json:"double_match,omitempty"`
	String_match  StringMatcher `json:"string_match,omitempty"`
	Bool_match    bool          `json:"bool_match,omitempty"`
	Present_match bool          `json:"present_match,omitempty"`
}

type WatchedDirectory struct {
	Path string `json:"path"`
}
