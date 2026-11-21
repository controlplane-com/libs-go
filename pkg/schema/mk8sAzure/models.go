/* auto-generated */

package mk8sAzure

import "github.com/controlplane-com/libs-go/pkg/schema/mk8sCommon"

type AzurePool struct {
	Name          string              `json:"name"`
	Labels        mk8sCommon.Labels   `json:"labels,omitzero"`
	Taints        mk8sCommon.Taints   `json:"taints,omitzero"`
	MaxPods       *mk8sCommon.MaxPods `json:"maxPods,omitzero"`
	Size          string              `json:"size"`
	SubnetId      string              `json:"subnetId"`
	Zones         []float64           `json:"zones"`
	OverrideImage *Image              `json:"overrideImage,omitzero"`
	BootDiskSize  float64             `json:"bootDiskSize"`
	MinSize       float64             `json:"minSize"`
	MaxSize       float64             `json:"maxSize"`
}

type AzureProviderNetworkingServiceNetwork string

const (
	AzureProviderNetworkingServiceNetwork10430016   AzureProviderNetworkingServiceNetwork = "10.43.0.0/16"
	AzureProviderNetworkingServiceNetwork1921680016 AzureProviderNetworkingServiceNetwork = "192.168.0.0/16"
)

type AzureProviderNetworkingPodNetwork string

const (
	AzureProviderNetworkingPodNetwork10420016  AzureProviderNetworkingPodNetwork = "10.42.0.0/16"
	AzureProviderNetworkingPodNetwork172160015 AzureProviderNetworkingPodNetwork = "172.16.0.0/15"
	AzureProviderNetworkingPodNetwork172180015 AzureProviderNetworkingPodNetwork = "172.18.0.0/15"
	AzureProviderNetworkingPodNetwork172200015 AzureProviderNetworkingPodNetwork = "172.20.0.0/15"
	AzureProviderNetworkingPodNetwork172220015 AzureProviderNetworkingPodNetwork = "172.22.0.0/15"
	AzureProviderNetworkingPodNetwork172240015 AzureProviderNetworkingPodNetwork = "172.24.0.0/15"
	AzureProviderNetworkingPodNetwork172260015 AzureProviderNetworkingPodNetwork = "172.26.0.0/15"
	AzureProviderNetworkingPodNetwork172280015 AzureProviderNetworkingPodNetwork = "172.28.0.0/15"
	AzureProviderNetworkingPodNetwork172300015 AzureProviderNetworkingPodNetwork = "172.30.0.0/15"
)

type AzureProviderNetworking struct {
	ServiceNetwork AzureProviderNetworkingServiceNetwork `json:"serviceNetwork,omitzero"`
	PodNetwork     AzureProviderNetworkingPodNetwork     `json:"podNetwork,omitzero"`
	DnsForwarder   string                                `json:"dnsForwarder,omitzero"`
}

type AzureProviderImageRecommended string

const (
	AzureProviderImageRecommendedUbuntuNoble2404  AzureProviderImageRecommended = "ubuntu/noble-24.04"
	AzureProviderImageRecommendedUbuntuJammy2204  AzureProviderImageRecommended = "ubuntu/jammy-22.04"
	AzureProviderImageRecommendedUbuntuFocal2004  AzureProviderImageRecommended = "ubuntu/focal-20.04"
	AzureProviderImageRecommendedDebianBookworm12 AzureProviderImageRecommended = "debian/bookworm-12"
	AzureProviderImageRecommendedDebianBullseye11 AzureProviderImageRecommended = "debian/bullseye-11"
)

type AzureProviderImageReference struct {
	Publisher string `json:"publisher"`
	Offer     string `json:"offer"`
	Sku       string `json:"sku"`
	Version   string `json:"version"`
}

type AzureProviderImage struct {
	Recommended AzureProviderImageRecommended `json:"recommended,omitzero"`
	Reference   *AzureProviderImageReference  `json:"reference,omitzero"`
}

type AzureProviderTags map[string]string

type AzureProvider struct {
	Location         string                       `json:"location"`
	SubscriptionId   string                       `json:"subscriptionId"`
	SdkSecretLink    string                       `json:"sdkSecretLink"`
	ResourceGroup    string                       `json:"resourceGroup"`
	Networking       *AzureProviderNetworking     `json:"networking,omitzero"`
	PreInstallScript string                       `json:"preInstallScript,omitzero"`
	Image            AzureProviderImage           `json:"image"`
	SshKeys          []mk8sCommon.SshPublicKey    `json:"sshKeys"`
	NetworkId        string                       `json:"networkId"`
	Tags             AzureProviderTags            `json:"tags,omitzero"`
	NodePools        []AzurePool                  `json:"nodePools,omitzero"`
	Autoscaler       *mk8sCommon.AutoscalerConfig `json:"autoscaler,omitzero"`
}

type AzureProviderStatus map[string]any

type ImageRecommended string

const (
	ImageRecommendedUbuntuNoble2404  ImageRecommended = "ubuntu/noble-24.04"
	ImageRecommendedUbuntuJammy2204  ImageRecommended = "ubuntu/jammy-22.04"
	ImageRecommendedUbuntuFocal2004  ImageRecommended = "ubuntu/focal-20.04"
	ImageRecommendedDebianBookworm12 ImageRecommended = "debian/bookworm-12"
	ImageRecommendedDebianBullseye11 ImageRecommended = "debian/bullseye-11"
)

type ImageReference struct {
	Publisher string `json:"publisher"`
	Offer     string `json:"offer"`
	Sku       string `json:"sku"`
	Version   string `json:"version"`
}

type Image struct {
	Recommended ImageRecommended `json:"recommended,omitzero"`
	Reference   *ImageReference  `json:"reference,omitzero"`
}
