package workload

// validate.go attaches validation to the canonical workload types via the
// Validator interface (ValidationRule) — the go-libs home of what used to be
// data-service/pkg/handlers/schema_workload*.go + schema_reflect_workload.go. The
// "workload" kind validates against Workload.ValidationRule(); nested object types
// carry their own object-level relations/refines, and the field patterns/customs
// become strongly-typed nominal types (Cpu/Memory/ContainerName/…).
//
// The cross-package nested subtrees (vm.VmSpec, workloadOptions.DefaultOptions/
// LocalOptions, cronjob.*, env.EnvCollection, port.Port, volumeSpec.VolumeSpec)
// keep resolving to their own package's rule (their Validator once migrated, else
// the transitional registry) — this file only ports what belongs to the workload
// package. runOptionsValidator is inlined here because the WorkloadSpec refine
// invokes it directly with the resolved workload type.

import (
	"math"
	"net"
	"path"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/controlplane-com/libs-go/pkg/schema/base"
	"github.com/controlplane-com/libs-go/pkg/schema/port"
	"github.com/controlplane-com/libs-go/pkg/schema/util"
	"github.com/controlplane-com/libs-go/pkg/schema/validate"
	"github.com/controlplane-com/libs-go/pkg/schema/vm"
	"github.com/controlplane-com/libs-go/pkg/schema/volumeSpec"
	"github.com/controlplane-com/libs-go/pkg/schema/workloadOptions"
)

// --- shared value helpers (ported from pkg/handlers, delegated to schema/util) ---

var (
	numOf   = util.NumOf
	asArray = util.AsArray
	truthy  = util.Truthy
)

func containsStr(ss []string, s string) bool { return util.ContainsStr(ss, s) }

// errString is a tiny error wrapper (mirrors the pkg/handlers helper).
type errString string

func (e errString) Error() string { return string(e) }

func label2(h *validate.Helper) string {
	if len(h.Path) == 0 {
		return "value"
	}
	return h.Path[len(h.Path)-1]
}

// --- regexes / constants (workload.ts) ---

const (
	maxCPURange    = 4000
	maxCPURatio    = 4
	maxMemoryRange = 4096
	maxMemoryRatio = 4

	slsMaxTimeoutSeconds = 600
)

var (
	wlMemFmtRe          = regexp.MustCompile(`(?i)^[0-9]+(\.[0-9]{1,3})?(G|M|k|Gi|Mi|Ki)?$`)
	wlMemSplitRe        = regexp.MustCompile(`^([0-9]+(?:\.[0-9]+)?)(.*)$`)
	wlCpuFmtRe          = regexp.MustCompile(`^([0-9]+)(m|(\.[0-9]{1,3}))?$`)
	wlIntPercentRe      = regexp.MustCompile(`^[0-9]+(\.[0-9]+)?%$`)
	wlCronRe            = regexp.MustCompile(`^(\*|(\d+(-\d+)?))(\/\d+)?(,(\*|(\d+(-\d+)?))(\/\d+)?)*(\s(\*|(\d+(-\d+)?))(\/\d+)?(,(\*|(\d+(-\d+)?))(\/\d+)?)*){4}$`)
	wlGpuResourceRe     = regexp.MustCompile(`^[a-zA-Z0-9./_-]*$`)
	wlGpuRuntimeClassRe = regexp.MustCompile(`^[a-zA-Z0-9./]*$`)
	wlWhitespaceRe      = regexp.MustCompile(`\s+`)
	wlNameRe            = regexp.MustCompile(`^[a-z]([-a-z0-9])*[a-z0-9]$`)

	dnsBodyRe     = regexp.MustCompile(`^[a-z0-9.-]+$`)
	allDigitsReWL = regexp.MustCompile(`^[0-9]+$`)
)

// dnsWildcardJSPattern is the JS source of DnsWildcardPrefix (workload.ts:111). RE2
// has no lookahead so the Go check is procedural, but the message must quote the
// pattern the TS schema was written with.
const dnsWildcardJSPattern = `/^(?![0-9]+$)(?!.*-$)([*]?)(?!-)[a-z0-9-.]+$/`

var wlExcludedContainerNames = []string{
	"istio-proxy", "queue-proxy", "istio-validation",
	"cpln-envoy-assassin", "cpln-writer-proxy",
	"cpln-reader-proxy", "cpln-dbaas-config",
}

var wlRetryConditions = []*regexp.Regexp{
	regexp.MustCompile(`^5[0-9][0-9]$`), regexp.MustCompile(`^gateway-error$`), regexp.MustCompile(`^reset$`),
	regexp.MustCompile(`^reset-before-request$`), regexp.MustCompile(`^connect-failure$`), regexp.MustCompile(`^envoy-ratelimited$`),
	regexp.MustCompile(`^retriable-4[0-9][0-9]$`), regexp.MustCompile(`^refused-stream$`), regexp.MustCompile(`^retriable-status-codes$`),
	regexp.MustCompile(`^http3-post-connect-failure$`), regexp.MustCompile(`^cancelled$`), regexp.MustCompile(`^deadline-exceeded$`),
	regexp.MustCompile(`^internal$`), regexp.MustCompile(`^resource-exhausted$`), regexp.MustCompile(`^unavailable$`),
}

// --- scalar nominal types (were custom=/pattern= field tags) ---

// Memory ports `Memory = Joi.string().regex(memFmt).max(20).custom(normalize)`.
func (Memory) ValidationRule() *validate.Rule {
	return validate.String().Regex(wlMemFmtRe).Max(20).Custom(memNormalizeCustom)
}

// Cpu ports `Cpu = Joi.string().regex(cpuFmt).max(20)`.
func (Cpu) ValidationRule() *validate.Rule {
	return validate.String().Regex(wlCpuFmtRe).Max(20)
}

// ScheduleType ports `ScheduleType = Joi.string().custom(cron)`.
func (ScheduleType) ValidationRule() *validate.Rule {
	return validate.String().Custom(scheduleTypeCustom)
}

func (GpuResourceName) ValidationRule() *validate.Rule {
	return validate.String().Regex(wlGpuResourceRe)
}

func (GpuRuntimeClassName) ValidationRule() *validate.Rule {
	return validate.String().Regex(wlGpuRuntimeClassRe)
}

func (ContainerName) ValidationRule() *validate.Rule {
	// `.disallow(...)` is Joi's own any.invalid, which carries the `invalids` list in
	// context and the field's real label; the custom it replaces hard-coded the label
	// as "name" and dropped the list entirely.
	return validate.String().Regex(wlNameRe).Invalid(excludedContainerNames()...).Custom(containerNameCustom)
}

func (WorkingDir) ValidationRule() *validate.Rule {
	return validate.String().Custom(workingDirCustom)
}

func (IntOrPercent) ValidationRule() *validate.Rule {
	// FromNumber: rolloutOptions maxUnavailable/maxSurge accept a JSON number (master
	// f94fb6d78 made IntegerOrPercent a JoiNumberToString).
	return validate.String().FromNumber().Custom(intOrPercentCustom)
}

// DisruptionBudget: oxor(minAvailable,maxUnavailable); each is a whole number or whole
// percentage (WholeNumberOrPercent) with its own eviction-sanity custom (workload.ts).
func (DisruptionBudget) ValidationRule() *validate.Rule {
	wholeNumberOrPercent := func() *validate.Rule {
		return validate.String().FromNumber().Max(5).Custom(wholeNumberOrPercentCustom)
	}
	return validate.Reflect(DisruptionBudget{}).
		Fork("minAvailable", func(*validate.Rule) *validate.Rule {
			return wholeNumberOrPercent().Custom(minAvailableEvictCustom)
		}).
		Fork("maxUnavailable", func(*validate.Rule) *validate.Rule {
			return wholeNumberOrPercent().Custom(maxUnavailableEvictCustom)
		}).
		Oxor("minAvailable", "maxUnavailable")
}

var wholeNumberOrPercentRe = regexp.MustCompile(`^(0|[1-9][0-9]*)%?$`)

func wholeNumberOrPercentCustom(value any, h *validate.Helper) (any, error) {
	v, _ := value.(string)
	if !wholeNumberOrPercentRe.MatchString(v) {
		return nil, h.Errorf("any.custom", "Invalid value %q. Valid values include: whole numbers, and whole percentages e.g. 2, 20%%", v)
	}
	return v, nil
}

func minAvailableEvictCustom(value any, h *validate.Helper) (any, error) {
	v, _ := value.(string)
	if strings.HasSuffix(v, "%") {
		if n, err := strconv.ParseFloat(strings.TrimSuffix(v, "%"), 64); err == nil && n >= 100 {
			return nil, h.Errorf("any.custom", "%q never allows a replica to be evicted; use less than 100%%", v)
		}
	}
	return v, nil
}

func maxUnavailableEvictCustom(value any, h *validate.Helper) (any, error) {
	v, _ := value.(string)
	n, _ := strconv.ParseFloat(strings.TrimSuffix(v, "%"), 64)
	if n == 0 {
		return nil, h.Errorf("any.custom", "%q never allows a replica to be evicted; use at least 1 or a percentage above 0%%", v)
	}
	if strings.HasSuffix(v, "%") && n > 100 {
		return nil, h.Errorf("any.custom", "%q must be at most 100%%", v)
	}
	return v, nil
}

func autoscalingMinScale(a *workloadOptions.OptionsAutoscaling) int {
	if a != nil && a.MinScale != nil {
		return int(*a.MinScale)
	}
	return 1
}

func defaultOptionsMinScale(d *workloadOptions.DefaultOptions) int {
	if d == nil {
		return 1
	}
	return autoscalingMinScale(d.Autoscaling)
}

// minAvailableBlocksDrains ports workload.ts: judged at one replica when minScale is 0,
// the fewest a workload runs while it has pods.
func minAvailableBlocksDrains(minAvailable string, minScale int) bool {
	replicas := minScale
	if replicas < 1 {
		replicas = 1
	}
	n, _ := strconv.Atoi(strings.TrimSuffix(minAvailable, "%"))
	if n == 0 {
		return false
	}
	if strings.HasSuffix(minAvailable, "%") {
		return int(math.Ceil(float64(n)/100*float64(replicas))) >= replicas
	}
	return n >= replicas
}

func (WorkloadName) ValidationRule() *validate.Rule {
	return validate.String().Regex(wlNameRe)
}

// --- leaf nominal types (slice elements — cannot carry field tags) ---

// CIDR ports `Joi.string().ip({version:[ipv4,ipv6], cidr:'optional'})`.
func (CIDR) ValidationRule() *validate.Rule {
	return validate.String().Custom(func(value any, h *validate.Helper) (any, error) {
		s, _ := value.(string)
		if ipOrCidrOK(s) {
			return s, nil
		}
		return nil, h.WithCtx(
			h.Errorf("string.ipVersion", "%s must be a valid ip address of one of the following versions [ipv4, ipv6] with a optional CIDR", quoteLabel(h)),
			"value", s, "cidr", "optional", "version", []string{"ipv4", "ipv6"})
	})
}

// ipOrCidrOK is Joi's ip({version:[ipv4,ipv6], cidr:'optional'}). Joi matches a
// PATTERN, so `010.1.1.1` (leading zeros) is a valid dotted quad to it; net.ParseIP
// has rejected leading zeros since Go 1.17, which turned addresses TS stores into a
// 400. Leading zeros are stripped before the parse rather than parsed by hand.
func ipOrCidrOK(s string) bool {
	host, mask, hasMask := strings.Cut(s, "/")
	norm, ok := jsIPv4Normalize(host)
	if ok {
		host = norm
	}
	if !hasMask {
		return net.ParseIP(host) != nil
	}
	if _, _, err := net.ParseCIDR(host + "/" + mask); err == nil {
		return true
	}
	return false
}

// jsIPv4Normalize strips leading zeros from each octet of a dotted quad, reporting
// whether the input looked like one at all.
func jsIPv4Normalize(s string) (string, bool) {
	parts := strings.Split(s, ".")
	if len(parts) != 4 {
		return s, false
	}
	out := make([]string, 4)
	for i, p := range parts {
		if p == "" || len(p) > 3 || !allDigitsReWL.MatchString(p) {
			return s, false
		}
		n, err := strconv.Atoi(p)
		if err != nil || n > 255 {
			return s, false
		}
		out[i] = strconv.Itoa(n)
	}
	return strings.Join(out, "."), true
}

// quoteLabel is the `"<full label>"` Joi puts at the head of its messages — the ported
// customs used to substitute the offending VALUE there, which named no field at all.
func quoteLabel(h *validate.Helper) string { return strconv.Quote(h.Label()) }

// DnsName ports DnsWildcardPrefix (workload.ts:111) — Go RE2 has no lookahead, so
// the not-all-digits / not-trailing-dash / optional-leading-* checks are procedural.
func (DnsName) ValidationRule() *validate.Rule {
	return validate.String().Max(128).Custom(func(value any, h *validate.Helper) (any, error) {
		s, _ := value.(string)
		fail := func() (any, error) {
			return nil, h.WithCtx(
				h.Errorf("string.pattern.base", "%s with value %q fails to match the required pattern: %s", quoteLabel(h), s, dnsWildcardJSPattern),
				"value", s, "regex", map[string]any{})
		}
		if allDigitsReWL.MatchString(s) || strings.HasSuffix(s, "-") {
			return fail()
		}
		body := s
		if strings.HasPrefix(body, "*") {
			body = body[1:]
		}
		if body == "" || strings.HasPrefix(body, "-") || !dnsBodyRe.MatchString(body) {
			return fail()
		}
		return s, nil
	})
}

// WorkloadLink ports LocalLink + ToKind("!workload","internal","agent").
func (WorkloadLink) ValidationRule() *validate.Rule {
	return base.LocalLinkRule().Custom(base.ToKind("!workload", "internal", "agent"))
}

// IpSetLink ports LocalLink().custom(toKind("ipset")): a localLink pointing to a
// resource of kind ipset.
func (IpSetLink) ValidationRule() *validate.Rule {
	return base.LocalLinkRule().Custom(base.ToKind("ipset"))
}

// IdentityLink ports LocalLink().custom(toKind("identity")): a localLink pointing to
// a resource of kind identity.
func (IdentityLink) ValidationRule() *validate.Rule {
	return base.LocalLinkRule().Custom(base.ToKind("identity"))
}

// RetryCondition ports the requestRetryPolicy.retryOn[] element check.
func (RetryCondition) ValidationRule() *validate.Rule {
	return validate.String().Max(32).Custom(retryOnCustom)
}

// --- collection nominal types (array-level customs: dedupe / unique) ---

func (CIDRList) ValidationRule() *validate.Rule {
	return validate.Array(validate.FromType(reflect.TypeOf(CIDR("")))).Dedupe()
}

func (WorkloadLinkList) ValidationRule() *validate.Rule {
	return validate.Array(validate.FromType(reflect.TypeOf(WorkloadLink("")))).Dedupe()
}

func (HeaderFilterList) ValidationRule() *validate.Rule {
	return validate.Array(validate.FromType(reflect.TypeOf(HeaderFilter{}))).Dedupe("key")
}

// SortBy precedes the dedupe, matching workload.ts:155 `.sort({by:'name'})
// .custom(makeUnique())` — the sort was missing, so headers persisted in
// submitted order.
func (ReadinessHttpHeaderList) ValidationRule() *validate.Rule {
	return validate.Array(validate.FromType(reflect.TypeOf(ContainerSpecReadinessProbeHttpGetHttpHeaders{}))).
		SortBy("name").Dedupe("name")
}

func (LivenessHttpHeaderList) ValidationRule() *validate.Rule {
	return validate.Array(validate.FromType(reflect.TypeOf(ContainerSpecLivenessProbeHttpGetHttpHeaders{}))).
		SortBy("name").Dedupe("name")
}

func (VolumeList) ValidationRule() *validate.Rule {
	return validate.Array(validate.FromType(reflect.TypeOf(volumeSpec.VolumeSpec{}))).CustomTyped(validate.CustomT(volumesUnique))
}

// --- field customs (bodies ported from pkg/handlers) ---

// memNormalizeCustom normalises the memory unit casing.
func memNormalizeCustom(value any, h *validate.Helper) (any, error) {
	v, _ := value.(string)
	m := wlMemSplitRe.FindStringSubmatch(v)
	if m == nil {
		return v, nil
	}
	number, suffix := m[1], strings.ToLower(m[2])
	if suffix == "" {
		return v, nil
	}
	if suffix != "k" {
		suffix = strings.ToUpper(suffix[:1]) + strings.ToLower(suffix[1:])
	}
	return number + suffix, nil
}

func scheduleTypeCustom(value any, h *validate.Helper) (any, error) {
	v, _ := value.(string)
	if !wlCronRe.MatchString(v) {
		return nil, h.Errorf("any.custom", "the given schedule does not match the standard cron format.")
	}
	if len(wlWhitespaceRe.Split(v, -1)) != 5 {
		return nil, h.Errorf("any.custom", "the given schedule does not match the standard cron format.")
	}
	if err := checkCronValues(v); err != nil {
		return nil, h.Errorf("any.custom", "%s", err.Error())
	}
	return v, nil
}

// excludedContainerNames is ExcludedContainerNames (workload.ts:256-264) IN ORDER —
// Joi reports the disallow list verbatim in the error context.
func excludedContainerNames() []any {
	out := make([]any, len(wlExcludedContainerNames))
	for i, n := range wlExcludedContainerNames {
		out[i] = n
	}
	return out
}

// containerNameCustom ports the cpln-/debugger- prefix check.
func containerNameCustom(value any, h *validate.Helper) (any, error) {
	s, _ := value.(string)
	if strings.HasPrefix(s, "cpln-") {
		return nil, h.Errorf("any.custom", "Container names cannot start with cpln-")
	}
	if strings.HasPrefix(s, "debugger-") {
		return nil, h.Errorf("any.custom", "Container names cannot start with debugger-")
	}
	return s, nil
}

// workingDirCustom ports ContainerSpec.workingDir (absolute path).
func workingDirCustom(value any, h *validate.Helper) (any, error) {
	s, _ := value.(string)
	s = path.Clean(s)
	if !path.IsAbs(s) {
		return nil, h.Errorf("any.custom", "workingDir must be an absolute path")
	}
	return s, nil
}

// intOrPercentCustom ports integerOrPercent (rolloutOptions maxUnavailable/maxSurge).
func intOrPercentCustom(value any, h *validate.Helper) (any, error) {
	v, _ := value.(string)
	if _, err := strconv.Atoi(v); err != nil && !wlIntPercentRe.MatchString(v) {
		return nil, h.Errorf("any.custom", "Invalid value for rolloutOptions: %q. Valid values include: integer values, and floating point percentages e.g. 2, 20%%, 10.5%%", v)
	}
	return v, nil
}

// retryOnCustom ports the requestRetryPolicy.retryOn[] element check.
func retryOnCustom(value any, h *validate.Helper) (any, error) {
	s, _ := value.(string)
	for _, r := range wlRetryConditions {
		if r.MatchString(s) {
			return s, nil
		}
	}
	return nil, h.Errorf("any.custom", "Invalid policy: %s", s)
}

// volumesUnique ports the container volumes .unique() + mount-path checks.
func volumesUnique(vols *VolumeList, h *validate.Helper) error {
	arr := *vols
	for i := 0; i < len(arr); i++ {
		for j := i + 1; j < len(arr); j++ {
			if reflect.DeepEqual(arr[i], arr[j]) {
				e := h.ErrorAt([]any{j}, arr[j], "array.unique", "%s contains a duplicate value", strconv.Quote(h.Label()+"["+strconv.Itoa(j)+"]"))
				return h.WithCtx(e, "pos", j, "dupePos", i, "dupeValue", arr[i])
			}
		}
	}
	for i := range arr {
		for j := range arr {
			if reflect.DeepEqual(arr[i], arr[j]) {
				continue
			}
			ap, bp := string(arr[i].Path), string(arr[j].Path)
			if ap != "" && ap == bp {
				return h.Errorf("any.custom", "Multiple volumes are mounted on the same path %s", ap)
			}
			if bp != "" && strings.HasPrefix(ap+"/", bp+"/") {
				return h.Errorf("any.custom", "%s cannot be parent of %s", bp, ap)
			}
		}
	}
	return nil
}

func (ContainerImage) ValidationRule() *validate.Rule { return base.ImageLinkRule() }

// ResolvedImageManifestsPlatform ports `Joi.object().pattern(Joi.string(),
// Joi.string().allow(”))` (workload.ts:575). The generic map rule rejected the empty
// string (Joi string.empty), so a manifest whose platform carries e.g. `variant: ""` —
// what a registry actually returns — was refused where TS stores it.
func (ResolvedImageManifestsPlatform) ValidationRule() *validate.Rule {
	return validate.Object(nil).Pattern(validate.String(), validate.String().Allow(""))
}

// --- object-level Validators (Reflect + the refines/relations) ---

// Workload is the schema the "workload" kind validates against.
func (Workload) ValidationRule() *validate.Rule {
	return validate.Reflect(Workload{}).CustomTyped(workloadRootRefine)
}

func (WorkloadSpec) ValidationRule() *validate.Rule {
	return validate.RefineT(validate.Reflect(WorkloadSpec{}), workloadSpecRefine)
}

func (WorkloadStatus) ValidationRule() *validate.Rule {
	return validate.Reflect(WorkloadStatus{}).Unknown(true)
}

func (ContainerSpecReadinessProbe) ValidationRule() *validate.Rule {
	return validate.Reflect(ContainerSpecReadinessProbe{}).Xor("exec", "grpc", "tcpSocket", "httpGet")
}

func (ContainerSpecLivenessProbe) ValidationRule() *validate.Rule {
	return validate.Reflect(ContainerSpecLivenessProbe{}).Xor("exec", "grpc", "tcpSocket", "httpGet")
}

func (HeaderFilter) ValidationRule() *validate.Rule {
	return validate.Reflect(HeaderFilter{}).Nand("allowedValues", "blockedValues")
}

// ValidationRule ports GpuResource's `.xor('nvidia','custom')` + `.error()` remap
// (workload.ts:289-300). TS remaps ONLY object.xor, so the "neither present" branch
// keeps Joi's own object.missing message — hence Or() for that branch and the refine
// for the conflict, rather than one Xor() whose message cannot be overridden.
func (ContainerSpecGpu) ValidationRule() *validate.Rule {
	return validate.RefineT(validate.Reflect(ContainerSpecGpu{}).Or("nvidia", "custom"), gpuResourceRefine)
}

func (WorkloadSpecLoadBalancer) ValidationRule() *validate.Rule {
	return validate.RefineT(validate.Reflect(WorkloadSpecLoadBalancer{}), loadBalancerRefine)
}

func (WorkloadSpecRequestRetryPolicy) ValidationRule() *validate.Rule {
	return validate.RefineT(validate.Reflect(WorkloadSpecRequestRetryPolicy{}), retryOnDefaultRefine)
}

// gpuResourceRefine is the object.xor arm of GpuResource's xor: both peers present.
// The neither-present arm is Or()'s object.missing, which TS leaves unremapped.
func gpuResourceRefine(g *ContainerSpecGpu, h *validate.Helper) error {
	if g.Nvidia != nil && g.Custom != nil {
		e := h.Errorf("object.xor", "The parameter spec.gpu must contain one of nvidia or custom and not both")
		return h.WithCtx(e, "value", g, "peers", []string{"nvidia", "custom"},
			"peersWithLabels", []string{"nvidia", "custom"}, "present", []string{"nvidia", "custom"},
			"presentWithLabels", []string{"nvidia", "custom"})
	}
	return nil
}

// loadBalancerRefine ports loadBalancer.custom (geoLocation headers).
func loadBalancerRefine(lb *WorkloadSpecLoadBalancer, h *validate.Helper) error {
	geo := lb.GeoLocation
	if geo != nil && util.Bool(geo.Enabled) {
		var present []string
		if hdrs := geo.Headers; hdrs != nil {
			for _, s := range []string{hdrs.Asn, hdrs.City, hdrs.Country, hdrs.Region} {
				if s != "" {
					present = append(present, s)
				}
			}
		}
		if len(present) == 0 {
			return h.Errorf("any.custom", "When geoLocation is enabled, at least one header for asn, city, country or region must be set")
		}
		seen := map[string]bool{}
		for _, s := range present {
			if seen[s] {
				return h.Errorf("any.custom", "When geoLocation is enabled, headers for asn/city/country/region must be unique")
			}
			seen[s] = true
		}
	}
	return nil
}

// retryOnDefaultRefine applies requestRetryPolicy.retryOn's default list when absent.
func retryOnDefaultRefine(p *WorkloadSpecRequestRetryPolicy, h *validate.Helper) error {
	if !h.Present("retryOn") {
		p.RetryOn = []RetryCondition{"connect-failure", "refused-stream", "unavailable", "cancelled", "resource-exhausted", "retriable-status-codes"}
	}
	return nil
}

// --- WorkloadSpec refine (workload.ts:946) — checks wholly within the spec ---

func workloadSpecRefine(spec *WorkloadSpec, h *validate.Helper) error {
	typ := string(spec.Type)
	containers := spec.Containers

	if typ == "stateful" && spec.RolloutOptions != nil {
		spec.RolloutOptions.MaxUnavailableReplicas = ""
	}
	if typ == "vm" && spec.RolloutOptions != nil {
		spec.RolloutOptions.MaxSurgeReplicas = ""
		spec.RolloutOptions.ScalingPolicy = ""
		spec.RolloutOptions.RolloutOnReferenceChange = nil
	}

	if spec.DefaultOptions != nil {
		if err := runOptionsValidator(typ, defaultOptionsView(spec.DefaultOptions)); err != nil {
			return h.Errorf("any.custom", "%s", err.Error())
		}
	}
	for i := range spec.LocalOptions {
		if err := runOptionsValidator(typ, localOptionsView(&spec.LocalOptions[i])); err != nil {
			return h.Errorf("any.custom", "%s", err.Error())
		}
	}

	if err := checkForDuplicatePortNumber(containers); err != nil {
		return h.Errorf("any.custom", "%s", err.Error())
	}
	capacityAIEnabled := spec.DefaultOptions != nil && util.Bool(spec.DefaultOptions.CapacityAI)
	for i := range spec.LocalOptions {
		if util.Bool(spec.LocalOptions[i].CapacityAI) {
			capacityAIEnabled = true
		}
	}

	switch typ {
	case "serverless":
		if err := checkPortEnv(containers); err != nil {
			return h.Errorf("any.custom", "%s", err.Error())
		}
		var withPorts []*ContainerSpec
		var withMulti []*ContainerSpec
		for i := range containers {
			c := &containers[i]
			// `_.filter(containers, c => c.port ?? c.ports)`: in JS an empty array is
			// truthy, so `ports: []` counts as exposing a port. It is a TS quirk, but it
			// is the accept/reject boundary clients were built against.
			if c.Port != nil || c.Ports != nil {
				withPorts = append(withPorts, c)
			}
			if len(c.Ports) > 1 {
				withMulti = append(withMulti, c)
			}
		}
		if len(withPorts) > 1 {
			return h.Errorf("any.custom", "a serverless workload can not expose a port in more than one container. The following containers expose ports: %s", portsMsg(withPorts))
		}
		if len(withPorts) == 0 {
			return h.Errorf("any.custom", "workloads of type 'serverless' must have at least one container port exposed.")
		}
		if len(withMulti) > 0 {
			return h.Errorf("any.custom", "a serverless workload can not expose multiple ports in any container. The following containers expose multiple ports: %s", portsMsg(withMulti))
		}
	case "cron":
		if spec.Job == nil {
			return h.Errorf("any.custom", "the job property is required for workloads of type 'cron'")
		}
		for i := range containers {
			containers[i].LivenessProbe = nil
			containers[i].ReadinessProbe = nil
		}
	case "vm":
		if err := validateVmWorkload(spec, containers); err != nil {
			return h.Errorf("any.custom", "%s", err.Error())
		}
	}

	if typ != "vm" {
		if spec.Vm != nil {
			return h.Errorf("any.custom", "the 'vm' property is only valid for workloads of type 'vm'")
		}
		for i := range containers {
			c := &containers[i]
			name := string(c.Name)
			if c.Image == "" {
				return h.Errorf("any.custom", "container '%s': image is required", name)
			}
			for _, vol := range c.Volumes {
				if vol.Path == "" {
					return h.Errorf("any.custom", "container '%s': volumes[].path is required", name)
				}
				var vmOnly []string
				if vol.Name != "" {
					vmOnly = append(vmOnly, "name")
				}
				if vol.Bus != "" {
					vmOnly = append(vmOnly, "bus")
				}
				if vol.BootOrder != nil {
					vmOnly = append(vmOnly, "bootOrder")
				}
				if len(vmOnly) > 0 {
					return h.Errorf("any.custom", "container '%s': volume fields %s are only valid for workloads of type 'vm'", name, strings.Join(vmOnly, ", "))
				}
			}
		}
	}

	if typ != "cron" && spec.Job != nil {
		return h.Errorf("any.custom", "the job property is only valid for workloads of type 'cron'")
	}

	if (typ == "cron" || typ == "vm") && spec.DisruptionBudget != nil {
		return h.Errorf("any.custom", "the disruptionBudget property is not supported for workloads of type '%s'", typ)
	}
	if spec.DisruptionBudget != nil && spec.DisruptionBudget.MinAvailable != "" {
		minAvailable := spec.DisruptionBudget.MinAvailable
		type scaleAt struct {
			where    string
			minScale int
		}
		scales := []scaleAt{{"defaultOptions", defaultOptionsMinScale(spec.DefaultOptions)}}
		for i := range spec.LocalOptions {
			o := &spec.LocalOptions[i]
			scales = append(scales, scaleAt{"localOptions " + string(o.Location), autoscalingMinScale(o.Autoscaling)})
		}
		for _, s := range scales {
			if minAvailableBlocksDrains(minAvailable, s.minScale) {
				return h.Errorf("any.custom",
					"disruptionBudget.minAvailable %s allows no evictions at minScale %d (%s), which blocks node drains; lower it or use maxUnavailable",
					minAvailable, s.minScale, s.where)
			}
		}
	}

	byName := map[string]int{}
	for i := range containers {
		name := string(containers[i].Name)
		if name == "" {
			return h.Errorf("any.custom", "all containers must be named")
		}
		byName[name]++
	}
	for g, n := range byName {
		if n > 1 {
			return h.Errorf("any.custom", "at most one container can be named %s. Instead there are %d", g, n)
		}
	}

	gpuCount := 0
	for i := range containers {
		if containers[i].Gpu != nil {
			gpuCount++
		}
	}
	if gpuCount == 1 && capacityAIEnabled {
		return h.Errorf("any.custom", "capacity AI must be disabled when using GPUs. Please remove the GPU selection from the containers or disable Capacity AI.")
	}
	if gpuCount > 1 {
		return h.Errorf("any.custom", "only one container can use a gpu")
	}

	if lb := spec.LoadBalancer; lb != nil && util.Bool(lb.ReplicaDirect) {
		if typ != "stateful" {
			return h.Errorf("any.custom", "replicaDirect is only available for stateful workloads")
		}
	}
	return nil
}

func portsMsg(cs []*ContainerSpec) string {
	var parts []string
	for _, c := range cs {
		portStr := ""
		if c.Port != nil {
			portStr = strconv.Itoa(int(*c.Port))
		} else {
			var nums []string
			for _, p := range c.Ports {
				nums = append(nums, strconv.Itoa(int(p.Number)))
			}
			portStr = strings.Join(nums, ",")
		}
		parts = append(parts, string(c.Name)+"="+portStr)
	}
	return strings.Join(parts, ";")
}

func checkForDuplicatePortNumber(containers []ContainerSpec) error {
	seen := map[float64]bool{}
	count := 0
	for i := range containers {
		c := &containers[i]
		for _, p := range c.Ports {
			n := float64(p.Number)
			count++
			seen[n] = true
			if len(seen) != count {
				return errString("The port number must be unique for all containers: duplicate found for container " + string(c.Name) + " port " + strconv.Itoa(int(n)))
			}
		}
	}
	return nil
}

func checkPortEnv(containers []ContainerSpec) error {
	for i := range containers {
		c := &containers[i]
		var port float64
		havePort := false
		if c.Port != nil {
			port, havePort = float64(*c.Port), true
		} else if len(c.Ports) > 0 {
			port, havePort = float64(c.Ports[0].Number), true
		}
		if !havePort {
			return nil
		}
		for _, e := range c.Env {
			if string(e.Name) == "PORT" && string(e.Value) != strconv.Itoa(int(port)) {
				return errString("Container environment variable with the name PORT is not allowed when container.port is set. The PORT environment variable will be provided at runtime automatically")
			}
		}
	}
	return nil
}

// --- Workload root refine (tag/type-dependent container + rollout checks) ---

// workloadRootRefine runs on the root object, which on the write path is the ROW
// (whose spec column wraps WorkloadSpec) and elsewhere the wire struct, so it reads
// its fields by json name rather than through one concrete type.
func workloadRootRefine(w any, h *validate.Helper) (any, error) {
	tags, _ := validate.Get(w, "tags").(map[string]any)
	spec := validate.Sub[WorkloadSpec](w, "spec")
	if spec != nil {
		typ := string(spec.Type)
		capAIEverywhere := capacityAIEverywhere(spec)
		for i := range spec.Containers {
			c := &spec.Containers[i]
			if err := validateContainerCross(c, typ, tags, capAIEverywhere); err != nil {
				// Blame the container, not the workload: these checks read the parent's
				// options (which is why they run here) but Joi reports them at
				// spec.containers[i], and that is the field the caller has to fix.
				return nil, h.ErrorAt([]any{"spec", "containers", i}, c, "any.custom", "%s", err.Error())
			}
		}
		if ro := spec.RolloutOptions; ro != nil && ro.TerminationGracePeriodSeconds != nil {
			maxGrace := 900.0
			if truthy(tags["cpln/relaxGracePeriodMax"]) {
				maxGrace = 3600
			}
			if *ro.TerminationGracePeriodSeconds > maxGrace {
				// TS attaches this custom to RolloutOptions (workload.ts:230-238), so Joi
				// blames spec.rolloutOptions — the field the caller has to fix.
				return nil, h.ErrorAt([]any{"spec", "rolloutOptions"}, ro, "any.custom", "terminationGracePeriodSeconds must be at most %d.", int(maxGrace))
			}
		}
		// concurrency/rps on kata are scaled by the kata-agent (single-metric only),
		// not KEDA/HPA, so they cannot appear in a multi set (workload.ts:1081-1098).
		if s, _ := tags["cpln/kata"].(string); s == "true" {
			if err := checkKataMultiMetrics(defaultOptionsMulti(spec.DefaultOptions)); err != nil {
				return nil, h.ErrorAt([]any{"spec"}, spec, "any.custom", "%s", err.Error())
			}
			for i := range spec.LocalOptions {
				if err := checkKataMultiMetrics(localOptionsMulti(&spec.LocalOptions[i])); err != nil {
					return nil, h.ErrorAt([]any{"spec"}, spec, "any.custom", "%s", err.Error())
				}
			}
		}
	}
	if strings.HasSuffix(validate.Str(w, "name"), "-headless") {
		return nil, h.Errorf("any.custom", "workload name cannot end with -headless")
	}
	if spec != nil {
		org, gvc := "", ""
		if h.Ctx != nil {
			org, gvc = h.Ctx.Org, h.Ctx.Gvc
		}
		if err := applySandboxSubtype(w, spec, org, gvc, h); err != nil {
			return nil, err
		}
	}
	return w, nil
}

// checkKataMultiMetrics ports the kata multi-metric guard (workload.ts:1082-1093).
func checkKataMultiMetrics(multi []workloadOptions.OptionsAutoscalingMulti) error {
	if multi == nil {
		return nil
	}
	var offending []string
	for _, m := range multi {
		if metric, _ := m.Metric.(string); metric == "concurrency" || metric == "rps" {
			offending = append(offending, "'"+metric+"'")
		}
	}
	if len(offending) == 0 {
		return nil
	}
	plural := ""
	if len(offending) > 1 {
		plural = "s"
	}
	return errString("autoscaling metric" + plural + " " + strings.Join(offending, ", ") +
		" cannot be used in a multi-metric set on a kata (cpln/kata tag) workload; concurrency/rps are scaled by the kata-agent and support single-metric only")
}

// autoscalingMulti reads `<options>.autoscaling.multi` off either options struct.
func defaultOptionsMulti(o *workloadOptions.DefaultOptions) []workloadOptions.OptionsAutoscalingMulti {
	if o == nil || o.Autoscaling == nil {
		return nil
	}
	return o.Autoscaling.Multi
}

func localOptionsMulti(o *workloadOptions.LocalOptionsItem) []workloadOptions.OptionsAutoscalingMulti {
	if o == nil || o.Autoscaling == nil {
		return nil
	}
	return o.Autoscaling.Multi
}

// capacityAIEverywhere ports the capacityAIEnabledEverywhere expression of
// ContainerSpec.custom (workload.ts:456-458, 6ea7a5804): capacityAI is on by default
// AND no per-location override turns it off. It reads the PARENT WORKLOAD's spec — the
// pre-6ea7a5804 expression read `value.defaultOptions`, where value is the CONTAINER, so
// it was always undefined and the flag was always false.
//
// An empty localOptions is `every` over nothing: true, i.e. defaultOptions alone decides.
func capacityAIEverywhere(spec *WorkloadSpec) bool {
	if spec.DefaultOptions == nil || !util.Bool(spec.DefaultOptions.CapacityAI) {
		return false
	}
	for i := range spec.LocalOptions {
		if !util.Bool(spec.LocalOptions[i].CapacityAI) {
			return false
		}
	}
	return true
}

// probePortRef is one probe section's port, addressable across the six generated probe
// types (readiness and liveness each inline their own httpGet/tcpSocket/grpc).
type probePortRef struct {
	kind string
	port **port.Port
}

// defaultProbePorts ports the probe loop of ContainerSpec.custom (workload.ts:425-440,
// 160c232a6): every probe kind takes the container's port when it names none, a grpc
// probe preferring a grpc port. Only the types that deploy probes as-is then require one
// — cron strips probes, and serverless drops them from a portless container.
func defaultProbePorts(c *ContainerSpec, specType string, firstPort func() *port.Port) error {
	var refs []probePortRef
	if p := c.ReadinessProbe; p != nil {
		if p.HttpGet != nil {
			refs = append(refs, probePortRef{"httpGet", &p.HttpGet.Port})
		}
		if p.TcpSocket != nil {
			refs = append(refs, probePortRef{"tcpSocket", &p.TcpSocket.Port})
		}
		if p.Grpc != nil {
			refs = append(refs, probePortRef{"grpc", &p.Grpc.Port})
		}
	}
	if p := c.LivenessProbe; p != nil {
		if p.HttpGet != nil {
			refs = append(refs, probePortRef{"httpGet", &p.HttpGet.Port})
		}
		if p.TcpSocket != nil {
			refs = append(refs, probePortRef{"tcpSocket", &p.TcpSocket.Port})
		}
		if p.Grpc != nil {
			refs = append(refs, probePortRef{"grpc", &p.Grpc.Port})
		}
	}
	probesDeployedAsIs := specType == "standard" || specType == "stateful" || specType == "vm"
	for _, r := range refs {
		if *r.port != nil {
			continue
		}
		if r.kind == "grpc" {
			for i := range c.Ports {
				if c.Ports[i].Protocol == ContainerSpecPortsProtocolGrpc {
					n := c.Ports[i].Number
					*r.port = &n
					break
				}
			}
		}
		if *r.port == nil {
			*r.port = firstPort()
		}
		if *r.port == nil && probesDeployedAsIs {
			return errString("Probes using " + r.kind + " must have a port defined.")
		}
	}
	return nil
}

// validateContainerCross ports ContainerSpec.custom (workload.ts:421-541).
func validateContainerCross(c *ContainerSpec, specType string, tags map[string]any, capacityAIEverywhere bool) error {
	firstPort := func() *port.Port {
		if len(c.Ports) > 0 {
			p := c.Ports[0].Number
			return &p
		}
		return c.Port
	}
	if err := defaultProbePorts(c, specType, firstPort); err != nil {
		return err
	}
	if c.Port != nil && c.Ports != nil {
		return errString("Only one of 'port' and 'ports' may be specified.")
	}

	memStr := string(c.Memory)
	cpuStr := string(c.Cpu)
	maxRatio := 8.0
	if truthy(tags["cpln/relaxMemoryToCpuRatio"]) {
		maxRatio = 32
	}
	cpuNum := wlCpuNum(cpuStr)
	if cpuNum != 0 && base.MemoryToMiB(memStr)/cpuNum > maxRatio {
		return errString("Memory (MiB) to CPU (millicore) ratio must be less than " + strconv.Itoa(int(maxRatio)) + ", please decrease memory or increase CPU.")
	}
	if wlCpuNum(cpuStr) < 25 {
		return errString("CPU must be at least 25 millicore.")
	}
	if base.MemoryToMiB(memStr) < 32 {
		return errString("Memory must be at least 32 MiB.")
	}

	if minCpuStr := string(c.MinCpu); minCpuStr != "" {
		minCpu := wlCpuNum(minCpuStr)
		cpu := wlCpuNum(cpuStr)
		cpuRange := cpu - minCpu
		if minCpu < 25 {
			return errString("Minimum CPU must be at least 25 millicore.")
		}
		if cpuRange < 0 {
			return errString("Minimum CPU can not be greater than CPU.")
		}
		// With capacityAI on everywhere the scheduler sizes the pod from live usage, so
		// the stateful min/max spread limits do not apply (workload.ts:471, 6ea7a5804).
		if specType == "stateful" && !capacityAIEverywhere {
			if cpu/minCpu > maxCPURatio {
				return errString("The ratio between cpu and minCpu must be less than " + strconv.Itoa(maxCPURatio) + ":1, please increase minCpu or decrease cpu.")
			}
			if cpuRange > maxCPURange {
				return errString("The difference between minCpu and cpu must be less than " + strconv.Itoa(maxCPURange) + " millicores, please increase minCpu or decrease cpu.")
			}
		}
	}
	if minMemStr := string(c.MinMemory); minMemStr != "" {
		minMem := base.MemoryToMiB(minMemStr)
		mem := base.MemoryToMiB(memStr)
		memRange := mem - minMem
		if minMem < 32 {
			return errString("Minimum memory must be at least 32 MiB.")
		}
		if memRange < 0 {
			return errString("Minimum memory can not be greater than memory.")
		}
		if specType == "stateful" && !capacityAIEverywhere {
			if mem/minMem > maxMemoryRatio {
				return errString("The ratio between minMemory and memory must be less than " + strconv.Itoa(maxMemoryRatio) + ", please increase minMemory or decrease memory.")
			}
			if memRange > maxMemoryRange {
				return errString("The difference between minMemory and memory must be less than " + strconv.Itoa(maxMemoryRange) + " MiB, please increase minMemory or decrease memory.")
			}
		}
	}
	return validateGpuBounds(c, cpuStr, memStr)
}

func validateGpuBounds(c *ContainerSpec, cpuStr, memStr string) error {
	if c.Gpu == nil || c.Gpu.Nvidia == nil {
		return nil
	}
	nvidia := c.Gpu.Nvidia
	qty := util.Float64(nvidia.Quantity)
	if qty == 0 {
		qty = 1
	}
	minCpu := 2000.0
	maxCpu := 8000 * qty
	minMemoryGB := 7 * qty
	maxMemoryGB := 31 * qty
	model, _ := nvidia.Model.(string)
	if model == "t4" {
		if qty > 1 || base.MemoryToMiB(memStr) > 14000 {
			minCpu = qty * 4000
		}
	}
	if model == "a10g" {
		minCpu = 6000
		minMemoryGB = 8
		maxMemoryGB = 62
		if qty > 1 {
			return errString("A maximum GPU quantity of 1 is currently supported for a10g GPUs.")
		}
	}
	curCPUStr := cpuStr
	if mc := string(c.MinCpu); mc != "" {
		curCPUStr = mc
	}
	currentCpu := wlCpuNum(curCPUStr)
	if currentCpu < minCpu {
		return errString("CPU must be at least " + strconv.Itoa(int(minCpu)) + " millicores based on your GPU selection, please increase CPU.")
	}
	if currentCpu > maxCpu {
		return errString("CPU must be at most " + strconv.Itoa(int(minCpu)) + " millicores based on your GPU selection, please decrease CPU.")
	}
	curMemStr := memStr
	if mm := string(c.MinMemory); mm != "" {
		curMemStr = mm
	}
	currentMemory := base.MemoryToMiB(curMemStr)
	if currentMemory < minMemoryGB*1024 {
		return errString("Memory must be at least " + strconv.Itoa(int(minMemoryGB)) + "Gi based on your GPU selection, please increase Memory.")
	}
	if currentMemory > maxMemoryGB*1024 {
		return errString("Memory must be at most " + strconv.Itoa(int(minMemoryGB)) + "Gi based on your GPU selection, please decrease Memory.")
	}
	return nil
}

// --- vm/workload cross-validation (workload.ts:1091 validateVmWorkload) ---

func validateVmWorkload(spec *WorkloadSpec, containers []ContainerSpec) error {
	vmSpec := spec.Vm
	if vmSpec == nil {
		return errString("the 'vm' property is required for workloads of type 'vm'")
	}
	if len(containers) != 1 {
		return errString("workloads of type 'vm' must have exactly one container entry (got " + strconv.Itoa(len(containers)) + ")")
	}
	if spec.SecurityOptions != nil {
		return errString("spec.securityOptions is not valid for workloads of type 'vm' (the pod-level security context does not apply to the guest OS)")
	}
	if lb := spec.LoadBalancer; lb != nil && util.Bool(lb.ReplicaDirect) {
		return errString("loadBalancer.replicaDirect is not valid for workloads of type 'vm' (replicas are addressable via the VirtualMachinePool name)")
	}

	c := &containers[0]
	disallowed := []struct {
		name string
		set  bool
	}{
		{"port", c.Port != nil}, {"command", c.Command != ""}, {"args", c.Args != nil},
		{"lifecycle", c.Lifecycle != nil}, {"workingDir", c.WorkingDir != ""},
	}
	for _, f := range disallowed {
		if f.set {
			return errString("containers[0]." + f.name + " is not valid for workloads of type 'vm' (use spec.vm.networks for ports; the guest OS owns process lifecycle)")
		}
	}
	if c.Image != "" {
		return errString("containers[0].image is not valid for workloads of type 'vm' (specify the boot source in spec.vm.bootDisk.source instead)")
	}

	cpuStr := string(c.Cpu)
	cpuMillis := wlCpuNum(cpuStr)
	if cpuMillis < 1000 || math.Mod(cpuMillis, 1000) != 0 {
		return errString("containers[0].cpu must be a whole-core multiple >= 1000m for workloads of type 'vm' (got " + cpuStr + ")")
	}

	if p := c.ReadinessProbe; p != nil {
		if p.Exec != nil {
			return errString("containers[0].readinessProbe.exec is not supported for workloads of type 'vm' (only tcpSocket and httpGet are valid)")
		}
		if p.Grpc != nil {
			return errString("containers[0].readinessProbe.grpc is not supported for workloads of type 'vm' (only tcpSocket and httpGet are valid)")
		}
	}
	if p := c.LivenessProbe; p != nil {
		if p.Exec != nil {
			return errString("containers[0].livenessProbe.exec is not supported for workloads of type 'vm' (only tcpSocket and httpGet are valid)")
		}
		if p.Grpc != nil {
			return errString("containers[0].livenessProbe.grpc is not supported for workloads of type 'vm' (only tcpSocket and httpGet are valid)")
		}
	}

	bootDisk := vmSpec.BootDisk
	if bootDisk == nil {
		bootDisk = &vm.VmSpecBootDisk{}
	}
	if bootDisk.Source == nil {
		return errString("spec.vm.bootDisk.source is required for workloads of type 'vm' (oci or http)")
	}
	if bootDisk.Persist == nil || bootDisk.Persist.VolumeSet == "" {
		return errString("spec.vm.bootDisk.persist.volumeSet is required (the boot disk is always backed by a per-replica PVC populated via CDI)")
	}

	minScale := 0.0
	if do := spec.DefaultOptions; do != nil && do.Autoscaling != nil && do.Autoscaling.MinScale != nil {
		minScale = *do.Autoscaling.MinScale
	}
	if vmSpec.RunStrategy == "Halted" && minScale > 0 {
		return errString("spec.vm.runStrategy='Halted' conflicts with defaultOptions.autoscaling.minScale > 0 (a halted pool cannot have running replicas)")
	}

	bootOrderOwners := map[int]string{}
	claim := func(order int, owner string) error {
		if existing, ok := bootOrderOwners[order]; ok {
			return errString("bootOrder " + strconv.Itoa(order) + " is assigned to both " + existing + " and " + owner + "; each disk must have a unique bootOrder (the boot disk defaults to 1)")
		}
		bootOrderOwners[order] = owner
		return nil
	}
	bootDiskOrder := 1
	if bootDisk.BootOrder != nil {
		bootDiskOrder = int(*bootDisk.BootOrder)
	}
	if err := claim(bootDiskOrder, "spec.vm.bootDisk"); err != nil {
		return err
	}

	for _, e := range c.Env {
		if v := string(e.Value); strings.HasPrefix(v, "cpln://reference/") {
			return errString("containers[0].env '" + string(e.Name) + "' uses a manifest reference (" + v + ") which is not supported for workloads of type 'vm' (no pod fields are exposed to the guest)")
		}
	}

	for _, vol := range c.Volumes {
		name := string(vol.Name)
		if name == "" {
			return errString("containers[0].volumes[].name is required for workloads of type 'vm'")
		}
		if vol.Path != "" {
			return errString("containers[0].volumes[].path is not used for workloads of type 'vm' (volume '" + name + "' is attached to the VM as a block device)")
		}
		if vol.BootOrder != nil {
			uri := string(vol.Uri)
			if strings.HasPrefix(uri, "cpln://secret/") || strings.HasPrefix(uri, "k8s://secret/") {
				return errString("containers[0].volumes[].bootOrder is not valid for secret volume '" + name + "' (a secret is a non-bootable config/data disk)")
			}
			if err := claim(int(*vol.BootOrder), "volume '"+name+"'"); err != nil {
				return err
			}
		}
	}
	return nil
}

// --- unit conversions (workload.ts:32-82) ---

func wlCpuNum(s string) float64 {
	cpu := 0.0
	if s != "" {
		parts := strings.Split(s, "m")
		if len(parts) == 2 {
			n, _ := strconv.Atoi(parts[0])
			cpu += float64(n)
		}
		if len(parts) == 1 {
			f, _ := strconv.ParseFloat(parts[0], 64)
			cpu += f * 1000
		}
	}
	return cpu
}

// --- cron value checks (workload.ts:1205-1237) ---

func checkCronValues(expr string) error {
	pieces := wlWhitespaceRe.Split(expr, -1)
	ranges := []struct {
		name     string
		min, max int
	}{{"minute", 0, 59}, {"hour", 0, 23}, {"day (month)", 1, 31}, {"month", 1, 12}, {"day (week)", 0, 6}}
	for i, r := range ranges {
		vals, err := extractCronValues(pieces[i])
		if err != nil {
			return err
		}
		for _, v := range vals {
			if math.IsNaN(v) {
				continue
			}
			if v < float64(r.min) || v > float64(r.max) {
				return errString("an invalid " + r.name + " was found: " + strconv.Itoa(int(v)) + ". Valid " + r.name + " values are in [" + strconv.Itoa(r.min) + ", " + strconv.Itoa(r.max) + "]")
			}
		}
	}
	return nil
}

func jsParseInt(s string) float64 {
	i := 0
	if i < len(s) && (s[i] == '+' || s[i] == '-') {
		i++
	}
	start := i
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	if i == start {
		return math.NaN()
	}
	n, err := strconv.Atoi(s[:i])
	if err != nil {
		return math.NaN()
	}
	return float64(n)
}

func extractCronValues(seg string) ([]float64, error) {
	var out []float64
	for _, piece := range strings.Split(seg, ",") {
		p := strings.Replace(piece, "*/", "", 1)
		pv := strings.Split(p, "-")
		lower := jsParseInt(pv[0])
		if len(pv) > 1 {
			upper := jsParseInt(pv[1])
			if upper <= lower {
				return nil, errString("an invalid range was found in the " + seg + " segment: " + pv[0] + "-" + pv[1])
			}
		}
		for _, x := range pv {
			out = append(out, jsParseInt(x))
		}
	}
	return out, nil
}

// --- workload options validator (schema_workloadoptions.go runOptionsValidator) ---

type wlValidationOptions struct {
	metrics           []string
	multiMetrics      []string // nil = multi not supported
	scaleToZero       any      // true | false | "keda"
	maxTimeoutSeconds int      // 0 = no cap
	defaultCapacityAI bool
}

var workloadValidationMap = map[string]wlValidationOptions{
	"serverless": {metrics: []string{"concurrency", "cpu", "memory", "rps", "disabled"}, scaleToZero: true, maxTimeoutSeconds: slsMaxTimeoutSeconds, defaultCapacityAI: true},
	"standard":   {metrics: []string{"cpu", "memory", "latency", "rps", "keda", "disabled"}, multiMetrics: []string{"cpu", "memory", "rps"}, scaleToZero: "keda", defaultCapacityAI: true},
	"cron":       {metrics: []string{"concurrency", "cpu", "memory", "rps", "latency", "keda", "disabled"}, scaleToZero: false, defaultCapacityAI: true},
	// concurrency leads the stateful list because metrics[0] is also the DEFAULT
	// metric when none is sent (workloadOptions.ts:43-48).
	"stateful": {metrics: []string{"concurrency", "cpu", "memory", "latency", "rps", "keda", "disabled"}, multiMetrics: []string{"cpu", "memory", "rps"}, scaleToZero: "keda", defaultCapacityAI: false},
	"vm":       {metrics: []string{"disabled"}, scaleToZero: false, defaultCapacityAI: false},
}

// optionsView is the part of DefaultOptions / LocalOptionsItem the type-dependent
// validation reads and writes — the two structs are distinct generated types with
// the same option fields.
type optionsView struct {
	autoscaling    **workloadOptions.OptionsAutoscaling
	capacityAI     **bool
	timeoutSeconds *float64
}

func defaultOptionsView(o *workloadOptions.DefaultOptions) optionsView {
	return optionsView{&o.Autoscaling, &o.CapacityAI, o.TimeoutSeconds}
}

func localOptionsView(o *workloadOptions.LocalOptionsItem) optionsView {
	return optionsView{&o.Autoscaling, &o.CapacityAI, o.TimeoutSeconds}
}

func runOptionsValidator(workloadType string, options optionsView) error {
	settings, ok := workloadValidationMap[workloadType]
	if !ok {
		return nil
	}
	if *options.capacityAI == nil {
		*options.capacityAI = util.Ptr(settings.defaultCapacityAI)
		if as := *options.autoscaling; **options.capacityAI && as != nil && string(as.Metric) == "cpu" {
			*options.capacityAI = util.Ptr(false)
		}
	}

	as := *options.autoscaling
	metricPresent := as != nil && as.Metric != ""
	if !metricPresent {
		multiPresent := as != nil && as.Multi != nil
		if !multiPresent {
			if as == nil {
				as = &workloadOptions.OptionsAutoscaling{}
				*options.autoscaling = as
			}
			metric := "disabled"
			if len(settings.metrics) > 0 {
				metric = settings.metrics[0]
			}
			if **options.capacityAI && metric == "cpu" {
				metric = "disabled"
			}
			as.Metric = workloadOptions.OptionsAutoscalingMetric(metric)
		}
	} else {
		metric := string(as.Metric)
		if !containsStr(settings.metrics, metric) {
			return errString("Workloads of type '" + workloadType + "' do not support autoscaling based on metric: '" + metric + "'")
		}
	}

	if as != nil && as.Multi != nil {
		if settings.multiMetrics == nil {
			return errString("Workloads of type '" + workloadType + "' do not support multi metrics")
		}
		for _, m := range as.Multi {
			metric, _ := m.Metric.(string)
			if !containsStr(settings.multiMetrics, metric) {
				return errString("Workloads of type '" + workloadType + "' do not support multi-target autoscaling based on metric: " + metric)
			}
		}
	}

	if ts := options.timeoutSeconds; ts != nil && *ts > 0 && settings.maxTimeoutSeconds > 0 && *ts > float64(settings.maxTimeoutSeconds) {
		return errString("Workloads of type '" + workloadType + "' do not support timeoutSeconds greater than " + strconv.Itoa(settings.maxTimeoutSeconds))
	}

	if as != nil {
		minScale := util.Float64(as.MinScale)
		maxScale := util.Float64(as.MaxScale)
		if minScale == 0 && maxScale > 0 {
			if settings.scaleToZero == "keda" {
				// minScale 0 is allowed via KEDA or kata scale-to-zero (metric
				// concurrency/rps); the actuator gates kata behavior on the cpln/kata
				// tag, so non-kata is inert (workloadOptions.ts:352-366).
				metric := string(as.Metric)
				kataScaleToZeroMetric := metric == "concurrency" || metric == "rps"
				if metric != "keda" && !kataScaleToZeroMetric {
					return errString("workloads of type '" + workloadType + "' can only scale to zero via KEDA (metric 'keda') or kata (metric 'concurrency'/'rps'); please enable KEDA or set a kata scale-to-zero metric")
				}
			}
			if settings.scaleToZero == false {
				return errString("workloads of type '" + workloadType + "' are unable to scale to zero, please set minScale > 0")
			}
		}
		if as.Keda != nil && !containsStr(settings.metrics, "keda") {
			return errString("Workloads of type '" + workloadType + "' do not support KEDA autoscaling")
		}
	}
	return nil
}

// --- Extras (BYOK pod customizations, workload.ts:843-873 / 2e07953ae) -------

var (
	// A k8s label key: an optional DNS-subdomain prefix plus the name segment.
	k8sLabelKeyRe = regexp.MustCompile(`^([a-z0-9]([-a-z0-9.]{0,251}[a-z0-9])?/)?[A-Za-z0-9]([-A-Za-z0-9_.]{0,61}[A-Za-z0-9])?$`)
	// A k8s label value, which may be empty.
	k8sLabelValueRe = regexp.MustCompile(`^([A-Za-z0-9]([-A-Za-z0-9_.]{0,61}[A-Za-z0-9])?)?$`)
	// DNS-1123 subdomain.
	schedulerNameRe = regexp.MustCompile(`^[a-z0-9]([-a-z0-9.]*[a-z0-9])?$`)
)

func (SchedulerName) ValidationRule() *validate.Rule {
	return validate.String().Max(253).Regex(schedulerNameRe)
}

// extrasLabelsRule ports Extras.metadata.labels (workload.ts:860-863). Unusually for
// this schema the KEY carries constraints too, because these labels are merged into a
// pod template: a key k8s rejects would be found by the actuator at apply time, on a
// workload the API had already accepted.
func extrasLabelsRule() *validate.Rule {
	return validate.Object(nil).Pattern(
		validate.String().Max(317).Regex(k8sLabelKeyRe),
		validate.String().Max(63).Regex(k8sLabelValueRe).Allow(""),
	)
}

// The generator emits the extras subtree twice (once for the exported Extras type,
// once inlined under WorkloadSpec); both carry the same rule.
func (ExtrasMetadataLabels) ValidationRule() *validate.Rule             { return extrasLabelsRule() }
func (WorkloadSpecExtrasMetadataLabels) ValidationRule() *validate.Rule { return extrasLabelsRule() }
