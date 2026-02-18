/* auto-generated */

package mk8sLambdalabs

import "github.com/controlplane-com/libs-go/pkg/schema/mk8sCommon"

type LambdalabsJoinParams struct {
	NodePoolName string `json:"nodePoolName"`

	/* WARNING!! Arbitrary properties are being ignored! */
}

type LambdalabsPoolInstanceType string

const (
	LambdalabsPoolInstanceTypeCpu4XGeneral      LambdalabsPoolInstanceType = "cpu_4x_general"
	LambdalabsPoolInstanceTypeGpu1XA10          LambdalabsPoolInstanceType = "gpu_1x_a10"
	LambdalabsPoolInstanceTypeGpu1XA100         LambdalabsPoolInstanceType = "gpu_1x_a100"
	LambdalabsPoolInstanceTypeGpu1XA100Sxm4     LambdalabsPoolInstanceType = "gpu_1x_a100_sxm4"
	LambdalabsPoolInstanceTypeGpu1XA6000        LambdalabsPoolInstanceType = "gpu_1x_a6000"
	LambdalabsPoolInstanceTypeGpu1XH100Pcie     LambdalabsPoolInstanceType = "gpu_1x_h100_pcie"
	LambdalabsPoolInstanceTypeGpu1XRtx6000      LambdalabsPoolInstanceType = "gpu_1x_rtx6000"
	LambdalabsPoolInstanceTypeGpu2XA100         LambdalabsPoolInstanceType = "gpu_2x_a100"
	LambdalabsPoolInstanceTypeGpu2XA6000        LambdalabsPoolInstanceType = "gpu_2x_a6000"
	LambdalabsPoolInstanceTypeGpu4XA100         LambdalabsPoolInstanceType = "gpu_4x_a100"
	LambdalabsPoolInstanceTypeGpu4XA6000        LambdalabsPoolInstanceType = "gpu_4x_a6000"
	LambdalabsPoolInstanceTypeGpu8XA100         LambdalabsPoolInstanceType = "gpu_8x_a100"
	LambdalabsPoolInstanceTypeGpu8XA10080GbSxm4 LambdalabsPoolInstanceType = "gpu_8x_a100_80gb_sxm4"
	LambdalabsPoolInstanceTypeGpu8XH100Sxm5     LambdalabsPoolInstanceType = "gpu_8x_h100_sxm5"
	LambdalabsPoolInstanceTypeGpu8XV100         LambdalabsPoolInstanceType = "gpu_8x_v100"
)

type LambdalabsPool struct {
	Name         string                     `json:"name"`
	Labels       mk8sCommon.Labels          `json:"labels,omitempty"`
	Taints       mk8sCommon.Taints          `json:"taints,omitempty"`
	MinSize      float32                    `json:"minSize"`
	MaxSize      float32                    `json:"maxSize"`
	InstanceType LambdalabsPoolInstanceType `json:"instanceType,omitempty"`
}

type LambdalabsProviderRegion string

const (
	LambdalabsProviderRegionAsiaNortheast1      LambdalabsProviderRegion = "asia-northeast-1"
	LambdalabsProviderRegionAsiaNortheast2      LambdalabsProviderRegion = "asia-northeast-2"
	LambdalabsProviderRegionAsiaSouth1          LambdalabsProviderRegion = "asia-south-1"
	LambdalabsProviderRegionAustraliaSoutheast1 LambdalabsProviderRegion = "australia-southeast-1"
	LambdalabsProviderRegionEuropeCentral1      LambdalabsProviderRegion = "europe-central-1"
	LambdalabsProviderRegionEuropeSouth1        LambdalabsProviderRegion = "europe-south-1"
	LambdalabsProviderRegionMeWest1             LambdalabsProviderRegion = "me-west-1"
	LambdalabsProviderRegionUsEast1             LambdalabsProviderRegion = "us-east-1"
	LambdalabsProviderRegionUsMidwest1          LambdalabsProviderRegion = "us-midwest-1"
	LambdalabsProviderRegionUsSouth1            LambdalabsProviderRegion = "us-south-1"
	LambdalabsProviderRegionUsWest1             LambdalabsProviderRegion = "us-west-1"
	LambdalabsProviderRegionUsWest2             LambdalabsProviderRegion = "us-west-2"
	LambdalabsProviderRegionUsWest3             LambdalabsProviderRegion = "us-west-3"
)

type LambdalabsProvider struct {
	Region             LambdalabsProviderRegion    `json:"region,omitempty"`
	TokenSecretLink    string                      `json:"tokenSecretLink"`
	NodePools          []LambdalabsPool            `json:"nodePools,omitempty"`
	SshKey             string                      `json:"sshKey"`
	UnmanagedNodePools []mk8sCommon.UnmanagedPool  `json:"unmanagedNodePools,omitempty"`
	Autoscaler         mk8sCommon.AutoscalerConfig `json:"autoscaler,omitempty"`
	FileSystems        []string                    `json:"fileSystems,omitempty"`
	PreInstallScript   mk8sCommon.PreInstallScript `json:"preInstallScript,omitempty"`
}

type LambdalabsProviderStatus map[string]any
