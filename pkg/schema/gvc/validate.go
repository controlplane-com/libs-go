package gvc

// validate.go attaches validation to the canonical gvc types via the Validator
// interface — the go-libs home of what used to be pkg/handlers/schema_gvc.go +
// the gvc registrations in schema_reflect_orggvc.go. gvc.ts reflects from Gvc; the
// DB/authz-dependent seam logic (name-required fork, locationQuery kind stamping,
// loadBalancer permission gate) stays in the handler and is NOT part of this schema.
//
// Irreducible Joi parts ported here:
//   - GvcSpec.endpointNamingFormat POST default (gvcSpecRefine);
//   - GvcLoadBalancerConfig minScale<=maxScale (gvcLbScaleRefine);
//   - the LocationLink/SecretLink localLink leaf rules (were RegisterType);
//   - the dedupe collection customs and the cpu/mem format+normalize customs
//     (were custom=/pattern= tags) as strongly-typed field types.

import (
	"regexp"
	"strings"

	"github.com/controlplane-com/libs-go/pkg/schema/base"
	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// --- top-level + nested object rules ---

// ValidationRule reflects Gvc.
func (Gvc) ValidationRule() *validate.Rule {
	return validate.Reflect(Gvc{})
}

// ValidationRule reflects GvcSpec and layers the endpointNamingFormat POST default
// (gvc.ts:98 — on POST, if unset, default to 'org').
func (GvcSpec) ValidationRule() *validate.Rule {
	return validate.RefineT(validate.Reflect(GvcSpec{}), gvcSpecRefine)
}

// ValidationRule reflects GvcLoadBalancerConfig and layers the minScale<=maxScale
// cross-field check (gvc.ts:190 .custom).
func (GvcLoadBalancerConfig) ValidationRule() *validate.Rule {
	return validate.RefineT(validate.Reflect(GvcLoadBalancerConfig{}), gvcLbScaleRefine)
}

// gvcSpecRefine ports gvc.ts:98's endpointNamingFormat default function: on POST, if
// unset, default to 'org'; otherwise leave as-is (Helper defaults cannot see the
// request method, so this lives on the parent object rule).
func gvcSpecRefine(spec *GvcSpec, h *validate.Helper) error {
	if h.Present("endpointNamingFormat") {
		return nil
	}
	if h.Ctx != nil && h.Ctx.Method == "POST" {
		spec.EndpointNamingFormat = GvcSpecEndpointNamingFormatOrg
	}
	return nil
}

// ValidationRule reflects GvcConfig and layers the POST-only `spire` default
// (gvc.ts:242).
func (GvcConfig) ValidationRule() *validate.Rule {
	return validate.RefineT(validate.Reflect(GvcConfig{}), gvcConfigRefine)
}

// gvcConfigRefine ports gvc.ts:242's spire default function: on POST, if unset,
// default to true; otherwise leave it absent. Same shape as gvcSpecRefine — a Joi
// default() that reads helpers.prefs.context.method cannot be expressed as a struct
// tag, because a leaf default has no view of the request method.
//
// It matters that this is create-only: defaulting on PATCH would silently turn spire
// on for every gvc whose config was touched for any other reason.
func gvcConfigRefine(cfg *GvcConfig, h *validate.Helper) error {
	if cfg.Spire != nil {
		return nil
	}
	if h.Ctx != nil && h.Ctx.Method == "POST" {
		t := true
		cfg.Spire = &t
	}
	return nil
}

// gvcLbScaleRefine ports GvcLoadBalancerConfig's .custom: minScale <= maxScale.
func gvcLbScaleRefine(lb *GvcLoadBalancerConfig, h *validate.Helper) error {
	var minS, maxS float64
	if lb.MinScale != nil {
		minS = *lb.MinScale
	}
	if lb.MaxScale != nil {
		maxS = *lb.MaxScale
	}
	if minS > maxS {
		return h.Errorf("any.custom", "minScale must be less than or equal to maxScale")
	}
	return nil
}

// --- localLink leaf types (were RegisterType in schema_reflect_orggvc.go) ---

// ValidationRule ports LocalLink().custom(toKind("location")): a localLink pointing
// to a resource of kind location.
func (LocationLink) ValidationRule() *validate.Rule {
	return base.LocalLinkRule().Custom(base.ToKind("location"))
}

// ValidationRule ports LocalLink().custom(toKind("secret")): a localLink pointing to
// a resource of kind secret.
func (SecretLink) ValidationRule() *validate.Rule {
	return base.LocalLinkRule().Custom(base.ToKind("secret"))
}

// ValidationRule ports LocalLink().custom(toKind("ipset")): a localLink pointing to
// a resource of kind ipset.
func (IpSetLink) ValidationRule() *validate.Rule {
	return base.LocalLinkRule().Custom(base.ToKind("ipset"))
}

// ValidationRule ports LocalLink().custom(toKind("identity")): a localLink pointing
// to a resource of kind identity.
func (IdentityLink) ValidationRule() *validate.Rule {
	return base.LocalLinkRule().Custom(base.ToKind("identity"))
}

// ValidationRule ports LocalLink().custom(toKind("workload")): a localLink pointing
// to a resource of kind workload.
func (WorkloadLink) ValidationRule() *validate.Rule {
	return base.LocalLinkRule().Custom(base.ToKind("workload"))
}

// --- collection nominal types (were sort,custom=dedupe* tags) ---

func (LocationLinkSet) ValidationRule() *validate.Rule {
	return validate.Array(LocationLink("").ValidationRule()).Dedupe()
}

func (SecretLinkSet) ValidationRule() *validate.Rule {
	return validate.Array(SecretLink("").ValidationRule()).Dedupe()
}

// ValidationRule ports LocationOptionsRule (gvc.ts:40): the array deduplicated by
// each item's locationLink field (was custom=dedupeLocationLink).
func (LocationOptions) ValidationRule() *validate.Rule {
	return validate.Array(validate.Reflect(LocationOptionsItem{})).Dedupe("locationLink")
}

// --- cpu / memory format nominal types (were pattern=/custom= tags) ---

var (
	memFmtRe   = regexp.MustCompile(`(?i)^[0-9]+(\.[0-9]{1,3})?(G|M|k|Gi|Mi|Ki)?$`)
	memSplitRe = regexp.MustCompile(`^([0-9]+(?:\.[0-9]+)?)(.*)$`)
	cpuFmtRe   = regexp.MustCompile(`^([0-9]+)(m|(\.[0-9]{1,3}))?$`)
)

func (MemFmt) ValidationRule() *validate.Rule {
	return validate.String().Regex(memFmtRe).Custom(memNormalizeCustom)
}

func (CpuFmt) ValidationRule() *validate.Rule {
	return validate.String().Regex(cpuFmtRe)
}

// memNormalizeCustom ports handlers.memNormalizeCustom: normalise the memory unit
// casing (k stays lower; otherwise upper-first, e.g. mi -> Mi, GI -> Gi).
func memNormalizeCustom(value any, h *validate.Helper) (any, error) {
	v, _ := value.(string)
	m := memSplitRe.FindStringSubmatch(v)
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
