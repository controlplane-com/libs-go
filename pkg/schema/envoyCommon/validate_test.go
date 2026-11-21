package envoyCommon

import (
	"reflect"
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

func ruleFor[T any]() *validate.Rule { return validate.FromType(reflect.TypeOf((*T)(nil)).Elem()) }

func TestDuration(t *testing.T) {
	run(t, DurationRule(), []validate.Case{
		{Name: "seconds string", Input: "10s", Want: "10s"},
		{Name: "fractional string", Input: "15.5s", Want: "15.5s"},
		{Name: "object", Input: map[string]any{"seconds": 5.0}},
		{Name: "bogus string", Input: "BOGUS", WantErr: "fails to match the required pattern"},
		{Name: "negative seconds", Input: map[string]any{"seconds": -1.0}, WantErr: "must be greater than or equal to 0"},
		{Name: "unknown key", Input: map[string]any{"bogus": 1.0}, WantErr: "is not allowed"},
		{Name: "number", Input: 123.0, WantErr: "must be one of [string, object]"},
	})
}

func TestTimeInterval(t *testing.T) {
	run(t, ruleFor[TimeInterval](), []validate.Case{
		{Name: "ok", Input: map[string]any{"base_interval": "1s", "max_interval": "10s"}},
		{Name: "base required", Input: map[string]any{"max_interval": "10s"}, WantErr: `"base_interval" is required`},
		// the >0 check is a `.custom()` on base_interval in TS; it must still fire
		// even though base_interval's rule is an alternatives.
		{Name: "base zero", Input: map[string]any{"base_interval": "0s"},
			WantErr: `"base_interval" failed custom validation because duration not greater than 0`},
		{Name: "base zero object", Input: map[string]any{"base_interval": map[string]any{"seconds": 0.0}},
			WantErr: "duration not greater than 0"},
		{Name: "max below base", Input: map[string]any{"base_interval": "10s", "max_interval": "1s"},
			WantErr: "max_interval must be greater than or equal to base_interval"},
	})
}

func TestBinary(t *testing.T) {
	// Joi.binary() coerces an accepted string with Buffer.from(value); JSON.stringify
	// then renders it as {"type":"Buffer","data":[...]}, which is what TS stores.
	run(t, BinaryRule(), []validate.Case{
		{Name: "string becomes a buffer", Input: "ab",
			Want: map[string]any{"type": "Buffer", "data": []any{float64('a'), float64('b')}}},
		{Name: "object", Input: map[string]any{"not": "bytes"}, WantErr: "must be a buffer or a string"},
		{Name: "number", Input: 5.0, WantErr: "must be a buffer or a string"},
	})
}

func TestDataSourceXor(t *testing.T) {
	run(t, ruleFor[DataSource](), []validate.Case{
		{Name: "one", Input: map[string]any{"filename": "f"}},
		{Name: "none", Input: map[string]any{}, WantErr: "must contain at least one of [filename, inline_bytes, inline_string, environment_variable]"},
		{Name: "two", Input: map[string]any{"filename": "f", "inline_string": "s"},
			WantErr: "contains a conflict between exclusive peers"},
	})
}

func TestGrpcService(t *testing.T) {
	r := ruleFor[GrpcService]()
	run(t, r, []validate.Case{
		{Name: "xor none", Input: map[string]any{"timeout": "5s"},
			WantErr: "must contain at least one of [envoy_grpc, google_grpc]"},
		{Name: "xor both", Input: map[string]any{
			"envoy_grpc":  map[string]any{"cluster_name": "c"},
			"google_grpc": map[string]any{"target_uri": "u", "stat_prefix": "s"}},
			WantErr: "contains a conflict between exclusive peers"},
		{Name: "cluster_name required", Input: map[string]any{"envoy_grpc": map[string]any{}},
			WantErr: `"envoy_grpc.cluster_name" is required`},
		{Name: "target_uri required", Input: map[string]any{"google_grpc": map[string]any{"stat_prefix": "s"}},
			WantErr: `"google_grpc.target_uri" is required`},
		{Name: "channel_credentials xor", Input: map[string]any{"google_grpc": map[string]any{
			"target_uri": "u", "stat_prefix": "s",
			"channel_credentials": map[string]any{"ssl_credentials": map[string]any{}, "google_default": map[string]any{}}}},
			WantErr: "contains a conflict between exclusive peers [ssl_credentials, google_default]"},
		{Name: "sts subject_token_path required", Input: map[string]any{"google_grpc": map[string]any{
			"target_uri": "u", "stat_prefix": "s",
			"call_credentials": []any{map[string]any{"sts_service": map[string]any{"subject_token_type": "t"}}}}},
			WantErr: "subject_token_path"},
		{Name: "initial_metadata key required", Input: map[string]any{
			"envoy_grpc":       map[string]any{"cluster_name": "c"},
			"initial_metadata": []any{map[string]any{"value": "v"}}},
			WantErr: `"initial_metadata[0].key" is required`},
	})

	// defaults: timeout `0.200s` and authority defaulting to Joi.ref('cluster_name')
	out, verr := validate.Validate(r, map[string]any{"envoy_grpc": map[string]any{"cluster_name": "c"}}, nil, validate.Prefs{})
	if verr != nil {
		t.Fatalf("unexpected error: %s", verr.Message)
	}
	m := out.(map[string]any)
	if m["timeout"] != "0.200s" {
		t.Errorf("timeout default: got %#v want %q", m["timeout"], "0.200s")
	}
	if got := m["envoy_grpc"].(map[string]any)["authority"]; got != "c" {
		t.Errorf("authority default: got %#v want %q", got, "c")
	}
}

func TestRouteMatchXor(t *testing.T) {
	run(t, ruleFor[RouteMatch](), []validate.Case{
		{Name: "ok", Input: map[string]any{"prefix": "/a"}},
		{Name: "none", Input: map[string]any{"case_sensitive": true},
			WantErr: "must contain at least one of [prefix, path, safe_regex, connect_matcher, path_separated_prefix, path_match_policy]"},
		{Name: "two", Input: map[string]any{"prefix": "/a", "path": "/b"},
			WantErr: "contains a conflict between exclusive peers"},
		{Name: "header name required", Input: map[string]any{"prefix": "/a",
			"headers": []any{map[string]any{"present_match": true}}},
			WantErr: `"headers[0].name" is required`},
		{Name: "string_match xor", Input: map[string]any{"prefix": "/a",
			"headers": []any{map[string]any{"name": "h", "string_match": map[string]any{"exact": "a", "prefix": "b"}}}},
			WantErr: "contains a conflict between exclusive peers [exact, prefix, suffix, safe_regex, contains]"},
		{Name: "safe_regex regex required", Input: map[string]any{"prefix": "/a",
			"headers": []any{map[string]any{"name": "h", "safe_regex_match": map[string]any{}}}},
			WantErr: "regex"},
		{Name: "max_program_size integer", Input: map[string]any{"prefix": "/a",
			"headers": []any{map[string]any{"name": "h", "safe_regex_match": map[string]any{
				"regex": "r", "google_re2": map[string]any{"max_program_size": 1.7}}}}},
			WantErr: "must be an integer"},
		{Name: "range_match integer", Input: map[string]any{"prefix": "/a",
			"headers": []any{map[string]any{"name": "h", "range_match": map[string]any{"start": 1.5}}}},
			WantErr: "must be an integer"},
		{Name: "range_match end below start", Input: map[string]any{"prefix": "/a",
			"headers": []any{map[string]any{"name": "h", "range_match": map[string]any{"start": 100.0, "end": 5.0}}}},
			WantErr: "must be greater than or equal to"},
		{Name: "query_parameter xor", Input: map[string]any{"prefix": "/a",
			"query_parameters": []any{map[string]any{"name": "q"}}},
			WantErr: "must contain at least one of [string_match, present_match]"},
		{Name: "runtime_key required", Input: map[string]any{"prefix": "/a",
			"runtime_fraction": map[string]any{"default_value": map[string]any{"numerator": 1.0}}},
			WantErr: `"runtime_fraction.runtime_key" is required`},
		{Name: "denominator enum", Input: map[string]any{"prefix": "/a",
			"runtime_fraction": map[string]any{"runtime_key": "k",
				"default_value": map[string]any{"denominator": 7.0}}},
			WantErr: "must be one of"},
		{Name: "dynamic_metadata value required", Input: map[string]any{"prefix": "/a",
			"dynamic_metadata": []any{map[string]any{"filter": "f"}}},
			WantErr: `"dynamic_metadata[0].value" is required`},
		{Name: "value matcher xor names all seven peers", Input: map[string]any{"prefix": "/a",
			"dynamic_metadata": []any{map[string]any{"filter": "f", "value": map[string]any{}}}},
			WantErr: "[null_match, double_match, string_match, bool_match, present_match, list_match, or_match]"},
		{Name: "double matcher xor", Input: map[string]any{"prefix": "/a",
			"dynamic_metadata": []any{map[string]any{"filter": "f", "value": map[string]any{
				"double_match": map[string]any{"range": map[string]any{"start": 1.0}, "exact": 2.0}}}}},
			WantErr: "contains a conflict between exclusive peers [range, exact]"},
		{Name: "path key required", Input: map[string]any{"prefix": "/a",
			"dynamic_metadata": []any{map[string]any{"filter": "f", "path": []any{map[string]any{}},
				"value": map[string]any{"bool_match": true}}}},
			WantErr: `"dynamic_metadata[0].path[0].key" is required`},
	})
}

func TestSocketAddress(t *testing.T) {
	run(t, ruleFor[SocketAddress](), []validate.Case{
		{Name: "ok", Input: map[string]any{"address": "1.1.1.1", "port_value": 53.0}},
		// Joi validates keys before dependencies, so `resolver_name` (which named_port
		// makes required) is reported ahead of the xor — verified against the oracle.
		{Name: "xor both", Input: map[string]any{"address": "a", "port_value": 1.0, "named_port": "p", "resolver_name": "r"},
			WantErr: "contains a conflict between exclusive peers [port_value, named_port]"},
		{Name: "xor none", Input: map[string]any{"address": "a"},
			WantErr: "must contain at least one of [port_value, named_port]"},
		{Name: "address required", Input: map[string]any{"port_value": 1.0}, WantErr: `"address" is required`},
		{Name: "address may not be empty", Input: map[string]any{"address": "", "port_value": 1.0},
			WantErr: "contains an invalid value"},
		{Name: "named_port needs resolver_name", Input: map[string]any{"address": "a", "named_port": "p"},
			WantErr: `"resolver_name" is required`},
		{Name: "named_port with resolver_name", Input: map[string]any{"address": "a", "named_port": "p", "resolver_name": "r"}},
		{Name: "protocol enum", Input: map[string]any{"address": "a", "port_value": 1.0, "protocol": "SCTP"},
			WantErr: "must be one of [TCP, UDP]"},
	})
}

func TestOtherXors(t *testing.T) {
	run(t, ruleFor[Address](), []validate.Case{
		{Name: "address none", Input: map[string]any{},
			WantErr: "must contain at least one of [socket_address, pipe, envoy_internal_address]"},
	})
	run(t, ruleFor[ConfigSource](), []validate.Case{
		{Name: "config source two", Input: map[string]any{"path": "p", "path_config_source": map[string]any{"path": "q"}},
			WantErr: "contains a conflict between exclusive peers"},
	})
	run(t, ruleFor[HealthCheck](), []validate.Case{
		{Name: "health check none", Input: map[string]any{"timeout": "5s"},
			WantErr: "must contain at least one of [http_health_check, tcp_health_check, grpc_health_check, custom_health_check]"},
		{Name: "health check two", Input: map[string]any{
			"http_health_check": map[string]any{"path": "/"}, "tcp_health_check": map[string]any{}},
			WantErr: "contains a conflict between exclusive peers"},
		{Name: "http path required", Input: map[string]any{"http_health_check": map[string]any{}},
			WantErr: `"http_health_check.path" is required`},
	})
	run(t, ruleFor[HealthCheckPayload](), []validate.Case{
		{Name: "payload xor", Input: map[string]any{"text": "t", "binary": "b"},
			WantErr: "contains a conflict between exclusive peers [text, binary]"},
	})
	run(t, ruleFor[KeepaliveSettings](), []validate.Case{
		{Name: "timeout required", Input: map[string]any{"interval": "5s"}, WantErr: `"timeout" is required`},
	})
	run(t, ruleFor[RuntimeFeatureFlag](), []validate.Case{
		{Name: "default_value required", Input: map[string]any{"runtime_key": "k"},
			WantErr: `"default_value" is required`},
	})
}

func TestScalarRules(t *testing.T) {
	run(t, ruleFor[UInt32](), []validate.Case{
		{Name: "integer", Input: 1.5, WantErr: "must be an integer"},
		{Name: "min", Input: -1.0, WantErr: "must be greater than or equal to 0"},
		{Name: "max", Input: 99999999999.0, WantErr: "must be less than or equal to"},
	})
	run(t, ruleFor[Priority](), []validate.Case{
		{Name: "max", Input: 500.0, WantErr: "must be less than or equal to 100"},
		{Name: "fraction allowed", Input: 1.5},
	})
	run(t, ruleFor[Percent](), []validate.Case{
		{Name: "integer", Input: 1.5, WantErr: "must be an integer"},
	})
	run(t, ruleFor[IntRange](), []validate.Case{
		{Name: "integer", Input: map[string]any{"start": 1.5}, WantErr: "must be an integer"},
		{Name: "end below start", Input: map[string]any{"start": 10.0, "end": 1.0},
			WantErr: "must be greater than or equal to"},
		{Name: "ok", Input: map[string]any{"start": 1.0, "end": 10.0}},
	})
	run(t, ruleFor[DoubleRange](), []validate.Case{
		{Name: "fraction allowed", Input: map[string]any{"start": 1.5, "end": 2.5}},
		{Name: "end below start", Input: map[string]any{"start": 10.0, "end": 1.0},
			WantErr: "must be greater than or equal to"},
	})
	run(t, ruleFor[ApiVersion](), []validate.Case{
		{Name: "bad", Input: "V99", WantErr: "must be one of [AUTO, V2, V3]"},
		{Name: "ok", Input: "V3"},
	})
	run(t, ruleFor[StatusCode](), []validate.Case{
		{Name: "bad", Input: "Nope", WantErr: "must be one of"},
		{Name: "ok", Input: "TooManyRequests"},
	})
}

func TestRetryPolicyNumRetriesDefault(t *testing.T) {
	out, verr := validate.Validate(ruleFor[RetryPolicy](), map[string]any{}, nil, validate.Prefs{})
	if verr != nil {
		t.Fatalf("unexpected error: %s", verr.Message)
	}
	if got := out.(map[string]any)["num_retries"]; got != 1.0 {
		t.Errorf("num_retries default: got %#v want 1", got)
	}
}
