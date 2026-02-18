/* auto-generated */

package mk8sLinode

import "github.com/controlplane-com/libs-go/pkg/schema/mk8sCommon"

type LinodeJoinParams struct {
	IpAddress    string `json:"ipAddress,omitempty"`
	NodePoolName string `json:"nodePoolName"`

	/* WARNING!! Arbitrary properties are being ignored! */
}

type LinodePool struct {
	Name          string            `json:"name"`
	Labels        mk8sCommon.Labels `json:"labels,omitempty"`
	Taints        mk8sCommon.Taints `json:"taints,omitempty"`
	ServerType    string            `json:"serverType"`
	OverrideImage string            `json:"overrideImage,omitempty"`
	SubnetId      string            `json:"subnetId"`
	MinSize       float32           `json:"minSize"`
	MaxSize       float32           `json:"maxSize"`
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
	ServiceNetwork LinodeProviderNetworkingServiceNetwork `json:"serviceNetwork,omitempty"`
	PodNetwork     LinodeProviderNetworkingPodNetwork     `json:"podNetwork,omitempty"`
	DnsForwarder   string                                 `json:"dnsForwarder,omitempty"`
}

type LinodeProvider struct {
	Region           string                      `json:"region"`
	TokenSecretLink  string                      `json:"tokenSecretLink"`
	FirewallId       string                      `json:"firewallId,omitempty"`
	NodePools        []LinodePool                `json:"nodePools,omitempty"`
	Image            string                      `json:"image"`
	AuthorizedUsers  []string                    `json:"authorizedUsers,omitempty"`
	AuthorizedKeys   []string                    `json:"authorizedKeys,omitempty"`
	VpcId            string                      `json:"vpcId"`
	PreInstallScript mk8sCommon.PreInstallScript `json:"preInstallScript,omitempty"`
	Networking       LinodeProviderNetworking    `json:"networking,omitempty"`
	Autoscaler       mk8sCommon.AutoscalerConfig `json:"autoscaler,omitempty"`
}

type LinodeProviderStatus map[string]any
