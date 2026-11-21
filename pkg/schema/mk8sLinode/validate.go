package mk8sLinode

// validate.go attaches the LinodeProvider validation rule (ported from
// pkg/handlers/schema_mk8s.go LinodeProviderRule / mk8sLinode.ts).

import (
	"github.com/controlplane-com/libs-go/pkg/schema/base"
	"github.com/controlplane-com/libs-go/pkg/schema/mk8sCommon"
	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

func linodePoolRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"name":          mk8sCommon.NodePoolNameRule().Required(),
		"labels":        mk8sCommon.LabelsRule(),
		"taints":        mk8sCommon.TaintsRule(),
		"maxPods":       mk8sCommon.MaxPodsRule(),
		"serverType":    validate.String().Required(),
		"overrideImage": validate.String(),
		"subnetId":      validate.String().Required(),
		"minSize":       validate.Number().Integer().Min(0).Required(),
		"maxSize":       validate.Number().Integer().Required().MinRef("minSize"),
	})
}

// ValidationRule ports mk8sLinode.ts:69 LinodeProvider.
func (LinodeProvider) ValidationRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"region":           validate.String().Required().Unmodifiable(),
		"tokenSecretLink":  base.LocalLinkRule().Custom(base.ToKind("secret")).Required(),
		"firewallId":       validate.String(),
		"nodePools":        validate.Array(linodePoolRule()),
		"image":            validate.String().Required(),
		"authorizedUsers":  validate.Array(validate.String()).Sort().Dedupe(),
		"authorizedKeys":   validate.Array(validate.String()).Sort().Dedupe(),
		"vpcId":            validate.String().Required().Unmodifiable(),
		"preInstallScript": mk8sCommon.PreInstallScriptRule(false),
		"networking":       mk8sCommon.NetworkingConfigRule(false),
		"autoscaler":       mk8sCommon.AutoscalerConfigRule(),
	}).CustomTyped(mk8sCommon.PoolNamesUniqueObjectCustom("nodePools"))
}
