package vm

// validate.go attaches validation to the canonical vm leaf types via the
// Validator interface — the go-libs home of what used to be the vm
// pattern=/custom=/imageLink registrations in pkg/handlers (schema_vm.go +
// schema_a_reflect_workload.go). Each strongly-typed field replaces a
// registry-backed struct tag:
//
//   VmImageLink    → base.ImageLinkRule()            (was `imageLink`)
//   VmHttpUrl      → String().URIScheme("http","https") (was `custom=vmHttpUrl`)
//   VmChecksum     → String().Regex(vmChecksumRe)    (was `pattern=vmChecksum`)
//   VmVolumeSetUri → String().Regex(vmVolumeSetUriRe)(was `pattern=vmVolumeSetUri`)
//   VmNetName      → String().Regex(vmNetNameRe)     (was `pattern=vmNetName`)
//   VmHost         → String().Regex(vmHostRe)        (was `pattern=vmHost`)
//   VmBase64       → String().Base64()                 (was `custom=vmBase64`)
//   SecretLink     → LocalLink.custom(toKind('secret'))
//   VmUser         → String().Max(32).Regex(vmUserRe)
//
// It also carries the object-level relations vm.ts declares — the bootDisk.source
// and cloudInit xors and the firmware secureBoot throw — which the port dropped.
//
// Declarative modifiers (required/max/default) stay on the field tags and layer
// on top of the Validator rule via the reflection engine.

import (
	"regexp"
	"strconv"

	"github.com/controlplane-com/libs-go/pkg/schema/base"
	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// --- nominal leaf type rules (type declarations live in models.go) ---

func (VmImageLink) ValidationRule() *validate.Rule { return base.ImageLinkRule() }

// VmHttpUrl ports `Joi.string().uri({scheme:['http','https']})` (vm.ts:24-26). The
// hand-rolled prefix check it replaces accepted any malformed URI that started with
// http(s):// and reported a message/type pair Joi never emits.
func (VmHttpUrl) ValidationRule() *validate.Rule {
	return validate.String().URIScheme("http", "https")
}

func (VmChecksum) ValidationRule() *validate.Rule { return validate.String().Regex(vmChecksumRe) }

// VmVolumeSetUri carries VolumeSetUri's `.message()` override (vm.ts:5-7).
func (VmVolumeSetUri) ValidationRule() *validate.Rule {
	return validate.String().Regex(vmVolumeSetUriRe).Message("volumeSet must be in the form cpln://volumeset/<name>")
}

func (VmNetName) ValidationRule() *validate.Rule { return validate.String().Regex(vmNetNameRe) }

func (VmHost) ValidationRule() *validate.Rule { return validate.String().Regex(vmHostRe) }

// VmBase64 ports `Joi.string().base64()` (vm.ts:119). The custom it replaces spelled
// its own message, so it named the field differently from every other string rule.
func (VmBase64) ValidationRule() *validate.Rule { return validate.String().Base64() }

// SecretLink ports `LocalLink.custom(toKind('secret'))` (vm.ts:124,132,139). Without
// it every SSH-key / cloud-init secret reference under spec.vm was a free string:
// a malformed link reached the stored document unchallenged.
func (SecretLink) ValidationRule() *validate.Rule {
	return base.LocalLinkRule().Custom(base.ToKind("secret"))
}

// VmUser ports `Joi.string().max(32).regex(/^[a-z_][a-z0-9_-]*$/)` (vm.ts:141-147) —
// the guest account names cloud-init injects.
func (VmUser) ValidationRule() *validate.Rule {
	return validate.String().Max(32).Regex(vmUserRe)
}

// --- object-level relations (vm.ts) ---

// VmSpecBootDiskSource ports VmBootSource's `.xor('oci','http')` (vm.ts:41).
func (VmSpecBootDiskSource) ValidationRule() *validate.Rule {
	return validate.Reflect(VmSpecBootDiskSource{}).Xor("oci", "http")
}

// VmSpecCloudInit ports `.xor('userData','userDataBase64','userDataSecret')` (vm.ts:137).
func (VmSpecCloudInit) ValidationRule() *validate.Rule {
	return validate.Reflect(VmSpecCloudInit{}).Xor("userData", "userDataBase64", "userDataSecret")
}

// VmSpecFirmware ports the firmware custom (vm.ts:90-98) plus the uuidv4-only GUID
// check (vm.ts:79-82). Secure Boot needs persistent EFI NVRAM the clusters do not yet
// provide; without it the guest loses its enrolled keys on reboot and cannot boot, so
// TS refuses the flag rather than storing an unbootable config.
func (VmSpecFirmware) ValidationRule() *validate.Rule {
	return validate.RefineT(
		validate.Reflect(VmSpecFirmware{}).Fork("uuid", func(r *validate.Rule) *validate.Rule {
			return r.Custom(uuidV4Custom)
		}),
		secureBootRefine)
}

// VmSpec ports the deep `.default()`s (vm.ts:166-181): bootDisk/firmware/clock default
// to {} (so their own field defaults materialise) and networks to [{name:'default'}].
// Without them a minimal type=vm workload stored none of the structural defaults TS
// stores, so the two services disagreed on the spec they had just accepted.
func (VmSpec) ValidationRule() *validate.Rule {
	return validate.RefineT(validate.Reflect(VmSpec{}), vmSpecDefaults)
}

func secureBootRefine(f *VmSpecFirmware, h *validate.Helper) error {
	if f.SecureBoot != nil && *f.SecureBoot {
		return h.Errorf("any.custom", "firmware.secureBoot is not yet supported")
	}
	return nil
}

func vmSpecDefaults(s *VmSpec, h *validate.Helper) error {
	if s.Networks == nil {
		s.Networks = []VmSpecNetworks{{Name: "default"}}
	}
	return nil
}

// uuidV4Custom narrows Joi's `guid({version:['uuidv4']})`: Rule.UUID() accepts any
// version, so a v1 GUID TS rejects was stored verbatim.
func uuidV4Custom(value any, h *validate.Helper) (any, error) {
	s, _ := value.(string)
	if !vmUuidV4Re.MatchString(s) {
		return nil, h.WithCtx(h.Errorf("string.guid", "%s must be a valid GUID", strconv.Quote(h.Label())), "value", s)
	}
	return s, nil
}

// --- named regexes (ported from pkg/handlers/schema_vm.go) ---

var (
	vmVolumeSetUriRe = regexp.MustCompile(`^cpln://volumeset/[a-z0-9][a-z0-9-]{0,63}$`)
	vmChecksumRe     = regexp.MustCompile(`(?i)^sha(256|512):[a-f0-9]+$`)
	vmNetNameRe      = regexp.MustCompile(`^[a-z0-9-]+$`)
	vmHostRe         = regexp.MustCompile(`^[a-z0-9-]+$`)
	vmUserRe         = regexp.MustCompile(`^[a-z_][a-z0-9_-]*$`)
	// Joi's uuidv4 guid: the version nibble is 4 and the variant is 8/9/a/b.
	vmUuidV4Re = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
)
