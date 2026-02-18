/* auto-generated */

package mk8sGeneric

import "github.com/controlplane-com/libs-go/pkg/schema/mk8sCommon"

type GenericJoinParams struct {
	NodePoolName string `json:"nodePoolName"`

	/* WARNING!! Arbitrary properties are being ignored! */
}

type GenericPool struct {
	Name   string            `json:"name"`
	Labels mk8sCommon.Labels `json:"labels,omitempty"`
	Taints mk8sCommon.Taints `json:"taints,omitempty"`
}

type GenericProviderLocation string

const (
	GenericProviderLocationAwsEuCentral1 GenericProviderLocation = "aws-eu-central-1"
	GenericProviderLocationAwsUsEast2    GenericProviderLocation = "aws-us-east-2"
	GenericProviderLocationAwsUsWest2    GenericProviderLocation = "aws-us-west-2"
	GenericProviderLocationGcpMeWest1    GenericProviderLocation = "gcp-me-west1"
	GenericProviderLocationGcpUsEast1    GenericProviderLocation = "gcp-us-east1"
)

type GenericProviderNetworkingServiceNetwork string

const (
	GenericProviderNetworkingServiceNetwork10430016   GenericProviderNetworkingServiceNetwork = "10.43.0.0/16"
	GenericProviderNetworkingServiceNetwork1921680016 GenericProviderNetworkingServiceNetwork = "192.168.0.0/16"
)

type GenericProviderNetworkingPodNetwork string

const (
	GenericProviderNetworkingPodNetwork10420016  GenericProviderNetworkingPodNetwork = "10.42.0.0/16"
	GenericProviderNetworkingPodNetwork172160015 GenericProviderNetworkingPodNetwork = "172.16.0.0/15"
	GenericProviderNetworkingPodNetwork172180015 GenericProviderNetworkingPodNetwork = "172.18.0.0/15"
	GenericProviderNetworkingPodNetwork172200015 GenericProviderNetworkingPodNetwork = "172.20.0.0/15"
	GenericProviderNetworkingPodNetwork172220015 GenericProviderNetworkingPodNetwork = "172.22.0.0/15"
	GenericProviderNetworkingPodNetwork172240015 GenericProviderNetworkingPodNetwork = "172.24.0.0/15"
	GenericProviderNetworkingPodNetwork172260015 GenericProviderNetworkingPodNetwork = "172.26.0.0/15"
	GenericProviderNetworkingPodNetwork172280015 GenericProviderNetworkingPodNetwork = "172.28.0.0/15"
	GenericProviderNetworkingPodNetwork172300015 GenericProviderNetworkingPodNetwork = "172.30.0.0/15"
)

type GenericProviderNetworking struct {
	ServiceNetwork GenericProviderNetworkingServiceNetwork `json:"serviceNetwork,omitempty"`
	PodNetwork     GenericProviderNetworkingPodNetwork     `json:"podNetwork,omitempty"`
	DnsForwarder   string                                  `json:"dnsForwarder,omitempty"`
}

type GenericProvider struct {
	Location   GenericProviderLocation   `json:"location,omitempty"`
	Networking GenericProviderNetworking `json:"networking,omitempty"`
	NodePools  []GenericPool             `json:"nodePools,omitempty"`
}

type GenericProviderStatus map[string]any
