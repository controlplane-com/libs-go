/* auto-generated */

package mk8sLinode

import "github.com/controlplane-com/libs-go/pkg/schema/mk8sCommon"

type LinodeJoinParams struct {
	IpAddress    string `json:"ipAddress,omitzero"`
	NodePoolName string `json:"nodePoolName"`

	/* WARNING!! Arbitrary properties are being ignored! */
}

type LinodePool struct {
	Name          string              `json:"name"`
	Labels        mk8sCommon.Labels   `json:"labels,omitzero"`
	Taints        mk8sCommon.Taints   `json:"taints,omitzero"`
	MaxPods       *mk8sCommon.MaxPods `json:"maxPods,omitzero"`
	ServerType    string              `json:"serverType"`
	OverrideImage string              `json:"overrideImage,omitzero"`
	SubnetId      string              `json:"subnetId"`
	MinSize       float64             `json:"minSize"`
	MaxSize       float64             `json:"maxSize"`
}

type LinodeProviderNetworkingServiceNetwork string

const (
	LinodeProviderNetworkingServiceNetwork10430016   LinodeProviderNetworkingServiceNetwork = "10.43.0.0/16"
	LinodeProviderNetworkingServiceNetwork1921680016 LinodeProviderNetworkingServiceNetwork = "192.168.0.0/16"
)

type LinodeProviderNetworkingPodNetwork string

const (
	LinodeProviderNetworkingPodNetwork10420016  LinodeProviderNetworkingPodNetwork = "10.42.0.0/16"
	LinodeProviderNetworkingPodNetwork172160015 LinodeProviderNetworkingPodNetwork = "172.16.0.0/15"
	LinodeProviderNetworkingPodNetwork172180015 LinodeProviderNetworkingPodNetwork = "172.18.0.0/15"
	LinodeProviderNetworkingPodNetwork172200015 LinodeProviderNetworkingPodNetwork = "172.20.0.0/15"
	LinodeProviderNetworkingPodNetwork172220015 LinodeProviderNetworkingPodNetwork = "172.22.0.0/15"
	LinodeProviderNetworkingPodNetwork172240015 LinodeProviderNetworkingPodNetwork = "172.24.0.0/15"
	LinodeProviderNetworkingPodNetwork172260015 LinodeProviderNetworkingPodNetwork = "172.26.0.0/15"
	LinodeProviderNetworkingPodNetwork172280015 LinodeProviderNetworkingPodNetwork = "172.28.0.0/15"
	LinodeProviderNetworkingPodNetwork172300015 LinodeProviderNetworkingPodNetwork = "172.30.0.0/15"
)

type LinodeProviderNetworking struct {
	ServiceNetwork LinodeProviderNetworkingServiceNetwork `json:"serviceNetwork,omitzero"`
	PodNetwork     LinodeProviderNetworkingPodNetwork     `json:"podNetwork,omitzero"`
	DnsForwarder   string                                 `json:"dnsForwarder,omitzero"`
}

type LinodeProvider struct {
	Region           string                       `json:"region"`
	TokenSecretLink  string                       `json:"tokenSecretLink"`
	FirewallId       string                       `json:"firewallId,omitzero"`
	NodePools        []LinodePool                 `json:"nodePools,omitzero"`
	Image            string                       `json:"image"`
	AuthorizedUsers  []string                     `json:"authorizedUsers,omitzero"`
	AuthorizedKeys   []string                     `json:"authorizedKeys,omitzero"`
	VpcId            string                       `json:"vpcId"`
	PreInstallScript mk8sCommon.PreInstallScript  `json:"preInstallScript,omitzero"`
	Networking       *LinodeProviderNetworking    `json:"networking,omitzero"`
	Autoscaler       *mk8sCommon.AutoscalerConfig `json:"autoscaler,omitzero"`
}

type LinodeProviderStatus map[string]any
