package mk8sCommon

// validate.go is the go-libs home of what used to be
// data-service/pkg/handlers/schema_mk8sCommon.go — the shared building blocks every
// mk8s provider sub-schema is composed from (LabelName/LabelValue, Labels, Taint(s),
// NodePoolName, SshPublicKey, AutoscalerConfig, UnmanagedPool, PreInstallScript,
// GoDuration, NetworkingConfig) plus the pod/service network constants and the small
// Joi primitives the engine has no dedicated builder for (string.ip({version,cidr}),
// string.hostname(), string.uri({scheme})), implemented as .Custom() closures.
//
// These are exported building blocks (not Validator methods) because the per-provider
// packages compose them by hand exactly as the old prebuilt provider rules did. The
// twelve provider bodies attach their own ValidationRule methods that call into here.

import (
	"reflect"
	"regexp"
	"strings"

	"github.com/controlplane-com/libs-go/pkg/schema/base"
	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// --- mk8sCommon regexes ---

var (
	// LabelName's domain part (mk8sCommon.ts:21).
	labelDomainRe = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`)
	// LabelName's name part / LabelValue (mk8sCommon.ts:31,46).
	labelNameRe = regexp.MustCompile(`^([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]$`)
	// NodePoolName (mk8sCommon.ts:70).
	nodePoolNameRe = regexp.MustCompile(`^([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]$`)
	// GoDuration (mk8sCommon.ts:10).
	goDurationRe = regexp.MustCompile(`^(\d+(?:\.\d+)?)([hms]+)$`)
)

const labelValueMaxLength = 63

// CplnHostLocations is the Control Plane location enum shared by the ephemeral,
// generic and triton providers (mk8sEphemeral.ts:33 etc.).
var CplnHostLocations = []any{"aws-eu-central-1", "aws-us-east-2", "aws-us-west-2", "gcp-me-west1", "gcp-us-east1"}

// --- accepted networks (mk8sCommon.ts:121-136) ---

var acceptedPodNetworks = []string{
	"10.42.0.0/16", // k3s default, keep it at index 0
	"172.16.0.0/15", "172.18.0.0/15", "172.20.0.0/15", "172.22.0.0/15",
	"172.24.0.0/15", "172.26.0.0/15", "172.28.0.0/15", "172.30.0.0/15",
}

var acceptedServiceNetworks = []string{
	"10.43.0.0/16",   // k3s default, keep it at index 0
	"192.168.0.0/16", // max 65k services
}

// defaultNetworkingConfigMap ports mk8sCommon.ts:138 defaultNetworkingConfig().
func defaultNetworkingConfigMap() map[string]any {
	return map[string]any{
		"serviceNetwork": acceptedServiceNetworks[0],
		"podNetwork":     acceptedPodNetworks[0],
	}
}

func strAnySlice(ss []string) []any {
	out := make([]any, len(ss))
	for i, s := range ss {
		out[i] = s
	}
	return out
}

// --- Joi primitives the engine lacks a dedicated builder for ---

// IpStringRule ports Joi.string().ip({version, cidr}). versions is a subset of
// {"ipv4","ipv6"}; cidr is "optional" | "forbidden" | "required". The check itself is
// base.IPRule — the single Joi-faithful implementation (this wrapper stays because
// twelve provider packages name it).
func IpStringRule(versions []string, cidr string) *validate.Rule {
	return base.IPRule(versions, cidr)
}

// HostnameStringRule ports Joi.string().hostname() — base.HostnameRule is the one
// implementation (it is Joi's domain-analyze, not an RFC-1123 regex, so a TLD that
// starts with a digit is rejected the way Joi rejects it).
func HostnameStringRule() *validate.Rule { return base.HostnameRule() }

// UriStringRule ports Joi.string().uri({scheme}). With no schemes it is the plain
// .URI(); with schemes the value must additionally carry one of them.
func UriStringRule(schemes ...string) *validate.Rule {
	r := validate.String().URI()
	if len(schemes) == 0 {
		return r
	}
	return r.Custom(func(value any, h *validate.Helper) (any, error) {
		s, _ := value.(string)
		for _, sch := range schemes {
			if strings.HasPrefix(s, sch+"://") {
				return s, nil
			}
		}
		// Joi renders the scheme option as an alternation ("https", or "http|https"),
		// and the message quotes the LABEL, not the value.
		scheme := strings.Join(schemes, "|")
		return nil, h.WithCtx(
			h.Errorf("string.uriCustomScheme", "%q must be a valid uri with a scheme matching the %s pattern", h.Label(), scheme),
			"scheme", scheme)
	})
}

// GoDurationRule ports mk8sCommon.ts:10 GoDuration.
func GoDurationRule() *validate.Rule { return validate.String().Regex(goDurationRe) }

// --- label primitives (mk8sCommon.ts:12-71) ---

// labelNameRule ports LabelName (mk8sCommon.ts:12).
func labelNameRule() *validate.Rule {
	return validate.String().Custom(func(value any, h *validate.Helper) (any, error) {
		v, _ := value.(string)
		parts := strings.Split(v, "/")
		if len(parts) != 1 && len(parts) != 2 {
			return nil, h.Errorf("any.custom", "label name invalid format")
		}
		if len(parts) == 2 {
			if !labelDomainRe.MatchString(parts[0]) {
				return nil, h.Errorf("any.custom", "invalid domain part for label name")
			}
		}
		name := parts[0]
		if len(parts) == 2 {
			name = parts[1]
		}
		if !labelNameRe.MatchString(name) {
			return nil, h.Errorf("any.custom", "invalid label name")
		}
		return v, nil
	})
}

// labelValueRule ports LabelValue (mk8sCommon.ts:37).
func labelValueRule() *validate.Rule {
	return validate.String().Allow("").Custom(func(value any, h *validate.Helper) (any, error) {
		v, _ := value.(string)
		if validate.JSLen(v) > labelValueMaxLength {
			return nil, h.Errorf("any.custom", "label value must be max %d long", labelValueMaxLength)
		}
		if !labelNameRe.MatchString(v) {
			return nil, h.Errorf("any.custom", "label value is invalid")
		}
		return v, nil
	})
}

// LabelsRule ports Labels (mk8sCommon.ts:53).
func LabelsRule() *validate.Rule {
	return validate.Object(nil).Pattern(labelNameRule().Required(), labelValueRule())
}

// taintRule ports Taint (mk8sCommon.ts:57).
func taintRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"key":    labelNameRule(),
		"value":  labelValueRule(),
		"effect": validate.String().Valid("NoSchedule", "PreferNoSchedule", "NoExecute"),
	})
}

// TaintsRule ports Taints (mk8sCommon.ts:63).
func TaintsRule() *validate.Rule { return validate.Array(taintRule()).Dedupe() }

// NodePoolNameRule ports NodePoolName (mk8sCommon.ts:69).
func NodePoolNameRule() *validate.Rule {
	return validate.String().Regex(nodePoolNameRe).Max(63)
}

// MaxPodsRule ports MaxPods (mk8sCommon.ts:68): an optional per-nodepool pod cap.
func MaxPodsRule() *validate.Rule {
	return validate.Number().Integer().Min(16).Max(250)
}

// SshPublicKeyRule ports SshPublicKey (mk8sCommon.ts:73).
func SshPublicKeyRule() *validate.Rule {
	return validate.String().Custom(func(value any, h *validate.Helper) (any, error) {
		v, _ := value.(string)
		for _, p := range []string{"ssh-rsa", "ssh-ed25519", "ecdsa-sha2-nistp256", "ssh-ecdsa", "ssh-dsa"} {
			if strings.HasPrefix(v, p+" ") {
				return v, nil
			}
		}
		return nil, h.Errorf("any.custom", "invalid SSH public key")
	})
}

// AutoscalerConfigRule ports AutoscalerConfig (mk8sCommon.ts:98).
func AutoscalerConfigRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"expander": validate.Array(validate.String().Valid("random", "most-pods", "least-waste")).
			Max(3).Min(1).Unique().Default(func() any { return []any{"most-pods"} }),
		"unneededTime":         GoDurationRule().Default("10m"),
		"unreadyTime":          GoDurationRule().Default("20m"),
		"utilizationThreshold": validate.Number().NumMin(0.1).NumMax(1).Default(float64(0.7)),
	}).Default(func() any {
		return map[string]any{
			"expander":             []any{"most-pods"},
			"unneededTime":         "10m",
			"unreadyTime":          "20m",
			"utilizationThreshold": float64(0.7),
		}
	})
}

// UnmanagedPoolRule ports UnmanagedPool (mk8sCommon.ts:115).
func UnmanagedPoolRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"name":   NodePoolNameRule().Required(),
		"labels": LabelsRule(),
		"taints": TaintsRule(),
	})
}

// PreInstallScriptRule ports PreInstallScript (mk8sCommon.ts:145 / mk8sAws.ts:153
// when allowSSM). ” is .allow()'d (bypasses the custom); a non-” script must
// start with '#!' (or 'resolve:ssm:' on AWS).
func PreInstallScriptRule(allowSSM bool) *validate.Rule {
	return validate.String().Allow("").Trim().Custom(func(value any, h *validate.Helper) (any, error) {
		v, _ := value.(string)
		if v == "" {
			return nil, nil
		}
		if strings.HasPrefix(v, "#!") {
			return v, nil
		}
		if allowSSM && strings.HasPrefix(v, "resolve:ssm:") {
			return v, nil
		}
		return nil, h.Errorf("any.custom", "Script must start with #!")
	})
}

// NetworkingConfigRule ports the per-provider NetworkingConfig object. podVpc adds
// the 'vpc' pod-network option (aws/gcp only); every variant defaults to
// defaultNetworkingConfig and enforces serviceNetwork != podNetwork.
func NetworkingConfigRule(podVpc bool) *validate.Rule {
	podValids := []any{}
	if podVpc {
		podValids = append(podValids, "vpc")
	}
	podValids = append(podValids, strAnySlice(acceptedPodNetworks)...)
	return validate.Object(map[string]*validate.Rule{
		"serviceNetwork": validate.String().Valid(strAnySlice(acceptedServiceNetworks)...).Default(acceptedServiceNetworks[0]).Unmodifiable(),
		"podNetwork":     validate.String().Valid(podValids...).Default(acceptedPodNetworks[0]).Unmodifiable(),
		"dnsForwarder":   validate.String().Max(100).Optional(),
	}).Default(func() any { return defaultNetworkingConfigMap() }).CustomTyped(networkingDistinctCustom)
}

// networkingDistinctCustom ports the NetworkingConfig .custom: serviceNetwork must
// differ from podNetwork.
func networkingDistinctCustom(value any, h *validate.Helper) (any, error) {
	if sn := StrOf(value, "serviceNetwork"); sn != "" && StrOf(value, "podNetwork") == sn {
		return nil, h.Errorf("any.custom", "serviceNetwork must be different from the podNetwork")
	}
	return value, nil
}

// StrOf reads a string field off an object in either form a container custom sees:
// a map (generic evaluator) or a pointer to a struct (typed walk), by json name.
func StrOf(obj any, key string) string {
	if m, ok := obj.(map[string]any); ok {
		s, _ := m[key].(string)
		return s
	}
	if f, ok := validate.Field(obj, key); ok && f.Kind() == reflect.String {
		return f.String()
	}
	return ""
}

// elemsOf iterates the elements of an array field (or the array itself) in either
// form, handing each element in a form StrOf accepts.
func elemsOf(obj any, key string, fn func(el any) error) error {
	var arr reflect.Value
	if key == "" {
		arr = reflect.ValueOf(obj)
	} else if m, ok := obj.(map[string]any); ok {
		arr = reflect.ValueOf(m[key])
	} else if f, ok := validate.Field(obj, key); ok {
		arr = f
	}
	for arr.IsValid() && (arr.Kind() == reflect.Ptr || arr.Kind() == reflect.Interface) {
		if arr.IsNil() {
			return nil
		}
		arr = arr.Elem()
	}
	if !arr.IsValid() || arr.Kind() != reflect.Slice {
		return nil
	}
	for i := 0; i < arr.Len(); i++ {
		el := arr.Index(i)
		var v any
		if el.Kind() == reflect.Struct {
			v = el.Addr().Interface()
		} else {
			v = el.Interface()
		}
		if err := fn(v); err != nil {
			return err
		}
	}
	return nil
}

// PoolNamesUniqueObjectCustom ports the provider-level `.custom` that rejects
// duplicate node-pool names across the given array keys (e.g. nodePools +
// dedicatedServerNodePools + unmanagedNodePools).
func PoolNamesUniqueObjectCustom(keys ...string) validate.CustomFn {
	return func(value any, h *validate.Helper) (any, error) {
		counts := map[string]int{}
		for _, k := range keys {
			if err := elemsOf(value, k, func(p any) error {
				n := StrOf(p, "name")
				counts[n]++
				if counts[n] > 1 {
					return h.Errorf("any.custom", "multiple node pools named %s", n)
				}
				return nil
			}); err != nil {
				return nil, err
			}
		}
		return value, nil
	}
}

// PoolNamesUniqueArrayCustom ports the `.custom` some providers attach directly to
// their nodePools array (aws/azure/gcp/ephemeral): reject duplicate pool names.
func PoolNamesUniqueArrayCustom(value any, h *validate.Helper) (any, error) {
	counts := map[string]int{}
	if err := elemsOf(value, "", func(p any) error {
		n := StrOf(p, "name")
		counts[n]++
		if counts[n] > 1 {
			return h.Errorf("any.custom", "multiple node pools named %s", n)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return value, nil
}
