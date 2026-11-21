package mk8sGcp

// validate.go attaches the GcpProvider validation rule (ported from
// pkg/handlers/schema_mk8s.go GcpProviderRule / mk8sGcp.ts).

import (
	"regexp"
	"strings"

	"github.com/controlplane-com/libs-go/pkg/schema/base"
	"github.com/controlplane-com/libs-go/pkg/schema/mk8sCommon"
	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

var (
	gcpLabelRe      = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]+$`)
	gcpTagRe        = regexp.MustCompile(`[a-z]([-a-z0-9]*[a-z0-9])?`)
	gcpMetaKeyRe    = regexp.MustCompile(`[a-zA-Z0-9_-]+`)
	gcpImageExactRe = regexp.MustCompile(`^projects/[a-z][-a-z0-9]{4,28}[a-z0-9]/global/images/(?:family/)?[a-z][-a-z0-9]{0,62}$`)
)

func gcpLabelsRule() *validate.Rule {
	// Joi's GcpLabels is a pattern-map with .max(10) — up to 10 extra tags.
	return validate.Object(nil).Pattern(
		validate.String().Regex(gcpLabelRe).Min(1).Max(63),
		validate.String().Regex(gcpLabelRe).Allow("").Max(63),
	).Max(10)
}

func gcpMetadataKeyRule() *validate.Rule {
	return validate.String().Max(128).Regex(gcpMetaKeyRe).Custom(func(value any, h *validate.Helper) (any, error) {
		s, _ := value.(string)
		if strings.HasPrefix(s, "goog-") {
			return nil, h.Errorf("any.custom", "metadata key cannot start with goog-")
		}
		if strings.HasPrefix(s, "google-") {
			return nil, h.Errorf("any.custom", "metadata key cannot start with google-")
		}
		return s, nil
	})
}

func gcpImageRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"recommended": validate.String().Valid("ubuntu/jammy-22.04", "ubuntu/noble-24.04", "debian/bookworm-12", "debian/trixie-13"),
		"family": validate.Object(map[string]*validate.Rule{
			"project": validate.String().Required(),
			"family":  validate.String().Required(),
		}),
		"exact": validate.String().Regex(gcpImageExactRe),
	}).Xor("recommended", "exact", "family")
}

func gcpPoolRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"name":                 mk8sCommon.NodePoolNameRule().Required(),
		"labels":               mk8sCommon.LabelsRule(),
		"taints":               mk8sCommon.TaintsRule(),
		"maxPods":              mk8sCommon.MaxPodsRule(),
		"machineType":          validate.String().Required(),
		"assignPublicIP":       validate.Boolean(),
		"zone":                 validate.String().Required(),
		"overrideImage":        gcpImageRule(),
		"bootDiskSize":         validate.Number().Integer().Min(20).Default(float64(20)).Required(),
		"minSize":              validate.Number().Integer().Min(0).Required(),
		"maxSize":              validate.Number().Integer().Required().MinRef("minSize"),
		"preemptible":          validate.Boolean(),
		"subnet":               validate.String().Required(),
		"localPersistentDisks": validate.Number().Integer().Min(0),
	})
}

// ValidationRule ports mk8sGcp.ts:128 GcpProvider.
func (GcpProvider) ValidationRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"projectId":        validate.String().Required(),
		"region":           validate.String().Required().Unmodifiable(),
		"labels":           gcpLabelsRule(),
		"tags":             validate.Array(validate.String().Max(63).Regex(gcpTagRe)).Sort().Dedupe().Max(61),
		"metadata":         validate.Object(nil).Pattern(gcpMetadataKeyRule(), validate.String()),
		"network":          validate.String().Required().Unmodifiable(),
		"saKeyLink":        base.LocalLinkRule().Custom(base.ToKind("secret")).Required(),
		"networking":       mk8sCommon.NetworkingConfigRule(true),
		"preInstallScript": mk8sCommon.PreInstallScriptRule(false),
		"image":            gcpImageRule().Required(),
		"nodePools":        validate.Array(gcpPoolRule()).CustomTyped(mk8sCommon.PoolNamesUniqueArrayCustom),
		"autoscaler":       mk8sCommon.AutoscalerConfigRule(),
	})
}
