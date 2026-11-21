package envoyHttp

// EnumValues makes the generated string enums reflect to Joi.string().valid(...).
// The generator emits the constants but not the method, so without these the types
// validate as free-form strings.
func (ConnectRpcName) EnumValues() []string {
	return []string{string(ConnectRpcNameEnvoyFiltersHttpConnectGrpcBridge)}
}

func (ConnectRpcTypedConfigType) EnumValues() []string {
	return []string{string(ConnectRpcTypedConfigTypeTypeGoogleapisComEnvoyExtensionsFiltersHttpConnectGrpcBridgeV3FilterConfig)}
}

func (CorsName) EnumValues() []string {
	return []string{string(CorsNameEnvoyFiltersHttpCors)}
}

func (CorsTypedConfigType) EnumValues() []string {
	return []string{string(CorsTypedConfigTypeTypeGoogleapisComEnvoyExtensionsFiltersHttpCorsV3Cors)}
}

func (DescriptorRateLimitUnit) EnumValues() []string {
	return []string{string(DescriptorRateLimitUnitSecond), string(DescriptorRateLimitUnitMinute), string(DescriptorRateLimitUnitHour), string(DescriptorRateLimitUnitDay)}
}

func (ExtAuthzName) EnumValues() []string {
	return []string{string(ExtAuthzNameEnvoyFiltersHttpExtAuthz)}
}

func (HeaderMutationName) EnumValues() []string {
	return []string{string(HeaderMutationNameEnvoyFiltersHttpHeaderMutation)}
}

func (HeaderMutationTypedConfigType) EnumValues() []string {
	return []string{string(HeaderMutationTypedConfigTypeHeaderMutation)}
}

func (ExtAuthzTypedConfigType) EnumValues() []string {
	return []string{string(ExtAuthzTypedConfigTypeTypeGoogleapisComEnvoyExtensionsFiltersHttpExtAuthzV3ExtAuthz)}
}

func (GrpcJsonTranscoderName) EnumValues() []string {
	return []string{string(GrpcJsonTranscoderNameEnvoyFiltersHttpGrpcJsonTranscoder)}
}

func (GrpcJsonTranscoderTypedConfigUrlUnescapeSpec) EnumValues() []string {
	return []string{string(GrpcJsonTranscoderTypedConfigUrlUnescapeSpecAllCharactersExceptReserved), string(GrpcJsonTranscoderTypedConfigUrlUnescapeSpecAllCharactersExceptSlash), string(GrpcJsonTranscoderTypedConfigUrlUnescapeSpecAllCharacters)}
}

func (GrpcJsonTranscoderTypedConfigType) EnumValues() []string {
	return []string{string(GrpcJsonTranscoderTypedConfigTypeTypeGoogleapisComEnvoyExtensionsFiltersHttpGrpcJsonTranscoderV3GrpcJsonTranscoder)}
}

func (GrpcWebName) EnumValues() []string {
	return []string{string(GrpcWebNameEnvoyFiltersHttpGrpcWeb)}
}

func (GrpcWebTypedConfigType) EnumValues() []string {
	return []string{string(GrpcWebTypedConfigTypeTypeGoogleapisComEnvoyExtensionsFiltersHttpGrpcWebV3GrpcWeb)}
}

func (JwtAuthnName) EnumValues() []string {
	return []string{string(JwtAuthnNameEnvoyFiltersHttpJwtAuthn)}
}

func (JwtAuthnTypedConfigType) EnumValues() []string {
	return []string{string(JwtAuthnTypedConfigTypeTypeGoogleapisComEnvoyExtensionsFiltersHttpJwtAuthnV3JwtAuthentication)}
}

func (RateLimitName) EnumValues() []string {
	return []string{string(RateLimitNameEnvoyFiltersHttpRatelimit)}
}

func (RateLimitTypedConfigRequestType) EnumValues() []string {
	return []string{string(RateLimitTypedConfigRequestTypeInternal), string(RateLimitTypedConfigRequestTypeExternal), string(RateLimitTypedConfigRequestTypeBoth)}
}

func (RateLimitTypedConfigEnableXRatelimitHeaders) EnumValues() []string {
	return []string{string(RateLimitTypedConfigEnableXRatelimitHeadersOff), string(RateLimitTypedConfigEnableXRatelimitHeadersDraftVersion03)}
}

func (RateLimitTypedConfigType) EnumValues() []string {
	return []string{string(RateLimitTypedConfigTypeTypeGoogleapisComEnvoyExtensionsFiltersHttpRatelimitV3RateLimit)}
}
