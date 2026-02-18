/* auto-generated */

package mk8sGcp

import "github.com/controlplane-com/libs-go/pkg/schema/mk8sCommon"

type GcpPool struct {
	Name                 string            `json:"name"`
	Labels               mk8sCommon.Labels `json:"labels,omitempty"`
	Taints               mk8sCommon.Taints `json:"taints,omitempty"`
	MachineType          string            `json:"machineType"`
	AssignPublicIP       bool              `json:"assignPublicIP,omitempty"`
	Zone                 string            `json:"zone"`
	OverrideImage        Image             `json:"overrideImage,omitempty"`
	BootDiskSize         float32           `json:"bootDiskSize"`
	MinSize              float32           `json:"minSize"`
	MaxSize              float32           `json:"maxSize"`
	Preemptible          bool              `json:"preemptible,omitempty"`
	Subnet               string            `json:"subnet"`
	LocalPersistentDisks float32           `json:"localPersistentDisks"`
}

type GcpProviderLabels map[string]string

type GcpProviderMetadata map[string]string

type GcpProviderNetworkingServiceNetwork string

const (
	GcpProviderNetworkingServiceNetwork10430016   GcpProviderNetworkingServiceNetwork = "10.43.0.0/16"
	GcpProviderNetworkingServiceNetwork1921680016 GcpProviderNetworkingServiceNetwork = "192.168.0.0/16"
)

type GcpProviderNetworkingPodNetwork string

const (
	GcpProviderNetworkingPodNetworkVpc       GcpProviderNetworkingPodNetwork = "vpc"
	GcpProviderNetworkingPodNetwork10420016  GcpProviderNetworkingPodNetwork = "10.42.0.0/16"
	GcpProviderNetworkingPodNetwork172160015 GcpProviderNetworkingPodNetwork = "172.16.0.0/15"
	GcpProviderNetworkingPodNetwork172180015 GcpProviderNetworkingPodNetwork = "172.18.0.0/15"
	GcpProviderNetworkingPodNetwork172200015 GcpProviderNetworkingPodNetwork = "172.20.0.0/15"
	GcpProviderNetworkingPodNetwork172220015 GcpProviderNetworkingPodNetwork = "172.22.0.0/15"
	GcpProviderNetworkingPodNetwork172240015 GcpProviderNetworkingPodNetwork = "172.24.0.0/15"
	GcpProviderNetworkingPodNetwork172260015 GcpProviderNetworkingPodNetwork = "172.26.0.0/15"
	GcpProviderNetworkingPodNetwork172280015 GcpProviderNetworkingPodNetwork = "172.28.0.0/15"
	GcpProviderNetworkingPodNetwork172300015 GcpProviderNetworkingPodNetwork = "172.30.0.0/15"
)

type GcpProviderNetworking struct {
	ServiceNetwork GcpProviderNetworkingServiceNetwork `json:"serviceNetwork,omitempty"`
	PodNetwork     GcpProviderNetworkingPodNetwork     `json:"podNetwork,omitempty"`
	DnsForwarder   string                              `json:"dnsForwarder,omitempty"`
}

type GcpProviderImageRecommended string

const (
	GcpProviderImageRecommendedUbuntuJammy2204  GcpProviderImageRecommended = "ubuntu/jammy-22.04"
	GcpProviderImageRecommendedUbuntuNoble2404  GcpProviderImageRecommended = "ubuntu/noble-24.04"
	GcpProviderImageRecommendedDebianBookworm12 GcpProviderImageRecommended = "debian/bookworm-12"
	GcpProviderImageRecommendedDebianTrixie13   GcpProviderImageRecommended = "debian/trixie-13"
	GcpProviderImageRecommendedGoogleCosStable  GcpProviderImageRecommended = "google/cos-stable"
)

type GcpProviderImageFamily struct {
	Project string `json:"project"`
	Family  string `json:"family"`
}

type GcpProviderImage struct {
	Recommended GcpProviderImageRecommended `json:"recommended,omitempty"`
	Family      GcpProviderImageFamily      `json:"family,omitempty"`
	Exact       string                      `json:"exact,omitempty"`
}

type GcpProvider struct {
	ProjectId        string                      `json:"projectId"`
	Region           string                      `json:"region"`
	Labels           GcpProviderLabels           `json:"labels,omitempty"`
	Tags             []string                    `json:"tags,omitempty"`
	Metadata         GcpProviderMetadata         `json:"metadata,omitempty"`
	Network          string                      `json:"network"`
	SaKeyLink        string                      `json:"saKeyLink"`
	Networking       GcpProviderNetworking       `json:"networking,omitempty"`
	PreInstallScript mk8sCommon.PreInstallScript `json:"preInstallScript,omitempty"`
	Image            GcpProviderImage            `json:"image"`
	NodePools        []GcpPool                   `json:"nodePools,omitempty"`
	Autoscaler       mk8sCommon.AutoscalerConfig `json:"autoscaler,omitempty"`
}

type GcpProviderStatus map[string]any

type ImageRecommended string

const (
	ImageRecommendedUbuntuJammy2204  ImageRecommended = "ubuntu/jammy-22.04"
	ImageRecommendedUbuntuNoble2404  ImageRecommended = "ubuntu/noble-24.04"
	ImageRecommendedDebianBookworm12 ImageRecommended = "debian/bookworm-12"
	ImageRecommendedDebianTrixie13   ImageRecommended = "debian/trixie-13"
	ImageRecommendedGoogleCosStable  ImageRecommended = "google/cos-stable"
)

type ImageFamily struct {
	Project string `json:"project"`
	Family  string `json:"family"`
}

type Image struct {
	Recommended ImageRecommended `json:"recommended,omitempty"`
	Family      ImageFamily      `json:"family,omitempty"`
	Exact       string           `json:"exact,omitempty"`
}
