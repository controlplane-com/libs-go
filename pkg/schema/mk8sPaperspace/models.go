/* auto-generated */

package mk8sPaperspace

import "github.com/controlplane-com/libs-go/pkg/schema/mk8sCommon"

type PaperspaceJoinParams struct {
	NodePoolName string `json:"nodePoolName"`

	/* WARNING!! Arbitrary properties are being ignored! */
}

type PaperspacePoolPublicIpType string

const (
	PaperspacePoolPublicIpTypeNone    PaperspacePoolPublicIpType = "none"
	PaperspacePoolPublicIpTypeDynamic PaperspacePoolPublicIpType = "dynamic"
	PaperspacePoolPublicIpTypeStatic  PaperspacePoolPublicIpType = "static"
)

type PaperspacePoolMachineType string

const (
	PaperspacePoolMachineTypeA100      PaperspacePoolMachineType = "A100"
	PaperspacePoolMachineTypeA10080G   PaperspacePoolMachineType = "A100-80G"
	PaperspacePoolMachineTypeA10080Gx2 PaperspacePoolMachineType = "A100-80Gx2"
	PaperspacePoolMachineTypeA10080Gx4 PaperspacePoolMachineType = "A100-80Gx4"
	PaperspacePoolMachineTypeA10080Gx8 PaperspacePoolMachineType = "A100-80Gx8"
	PaperspacePoolMachineTypeA100X2    PaperspacePoolMachineType = "A100x2"
	PaperspacePoolMachineTypeA100X4    PaperspacePoolMachineType = "A100x4"
	PaperspacePoolMachineTypeA100X8    PaperspacePoolMachineType = "A100x8"
	PaperspacePoolMachineTypeA4000     PaperspacePoolMachineType = "A4000"
	PaperspacePoolMachineTypeA4000X2   PaperspacePoolMachineType = "A4000x2"
	PaperspacePoolMachineTypeA4000X4   PaperspacePoolMachineType = "A4000x4"
	PaperspacePoolMachineTypeA5000     PaperspacePoolMachineType = "A5000"
	PaperspacePoolMachineTypeA5000X2   PaperspacePoolMachineType = "A5000x2"
	PaperspacePoolMachineTypeA5000X4   PaperspacePoolMachineType = "A5000x4"
	PaperspacePoolMachineTypeA6000     PaperspacePoolMachineType = "A6000"
	PaperspacePoolMachineTypeA6000X2   PaperspacePoolMachineType = "A6000x2"
	PaperspacePoolMachineTypeA6000X4   PaperspacePoolMachineType = "A6000x4"
	PaperspacePoolMachineTypeC10       PaperspacePoolMachineType = "C10"
	PaperspacePoolMachineTypeC2        PaperspacePoolMachineType = "C2"
	PaperspacePoolMachineTypeC3        PaperspacePoolMachineType = "C3"
	PaperspacePoolMachineTypeC4        PaperspacePoolMachineType = "C4"
	PaperspacePoolMachineTypeC5        PaperspacePoolMachineType = "C5"
	PaperspacePoolMachineTypeC6        PaperspacePoolMachineType = "C6"
	PaperspacePoolMachineTypeC7        PaperspacePoolMachineType = "C7"
	PaperspacePoolMachineTypeC8        PaperspacePoolMachineType = "C8"
	PaperspacePoolMachineTypeC9        PaperspacePoolMachineType = "C9"
	PaperspacePoolMachineTypeGpu       PaperspacePoolMachineType = "GPU+"
	PaperspacePoolMachineTypeP4000     PaperspacePoolMachineType = "P4000"
	PaperspacePoolMachineTypeP4000X2   PaperspacePoolMachineType = "P4000x2"
	PaperspacePoolMachineTypeP4000X4   PaperspacePoolMachineType = "P4000x4"
	PaperspacePoolMachineTypeP5000     PaperspacePoolMachineType = "P5000"
	PaperspacePoolMachineTypeP5000X2   PaperspacePoolMachineType = "P5000x2"
	PaperspacePoolMachineTypeP6000     PaperspacePoolMachineType = "P6000"
	PaperspacePoolMachineTypeP6000X2   PaperspacePoolMachineType = "P6000x2"
	PaperspacePoolMachineTypeRtx4000   PaperspacePoolMachineType = "RTX4000"
	PaperspacePoolMachineTypeRtx5000   PaperspacePoolMachineType = "RTX5000"
	PaperspacePoolMachineTypeRtx5000X2 PaperspacePoolMachineType = "RTX5000x2"
	PaperspacePoolMachineTypeV100      PaperspacePoolMachineType = "V100"
	PaperspacePoolMachineTypeV10032G   PaperspacePoolMachineType = "V100-32G"
	PaperspacePoolMachineTypeV10032Gx2 PaperspacePoolMachineType = "V100-32Gx2"
	PaperspacePoolMachineTypeV10032Gx4 PaperspacePoolMachineType = "V100-32Gx4"
)

type PaperspacePool struct {
	Name         string                     `json:"name"`
	Labels       mk8sCommon.Labels          `json:"labels,omitempty"`
	Taints       mk8sCommon.Taints          `json:"taints,omitempty"`
	MinSize      float32                    `json:"minSize"`
	MaxSize      float32                    `json:"maxSize"`
	PublicIpType PaperspacePoolPublicIpType `json:"publicIpType,omitempty"`
	BootDiskSize *float32                   `json:"bootDiskSize,omitempty"`
	MachineType  PaperspacePoolMachineType  `json:"machineType,omitempty"`
}

type PaperspaceProviderRegion string

const (
	PaperspaceProviderRegionAms1 PaperspaceProviderRegion = "AMS1"
	PaperspaceProviderRegionNy2  PaperspaceProviderRegion = "NY2"
	PaperspaceProviderRegionCa1  PaperspaceProviderRegion = "CA1"
)

type PaperspaceProvider struct {
	Region             PaperspaceProviderRegion    `json:"region,omitempty"`
	TokenSecretLink    string                      `json:"tokenSecretLink"`
	SharedDrives       []string                    `json:"sharedDrives,omitempty"`
	NodePools          []PaperspacePool            `json:"nodePools,omitempty"`
	Autoscaler         mk8sCommon.AutoscalerConfig `json:"autoscaler,omitempty"`
	UnmanagedNodePools []mk8sCommon.UnmanagedPool  `json:"unmanagedNodePools,omitempty"`
	PreInstallScript   mk8sCommon.PreInstallScript `json:"preInstallScript,omitempty"`
	UserIds            []string                    `json:"userIds,omitempty"`
	NetworkId          string                      `json:"networkId"`
}

type PaperspaceProviderStatus map[string]any
