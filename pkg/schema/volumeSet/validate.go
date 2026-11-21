package volumeSet

// validate.go attaches validation to the canonical volumeSet types via the
// Validator interface — the go-libs home of what used to be
// pkg/handlers/schema_volumeset.go + the volumeSet registrations in
// schema_reflect_datanet.go. volumeSet.ts reflects
// `VolumeSet = Base.keys({spec: required, status, gvc}).custom(volumeSetTopCustom)`.
//
// Leaf value types formerly wired via RegisterType now carry their own rule:
//   VsCpu → cpuRule (workload Cpu), VsMemory → memoryRule (workload Memory),
//   SnapshotSchedule → scheduleTypeRule + the snapshotSchedule frequency custom,
//   StatusWorkloadLink → LocalLink.custom(toKind('workload')),
//   PersistentVolumeStatusAttributes → pattern(String, String.allow('')).
// The `pattern=<name>` tags (vsStorageClassSuffix / vsSnapshotRetention) are
// replaced by the strongly-typed StorageClassSuffix / SnapshotRetentionDuration
// field types. The four cross-field .custom()s (volumeSetSpecCustom,
// mountResourcesCustom, volumeSetStatusDuplicateLocations, volumeSetTopCustom) are
// layered as object-level .Refine()s on the reflected nested rules.
//
// The Cpu/Memory/ScheduleType leaf rules are cross-partition workload symbols
// (schema_workload.go CpuRule/MemoryRule/ScheduleTypeRule); until the workload
// package is migrated they are ported here verbatim as unexported builders.

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/controlplane-com/libs-go/pkg/schema/base"
	"github.com/controlplane-com/libs-go/pkg/schema/util"
	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// --- resource + nested object rules ---

// ValidationRule reflects VolumeSet and layers the top-level performanceClass/
// fileSystemType consistency custom — the schema the "volumeset" kind validates
// against (Base.keys({spec: required, status, gvc}).custom(volumeSetTopCustom)).
func (VolumeSet) ValidationRule() *validate.Rule {
	return validate.Reflect(VolumeSet{}).CustomTyped(volumeSetTopRefine)
}

// ValidationRule reflects VolumeSetSpec and layers the fileSystemType/
// performanceClass/initialCapacity/autoscaling cross-field custom
// (volumeSet.ts:347-397). The parent VolumeSet.Spec field tag carries `required`.
func (VolumeSetSpec) ValidationRule() *validate.Rule {
	return validate.RefineT(validate.Reflect(VolumeSetSpec{}), volumeSetSpecRefine)
}

// ValidationRule reflects VolumeSetStatus and layers the duplicate-location check
// (volumeSet.ts:172-185).
func (VolumeSetStatus) ValidationRule() *validate.Rule {
	return validate.RefineT(validate.Reflect(VolumeSetStatus{}), volumeSetStatusDuplicateLocations)
}

// ValidationRule reflects the mount resources and layers the cpu/memory
// ratio+range custom (volumeSet.ts:273-308).
func (VolumeSetSpecMountOptionsResources) ValidationRule() *validate.Rule {
	return validate.RefineT(validate.Reflect(VolumeSetSpecMountOptionsResources{}), mountResourcesRefine)
}

// --- leaf value types ---

// ValidationRule for the mount CPU quantity — workload Cpu
// (String.regex(cpuFmt).max(20)).
func (VsCpu) ValidationRule() *validate.Rule { return cpuRule() }

// ValidationRule for the mount memory quantity — workload Memory
// (String.regex(memFmt).max(20).custom(normalize)).
func (VsMemory) ValidationRule() *validate.Rule { return memoryRule() }

// ValidationRule for the snapshot schedule — workload ScheduleType (cron custom)
// plus the snapshot-frequency custom (may not be more frequent than hourly).
func (SnapshotSchedule) ValidationRule() *validate.Rule {
	return scheduleTypeRule().Custom(snapshotScheduleCustom)
}

// ValidationRule for the status workload link — LocalLink.custom(toKind('workload')).
func (StatusWorkloadLink) ValidationRule() *validate.Rule {
	return base.LocalLinkRule().Custom(base.ToKind("workload"))
}

// ValidationRule for the persistent-volume attributes map — pattern(String,
// String().allow(”)): the value's allow(”) cannot be expressed by a field tag.
func (PersistentVolumeStatusAttributes) ValidationRule() *validate.Rule {
	return validate.Object(nil).Optional().Pattern(validate.String(), validate.String().Allow(""))
}

func (StorageClassSuffix) ValidationRule() *validate.Rule {
	return validate.String().Regex(storageClassSuffixRe)
}

func (SnapshotRetentionDuration) ValidationRule() *validate.Rule {
	return validate.String().Regex(snapshotRetentionRe)
}

// --- named regexes ---

var (
	storageClassSuffixRe = regexp.MustCompile(`^[a-zA-Z][0-9a-zA-Z\-_]*$`)
	snapshotRetentionRe  = regexp.MustCompile(`^([0-9]+(\.[0-9]+)?[dhm])$`)
)

// --- workload Cpu / Memory / ScheduleType leaf rules (ported from schema_workload.go) ---

var (
	wlMemFmtRe     = regexp.MustCompile(`(?i)^[0-9]+(\.[0-9]{1,3})?(G|M|k|Gi|Mi|Ki)?$`)
	wlMemSplitRe   = regexp.MustCompile(`^([0-9]+(?:\.[0-9]+)?)(.*)$`)
	wlCpuFmtRe     = regexp.MustCompile(`^([0-9]+)(m|(\.[0-9]{1,3}))?$`)
	wlCronRe       = regexp.MustCompile(`^(\*|(\d+(-\d+)?))(\/\d+)?(,(\*|(\d+(-\d+)?))(\/\d+)?)*(\s(\*|(\d+(-\d+)?))(\/\d+)?(,(\*|(\d+(-\d+)?))(\/\d+)?)*){4}$`)
	wlWhitespaceRe = regexp.MustCompile(`\s+`)
)

// memoryRule ports `Memory = Joi.string().regex(...).max(20).custom(normalize)`.
func memoryRule() *validate.Rule {
	return validate.String().Regex(wlMemFmtRe).Max(20).Custom(memNormalizeCustom)
}

// memNormalizeCustom normalises the memory unit casing.
func memNormalizeCustom(value any, _ *validate.Helper) (any, error) {
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

// cpuRule ports `Cpu = Joi.string().regex(...).max(20)`.
func cpuRule() *validate.Rule { return validate.String().Regex(wlCpuFmtRe).Max(20) }

// scheduleTypeRule ports `ScheduleType = Joi.string().custom(cron)`.
func scheduleTypeRule() *validate.Rule { return validate.String().Custom(scheduleTypeCustom) }

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
				return fmt.Errorf("an invalid %s was found: %d. Valid %s values are in [%d, %d]", r.name, int(v), r.name, r.min, r.max)
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
				return nil, fmt.Errorf("an invalid range was found in the %s segment: %s-%s", seg, pv[0], pv[1])
			}
		}
		for _, x := range pv {
			out = append(out, jsParseInt(x))
		}
	}
	return out, nil
}

// --- snapshot-frequency custom (volumeSet.ts:197-207) ---

// snapshotScheduleCustom rejects a snapshot schedule more frequent than hourly —
// the minute field must be a single fixed value (no wildcard/list/range/step).
func snapshotScheduleCustom(value any, h *validate.Helper) (any, error) {
	s, _ := value.(string)
	fields := strings.Fields(s)
	tooFrequent := true
	if len(fields) == 5 {
		minute := fields[0]
		if !strings.ContainsAny(minute, "*,-/") {
			tooFrequent = false
		}
	}
	if tooFrequent {
		return nil, h.Errorf("any.custom", "snapshots cannot be scheduled more frequently than once an hour")
	}
	return s, nil
}

// --- cross-field refines ---

// workload.ts MAX_* resource-ratio/range bounds (workload.ts:25-28).
const (
	vsMaxCPURange    = 4000.0
	vsMaxCPURatio    = 4.0
	vsMaxMemoryRange = 4096.0
	vsMaxMemoryRatio = 4.0
)

// vsFileSystemNames — the known file-system names (volumeSet.ts:37).
var vsFileSystemNames = []string{"ext4", "xfs", "shared"}

// perfClass capacity table (volumeSet.ts PerformanceClasses).
type perfClass struct{ min, max float64 }

var perfClasses = map[string]perfClass{
	"general-purpose-ssd": {10, 65536},
	"high-throughput-ssd": {200, 65536},
	"shared":              {10, 65536},
}

// mountResourcesRefine ports the MountResources cpu/memory ratio+range custom
// (volumeSet.ts:273-308).
func mountResourcesRefine(res *VolumeSetSpecMountOptionsResources, h *validate.Helper) error {
	or := func(a, b string) string {
		if a != "" {
			return a
		}
		return b
	}
	minCPU := convertCPUStringToNumber(or(string(res.MinCpu), string(res.MaxCpu)))
	maxCPU := convertCPUStringToNumber(or(string(res.MaxCpu), string(res.MinCpu)))
	if maxCPU < minCPU {
		return h.Errorf("any.custom", "maxCpu must be greater than or equal to minCpu")
	}
	// No zero guard: JS evaluates 2000/0 as Infinity (which trips the ratio) and 0/0 as
	// NaN (which compares false, so minCpu==maxCpu=="0" stays accepted). Go's float
	// division does exactly the same, and guarding on minCPU > 0 instead SKIPPED the
	// check — accepting `minCpu:"0"` with a 2000m max, which TS rejects.
	if maxCPU/minCPU > vsMaxCPURatio {
		return h.Errorf("any.custom", "The ratio between maxCpu and minCpu must be less than %d:1, please increase minCpu or decrease maxCpu.", int(vsMaxCPURatio))
	}
	if maxCPU-minCPU > vsMaxCPURange {
		return h.Errorf("any.custom", "The difference between minCpu and maxCpu must be less than %d millicores, please increase minCpu or decrease maxCpu.", int(vsMaxCPURange))
	}
	minMem := base.MemoryToMiB(or(string(res.MinMemory), string(res.MaxMemory)))
	maxMem := base.MemoryToMiB(or(string(res.MaxMemory), string(res.MinMemory)))
	if maxMem < minMem {
		return h.Errorf("any.custom", "maxMemory must be greater than or equal to minMemory")
	}
	if maxMem/minMem > vsMaxMemoryRatio {
		return h.Errorf("any.custom", "The ratio between maxMemory and minMemory must be less than %d:1, please increase minMemory or decrease maxMemory.", int(vsMaxMemoryRatio))
	}
	if maxMem-minMem > vsMaxMemoryRange {
		return h.Errorf("any.custom", "The difference between minMemory and maxMemory must be less than %dMi, please increase minMemory or decrease maxMemory.", int(vsMaxMemoryRange))
	}
	return nil
}

// volumeSetSpecRefine ports the VolumeSetSpec `.custom()` (volumeSet.ts:347-397).
func volumeSetSpecRefine(v *VolumeSetSpec, h *validate.Helper) error {
	pcName := string(v.PerformanceClass)
	fsType := string(v.FileSystemType)
	// `value.performanceClass == ''` is false for an ABSENT key in JS, so an omitted
	// performanceClass falls through to the "(undefined) is invalid" message below.
	// Go's zero value conflates the two, so the branch has to ask whether the key was
	// sent. (With PerformanceClassName's .valid() set excluding '', the branch is in
	// fact unreachable on both sides — it is kept to mirror the source.)
	if h.Present("performanceClass") && pcName == "" && fsType != "shared" {
		return h.Errorf("any.custom", "performanceClass is required when using a fileSystemType other than shared")
	}
	if pcName == "shared" && fsType != "shared" {
		return h.Errorf("any.custom", "the shared performance class can only be used in combination with a shared file system")
	}
	if fsType == "shared" {
		v.PerformanceClass = "shared"
		pcName = "shared"
	}
	pc, ok := perfClasses[pcName]
	if !ok {
		// An absent key interpolates as `undefined`, not as the empty string.
		shown := pcName
		if !h.Present("performanceClass") && pcName == "" {
			shown = "undefined"
		}
		return h.Errorf("any.custom", "spec.performanceClass (%s) is invalid", shown)
	}
	initCap := v.InitialCapacity
	if initCap > pc.max {
		return h.Errorf("any.custom", "initialCapacity cannot be greater than the maxCapacity of its performance class (%v GB)", pc.max)
	}
	if pc.min > 0 && initCap < pc.min {
		return h.Errorf("any.custom", "initialCapacity cannot be less than the minCapacity of its performance class (%v GB)", pc.min)
	}
	as := v.Autoscaling
	var maxCap float64
	hasMax := as != nil && as.MaxCapacity != nil
	if hasMax {
		maxCap = *as.MaxCapacity
	}
	if hasMax && maxCap > 0 && initCap > maxCap {
		return h.Errorf("any.custom", "initialCapacity cannot be greater than autoscaling.maxCapacity")
	}
	if hasMax && maxCap > 0 && maxCap > pc.max {
		return h.Errorf("any.custom", "autoscaling.maxCapacity cannot be greater than the maxCapacity of its performance class (%v GB)", pc.max)
	}
	if hasMax && maxCap > 0 && pc.min > 0 && maxCap < pc.min {
		return h.Errorf("any.custom", "autoscaling.maxCapacity cannot be less than the minCapacity of its performance class (%v GB)", pc.max)
	}
	if as != nil {
		if pred := as.Predictive; pred != nil {
			if util.Bool(pred.Enabled) {
				if util.Float64(as.MinFreePercentage) <= 0 {
					return h.Errorf("any.custom", "minFreePercentage must be set when predictive scaling is enabled")
				}
				if util.Float64(as.ScalingFactor) < 1.1 {
					return h.Errorf("any.custom", "scalingFactor must be >= 1.1 when predictive scaling is enabled")
				}
			}
		}
	}
	return nil
}

// volumeSetStatusDuplicateLocations ports the VolumeSetStatus custom
// (volumeSet.ts:172-185): status.locations names must be unique.
func volumeSetStatusDuplicateLocations(st *VolumeSetStatus, h *validate.Helper) error {
	seen := map[string]int{}
	var order []string
	for _, l := range st.Locations {
		name := l.Name
		if seen[name] == 0 {
			order = append(order, name)
		}
		seen[name]++
	}
	var dups []string
	for _, n := range order {
		if seen[n] > 1 {
			dups = append(dups, n)
		}
	}
	if len(dups) > 0 {
		plural := ""
		if len(dups) != 1 {
			plural = "s"
		}
		return h.Errorf("any.custom", "locations cannot contain duplicates. Duplicated location%s: [%s]", plural, strings.Join(dups, ", "))
	}
	return nil
}

// volumeSetTopRefine ports the VolumeSet top-level custom (volumeSet.ts:403-413):
// performanceClass + fileSystemType must resolve to known descriptors.
func volumeSetTopRefine(v any, h *validate.Helper) (any, error) {
	spec := validate.Sub[VolumeSetSpec](v, "spec")
	if spec == nil {
		spec = &VolumeSetSpec{}
	}
	pcName := string(spec.PerformanceClass)
	if _, ok := perfClasses[pcName]; !ok {
		return nil, h.Errorf("any.custom", "spec.performanceClass (%s) is invalid", pcName)
	}
	fsType := string(spec.FileSystemType)
	if !util.ContainsStr(vsFileSystemNames, fsType) {
		return nil, h.Errorf("any.custom", "spec.performanceClass (%s) is invalid", pcName)
	}
	return v, nil
}

// --- cpu/memory string conversions (volumeSet.ts:416/433) ---

// convertCPUStringToNumber ports volumeSet.ts:416.
func convertCPUStringToNumber(cpuString string) float64 {
	cpu := 0.0
	if cpuString == "" {
		return cpu
	}
	m := strings.Split(cpuString, "m")
	if len(m) == 2 {
		n, _ := strconv.Atoi(m[0])
		cpu += float64(n)
	}
	if len(m) == 1 {
		f, _ := strconv.ParseFloat(m[0], 64)
		cpu += f * 1000
	}
	return cpu
}
