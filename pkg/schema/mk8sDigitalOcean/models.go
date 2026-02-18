/* auto-generated */

package mk8sDigitalOcean

import "github.com/controlplane-com/libs-go/pkg/schema/mk8sCommon"

type DigitalOceanPool struct {
	Name          string            `json:"name"`
	Labels        mk8sCommon.Labels `json:"labels,omitempty"`
	Taints        mk8sCommon.Taints `json:"taints,omitempty"`
	DropletSize   string            `json:"dropletSize"`
	OverrideImage string            `json:"overrideImage,omitempty"`
	MinSize       float32           `json:"minSize"`
	MaxSize       float32           `json:"maxSize"`
}

type DigitalOceanProviderRegion string

const (
	DigitalOceanProviderRegionAms3 DigitalOceanProviderRegion = "ams3"
	DigitalOceanProviderRegionBlr1 DigitalOceanProviderRegion = "blr1"
	DigitalOceanProviderRegionFra1 DigitalOceanProviderRegion = "fra1"
	DigitalOceanProviderRegionLon1 DigitalOceanProviderRegion = "lon1"
	DigitalOceanProviderRegionNyc1 DigitalOceanProviderRegion = "nyc1"
	DigitalOceanProviderRegionNyc2 DigitalOceanProviderRegion = "nyc2"
	DigitalOceanProviderRegionNyc3 DigitalOceanProviderRegion = "nyc3"
	DigitalOceanProviderRegionSfo2 DigitalOceanProviderRegion = "sfo2"
	DigitalOceanProviderRegionSfo3 DigitalOceanProviderRegion = "sfo3"
	DigitalOceanProviderRegionSgp1 DigitalOceanProviderRegion = "sgp1"
	DigitalOceanProviderRegionSyd1 DigitalOceanProviderRegion = "syd1"
	DigitalOceanProviderRegionTor1 DigitalOceanProviderRegion = "tor1"
)

type DigitalOceanProviderNetworkingServiceNetwork string

const (
	DigitalOceanProviderNetworkingServiceNetwork10430016   DigitalOceanProviderNetworkingServiceNetwork = "10.43.0.0/16"
	DigitalOceanProviderNetworkingServiceNetwork1921680016 DigitalOceanProviderNetworkingServiceNetwork = "192.168.0.0/16"
)

type DigitalOceanProviderNetworkingPodNetwork string

const (
	DigitalOceanProviderNetworkingPodNetwork10420016  DigitalOceanProviderNetworkingPodNetwork = "10.42.0.0/16"
	DigitalOceanProviderNetworkingPodNetwork172160015 DigitalOceanProviderNetworkingPodNetwork = "172.16.0.0/15"
	DigitalOceanProviderNetworkingPodNetwork172180015 DigitalOceanProviderNetworkingPodNetwork = "172.18.0.0/15"
	DigitalOceanProviderNetworkingPodNetwork172200015 DigitalOceanProviderNetworkingPodNetwork = "172.20.0.0/15"
	DigitalOceanProviderNetworkingPodNetwork172220015 DigitalOceanProviderNetworkingPodNetwork = "172.22.0.0/15"
	DigitalOceanProviderNetworkingPodNetwork172240015 DigitalOceanProviderNetworkingPodNetwork = "172.24.0.0/15"
	DigitalOceanProviderNetworkingPodNetwork172260015 DigitalOceanProviderNetworkingPodNetwork = "172.26.0.0/15"
	DigitalOceanProviderNetworkingPodNetwork172280015 DigitalOceanProviderNetworkingPodNetwork = "172.28.0.0/15"
	DigitalOceanProviderNetworkingPodNetwork172300015 DigitalOceanProviderNetworkingPodNetwork = "172.30.0.0/15"
)

type DigitalOceanProviderNetworking struct {
	ServiceNetwork DigitalOceanProviderNetworkingServiceNetwork `json:"serviceNetwork,omitempty"`
	PodNetwork     DigitalOceanProviderNetworkingPodNetwork     `json:"podNetwork,omitempty"`
	DnsForwarder   string                                       `json:"dnsForwarder,omitempty"`
}

type DigitalOceanProvider struct {
	Region           DigitalOceanProviderRegion     `json:"region,omitempty"`
	DigitalOceanTags []ValidTag                     `json:"digitalOceanTags,omitempty"`
	Networking       DigitalOceanProviderNetworking `json:"networking,omitempty"`
	PreInstallScript mk8sCommon.PreInstallScript    `json:"preInstallScript,omitempty"`
	TokenSecretLink  string                         `json:"tokenSecretLink"`
	VpcId            string                         `json:"vpcId"`
	NodePools        []DigitalOceanPool             `json:"nodePools,omitempty"`
	Image            string                         `json:"image"`
	SshKeys          []string                       `json:"sshKeys"`
	ExtraSshKeys     []mk8sCommon.SshPublicKey      `json:"extraSshKeys,omitempty"`
	Autoscaler       mk8sCommon.AutoscalerConfig    `json:"autoscaler,omitempty"`
	ReservedIPs      []string                       `json:"reservedIPs,omitempty"`
}

type DigitalOceanProviderStatus map[string]any

type ValidTag string
