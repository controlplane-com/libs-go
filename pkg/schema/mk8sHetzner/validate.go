package mk8sHetzner

// validate.go attaches the HetznerProvider validation rule (ported from
// pkg/handlers/schema_mk8s.go HetznerProviderRule / mk8sHetzner.ts).

import (
	"regexp"
	"strings"

	"github.com/controlplane-com/libs-go/pkg/schema/base"
	"github.com/controlplane-com/libs-go/pkg/schema/mk8sCommon"
	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

var hetznerLabelKeyRe = regexp.MustCompile(`^[a-zA-Z0-9/_-]+$`)

func hetznerLabelsRule() *validate.Rule {
	return validate.Object(nil).Pattern(
		validate.String().Regex(hetznerLabelKeyRe).Min(1).Max(63),
		validate.String().Max(255).Allow(""),
	)
}

// hetznerValidLabelRule ports mk8sHetzner.ts:32 ValidLabel.
func hetznerValidLabelRule() *validate.Rule {
	return validate.String().Custom(func(value any, h *validate.Helper) (any, error) {
		s, _ := value.(string)
		if strings.Contains(s, ",") {
			return nil, h.Errorf("any.custom", "label cannot contain ,")
		}
		if strings.Contains(s, "=") {
			return nil, h.Errorf("any.custom", "label cannot contain =")
		}
		return s, nil
	})
}

func hetznerPoolRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"name":          mk8sCommon.NodePoolNameRule().Required(),
		"labels":        mk8sCommon.LabelsRule(),
		"taints":        mk8sCommon.TaintsRule(),
		"maxPods":       mk8sCommon.MaxPodsRule(),
		"serverType":    validate.String().Required(),
		"overrideImage": validate.String(),
		"minSize":       validate.Number().Integer().Min(0).Required(),
		"maxSize":       validate.Number().Integer().Required().MinRef("minSize"),
	})
}

func hetznerDedicatedPoolRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"name":    mk8sCommon.NodePoolNameRule().Required(),
		"labels":  mk8sCommon.LabelsRule(),
		"taints":  mk8sCommon.TaintsRule(),
		"maxPods": mk8sCommon.MaxPodsRule(),
	})
}

// ValidationRule ports mk8sHetzner.ts:92 HetznerProvider.
func (HetznerProvider) ValidationRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"region":                   validate.String().Valid("fsn1", "nbg1", "hel1", "ash", "hil").Required().Unmodifiable(),
		"hetznerLabels":            hetznerLabelsRule(),
		"networking":               mk8sCommon.NetworkingConfigRule(false),
		"preInstallScript":         mk8sCommon.PreInstallScriptRule(false),
		"tokenSecretLink":          base.LocalLinkRule().Custom(base.ToKind("secret")).Required(),
		"networkId":                validate.String().Required().Unmodifiable(),
		"firewallId":               validate.String(),
		"nodePools":                validate.Array(hetznerPoolRule()),
		"dedicatedServerNodePools": validate.Array(hetznerDedicatedPoolRule()),
		"image":                    validate.String().Default("ubuntu-20.04").Required(),
		"sshKey":                   validate.String(),
		"autoscaler":               mk8sCommon.AutoscalerConfigRule(),
		"floatingIPSelector":       validate.Object(nil).Pattern(hetznerValidLabelRule(), hetznerValidLabelRule().Allow("")),
	}).CustomTyped(mk8sCommon.PoolNamesUniqueObjectCustom("nodePools", "dedicatedServerNodePools"))
}
