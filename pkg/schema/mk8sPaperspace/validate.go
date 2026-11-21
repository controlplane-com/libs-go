package mk8sPaperspace

// validate.go attaches the PaperspaceProvider validation rule (ported from
// pkg/handlers/schema_mk8s.go PaperspaceProviderRule / mk8sPaperspace.ts).

import (
	"github.com/controlplane-com/libs-go/pkg/schema/base"
	"github.com/controlplane-com/libs-go/pkg/schema/mk8sCommon"
	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

func paperspacePoolRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"name":         mk8sCommon.NodePoolNameRule().Required(),
		"labels":       mk8sCommon.LabelsRule(),
		"taints":       mk8sCommon.TaintsRule(),
		"maxPods":      mk8sCommon.MaxPodsRule(),
		"minSize":      validate.Number().Integer().Min(0).Required(),
		"maxSize":      validate.Number().Integer().Required().MinRef("minSize"),
		"publicIpType": validate.String().Required().Valid("none", "dynamic", "static"),
		"bootDiskSize": validate.Number().Integer().Valid(float64(50), float64(100), float64(250), float64(500), float64(1000), float64(2000)),
		"machineType": validate.String().Required().Valid(
			"A100", "A100-80G", "A100-80Gx2", "A100-80Gx4", "A100-80Gx8", "A100x2", "A100x4", "A100x8",
			"A4000", "A4000x2", "A4000x4", "A5000", "A5000x2", "A5000x4", "A6000", "A6000x2", "A6000x4",
			"C10", "C2", "C3", "C4", "C5", "C6", "C7", "C8", "C9", "GPU+",
			"P4000", "P4000x2", "P4000x4", "P5000", "P5000x2", "P6000", "P6000x2",
			"RTX4000", "RTX5000", "RTX5000x2", "V100", "V100-32G", "V100-32Gx2", "V100-32Gx4",
		),
	})
}

// ValidationRule ports mk8sPaperspace.ts:64 PaperspaceProvider.
func (PaperspaceProvider) ValidationRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"region":             validate.String().Required().Valid("AMS1", "NY2", "CA1").Unmodifiable(),
		"tokenSecretLink":    base.LocalLinkRule().Custom(base.ToKind("secret")).Required(),
		"sharedDrives":       validate.Array(validate.String()),
		"nodePools":          validate.Array(paperspacePoolRule()),
		"autoscaler":         mk8sCommon.AutoscalerConfigRule(),
		"unmanagedNodePools": validate.Array(mk8sCommon.UnmanagedPoolRule()),
		"preInstallScript":   mk8sCommon.PreInstallScriptRule(false),
		"userIds":            validate.Array(validate.String()).Sort().Dedupe(),
		"networkId":          validate.String().Required(),
	}).CustomTyped(mk8sCommon.PoolNamesUniqueObjectCustom("unmanagedNodePools", "nodePools"))
}
