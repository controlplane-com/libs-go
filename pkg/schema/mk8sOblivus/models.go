/* auto-generated */

package mk8sOblivus

import "github.com/controlplane-com/libs-go/pkg/schema/mk8sCommon"

type OblivusJoinParams struct {
	NodePoolName string `json:"nodePoolName"`

	/* WARNING!! Arbitrary properties are being ignored! */
}

type OblivusPoolFlavor string

const (
	OblivusPoolFlavorA100Nvlink80GbX8 OblivusPoolFlavor = "A100_NVLINK_80GB_x8"
	OblivusPoolFlavorA100Pcie80GbX1   OblivusPoolFlavor = "A100_PCIE_80GB_x1"
	OblivusPoolFlavorA100Pcie80GbX2   OblivusPoolFlavor = "A100_PCIE_80GB_x2"
	OblivusPoolFlavorA100Pcie80GbX4   OblivusPoolFlavor = "A100_PCIE_80GB_x4"
	OblivusPoolFlavorA100Pcie80GbX8   OblivusPoolFlavor = "A100_PCIE_80GB_x8"
	OblivusPoolFlavorH100Nvlink80GbX8 OblivusPoolFlavor = "H100_NVLINK_80GB_x8"
	OblivusPoolFlavorH100Pcie80GbX1   OblivusPoolFlavor = "H100_PCIE_80GB_x1"
	OblivusPoolFlavorH100Pcie80GbX2   OblivusPoolFlavor = "H100_PCIE_80GB_x2"
	OblivusPoolFlavorH100Pcie80GbX4   OblivusPoolFlavor = "H100_PCIE_80GB_x4"
	OblivusPoolFlavorH100Pcie80GbX8   OblivusPoolFlavor = "H100_PCIE_80GB_x8"
	OblivusPoolFlavorIntelXeonV3X16   OblivusPoolFlavor = "INTEL_XEON_V3_x16"
	OblivusPoolFlavorIntelXeonV3X4    OblivusPoolFlavor = "INTEL_XEON_V3_x4"
	OblivusPoolFlavorIntelXeonV3X8    OblivusPoolFlavor = "INTEL_XEON_V3_x8"
	OblivusPoolFlavorL40X1            OblivusPoolFlavor = "L40_x1"
	OblivusPoolFlavorL40X2            OblivusPoolFlavor = "L40_x2"
	OblivusPoolFlavorL40X4            OblivusPoolFlavor = "L40_x4"
	OblivusPoolFlavorL40X8            OblivusPoolFlavor = "L40_x8"
	OblivusPoolFlavorRtxA4000X1       OblivusPoolFlavor = "RTX_A4000_x1"
	OblivusPoolFlavorRtxA4000X2       OblivusPoolFlavor = "RTX_A4000_x2"
	OblivusPoolFlavorRtxA4000X4       OblivusPoolFlavor = "RTX_A4000_x4"
	OblivusPoolFlavorRtxA4000X8       OblivusPoolFlavor = "RTX_A4000_x8"
	OblivusPoolFlavorRtxA5000X1       OblivusPoolFlavor = "RTX_A5000_x1"
	OblivusPoolFlavorRtxA5000X2       OblivusPoolFlavor = "RTX_A5000_x2"
	OblivusPoolFlavorRtxA5000X4       OblivusPoolFlavor = "RTX_A5000_x4"
	OblivusPoolFlavorRtxA5000X8       OblivusPoolFlavor = "RTX_A5000_x8"
	OblivusPoolFlavorRtxA6000X1       OblivusPoolFlavor = "RTX_A6000_x1"
	OblivusPoolFlavorRtxA6000X2       OblivusPoolFlavor = "RTX_A6000_x2"
	OblivusPoolFlavorRtxA6000X4       OblivusPoolFlavor = "RTX_A6000_x4"
	OblivusPoolFlavorRtxA6000X8       OblivusPoolFlavor = "RTX_A6000_x8"
)

type OblivusPool struct {
	Name    string            `json:"name"`
	Labels  mk8sCommon.Labels `json:"labels,omitempty"`
	Taints  mk8sCommon.Taints `json:"taints,omitempty"`
	MinSize float32           `json:"minSize"`
	MaxSize float32           `json:"maxSize"`
	Flavor  OblivusPoolFlavor `json:"flavor,omitempty"`
}

type OblivusProviderDatacenter string

const (
	OblivusProviderDatacenterMon1 OblivusProviderDatacenter = "MON1"
	OblivusProviderDatacenterOsl1 OblivusProviderDatacenter = "OSL1"
)

type OblivusProvider struct {
	Datacenter         OblivusProviderDatacenter   `json:"datacenter,omitempty"`
	TokenSecretLink    string                      `json:"tokenSecretLink"`
	NodePools          []OblivusPool               `json:"nodePools,omitempty"`
	SshKeys            []string                    `json:"sshKeys,omitempty"`
	UnmanagedNodePools []mk8sCommon.UnmanagedPool  `json:"unmanagedNodePools,omitempty"`
	Autoscaler         mk8sCommon.AutoscalerConfig `json:"autoscaler,omitempty"`
	PreInstallScript   mk8sCommon.PreInstallScript `json:"preInstallScript,omitempty"`
}

type OblivusProviderStatus map[string]any
