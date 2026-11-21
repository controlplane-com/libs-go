/* auto-generated */

package mk8sEphemeral

import "github.com/controlplane-com/libs-go/pkg/schema/mk8sCommon"

type EphemeralPoolArch string

const (
	EphemeralPoolArchAny   EphemeralPoolArch = "any"
	EphemeralPoolArchArm64 EphemeralPoolArch = "arm64"
	EphemeralPoolArchAmd64 EphemeralPoolArch = "amd64"
)

type EphemeralPoolFlavor string

const (
	EphemeralPoolFlavorDebian EphemeralPoolFlavor = "debian"
	EphemeralPoolFlavorUbuntu EphemeralPoolFlavor = "ubuntu"
	EphemeralPoolFlavorFedora EphemeralPoolFlavor = "fedora"
)

type EphemeralPool struct {
	Name    string              `json:"name"`
	Labels  mk8sCommon.Labels   `json:"labels,omitzero"`
	Taints  mk8sCommon.Taints   `json:"taints,omitzero"`
	MaxPods *mk8sCommon.MaxPods `json:"maxPods,omitzero"`
	Count   float64             `json:"count"`
	Arch    EphemeralPoolArch   `json:"arch,omitzero" validate:"required"`
	Flavor  EphemeralPoolFlavor `json:"flavor,omitzero" validate:"required"`
	Cpu     string              `json:"cpu"`
	Memory  string              `json:"memory"`
}

type EphemeralProviderLocation string

const (
	EphemeralProviderLocationAwsEuCentral1 EphemeralProviderLocation = "aws-eu-central-1"
	EphemeralProviderLocationAwsUsEast2    EphemeralProviderLocation = "aws-us-east-2"
	EphemeralProviderLocationAwsUsWest2    EphemeralProviderLocation = "aws-us-west-2"
	EphemeralProviderLocationGcpMeWest1    EphemeralProviderLocation = "gcp-me-west1"
	EphemeralProviderLocationGcpUsEast1    EphemeralProviderLocation = "gcp-us-east1"
)

type EphemeralProvider struct {
	Location  EphemeralProviderLocation `json:"location,omitzero" validate:"required"`
	NodePools []EphemeralPool           `json:"nodePools,omitzero"`
}

type EphemeralProviderStatus map[string]any
