package envoyHttp

// Validation for `spec.sidecar.envoy.http[]`. Ports nodelibs/schema/src/envoyHttp.ts.
//
// `type HttpFilter any` (the generator's spelling of a Joi alternatives union) made
// this entire package unreachable: every concrete filter struct below was defined and
// referenced by nothing, so ~22 TS rules — including the three `.forbidden()` fields
// on jwt_authn that exist to keep tenants off the UI-managed JWT path — were not
// enforced at all. HttpFilter is a defined INTERFACE type, so it cannot carry a
// method; the union rule is registered against the type instead (the same seam
// envoyCommon.Duration uses), which is why no model type had to change.

import (
	"reflect"
	"regexp"
	"strings"

	"github.com/controlplane-com/libs-go/pkg/schema/envoyCommon"
	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

func ruleOf[T any]() *validate.Rule { return validate.FromType(reflect.TypeOf((*T)(nil)).Elem()) }

// forbid is `schema.forbidden()`: the key keeps its schema (so the describe document
// still carries it) and only its presence changes, as Joi does.
func forbid(r *validate.Rule) *validate.Rule { return r.Forbidden() }

// HttpFilterRule ports `HttpFilter = Joi.alternatives(ExtAuthz, JwtAuthn, GrpcWeb,
// ConnectRpc, GrpcJsonTranscoder, Cors, RateLimit, HeaderMutation)` (envoyHttp.ts), in order.
func HttpFilterRule() *validate.Rule {
	return validate.Alternatives(
		ruleOf[ExtAuthz](),
		ruleOf[JwtAuthn](),
		ruleOf[GrpcWeb](),
		ruleOf[ConnectRpc](),
		ruleOf[GrpcJsonTranscoder](),
		ruleOf[Cors](),
		ruleOf[RateLimit](),
		ruleOf[HeaderMutation](),
	)
}

func init() {
	validate.RegisterTypeRule(reflect.TypeOf((*HttpFilter)(nil)).Elem(), HttpFilterRule)
}

// --- header_mutation (envoyHttp.ts header_mutation, master 894c1be3c) ---

var (
	headerMutationNameRe = regexp.MustCompile("^[a-z0-9!#$%&'*+.^_`|~-]+$")
	headerMutationValRe  = regexp.MustCompile(`^[^\x00\r\n]*$`)
)

const headerMutationNameJS = "^[a-z0-9!#$%&'*+.^_`|~-]+$"

func headerMutationHeaderName() *validate.Rule {
	return validate.String().Max(16384).RegexAs(headerMutationNameRe, headerMutationNameJS).Invalid("host")
}

// HeaderMutationRuleRule ports the HeaderMutationRule alternatives (envoyHttp.ts):
// {remove: HeaderName} or {append: HeaderValueOption whose header.key is a HeaderName},
// the append option carrying Joi's .oxor('append','append_action').
func HeaderMutationRuleRule() *validate.Rule {
	header := validate.Reflect(envoyCommon.HeaderValue{}).
		Fork("key", func(*validate.Rule) *validate.Rule { return headerMutationHeaderName().Required() }).
		Fork("value", func(*validate.Rule) *validate.Rule {
			return validate.String().Max(16384).RegexAs(headerMutationValRe, `^[^\0\r\n]*$`)
		})
	appendOption := validate.Reflect(envoyCommon.HeaderValueOption{}).
		Fork("header", func(*validate.Rule) *validate.Rule { return header.Required() }).
		Oxor("append", "append_action").
		Required()
	remove := validate.Object(map[string]*validate.Rule{"remove": headerMutationHeaderName().Required()})
	appendRule := validate.Object(map[string]*validate.Rule{"append": appendOption})
	return validate.Alternatives(remove, appendRule)
}

func init() {
	validate.RegisterTypeRule(reflect.TypeOf((*HeaderMutationRule)(nil)).Elem(), HeaderMutationRuleRule)
}

// --- ext_authz (envoyHttp.ts:136) ---

// ValidationRule ports ExtAuthzSchema's `.nand('http_service','grpc_service')`.
func (ExtAuthzTypedConfig) ValidationRule() *validate.Rule {
	return validate.Reflect(ExtAuthzTypedConfig{}).Nand("http_service", "grpc_service")
}

// ValidationRule ports `path_prefix: Joi.string().allow(”)`.
func (ExtAuthzTypedConfigHttpService) ValidationRule() *validate.Rule {
	return validate.Reflect(ExtAuthzTypedConfigHttpService{}).
		Fork("path_prefix", func(r *validate.Rule) *validate.Rule { return r.Allow("") })
}

// --- jwt_authn (envoyHttp.ts:174) ---

// ValidationRule ports JwtAuthnSchema: the dual-pattern `providers` map plus the three
// fields the UI-restricted path forbids outright.
func (JwtAuthnTypedConfig) ValidationRule() *validate.Rule {
	return validate.Reflect(JwtAuthnTypedConfig{}).
		Fork("providers", func(*validate.Rule) *validate.Rule { return jwtProvidersRule() }).
		Fork("filter_state_rules", forbid).
		Fork("bypass_cors_preflight", forbid).
		Fork("requirement_map", forbid)
}

// jwtProvidersRule ports
// `Joi.object().pattern(/^(?!cpln_).*/, JwtProvider).pattern(/^cpln_.*/, JwtProviderUIRestricted)`.
//
// The engine's object rule holds ONE pattern (key schema + value schema), so a
// two-pattern object cannot be expressed declaratively; the key→schema choice is made
// in a custom instead. Keys are walked in the order the body wrote them, so when two
// providers are both invalid abortEarly blames the one JS would — cosmetic here (the
// schema a key is judged against comes from its own `cpln_` prefix, not its position),
// but it costs nothing now that the engine carries the order.
func jwtProvidersRule() *validate.Rule {
	return validate.Object(nil).CustomTyped(validateProviders)
}

func validateProviders(value any, h *validate.Helper) (any, error) {
	m, ok := value.(map[string]any)
	if !ok {
		return value, nil
	}
	out := make(map[string]any, len(m))
	for _, k := range validate.OrderedKeys(m, h) {
		rule := ruleOf[JwtProvider]()
		if strings.HasPrefix(k, "cpln_") {
			rule = ruleOf[JwtProviderUIRestricted]()
		}
		nv, err := rule.ValidateAt(append(append([]string(nil), h.Path...), k), m[k], h.Ctx, validate.Prefs{})
		if err != nil {
			return nil, err
		}
		out[k] = nv
	}
	return out, nil
}

// ValidationRule ports JwtProviderUIRestricted's remote_jwks custom: the UI path
// pins the JWKS cache to the fetch timeout (envoyHttp.ts:73).
func (JwtProviderUiRestrictedRemoteJwks) ValidationRule() *validate.Rule {
	return validate.Reflect(JwtProviderUiRestrictedRemoteJwks{}).CustomTyped(cacheMatchesTimeout)
}

func cacheMatchesTimeout(value any, h *validate.Helper) (any, error) {
	var timeout any
	if u := validate.Get(value, "http_uri"); u != nil {
		timeout = validate.Get(u, "timeout")
	}
	if !looseEqual(timeout, validate.Get(value, "cache_duration")) {
		return nil, h.Errorf("any.custom", "cache_duration must equal http_uri timeout")
	}
	return value, nil
}

// looseEqual is JS `==` for the two shapes this comparison can see: an absent value
// (undefined) or a duration string.
func looseEqual(a, b any) bool {
	as, aok := a.(string)
	bs, bok := b.(string)
	if !aok && !bok {
		return a == nil && b == nil
	}
	return aok && bok && as == bs
}

// ValidationRule ports JwtRequirement's six-peer `.xor()` (envoyHttp.ts:98).
func (JwtRequirement) ValidationRule() *validate.Rule {
	return validate.Reflect(JwtRequirement{}).Xor("provider_name", "provider_and_audiences",
		"requires_any", "requires_all", "allow_missing_or_failed", "allow_missing")
}

// ValidationRule ports JwtRequirementRule: `requirement_name` is forbidden outright
// and may not accompany `requires` (envoyHttp.ts:104).
func (JwtRequirementRule) ValidationRule() *validate.Rule {
	return validate.Reflect(JwtRequirementRule{}).
		Fork("requirement_name", forbid).
		Nand("requires", "requirement_name")
}

// ValidationRule ports JwtRequirementRuleRestricted's nand (envoyHttp.ts:113).
func (JwtRequirementRuleRestricted) ValidationRule() *validate.Rule {
	return validate.Reflect(JwtRequirementRuleRestricted{}).Nand("requires", "requirement_name")
}

// --- grpc_json_transcoder (envoyHttp.ts:192) ---

// ValidationRule ports `.xor('proto_descriptor','proto_descriptor_bin')`.
func (GrpcJsonTranscoderTypedConfig) ValidationRule() *validate.Rule {
	return validate.Reflect(GrpcJsonTranscoderTypedConfig{}).Xor("proto_descriptor", "proto_descriptor_bin")
}

// --- ratelimit (envoyHttp.ts:235) ---
//
// The generator duplicated envoyCommon's GrpcService into this package rather than
// referencing it, so the shared type's rules have to be restated on the copy.

// ValidationRule ports GrpcService's `.xor('envoy_grpc','google_grpc')` and
// `timeout: Duration.default('0.200s')`.
func (RateLimitServiceGrpcService) ValidationRule() *validate.Rule {
	return validate.Reflect(RateLimitServiceGrpcService{}).
		Fork("timeout", func(r *validate.Rule) *validate.Rule { return r.Default("0.200s") }).
		Xor("envoy_grpc", "google_grpc")
}

// ValidationRule ports `authority: Joi.string().default(Joi.ref('cluster_name'))`.
func (RateLimitServiceGrpcServiceEnvoyGrpc) ValidationRule() *validate.Rule {
	return validate.Reflect(RateLimitServiceGrpcServiceEnvoyGrpc{}).
		Fork("authority", func(r *validate.Rule) *validate.Rule { return r.Default(validate.RefTo("cluster_name")) })
}

// ValidationRule ports channel_credentials' `.xor('ssl_credentials','google_default')`.
func (RateLimitServiceGrpcServiceGoogleGrpcChannelCredentials) ValidationRule() *validate.Rule {
	return validate.Reflect(RateLimitServiceGrpcServiceGoogleGrpcChannelCredentials{}).
		Xor("ssl_credentials", "google_default")
}
