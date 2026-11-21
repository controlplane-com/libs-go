/* auto-generated */

package mk8sTriton

import "github.com/controlplane-com/libs-go/pkg/schema/mk8sCommon"

type LoadBalancerConfigNone struct {
}

type LoadBalancerConfigGateway struct {
}

type LoadBalancerConfig struct {
	Manual  *Manual                    `json:"manual,omitzero"`
	None    *LoadBalancerConfigNone    `json:"none,omitzero"`
	Gateway *LoadBalancerConfigGateway `json:"gateway,omitzero"`
}

type ManualMetadata map[string]string

type ManualTags map[string]string

type ManualLogging struct {
	NodePort       *float64 `json:"nodePort,omitzero"`
	ExternalSyslog string   `json:"externalSyslog,omitzero"`
}

type Manual struct {
	PackageId         string         `json:"packageId"`
	ImageId           string         `json:"imageId"`
	PublicNetworkId   string         `json:"publicNetworkId"`
	PrivateNetworkIds []string       `json:"privateNetworkIds,omitzero"`
	Metadata          ManualMetadata `json:"metadata,omitzero"`
	Tags              ManualTags     `json:"tags,omitzero"`
	Logging           *ManualLogging `json:"logging,omitzero"`
	Count             float64        `json:"count"`
	CnsInternalDomain string         `json:"cnsInternalDomain"`
	CnsPublicDomain   string         `json:"cnsPublicDomain"`
}

type SdcConnection struct {
	Url                  string `json:"url"`
	Account              string `json:"account"`
	User                 string `json:"user,omitzero"`
	PrivateKeySecretLink string `json:"privateKeySecretLink"`
}

type TritonJoinParams struct {
	NodePoolName string `json:"nodePoolName"`

	/* WARNING!! Arbitrary properties are being ignored! */
}

type TritonPoolTritonTags map[string]string

type TritonPool struct {
	Name              string               `json:"name"`
	Labels            mk8sCommon.Labels    `json:"labels,omitzero"`
	Taints            mk8sCommon.Taints    `json:"taints,omitzero"`
	MaxPods           *mk8sCommon.MaxPods  `json:"maxPods,omitzero"`
	PackageId         string               `json:"packageId"`
	OverrideImageId   string               `json:"overrideImageId,omitzero"`
	PublicNetworkId   string               `json:"publicNetworkId,omitzero"`
	PrivateNetworkIds []string             `json:"privateNetworkIds,omitzero"`
	TritonTags        TritonPoolTritonTags `json:"tritonTags,omitzero"`
	MinSize           float64              `json:"minSize"`
	MaxSize           float64              `json:"maxSize"`
}

type TritonProviderConnection struct {
	Url                  string `json:"url"`
	Account              string `json:"account"`
	User                 string `json:"user,omitzero"`
	PrivateKeySecretLink string `json:"privateKeySecretLink"`
}

type TritonProviderNetworkingServiceNetwork string

const (
	TritonProviderNetworkingServiceNetwork10430016   TritonProviderNetworkingServiceNetwork = "10.43.0.0/16"
	TritonProviderNetworkingServiceNetwork1921680016 TritonProviderNetworkingServiceNetwork = "192.168.0.0/16"
)

type TritonProviderNetworkingPodNetwork string

const (
	TritonProviderNetworkingPodNetwork10420016  TritonProviderNetworkingPodNetwork = "10.42.0.0/16"
	TritonProviderNetworkingPodNetwork172160015 TritonProviderNetworkingPodNetwork = "172.16.0.0/15"
	TritonProviderNetworkingPodNetwork172180015 TritonProviderNetworkingPodNetwork = "172.18.0.0/15"
	TritonProviderNetworkingPodNetwork172200015 TritonProviderNetworkingPodNetwork = "172.20.0.0/15"
	TritonProviderNetworkingPodNetwork172220015 TritonProviderNetworkingPodNetwork = "172.22.0.0/15"
	TritonProviderNetworkingPodNetwork172240015 TritonProviderNetworkingPodNetwork = "172.24.0.0/15"
	TritonProviderNetworkingPodNetwork172260015 TritonProviderNetworkingPodNetwork = "172.26.0.0/15"
	TritonProviderNetworkingPodNetwork172280015 TritonProviderNetworkingPodNetwork = "172.28.0.0/15"
	TritonProviderNetworkingPodNetwork172300015 TritonProviderNetworkingPodNetwork = "172.30.0.0/15"
)

type TritonProviderNetworking struct {
	ServiceNetwork TritonProviderNetworkingServiceNetwork `json:"serviceNetwork,omitzero"`
	PodNetwork     TritonProviderNetworkingPodNetwork     `json:"podNetwork,omitzero"`
	DnsForwarder   string                                 `json:"dnsForwarder,omitzero"`
}

type TritonProviderLocation string

const (
	TritonProviderLocationAwsEuCentral1 TritonProviderLocation = "aws-eu-central-1"
	TritonProviderLocationAwsUsEast2    TritonProviderLocation = "aws-us-east-2"
	TritonProviderLocationAwsUsWest2    TritonProviderLocation = "aws-us-west-2"
	TritonProviderLocationGcpMeWest1    TritonProviderLocation = "gcp-me-west1"
	TritonProviderLocationGcpUsEast1    TritonProviderLocation = "gcp-us-east1"
)

type TritonProvider struct {
	Connection       TritonProviderConnection     `json:"connection"`
	Networking       *TritonProviderNetworking    `json:"networking,omitzero"`
	PreInstallScript mk8sCommon.PreInstallScript  `json:"preInstallScript,omitzero"`
	Location         TritonProviderLocation       `json:"location,omitzero" validate:"required"`
	LoadBalancer     *LoadBalancerConfig          `json:"loadBalancer,omitzero"`
	PrivateNetworkId string                       `json:"privateNetworkId"`
	FirewallEnabled  *bool                        `json:"firewallEnabled,omitzero"`
	NodePools        []TritonPool                 `json:"nodePools,omitzero"`
	ImageId          string                       `json:"imageId"`
	SshKeys          []mk8sCommon.SshPublicKey    `json:"sshKeys,omitzero"`
	Autoscaler       *mk8sCommon.AutoscalerConfig `json:"autoscaler,omitzero"`
}

type TritonProviderStatus map[string]any
