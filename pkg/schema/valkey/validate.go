package valkey

// validate.go attaches validation to the canonical valkey types via the Validator
// interface — the go-libs home of what used to be pkg/handlers/schema_valkey.go +
// the valkey registrations in schema_reflect_datanet.go. valkey.ts reflects
// `Valkey = Joi.object({name, gvc, spec, status})` (a plain object, NOT Base.keys,
// with no createSchema description-defaults-to-name fork). The non-declarative
// pieces are all leaf-type rules:
//   - ValkeyMemory / ValkeyCpu     -> the workload Memory / Cpu rules
//     (formerly RegisterType(..., MemoryRule()/CpuRule())).
//   - ValkeyLocationLink           -> LocalLink.custom(toKind('location'))
//     (formerly RegisterType(..., locationLinkRule())).
//   - ValkeyReplicas               -> replicas-must-be-odd
//     (formerly the `custom=valkeyReplicasOdd` tag / RegisterCustom).

import (
	"regexp"
	"strings"

	"github.com/controlplane-com/libs-go/pkg/schema/base"
	"github.com/controlplane-com/libs-go/pkg/schema/util"
	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// ValidationRule reflects `Valkey = Joi.object({name, gvc, spec, status})`. This is
// the schema the "valkey" kind validates against (formerly RegisterReflect and the
// ValkeySchema FromType shim).
func (Valkey) ValidationRule() *validate.Rule {
	return validate.Reflect(Valkey{})
}

// --- leaf-type rules (formerly RegisterType bindings in schema_reflect_datanet.go) ---

var (
	// wlMemFmtRe / wlMemSplitRe / wlCpuFmtRe port the workload Memory/Cpu regexes.
	valkeyMemFmtRe   = regexp.MustCompile(`(?i)^[0-9]+(\.[0-9]{1,3})?(G|M|k|Gi|Mi|Ki)?$`)
	valkeyMemSplitRe = regexp.MustCompile(`^([0-9]+(?:\.[0-9]+)?)(.*)$`)
	valkeyCpuFmtRe   = regexp.MustCompile(`^([0-9]+)(m|(\.[0-9]{1,3}))?$`)
)

// ValidationRule ports workload `Memory = Joi.string().regex(...).max(20)
// .custom(normalize)`.
func (ValkeyMemory) ValidationRule() *validate.Rule {
	return validate.String().Regex(valkeyMemFmtRe).Max(20).Custom(valkeyMemNormalize)
}

// valkeyMemNormalize normalises the memory unit casing (workload memNormalizeCustom).
func valkeyMemNormalize(value any, h *validate.Helper) (any, error) {
	v, _ := value.(string)
	m := valkeyMemSplitRe.FindStringSubmatch(v)
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

// ValidationRule ports workload `Cpu = Joi.string().regex(...).max(20)`.
func (ValkeyCpu) ValidationRule() *validate.Rule {
	return validate.String().Regex(valkeyCpuFmtRe).Max(20)
}

// ValidationRule for spec.localOptions[].location: LocalLink constrained to kind
// 'location' — LocalLink.custom(toKind('location')) (valkey.ts), formerly
// locationLinkRule = LocalLinkRule().Custom(ToKind("location")).
func (ValkeyLocationLink) ValidationRule() *validate.Rule {
	return base.LocalLinkRule().Custom(base.ToKind("location"))
}

// ValidationRule ports the ValkeyOptions.replicas custom (valkey.ts:11-15): replicas
// must be an odd number. The declarative integer/min=3 constraints stay on the field
// tag; this rule carries only the odd check (formerly `custom=valkeyReplicasOdd`).
func (ValkeyReplicas) ValidationRule() *validate.Rule {
	return validate.Number().Custom(func(value any, h *validate.Helper) (any, error) {
		n, ok := util.NumOf(value)
		if ok && int(n)%2 == 0 {
			return nil, h.Errorf("any.custom", "replicas must be an odd number")
		}
		return value, nil
	})
}
