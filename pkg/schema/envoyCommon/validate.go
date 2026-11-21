package envoyCommon

// Validation for the shared envoy scalar/value types. The generator emits models.go
// only, so every Joi rule the struct-tag DSL cannot express (alternatives, xor,
// custom, a ref bound) had no home and simply did not exist — these types are reached
// from ~40 paths under gvc/workload `spec.sidecar.envoy`, so attaching the rules HERE
// is what closes those paths rather than tagging each call site.
//
// Ports nodelibs/schema/src/envoyCommon.ts.

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

var (
	// DurationString: `10s`, `15.5s` (envoyCommon.ts:108).
	durationStringRe = regexp.MustCompile(`^\d+(?:\.\d+)?s$`)
	// DurationRestricted: whole seconds only (envoyCommon.ts:114).
	durationRestrictedRe = regexp.MustCompile(`^\d+s$`)
)

const (
	maxUInt32          = 4294967295             // 2^32 - 1
	maxDurationSeconds = 315576000000           // envoyCommon.ts DurationObject
	maxDurationNanos   = 999999999              //
	maxUInt64          = 18446744073709551615.0 // 2^64-1, as Joi computes Math.pow(2,64)-1
)

// ValidationRule ports `UInt32 = Joi.number().integer().min(0).max(2^32-1)`.
func (UInt32) ValidationRule() *validate.Rule {
	return validate.Number().Integer().Min(0).Max(maxUInt32)
}

// ValidationRule ports `UInt64 = Joi.number().integer().min(0).max(2^64-1)`.
func (UInt64) ValidationRule() *validate.Rule {
	return validate.Number().Integer().Min(0).MaxNum(maxUInt64)
}

// ValidationRule ports `Percent = Joi.number().integer().min(0).max(100)`.
func (Percent) ValidationRule() *validate.Rule {
	return validate.Number().Integer().Min(0).Max(100)
}

// ValidationRule ports `Priority = Joi.number().min(0).max(100)`.
func (Priority) ValidationRule() *validate.Rule {
	return validate.Number().Min(0).Max(100)
}

// ValidationRule ports `DurationRestricted = Joi.string().pattern(/^\d+s$/)`.
func (DurationRestricted) ValidationRule() *validate.Rule {
	return validate.String().Regex(durationRestrictedRe)
}

// ValidationRule ports `DurationObject`.
func (DurationObject) ValidationRule() *validate.Rule {
	return validate.ObjectOrdered([]string{"seconds", "nanos"}, map[string]*validate.Rule{
		"seconds": validate.Number().Integer().Min(0).Max(maxDurationSeconds),
		"nanos":   validate.Number().Integer().Min(0).Max(maxDurationNanos),
	})
}

// DurationRule ports `Duration = Joi.alternatives(DurationString, DurationObject)`.
// `Duration` is a bare `any` in Go (a union), so nothing reflects onto it — the rule
// has to be handed to each field that uses it.
//
// The engine's alternatives currently reports a flat `alternatives.match` for every
// failure, where Joi surfaces the failing branch's own error (`string.pattern.base`,
// `object.unknown`, `number.min` on `seconds`) and only collapses to
// `alternatives.types` when the value is neither shape — see the engine note in the
// report. The SHAPE here is right either way; only the message differs.
func DurationRule() *validate.Rule {
	return validate.Alternatives(
		validate.String().Regex(durationStringRe),
		DurationObject{}.ValidationRule(),
	)
}

// durationNanos ports durationToNanos (envoyCommon.ts:93). ok=false when the value is
// neither duration form, which the schema has already rejected.
func durationNanos(v any) (float64, bool) {
	switch t := v.(type) {
	case string:
		var f float64
		if _, err := fmt.Sscanf(t, "%gs", &f); err != nil {
			return 0, false
		}
		return f * 1e9, true
	case map[string]any:
		num := func(k string) float64 {
			f, _ := validate.ToFloat(t[k])
			return f
		}
		return num("seconds")*1e9 + num("nanos"), true
	}
	return 0, false
}

// DurationGreaterThanZero ports durationGreaterThanZeroMethod (envoyCommon.ts:117).
func DurationGreaterThanZero(value any, h *validate.Helper) (any, error) {
	if n, ok := durationNanos(value); ok && n > 0 {
		return value, nil
	}
	return nil, h.Errorf("any.custom", "duration not greater than 0")
}

// ValidationRule ports TimeInterval (envoyCommon.ts:218): base_interval is a required
// Duration that must be > 0, and max_interval (when present) must not be smaller.
//
// TS spells the first check as a `.custom()` ON base_interval. The engine does not run
// a rule's customs after an ALTERNATIVES branch matches (validateAlternatives returns
// the matched value directly), so a custom attached to a Duration is silently dropped
// and `base_interval: "0s"` was accepted. Until that is fixed in the engine the check
// runs from the object custom and reports itself at the base_interval path, which is
// the field TS blames and the message TS writes.
func (TimeInterval) ValidationRule() *validate.Rule {
	return validate.ObjectOrdered([]string{"base_interval", "max_interval"}, map[string]*validate.Rule{
		"base_interval": DurationRule().Required(),
		"max_interval":  DurationRule(),
	}).CustomTyped(timeIntervalChecks)
}

// timeIntervalChecks is reflective so it serves BOTH evaluators: the typed walk hands
// it a *TimeInterval, the generic one a map.
func timeIntervalChecks(value any, h *validate.Helper) (any, error) {
	base := validate.Get(value, "base_interval")
	baseN, okBase := durationNanos(base)
	if okBase && baseN <= 0 {
		return nil, h.ErrorAt([]any{"base_interval"}, base, "any.custom", "duration not greater than 0")
	}
	maxV := validate.Get(value, "max_interval")
	if maxV == nil {
		return value, nil
	}
	maxN, okMax := durationNanos(maxV)
	if okMax && okBase && maxN < baseN {
		return nil, h.Errorf("any.custom", "max_interval must be greater than or equal to base_interval")
	}
	return value, nil
}

// ValidationRule ports RangeFactory(Joi.number().integer()) (envoyCommon.ts:34): end must be at
// least start.
func (IntRange) ValidationRule() *validate.Rule {
	return validate.ObjectOrdered([]string{"start", "end"}, map[string]*validate.Rule{
		"start": validate.Number().Integer(),
		"end":   validate.Number().Integer().MinRef("start"),
	})
}

// Duration is a defined INTERFACE type (a Joi alternatives union), so it cannot carry
// a ValidationRule method — register the rule against the type instead.
func init() {
	validate.RegisterTypeRule(reflect.TypeOf((*Duration)(nil)).Elem(), DurationRule)
}

// --- Joi.binary() ---

// Binary is Joi.binary(): the wire carries a string, which Joi (convert:true)
// turns into a Buffer. It is a defined interface type so a single RegisterTypeRule
// covers every binary field rather than tagging each one.
type Binary any

// BinaryRule ports Joi.binary(). Joi coerces an accepted string with
// `Buffer.from(value)` (utf8) and JSON.stringify then renders the Buffer as
// {"type":"Buffer","data":[...]} — which is what TS stores, so the port emits the
// same shape rather than keeping the string.
func BinaryRule() *validate.Rule { return validate.Any().Custom(toBuffer) }

func toBuffer(value any, h *validate.Helper) (any, error) {
	s, ok := value.(string)
	if !ok {
		return nil, h.WithCtx(h.Errorf("binary.base", "%s must be a buffer or a string", strconv.Quote(h.Label())))
	}
	data := make([]any, 0, len(s))
	for _, b := range []byte(s) {
		data = append(data, float64(b))
	}
	return map[string]any{"type": "Buffer", "data": data}, nil
}

// --- xor / nand dependencies (envoyCommon.ts) ---

// ValidationRule ports DataSource's `.xor('filename','inline_bytes','inline_string','environment_variable')`.
func (DataSource) ValidationRule() *validate.Rule {
	return validate.Reflect(DataSource{}).Xor("filename", "inline_bytes", "inline_string", "environment_variable")
}

// ValidationRule ports GrpcService: `.xor('envoy_grpc','google_grpc')` plus
// `timeout: Duration.default('0.200s')`.
func (GrpcService) ValidationRule() *validate.Rule {
	return validate.Reflect(GrpcService{}).
		Fork("timeout", func(r *validate.Rule) *validate.Rule { return r.Default("0.200s") }).
		Xor("envoy_grpc", "google_grpc")
}

// ValidationRule ports `authority: Joi.string().default(Joi.ref('cluster_name'))`.
func (GrpcServiceEnvoyGrpc) ValidationRule() *validate.Rule {
	return validate.Reflect(GrpcServiceEnvoyGrpc{}).
		Fork("authority", func(r *validate.Rule) *validate.Rule { return r.Default(validate.RefTo("cluster_name")) })
}

// ValidationRule ports channel_credentials' `.xor('ssl_credentials','google_default')`.
func (GrpcServiceGoogleGrpcChannelCredentials) ValidationRule() *validate.Rule {
	return validate.Reflect(GrpcServiceGoogleGrpcChannelCredentials{}).Xor("ssl_credentials", "google_default")
}

// RouteMatchXorPeers is RouteMatch's xor peer list. `connect_matcher` is named by the
// Joi `.xor()` although it is not a key of the object (envoyCommon.ts:196) — it must
// stay in the list so the conflict message matches byte for byte.
var RouteMatchXorPeers = []string{"prefix", "path", "safe_regex", "connect_matcher", "path_separated_prefix", "path_match_policy"}

// ValidationRule ports RouteMatch's xor.
func (RouteMatch) ValidationRule() *validate.Rule {
	return validate.Reflect(RouteMatch{}).Xor(RouteMatchXorPeers...)
}

// ValidationRule ports StringMatcher's `.xor('exact','prefix','suffix','safe_regex','contains')`.
func (StringMatcher) ValidationRule() *validate.Rule {
	return validate.Reflect(StringMatcher{}).Xor("exact", "prefix", "suffix", "safe_regex", "contains")
}

// ValidationRule ports QueryParameterMatcher's `.xor('string_match','present_match')`.
func (QueryParameterMatcher) ValidationRule() *validate.Rule {
	return validate.Reflect(QueryParameterMatcher{}).Xor("string_match", "present_match")
}

// ValidationRule ports DoubleMatcher's `.xor('range','exact')`.
func (DoubleMatcher) ValidationRule() *validate.Rule {
	return validate.Reflect(DoubleMatcher{}).Xor("range", "exact")
}

// valueMatcherXorPeers is ValueMatcher's xor list. `list_match`/`or_match` are
// commented out of the key set but still named by the `.xor()` call
// (envoyCommon.ts:160-165), so they stay here for message parity.
var valueMatcherXorPeers = []string{"null_match", "double_match", "string_match", "bool_match", "present_match", "list_match", "or_match"}

// ValidationRule ports ValueMatcher's xor.
func (ValueMatcher) ValidationRule() *validate.Rule {
	return validate.Reflect(ValueMatcher{}).Xor(valueMatcherXorPeers...)
}

// ValidationRule ports ValueMatcher's xor — MetadataMatcherValue is the generator's
// second name for the same Joi schema.
func (MetadataMatcherValue) ValidationRule() *validate.Rule {
	return validate.Reflect(MetadataMatcherValue{}).Xor(valueMatcherXorPeers...)
}

// ValidationRule ports Address' `.xor('socket_address','pipe','envoy_internal_address')`.
func (Address) ValidationRule() *validate.Rule {
	return validate.Reflect(Address{}).Xor("socket_address", "pipe", "envoy_internal_address")
}

// ValidationRule ports ConfigSource's `.xor('path','path_config_source','api_config_source')`.
func (ConfigSource) ValidationRule() *validate.Rule {
	return validate.Reflect(ConfigSource{}).Xor("path", "path_config_source", "api_config_source")
}

// ValidationRule ports HealthCheckPayload's `.xor('text','binary')`.
func (HealthCheckPayload) ValidationRule() *validate.Rule {
	return validate.Reflect(HealthCheckPayload{}).Xor("text", "binary")
}

// ValidationRule ports HealthCheck's xor over the four check kinds.
func (HealthCheck) ValidationRule() *validate.Rule {
	return validate.Reflect(HealthCheck{}).
		Xor("http_health_check", "tcp_health_check", "grpc_health_check", "custom_health_check")
}

// ValidationRule ports SocketAddress: `.xor('port_value','named_port')`, an
// `address` that may not be empty, and a `resolver_name` that becomes required once
// `named_port` is set.
func (SocketAddress) ValidationRule() *validate.Rule {
	return validate.Reflect(SocketAddress{}).
		Fork("address", func(r *validate.Rule) *validate.Rule { return r.Invalid("") }).
		Fork("resolver_name", func(*validate.Rule) *validate.Rule {
			return validate.Any().When("named_port", validate.When{
				Is:        validate.Any().Required(),
				Then:      validate.String().Required(),
				Otherwise: validate.String(),
			})
		}).
		Xor("port_value", "named_port")
}

// ValidationRule ports RangeFactory(Joi.number()) for DoubleRange (envoyCommon.ts:31).
func (DoubleRange) ValidationRule() *validate.Rule {
	return validate.ObjectOrdered([]string{"start", "end"}, map[string]*validate.Rule{
		"start": validate.Number(),
		"end":   validate.Number().MinRef("start"),
	})
}

func init() {
	validate.RegisterTypeRule(reflect.TypeOf((*Binary)(nil)).Elem(), BinaryRule)
}
