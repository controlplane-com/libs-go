package envoyCommon

// EnumValues makes the generated string enums reflect to Joi.string().valid(...).
// The generator emits the constants but not the method, so without these the types
// validate as free-form strings.
func (ApiType) EnumValues() []string {
	return []string{string(ApiTypeDeprecatedAndUnavailableDoNotUse), string(ApiTypeRest), string(ApiTypeGrpc), string(ApiTypeDeltaGrpc)}
}

func (ApiVersion) EnumValues() []string {
	return []string{string(ApiVersionAuto), string(ApiVersionV2), string(ApiVersionV3)}
}

func (CodecClientType) EnumValues() []string {
	return []string{string(CodecClientTypeHttp1), string(CodecClientTypeHttp2)}
}

func (HeaderAppendAction) EnumValues() []string {
	return []string{string(HeaderAppendActionAppendIfExistsOrAdd), string(HeaderAppendActionAddIfAbsent), string(HeaderAppendActionOverwriteIfExistsOrAdd), string(HeaderAppendActionOverwriteIfExists)}
}

func (HealthStatus) EnumValues() []string {
	return []string{string(HealthStatusUnknown), string(HealthStatusHealthy), string(HealthStatusUnhealthy), string(HealthStatusDraining), string(HealthStatusTimeout), string(HealthStatusDegraded)}
}

func (HttpProtocolOptionsHeadersWithUnderscoresAction) EnumValues() []string {
	return []string{string(HttpProtocolOptionsHeadersWithUnderscoresActionAllow), string(HttpProtocolOptionsHeadersWithUnderscoresActionRejectRequest), string(HttpProtocolOptionsHeadersWithUnderscoresActionDropHeader)}
}

func (RequestMethod) EnumValues() []string {
	return []string{string(RequestMethodMethodUnspecified), string(RequestMethodGet), string(RequestMethodHead), string(RequestMethodPost), string(RequestMethodPut), string(RequestMethodDelete), string(RequestMethodConnect), string(RequestMethodOptions), string(RequestMethodTrace), string(RequestMethodPatch)}
}

func (RoutingPriority) EnumValues() []string {
	return []string{string(RoutingPriorityDefault), string(RoutingPriorityHigh)}
}

func (SocketAddressProtocol) EnumValues() []string {
	return []string{string(SocketAddressProtocolTcp), string(SocketAddressProtocolUdp)}
}

func (SocketOptionState) EnumValues() []string {
	return []string{string(SocketOptionStateStatePrebind), string(SocketOptionStateStateBound), string(SocketOptionStateStateListening)}
}

func (StatusCode) EnumValues() []string {
	return []string{string(StatusCodeEmpty), string(StatusCodeContinue), string(StatusCodeOk), string(StatusCodeCreated), string(StatusCodeAccepted), string(StatusCodeNonAuthoritativeInformation), string(StatusCodeNoContent), string(StatusCodeResetContent), string(StatusCodePartialContent), string(StatusCodeMultiStatus), string(StatusCodeAlradyReported), string(StatusCodeImUsed), string(StatusCodeMultipleChoices), string(StatusCodeMovedPermanently), string(StatusCodeFound), string(StatusCodeSeeOther), string(StatusCodeNotModified), string(StatusCodeUseProxy), string(StatusCodeTemporaryRedirect), string(StatusCodePermanentRedirect), string(StatusCodeBadRequest), string(StatusCodeUnauthroized), string(StatusCodePaymentRequired), string(StatusCodeForbidden), string(StatusCodeNotFound), string(StatusCodeMethodNotAllowed), string(StatusCodeNotAcceptable), string(StatusCodeProxyAuthenticationRequired), string(StatusCodeRequestTimeout), string(StatusCodeConflict), string(StatusCodeGone), string(StatusCodeLengthRequired), string(StatusCodePreconditionFailed), string(StatusCodePayloadTooLarge), string(StatusCodeUriTooLong), string(StatusCodeUnsupportedMediaType), string(StatusCodeRangeNotSatisfiable), string(StatusCodeExpectationFailed), string(StatusCodeMisdirectedRequest), string(StatusCodeUnprocessableEntity), string(StatusCodeLocked), string(StatusCodeFailedDependency), string(StatusCodeUpgradeRequired), string(StatusCodePreconditionRequired), string(StatusCodeTooManyRequests), string(StatusCodeRequestHeaderFieldsTooLarge), string(StatusCodeInternalServerError), string(StatusCodeNotImplemented), string(StatusCodeBadGateway), string(StatusCodeServiceUnavailable), string(StatusCodeGatewayTimeout), string(StatusCodeHttpVersionNotSupported), string(StatusCodeVariantAlsoNegotiates), string(StatusCodeInsufficientStorage), string(StatusCodeLoopDetected), string(StatusCodeNotExtended), string(StatusCodeNetworkAuthenticationRequired)}
}

func (TransportSocketTypedConfigType) EnumValues() []string {
	return []string{string(TransportSocketTypedConfigTypeTypeGoogleapisComEnvoyExtensionsTransportSocketsTlsV3UpstreamTlsContext)}
}
