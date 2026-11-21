/* auto-generated */

package mk8sGcp

import "github.com/controlplane-com/libs-go/pkg/schema/mk8sCommon"

type GcpPool struct {
	Name                 string              `json:"name"`
	Labels               mk8sCommon.Labels   `json:"labels,omitzero"`
	Taints               mk8sCommon.Taints   `json:"taints,omitzero"`
	MaxPods              *mk8sCommon.MaxPods `json:"maxPods,omitzero"`
	MachineType          string              `json:"machineType"`
	AssignPublicIP       *bool               `json:"assignPublicIP,omitzero"`
	Zone                 string              `json:"zone"`
	OverrideImage        *Image              `json:"overrideImage,omitzero"`
	BootDiskSize         float64             `json:"bootDiskSize"`
	MinSize              float64             `json:"minSize"`
	MaxSize              float64             `json:"maxSize"`
	Preemptible          *bool               `json:"preemptible,omitzero"`
	Subnet               string              `json:"subnet"`
	LocalPersistentDisks *float64            `json:"localPersistentDisks,omitzero"`
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
	ServiceNetwork GcpProviderNetworkingServiceNetwork `json:"serviceNetwork,omitzero"`
	PodNetwork     GcpProviderNetworkingPodNetwork     `json:"podNetwork,omitzero"`
	DnsForwarder   string                              `json:"dnsForwarder,omitzero"`
}

type GcpProviderImageRecommended string

const (
	GcpProviderImageRecommendedUbuntuJammy2204  GcpProviderImageRecommended = "ubuntu/jammy-22.04"
	GcpProviderImageRecommendedUbuntuNoble2404  GcpProviderImageRecommended = "ubuntu/noble-24.04"
	GcpProviderImageRecommendedDebianBookworm12 GcpProviderImageRecommended = "debian/bookworm-12"
	GcpProviderImageRecommendedDebianTrixie13   GcpProviderImageRecommended = "debian/trixie-13"
)

type GcpProviderImageFamily struct {
	Project string `json:"project"`
	Family  string `json:"family"`
}

type GcpProviderImage struct {
	Recommended GcpProviderImageRecommended `json:"recommended,omitzero"`
	Family      *GcpProviderImageFamily     `json:"family,omitzero"`
	Exact       string                      `json:"exact,omitzero"`
}

type GcpProvider struct {
	ProjectId        string                       `json:"projectId"`
	Region           string                       `json:"region"`
	Labels           GcpProviderLabels            `json:"labels,omitzero"`
	Tags             []string                     `json:"tags,omitzero"`
	Metadata         GcpProviderMetadata          `json:"metadata,omitzero"`
	Network          string                       `json:"network"`
	SaKeyLink        string                       `json:"saKeyLink"`
	Networking       *GcpProviderNetworking       `json:"networking,omitzero"`
	PreInstallScript mk8sCommon.PreInstallScript  `json:"preInstallScript,omitzero"`
	Image            GcpProviderImage             `json:"image"`
	NodePools        []GcpPool                    `json:"nodePools,omitzero"`
	Autoscaler       *mk8sCommon.AutoscalerConfig `json:"autoscaler,omitzero"`
}

type GcpProviderStatus map[string]any

type ImageRecommended string

const (
	ImageRecommendedUbuntuJammy2204  ImageRecommended = "ubuntu/jammy-22.04"
	ImageRecommendedUbuntuNoble2404  ImageRecommended = "ubuntu/noble-24.04"
	ImageRecommendedDebianBookworm12 ImageRecommended = "debian/bookworm-12"
	ImageRecommendedDebianTrixie13   ImageRecommended = "debian/trixie-13"
)

type ImageFamily struct {
	Project string `json:"project"`
	Family  string `json:"family"`
}

type Image struct {
	Recommended ImageRecommended `json:"recommended,omitzero"`
	Family      *ImageFamily     `json:"family,omitzero"`
	Exact       string           `json:"exact,omitzero"`
}
