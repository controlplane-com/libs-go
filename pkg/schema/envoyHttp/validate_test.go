package envoyHttp

import (
	"testing"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

func run(t *testing.T, rule *validate.Rule, cases []validate.Case) {
	t.Helper()
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			if err := c.Run(rule); err != nil {
				t.Fatal(err)
			}
		})
	}
}

const (
	extAuthzType = "type.googleapis.com/envoy.extensions.filters.http.ext_authz.v3.ExtAuthz"
	jwtAuthnType = "type.googleapis.com/envoy.extensions.filters.http.jwt_authn.v3.JwtAuthentication"
	gjtType      = "type.googleapis.com/envoy.extensions.filters.http.grpc_json_transcoder.v3.GrpcJsonTranscoder"
	corsType     = "type.googleapis.com/envoy.extensions.filters.http.cors.v3.Cors"
	rlType       = "type.googleapis.com/envoy.extensions.filters.http.ratelimit.v3.RateLimit"
	grpcWebType  = "type.googleapis.com/envoy.extensions.filters.http.grpc_web.v3.GrpcWeb"
)

func filter(name, atType string, tc map[string]any) map[string]any {
	cfg := map[string]any{"@type": atType}
	for k, v := range tc {
		cfg[k] = v
	}
	return map[string]any{"name": name, "typed_config": cfg}
}

// TestHttpFilterUnion covers the union dispatch itself: before this the whole package
// was unreachable (`type HttpFilter any`) and every body below was accepted.
func TestHttpFilterUnion(t *testing.T) {
	run(t, HttpFilterRule(), []validate.Case{
		{Name: "bogus object", Input: map[string]any{"totally": "bogus", "not_a_filter": 123.0},
			WantErr: "does not match any of the allowed types"},
		{Name: "wrong name", Input: filter("x", extAuthzType, nil),
			WantErr: "does not match any of the allowed types"},
		{Name: "wrong @type", Input: filter("envoy.filters.http.cors", "nope", nil),
			WantErr: "does not match any of the allowed types"},
		{Name: "missing @type", Input: map[string]any{"name": "envoy.filters.http.cors",
			"typed_config": map[string]any{"allow_methods": "GET"}},
			WantErr: "does not match any of the allowed types"},
		{Name: "unknown key inside typed_config", Input: filter("envoy.filters.http.grpc_web", grpcWebType,
			map[string]any{"bogus": 1.0}), WantErr: "does not match any of the allowed types"},
		{Name: "cors ok", Input: filter("envoy.filters.http.cors", corsType, map[string]any{"allow_methods": "GET"})},
		{Name: "grpc_web ok", Input: filter("envoy.filters.http.grpc_web", grpcWebType, nil)},
		{Name: "connect_rpc ok", Input: filter("envoy.filters.http.connect_grpc_bridge",
			"type.googleapis.com/envoy.extensions.filters.http.connect_grpc_bridge.v3.FilterConfig", nil)},
		{Name: "priority bound", Input: map[string]any{"priority": 500.0,
			"name": "envoy.filters.http.cors", "typed_config": map[string]any{"@type": corsType}},
			WantErr: "does not match any of the allowed types"},
	})
}

func TestExtAuthz(t *testing.T) {
	run(t, ruleOf[ExtAuthz](), []validate.Case{
		{Name: "nand", Input: filter("envoy.filters.http.ext_authz", extAuthzType, map[string]any{
			"http_service": map[string]any{},
			"grpc_service": map[string]any{"envoy_grpc": map[string]any{"cluster_name": "c"}}}),
			WantErr: `"http_service" must not exist simultaneously with [grpc_service]`},
		{Name: "either alone is fine", Input: filter("envoy.filters.http.ext_authz", extAuthzType, map[string]any{
			"http_service": map[string]any{}})},
		{Name: "empty path_prefix allowed", Input: filter("envoy.filters.http.ext_authz", extAuthzType, map[string]any{
			"http_service": map[string]any{"path_prefix": ""}})},
		{Name: "server_uri timeout required", Input: filter("envoy.filters.http.ext_authz", extAuthzType,
			map[string]any{"http_service": map[string]any{"server_uri": map[string]any{"uri": "u", "cluster": "c"}}}),
			WantErr: `"typed_config.http_service.server_uri.timeout" is required`},
		{Name: "buffer max_request_bytes integer", Input: filter("envoy.filters.http.ext_authz", extAuthzType,
			map[string]any{"with_request_body": map[string]any{"max_request_bytes": 1.5}}),
			WantErr: "must be an integer"},
		{Name: "buffer max_request_bytes min 0", Input: filter("envoy.filters.http.ext_authz", extAuthzType,
			map[string]any{"with_request_body": map[string]any{"max_request_bytes": -1.0}}),
			WantErr: "must be greater than or equal to 0"},
	})
}

func TestJwtAuthn(t *testing.T) {
	jwt := func(tc map[string]any) map[string]any {
		return filter("envoy.filters.http.jwt_authn", jwtAuthnType, tc)
	}
	run(t, ruleOf[JwtAuthn](), []validate.Case{
		{Name: "filter_state_rules forbidden", Input: jwt(map[string]any{
			"filter_state_rules": map[string]any{"name": "n"}}), WantErr: `"typed_config.filter_state_rules" is not allowed`},
		{Name: "bypass_cors_preflight forbidden", Input: jwt(map[string]any{
			"bypass_cors_preflight": true}), WantErr: `"typed_config.bypass_cors_preflight" is not allowed`},
		{Name: "requirement_map forbidden", Input: jwt(map[string]any{
			"requirement_map": map[string]any{}}), WantErr: `"typed_config.requirement_map" is not allowed`},
		{Name: "restricted rule requires is single-key", Input: jwt(map[string]any{
			"rules": []any{map[string]any{"requires": map[string]any{
				"provider_name": "p", "provider_and_audiences": map[string]any{}}}}}),
			WantErr: "is not allowed"},
		// providers is a DUAL-pattern object: a cpln_ key gets the UI-restricted schema,
		// anything else the full one.
		{Name: "cpln provider rejects async_fetch", Input: jwt(map[string]any{
			"providers": map[string]any{"cpln_a": map[string]any{
				"remote_jwks": map[string]any{"async_fetch": map[string]any{}}}}}),
			WantErr: `"typed_config.providers.cpln_a.remote_jwks.async_fetch" is not allowed`},
		{Name: "plain provider allows async_fetch", Input: jwt(map[string]any{
			"providers": map[string]any{"plain": map[string]any{
				"remote_jwks": map[string]any{"async_fetch": map[string]any{}}}}})},
		{Name: "cpln provider pins cache_duration to the http_uri timeout", Input: jwt(map[string]any{
			"providers": map[string]any{"cpln_a": map[string]any{"remote_jwks": map[string]any{
				"http_uri":       map[string]any{"uri": "u", "cluster": "c", "timeout": "5s"},
				"cache_duration": "9s"}}}}),
			WantErr: "cache_duration must equal http_uri timeout"},
		{Name: "cpln provider matching durations", Input: jwt(map[string]any{
			"providers": map[string]any{"cpln_a": map[string]any{"remote_jwks": map[string]any{
				"http_uri":       map[string]any{"uri": "u", "cluster": "c", "timeout": "5s"},
				"cache_duration": "5s"}}}})},
		{Name: "claim_to_headers fields required", Input: jwt(map[string]any{
			"providers": map[string]any{"p": map[string]any{
				"claim_to_headers": []any{map[string]any{"header_name": "h"}}}}}),
			WantErr: "claim_name"},
	})
}

func TestJwtRequirementXor(t *testing.T) {
	run(t, ruleOf[JwtRequirement](), []validate.Case{
		{Name: "one", Input: map[string]any{"provider_name": "p"}},
		{Name: "two", Input: map[string]any{"provider_name": "p", "allow_missing": map[string]any{}},
			WantErr: "contains a conflict between exclusive peers [provider_name, provider_and_audiences, requires_any, requires_all, allow_missing_or_failed, allow_missing]"},
		{Name: "none", Input: map[string]any{}, WantErr: "must contain at least one of"},
	})
	run(t, ruleOf[JwtRequirementRule](), []validate.Case{
		{Name: "requirement_name forbidden", Input: map[string]any{"requirement_name": "r"},
			WantErr: `"requirement_name" is not allowed`},
	})
}

func TestGrpcJsonTranscoderXor(t *testing.T) {
	gjt := func(tc map[string]any) map[string]any {
		return filter("envoy.filters.http.grpc_json_transcoder", gjtType, tc)
	}
	run(t, ruleOf[GrpcJsonTranscoder](), []validate.Case{
		{Name: "both", Input: gjt(map[string]any{"proto_descriptor": "a", "proto_descriptor_bin": "b"}),
			WantErr: "contains a conflict between exclusive peers [proto_descriptor, proto_descriptor_bin]"},
		{Name: "neither", Input: gjt(map[string]any{}),
			WantErr: "must contain at least one of [proto_descriptor, proto_descriptor_bin]"},
		{Name: "one", Input: gjt(map[string]any{"proto_descriptor": "a"})},
		{Name: "url_unescape_spec enum", Input: gjt(map[string]any{"proto_descriptor": "a", "url_unescape_spec": "NOPE"}),
			WantErr: "must be one of [ALL_CHARACTERS_EXCEPT_RESERVED, ALL_CHARACTERS_EXCEPT_SLASH, ALL_CHARACTERS]"},
		{Name: "proto_descriptor_bin is binary", Input: gjt(map[string]any{
			"proto_descriptor_bin": map[string]any{"not": "bytes"}}), WantErr: "must be a buffer or a string"},
	})
}

func TestRateLimitFilter(t *testing.T) {
	rl := func(tc map[string]any) map[string]any { return filter("envoy.filters.http.ratelimit", rlType, tc) }
	run(t, ruleOf[RateLimit](), []validate.Case{
		{Name: "domain required", Input: rl(map[string]any{}), WantErr: `"typed_config.domain" is required`},
		{Name: "stage max", Input: rl(map[string]any{"domain": "d", "stage": 99.0}),
			WantErr: "must be less than or equal to 10"},
		{Name: "stage integer", Input: rl(map[string]any{"domain": "d", "stage": 1.5}),
			WantErr: "must be an integer"},
		{Name: "request_type enum", Input: rl(map[string]any{"domain": "d", "request_type": "nope"}),
			WantErr: "must be one of [internal, external, both]"},
		{Name: "enable_x_ratelimit_headers enum", Input: rl(map[string]any{
			"domain": "d", "enable_x_ratelimit_headers": "nope"}),
			WantErr: "must be one of [OFF, DRAFT_VERSION_03]"},
		{Name: "grpc_service required", Input: rl(map[string]any{"domain": "d",
			"rate_limit_service": map[string]any{}}), WantErr: `"typed_config.rate_limit_service.grpc_service" is required`},
		{Name: "grpc_service xor", Input: rl(map[string]any{"domain": "d",
			"rate_limit_service": map[string]any{"grpc_service": map[string]any{}}}),
			WantErr: "must contain at least one of [envoy_grpc, google_grpc]"},
		{Name: "ok", Input: rl(map[string]any{"domain": "d", "stage": 3.0, "request_type": "both"})},
	})
}

// TestProviderBlameFollowsWireOrder: abortEarly reports the FIRST failing provider, so
// with two bad ones the key order decides which name the message quotes. Joi walks
// insertion order; this used to walk sorted order and quote the other one.
//
// It runs the filter as `spec.sidecar.envoy.http[]` does — an `any` element judged by
// the union rule — because that is the only way a request reaches this schema.
//
// Wants captured 2026-09-16 from the built TS schema:
//
//	node -e "const {JwtAuthn}=require('./build/cjs/src/envoyHttp'); JwtAuthn.validate(body)"
//	  {cpln_z, cpln_a} -> typed_config.providers.cpln_z.bogus is not allowed
//	  {cpln_a, cpln_z} -> typed_config.providers.cpln_a.bogus is not allowed
//	  {b_z, b_a}       -> typed_config.providers.b_z.bogus is not allowed
func TestProviderBlameFollowsWireOrder(t *testing.T) {
	body := func(first, second string) string {
		return `{"f":{"name":"envoy.filters.http.jwt_authn","typed_config":{"@type":"` + jwtAuthnType + `",` +
			`"providers":{"` + first + `":{"issuer":"https://i","bogus":1},` +
			`"` + second + `":{"issuer":"https://i","bogus":1}}}}}`
	}
	// the alternatives union reports its own summary, so the providers rule is applied
	// on its own here — the same rule the union's JwtAuthn branch runs
	rule := validate.Object(map[string]*validate.Rule{"f": ruleOf[JwtAuthn]()})
	for _, tc := range [][2]string{{"cpln_z", "cpln_a"}, {"cpln_a", "cpln_z"}, {"b_z", "b_a"}} {
		t.Run(tc[0]+"-first", func(t *testing.T) {
			var dst struct {
				F any `json:"f"`
			}
			d, err := validate.DecodeInto(&dst, validate.Bytes(body(tc[0], tc[1])))
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			verr := d.Validate(rule, &validate.Context{}, validate.Prefs{})
			if verr == nil {
				t.Fatal("expected the unknown key to be rejected")
			}
			want := `"f.typed_config.providers.` + tc[0] + `.bogus" is not allowed`
			if verr.Message != want {
				t.Fatalf("got %q, want %q", verr.Message, want)
			}
		})
	}
}
