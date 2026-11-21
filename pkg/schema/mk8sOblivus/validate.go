package mk8sOblivus

// validate.go attaches the OblivusProvider validation rule (ported from
// pkg/handlers/schema_mk8s.go OblivusProviderRule / mk8sOblivus.ts).

import (
	"github.com/controlplane-com/libs-go/pkg/schema/base"
	"github.com/controlplane-com/libs-go/pkg/schema/mk8sCommon"
	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

func oblivusPoolRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"name":    mk8sCommon.NodePoolNameRule().Required(),
		"labels":  mk8sCommon.LabelsRule(),
		"taints":  mk8sCommon.TaintsRule(),
		"maxPods": mk8sCommon.MaxPodsRule(),
		"minSize": validate.Number().Integer().Min(0).Required(),
		"maxSize": validate.Number().Integer().Required().MinRef("minSize"),
		"flavor": validate.String().Required().Valid(
			"A100_NVLINK_80GB_x8", "A100_PCIE_80GB_x1", "A100_PCIE_80GB_x2", "A100_PCIE_80GB_x4", "A100_PCIE_80GB_x8",
			"H100_NVLINK_80GB_x8", "H100_PCIE_80GB_x1", "H100_PCIE_80GB_x2", "H100_PCIE_80GB_x4", "H100_PCIE_80GB_x8",
			"INTEL_XEON_V3_x16", "INTEL_XEON_V3_x4", "INTEL_XEON_V3_x8",
			"L40_x1", "L40_x2", "L40_x4", "L40_x8",
			"RTX_A4000_x1", "RTX_A4000_x2", "RTX_A4000_x4", "RTX_A4000_x8",
			"RTX_A5000_x1", "RTX_A5000_x2", "RTX_A5000_x4", "RTX_A5000_x8",
			"RTX_A6000_x1", "RTX_A6000_x2", "RTX_A6000_x4", "RTX_A6000_x8",
		),
	})
}

// ValidationRule ports mk8sOblivus.ts:54 OblivusProvider.
func (OblivusProvider) ValidationRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"datacenter":         validate.String().Required().Valid("MON1", "OSL1").Unmodifiable(),
		"tokenSecretLink":    base.LocalLinkRule().Custom(base.ToKind("secret")).Required(),
		"nodePools":          validate.Array(oblivusPoolRule()),
		"sshKeys":            validate.Array(validate.String()).Sort().Dedupe(),
		"unmanagedNodePools": validate.Array(mk8sCommon.UnmanagedPoolRule()),
		"autoscaler":         mk8sCommon.AutoscalerConfigRule(),
		"preInstallScript":   mk8sCommon.PreInstallScriptRule(false),
	}).CustomTyped(mk8sCommon.PoolNamesUniqueObjectCustom("unmanagedNodePools", "nodePools"))
}
