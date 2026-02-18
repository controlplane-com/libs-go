/* auto-generated */

package mk8sAws

import "github.com/controlplane-com/libs-go/pkg/schema/mk8sCommon"

type AmiRecommended string

const (
	AmiRecommendedUbuntuJammy2204       AmiRecommended = "ubuntu/jammy-22.04"
	AmiRecommendedUbuntuJammy2204Nvidia AmiRecommended = "ubuntu/jammy-22.04+nvidia"
	AmiRecommendedUbuntuFocal2004       AmiRecommended = "ubuntu/focal-20.04"
	AmiRecommendedUbuntuFocal2004Nvidia AmiRecommended = "ubuntu/focal-20.04+nvidia"
	AmiRecommendedUbuntuNoble2404       AmiRecommended = "ubuntu/noble-24.04"
	AmiRecommendedUbuntuNoble2404Nvidia AmiRecommended = "ubuntu/noble-24.04+nvidia"
	AmiRecommendedUbuntuBionic1804      AmiRecommended = "ubuntu/bionic-18.04"
	AmiRecommendedAmazonAmzn2           AmiRecommended = "amazon/amzn2"
	AmiRecommendedAmazonAl2023          AmiRecommended = "amazon/al2023"
)

type Ami struct {
	Recommended AmiRecommended `json:"recommended,omitempty"`
	Exact       string         `json:"exact,omitempty"`
}

type AssumeRoleLink struct {
	RoleArn           string `json:"roleArn"`
	ExternalId        string `json:"externalId,omitempty"`
	SessionNamePrefix string `json:"sessionNamePrefix,omitempty"`
}

type AwsPoolSpotAllocationStrategy string

const (
	AwsPoolSpotAllocationStrategyLowestPrice                  AwsPoolSpotAllocationStrategy = "lowest-price"
	AwsPoolSpotAllocationStrategyCapacityOptimized            AwsPoolSpotAllocationStrategy = "capacity-optimized"
	AwsPoolSpotAllocationStrategyCapacityOptimizedPrioritized AwsPoolSpotAllocationStrategy = "capacity-optimized-prioritized"
	AwsPoolSpotAllocationStrategyPriceCapacityOptimized       AwsPoolSpotAllocationStrategy = "price-capacity-optimized"
)

type AwsPoolCpuOptions struct {
	NestedVirtualization bool `json:"nestedVirtualization,omitempty"`
}

type AwsPool struct {
	Name                                string                        `json:"name"`
	Labels                              mk8sCommon.Labels             `json:"labels,omitempty"`
	Taints                              mk8sCommon.Taints             `json:"taints,omitempty"`
	InstanceTypes                       []string                      `json:"instanceTypes"`
	OverrideImage                       Ami                           `json:"overrideImage,omitempty"`
	BootDiskSize                        float32                       `json:"bootDiskSize"`
	MinSize                             float32                       `json:"minSize"`
	MaxSize                             float32                       `json:"maxSize"`
	OnDemandBaseCapacity                float32                       `json:"onDemandBaseCapacity"`
	OnDemandPercentageAboveBaseCapacity float32                       `json:"onDemandPercentageAboveBaseCapacity"`
	SpotAllocationStrategy              AwsPoolSpotAllocationStrategy `json:"spotAllocationStrategy,omitempty"`
	SubnetIds                           []string                      `json:"subnetIds"`
	ExtraSecurityGroupIds               []string                      `json:"extraSecurityGroupIds,omitempty"`
	CpuOptions                          AwsPoolCpuOptions             `json:"cpuOptions,omitempty"`
}

type AwsProviderRegion string

const (
	AwsProviderRegionAfSouth1     AwsProviderRegion = "af-south-1"
	AwsProviderRegionApEast1      AwsProviderRegion = "ap-east-1"
	AwsProviderRegionApNortheast1 AwsProviderRegion = "ap-northeast-1"
	AwsProviderRegionApNortheast2 AwsProviderRegion = "ap-northeast-2"
	AwsProviderRegionApNortheast3 AwsProviderRegion = "ap-northeast-3"
	AwsProviderRegionApSouth1     AwsProviderRegion = "ap-south-1"
	AwsProviderRegionApSouth2     AwsProviderRegion = "ap-south-2"
	AwsProviderRegionApSoutheast1 AwsProviderRegion = "ap-southeast-1"
	AwsProviderRegionApSoutheast2 AwsProviderRegion = "ap-southeast-2"
	AwsProviderRegionApSoutheast3 AwsProviderRegion = "ap-southeast-3"
	AwsProviderRegionApSoutheast4 AwsProviderRegion = "ap-southeast-4"
	AwsProviderRegionCaCentral1   AwsProviderRegion = "ca-central-1"
	AwsProviderRegionEuCentral1   AwsProviderRegion = "eu-central-1"
	AwsProviderRegionEuCentral2   AwsProviderRegion = "eu-central-2"
	AwsProviderRegionEuNorth1     AwsProviderRegion = "eu-north-1"
	AwsProviderRegionEuSouth1     AwsProviderRegion = "eu-south-1"
	AwsProviderRegionEuSouth2     AwsProviderRegion = "eu-south-2"
	AwsProviderRegionEuWest1      AwsProviderRegion = "eu-west-1"
	AwsProviderRegionEuWest2      AwsProviderRegion = "eu-west-2"
	AwsProviderRegionEuWest3      AwsProviderRegion = "eu-west-3"
	AwsProviderRegionMeCentral1   AwsProviderRegion = "me-central-1"
	AwsProviderRegionMeSouth1     AwsProviderRegion = "me-south-1"
	AwsProviderRegionSaEast1      AwsProviderRegion = "sa-east-1"
	AwsProviderRegionUsEast1      AwsProviderRegion = "us-east-1"
	AwsProviderRegionUsEast2      AwsProviderRegion = "us-east-2"
	AwsProviderRegionUsWest1      AwsProviderRegion = "us-west-1"
	AwsProviderRegionUsWest2      AwsProviderRegion = "us-west-2"
)

type AwsProviderAwsTags map[string]string

type AwsProviderNetworkingServiceNetwork string

const (
	AwsProviderNetworkingServiceNetwork10430016   AwsProviderNetworkingServiceNetwork = "10.43.0.0/16"
	AwsProviderNetworkingServiceNetwork1921680016 AwsProviderNetworkingServiceNetwork = "192.168.0.0/16"
)

type AwsProviderNetworkingPodNetwork string

const (
	AwsProviderNetworkingPodNetworkVpc       AwsProviderNetworkingPodNetwork = "vpc"
	AwsProviderNetworkingPodNetwork10420016  AwsProviderNetworkingPodNetwork = "10.42.0.0/16"
	AwsProviderNetworkingPodNetwork172160015 AwsProviderNetworkingPodNetwork = "172.16.0.0/15"
	AwsProviderNetworkingPodNetwork172180015 AwsProviderNetworkingPodNetwork = "172.18.0.0/15"
	AwsProviderNetworkingPodNetwork172200015 AwsProviderNetworkingPodNetwork = "172.20.0.0/15"
	AwsProviderNetworkingPodNetwork172220015 AwsProviderNetworkingPodNetwork = "172.22.0.0/15"
	AwsProviderNetworkingPodNetwork172240015 AwsProviderNetworkingPodNetwork = "172.24.0.0/15"
	AwsProviderNetworkingPodNetwork172260015 AwsProviderNetworkingPodNetwork = "172.26.0.0/15"
	AwsProviderNetworkingPodNetwork172280015 AwsProviderNetworkingPodNetwork = "172.28.0.0/15"
	AwsProviderNetworkingPodNetwork172300015 AwsProviderNetworkingPodNetwork = "172.30.0.0/15"
)

type AwsProviderNetworking struct {
	ServiceNetwork AwsProviderNetworkingServiceNetwork `json:"serviceNetwork,omitempty"`
	PodNetwork     AwsProviderNetworkingPodNetwork     `json:"podNetwork,omitempty"`
	DnsForwarder   string                              `json:"dnsForwarder,omitempty"`
}

type AwsProviderImageRecommended string

const (
	AwsProviderImageRecommendedUbuntuJammy2204       AwsProviderImageRecommended = "ubuntu/jammy-22.04"
	AwsProviderImageRecommendedUbuntuJammy2204Nvidia AwsProviderImageRecommended = "ubuntu/jammy-22.04+nvidia"
	AwsProviderImageRecommendedUbuntuFocal2004       AwsProviderImageRecommended = "ubuntu/focal-20.04"
	AwsProviderImageRecommendedUbuntuFocal2004Nvidia AwsProviderImageRecommended = "ubuntu/focal-20.04+nvidia"
	AwsProviderImageRecommendedUbuntuNoble2404       AwsProviderImageRecommended = "ubuntu/noble-24.04"
	AwsProviderImageRecommendedUbuntuNoble2404Nvidia AwsProviderImageRecommended = "ubuntu/noble-24.04+nvidia"
	AwsProviderImageRecommendedUbuntuBionic1804      AwsProviderImageRecommended = "ubuntu/bionic-18.04"
	AwsProviderImageRecommendedAmazonAmzn2           AwsProviderImageRecommended = "amazon/amzn2"
	AwsProviderImageRecommendedAmazonAl2023          AwsProviderImageRecommended = "amazon/al2023"
)

type AwsProviderImage struct {
	Recommended AwsProviderImageRecommended `json:"recommended,omitempty"`
	Exact       string                      `json:"exact,omitempty"`
}

type AwsProvider struct {
	Region               AwsProviderRegion           `json:"region,omitempty"`
	AwsTags              AwsProviderAwsTags          `json:"awsTags,omitempty"`
	SkipCreateRoles      bool                        `json:"skipCreateRoles,omitempty"`
	Networking           AwsProviderNetworking       `json:"networking,omitempty"`
	PreInstallScript     string                      `json:"preInstallScript,omitempty"`
	Image                AwsProviderImage            `json:"image"`
	DeployRoleArn        string                      `json:"deployRoleArn"`
	DeployRoleChain      []AssumeRoleLink            `json:"deployRoleChain,omitempty"`
	VpcId                string                      `json:"vpcId"`
	KeyPair              string                      `json:"keyPair,omitempty"`
	DiskEncryptionKeyArn string                      `json:"diskEncryptionKeyArn,omitempty"`
	SecurityGroupIds     []string                    `json:"securityGroupIds,omitempty"`
	ExtraNodePolicies    []PolicyArn                 `json:"extraNodePolicies,omitempty"`
	NodePools            []AwsPool                   `json:"nodePools,omitempty"`
	Autoscaler           mk8sCommon.AutoscalerConfig `json:"autoscaler,omitempty"`
}

type AwsProviderStatus map[string]any

type PolicyArn string
