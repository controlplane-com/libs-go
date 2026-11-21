package mk8sLambdalabs

// validate.go attaches the LambdalabsProvider validation rule (ported from
// pkg/handlers/schema_mk8s.go LambdalabsProviderRule / mk8sLambdalabs.ts).

import (
	"github.com/controlplane-com/libs-go/pkg/schema/base"
	"github.com/controlplane-com/libs-go/pkg/schema/mk8sCommon"
	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

func lambdalabsPoolRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"name":    mk8sCommon.NodePoolNameRule().Required(),
		"labels":  mk8sCommon.LabelsRule(),
		"taints":  mk8sCommon.TaintsRule(),
		"maxPods": mk8sCommon.MaxPodsRule(),
		"minSize": validate.Number().Integer().Min(0).Required(),
		"maxSize": validate.Number().Integer().Required().MinRef("minSize"),
		"instanceType": validate.String().Required().Valid(
			"cpu_4x_general", "gpu_1x_a10", "gpu_1x_a100", "gpu_1x_a100_sxm4", "gpu_1x_a6000",
			"gpu_1x_h100_pcie", "gpu_1x_rtx6000", "gpu_2x_a100", "gpu_2x_a6000", "gpu_4x_a100",
			"gpu_4x_a6000", "gpu_8x_a100", "gpu_8x_a100_80gb_sxm4", "gpu_8x_h100_sxm5", "gpu_8x_v100",
		),
	})
}

// ValidationRule ports mk8sLambdalabs.ts:34 LambdalabsProvider.
func (LambdalabsProvider) ValidationRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"region": validate.String().Required().Valid(
			"asia-northeast-1", "asia-northeast-2", "asia-south-1", "australia-southeast-1",
			"europe-central-1", "europe-south-1", "me-west-1", "us-east-1", "us-midwest-1",
			"us-south-1", "us-west-1", "us-west-2", "us-west-3",
		).Unmodifiable(),
		"tokenSecretLink":    base.LocalLinkRule().Custom(base.ToKind("secret")).Required(),
		"nodePools":          validate.Array(lambdalabsPoolRule()),
		"sshKey":             validate.String().Required(),
		"unmanagedNodePools": validate.Array(mk8sCommon.UnmanagedPoolRule()),
		"autoscaler":         mk8sCommon.AutoscalerConfigRule(),
		"fileSystems":        validate.Array(validate.String()),
		"preInstallScript":   mk8sCommon.PreInstallScriptRule(false),
	}).CustomTyped(mk8sCommon.PoolNamesUniqueObjectCustom("unmanagedNodePools", "nodePools"))
}
