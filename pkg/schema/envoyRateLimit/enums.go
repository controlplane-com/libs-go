package envoyRateLimit

// EnumValues makes the generated string enums reflect to Joi.string().valid(...).
// The generator emits the constants but not the method, so without these the types
// validate as free-form strings.
func (RateLimitRequestType) EnumValues() []string {
	return []string{string(RateLimitRequestTypeInternal), string(RateLimitRequestTypeExternal), string(RateLimitRequestTypeBoth)}
}

func (RateLimitEnableXRatelimitHeaders) EnumValues() []string {
	return []string{string(RateLimitEnableXRatelimitHeadersOff), string(RateLimitEnableXRatelimitHeadersDraftVersion03)}
}
