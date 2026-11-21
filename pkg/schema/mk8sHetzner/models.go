/* auto-generated */

package mk8sHetzner

import "github.com/controlplane-com/libs-go/pkg/schema/mk8sCommon"

type DedicatedServerHetznerPool struct {
	Name    string              `json:"name"`
	Labels  mk8sCommon.Labels   `json:"labels,omitzero"`
	Taints  mk8sCommon.Taints   `json:"taints,omitzero"`
	MaxPods *mk8sCommon.MaxPods `json:"maxPods,omitzero"`
}

type HetznerJoinParams struct {
	IpAddress    string `json:"ipAddress,omitzero"`
	NodePoolName string `json:"nodePoolName"`

	/* WARNING!! Arbitrary properties are being ignored! */
}

type HetznerPool struct {
	Name          string              `json:"name"`
	Labels        mk8sCommon.Labels   `json:"labels,omitzero"`
	Taints        mk8sCommon.Taints   `json:"taints,omitzero"`
	MaxPods       *mk8sCommon.MaxPods `json:"maxPods,omitzero"`
	ServerType    string              `json:"serverType"`
	OverrideImage string              `json:"overrideImage,omitzero"`
	MinSize       float64             `json:"minSize"`
	MaxSize       float64             `json:"maxSize"`
}

type HetznerProviderRegion string

const (
	HetznerProviderRegionFsn1 HetznerProviderRegion = "fsn1"
	HetznerProviderRegionNbg1 HetznerProviderRegion = "nbg1"
	HetznerProviderRegionHel1 HetznerProviderRegion = "hel1"
	HetznerProviderRegionAsh  HetznerProviderRegion = "ash"
	HetznerProviderRegionHil  HetznerProviderRegion = "hil"
)

type HetznerProviderHetznerLabels map[string]string

type HetznerProviderFloatingIpSelector map[string]string

type HetznerProvider struct {
	Region                   HetznerProviderRegion             `json:"region,omitzero" validate:"required"`
	HetznerLabels            HetznerProviderHetznerLabels      `json:"hetznerLabels,omitzero"`
	Networking               *NetworkingConfig                 `json:"networking,omitzero"`
	PreInstallScript         mk8sCommon.PreInstallScript       `json:"preInstallScript,omitzero"`
	TokenSecretLink          string                            `json:"tokenSecretLink"`
	NetworkId                string                            `json:"networkId"`
	FirewallId               string                            `json:"firewallId,omitzero"`
	NodePools                []HetznerPool                     `json:"nodePools,omitzero"`
	DedicatedServerNodePools []DedicatedServerHetznerPool      `json:"dedicatedServerNodePools,omitzero"`
	Image                    string                            `json:"image"`
	SshKey                   string                            `json:"sshKey,omitzero"`
	Autoscaler               *mk8sCommon.AutoscalerConfig      `json:"autoscaler,omitzero"`
	FloatingIPSelector       HetznerProviderFloatingIpSelector `json:"floatingIPSelector,omitzero"`
}

type HetznerProviderStatus map[string]any

type NetworkingConfigServiceNetwork string

const (
	NetworkingConfigServiceNetwork10430016   NetworkingConfigServiceNetwork = "10.43.0.0/16"
	NetworkingConfigServiceNetwork1921680016 NetworkingConfigServiceNetwork = "192.168.0.0/16"
)

type NetworkingConfigPodNetwork string

const (
	NetworkingConfigPodNetwork10420016  NetworkingConfigPodNetwork = "10.42.0.0/16"
	NetworkingConfigPodNetwork172160015 NetworkingConfigPodNetwork = "172.16.0.0/15"
	NetworkingConfigPodNetwork172180015 NetworkingConfigPodNetwork = "172.18.0.0/15"
	NetworkingConfigPodNetwork172200015 NetworkingConfigPodNetwork = "172.20.0.0/15"
	NetworkingConfigPodNetwork172220015 NetworkingConfigPodNetwork = "172.22.0.0/15"
	NetworkingConfigPodNetwork172240015 NetworkingConfigPodNetwork = "172.24.0.0/15"
	NetworkingConfigPodNetwork172260015 NetworkingConfigPodNetwork = "172.26.0.0/15"
	NetworkingConfigPodNetwork172280015 NetworkingConfigPodNetwork = "172.28.0.0/15"
	NetworkingConfigPodNetwork172300015 NetworkingConfigPodNetwork = "172.30.0.0/15"
)

type NetworkingConfig struct {
	ServiceNetwork NetworkingConfigServiceNetwork `json:"serviceNetwork,omitzero"`
	PodNetwork     NetworkingConfigPodNetwork     `json:"podNetwork,omitzero"`
	DnsForwarder   string                         `json:"dnsForwarder,omitzero"`
}
