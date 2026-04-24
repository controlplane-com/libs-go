/* auto-generated */

package mk8s

import "github.com/controlplane-com/libs-go/pkg/schema/base"
import "github.com/controlplane-com/libs-go/pkg/schema/mk8sGeneric"
import "github.com/controlplane-com/libs-go/pkg/schema/mk8sHetzner"
import "github.com/controlplane-com/libs-go/pkg/schema/mk8sAws"
import "github.com/controlplane-com/libs-go/pkg/schema/mk8sLinode"
import "github.com/controlplane-com/libs-go/pkg/schema/mk8sOblivus"
import "github.com/controlplane-com/libs-go/pkg/schema/mk8sLambdalabs"
import "github.com/controlplane-com/libs-go/pkg/schema/mk8sPaperspace"
import "github.com/controlplane-com/libs-go/pkg/schema/mk8sEphemeral"
import "github.com/controlplane-com/libs-go/pkg/schema/mk8sTriton"
import "github.com/controlplane-com/libs-go/pkg/schema/mk8sAzure"
import "github.com/controlplane-com/libs-go/pkg/schema/mk8sDigitalOcean"
import "github.com/controlplane-com/libs-go/pkg/schema/mk8sGcp"
import "github.com/controlplane-com/libs-go/pkg/schema/mk8sAddons"

type Mk8SClusterTags map[string]any

type Mk8SClusterSpecVersion string

const (
	Mk8SClusterSpecVersion1260  Mk8SClusterSpecVersion = "1.26.0"
	Mk8SClusterSpecVersion1264  Mk8SClusterSpecVersion = "1.26.4"
	Mk8SClusterSpecVersion1273  Mk8SClusterSpecVersion = "1.27.3"
	Mk8SClusterSpecVersion1282  Mk8SClusterSpecVersion = "1.28.2"
	Mk8SClusterSpecVersion1284  Mk8SClusterSpecVersion = "1.28.4"
	Mk8SClusterSpecVersion1297  Mk8SClusterSpecVersion = "1.29.7"
	Mk8SClusterSpecVersion1303  Mk8SClusterSpecVersion = "1.30.3"
	Mk8SClusterSpecVersion1315  Mk8SClusterSpecVersion = "1.31.5"
	Mk8SClusterSpecVersion13113 Mk8SClusterSpecVersion = "1.31.13"
	Mk8SClusterSpecVersion1321  Mk8SClusterSpecVersion = "1.32.1"
	Mk8SClusterSpecVersion1329  Mk8SClusterSpecVersion = "1.32.9"
	Mk8SClusterSpecVersion1335  Mk8SClusterSpecVersion = "1.33.5"
	Mk8SClusterSpecVersion1342  Mk8SClusterSpecVersion = "1.34.2"
	Mk8SClusterSpecVersion1353  Mk8SClusterSpecVersion = "1.35.3"
)

type Mk8SClusterSpecFirewall struct {
	SourceCIDR  string `json:"sourceCIDR"`
	Description string `json:"description,omitempty"`
}

type Mk8SClusterSpecProvider struct {
	Generic      mk8sGeneric.GenericProvider           `json:"generic,omitempty"`
	Hetzner      mk8sHetzner.HetznerProvider           `json:"hetzner,omitempty"`
	Aws          mk8sAws.AwsProvider                   `json:"aws,omitempty"`
	Linode       mk8sLinode.LinodeProvider             `json:"linode,omitempty"`
	Oblivus      mk8sOblivus.OblivusProvider           `json:"oblivus,omitempty"`
	Lambdalabs   mk8sLambdalabs.LambdalabsProvider     `json:"lambdalabs,omitempty"`
	Paperspace   mk8sPaperspace.PaperspaceProvider     `json:"paperspace,omitempty"`
	Ephemeral    mk8sEphemeral.EphemeralProvider       `json:"ephemeral,omitempty"`
	Triton       mk8sTriton.TritonProvider             `json:"triton,omitempty"`
	Azure        mk8sAzure.AzureProvider               `json:"azure,omitempty"`
	Digitalocean mk8sDigitalOcean.DigitalOceanProvider `json:"digitalocean,omitempty"`
	Gcp          mk8sGcp.GcpProvider                   `json:"gcp,omitempty"`
}

type Mk8SClusterSpecAddOns struct {
	Dashboard             mk8sAddons.NonCustomizableAddonConfig `json:"dashboard,omitempty"`
	Headlamp              mk8sAddons.NonCustomizableAddonConfig `json:"headlamp,omitempty"`
	AzureWorkloadIdentity mk8sAddons.AzureAddonConfig           `json:"azureWorkloadIdentity,omitempty"`
	AwsWorkloadIdentity   mk8sAddons.NonCustomizableAddonConfig `json:"awsWorkloadIdentity,omitempty"`
	LocalPathStorage      mk8sAddons.NonCustomizableAddonConfig `json:"localPathStorage,omitempty"`
	Metrics               mk8sAddons.MetricsAddonConfig         `json:"metrics,omitempty"`
	Logs                  mk8sAddons.LogsAddonConfig            `json:"logs,omitempty"`
	RegistryMirror        mk8sAddons.RegistryMirrorConfig       `json:"registryMirror,omitempty"`
	Nvidia                mk8sAddons.NvidiaAddonConfig          `json:"nvidia,omitempty"`
	AwsEFS                mk8sAddons.AwsEFSAddonConfig          `json:"awsEFS,omitempty"`
	AwsECR                mk8sAddons.AwsECRAddonConfig          `json:"awsECR,omitempty"`
	AwsELB                mk8sAddons.AwsELBAddonConfig          `json:"awsELB,omitempty"`
	AzureACR              mk8sAddons.AzureACRAddonConfig        `json:"azureACR,omitempty"`
	Sysbox                mk8sAddons.NonCustomizableAddonConfig `json:"sysbox,omitempty"`
	Byok                  mk8sAddons.ByokAddonConfig            `json:"byok,omitempty"`
}

type Mk8SClusterSpec struct {
	Version  Mk8SClusterSpecVersion    `json:"version,omitempty"`
	Firewall []Mk8SClusterSpecFirewall `json:"firewall,omitempty"`
	Provider Mk8SClusterSpecProvider   `json:"provider"`
	AddOns   Mk8SClusterSpecAddOns     `json:"addOns,omitempty"`
}

type Mk8sCluster struct {
	Id           string          `json:"id,omitempty"`
	Name         base.Name       `json:"name,omitempty"`
	Kind         base.Kind       `json:"kind,omitempty"`
	Version      *float32        `json:"version,omitempty"`
	Description  string          `json:"description,omitempty"`
	Tags         Mk8SClusterTags `json:"tags,omitempty"`
	Created      string          `json:"created,omitempty"`
	LastModified string          `json:"lastModified,omitempty"`
	Links        base.Links      `json:"links,omitempty"`
	Spec         Mk8SClusterSpec `json:"spec"`
	Alias        string          `json:"alias,omitempty"`
	Status       Mk8sStatus      `json:"status,omitempty"`
}

type Mk8SSpecVersion string

const (
	Mk8SSpecVersion1260  Mk8SSpecVersion = "1.26.0"
	Mk8SSpecVersion1264  Mk8SSpecVersion = "1.26.4"
	Mk8SSpecVersion1273  Mk8SSpecVersion = "1.27.3"
	Mk8SSpecVersion1282  Mk8SSpecVersion = "1.28.2"
	Mk8SSpecVersion1284  Mk8SSpecVersion = "1.28.4"
	Mk8SSpecVersion1297  Mk8SSpecVersion = "1.29.7"
	Mk8SSpecVersion1303  Mk8SSpecVersion = "1.30.3"
	Mk8SSpecVersion1315  Mk8SSpecVersion = "1.31.5"
	Mk8SSpecVersion13113 Mk8SSpecVersion = "1.31.13"
	Mk8SSpecVersion1321  Mk8SSpecVersion = "1.32.1"
	Mk8SSpecVersion1329  Mk8SSpecVersion = "1.32.9"
	Mk8SSpecVersion1335  Mk8SSpecVersion = "1.33.5"
	Mk8SSpecVersion1342  Mk8SSpecVersion = "1.34.2"
	Mk8SSpecVersion1353  Mk8SSpecVersion = "1.35.3"
)

type Mk8SSpecFirewall struct {
	SourceCIDR  string `json:"sourceCIDR"`
	Description string `json:"description,omitempty"`
}

type Mk8SSpecProvider struct {
	Generic      mk8sGeneric.GenericProvider           `json:"generic,omitempty"`
	Hetzner      mk8sHetzner.HetznerProvider           `json:"hetzner,omitempty"`
	Aws          mk8sAws.AwsProvider                   `json:"aws,omitempty"`
	Linode       mk8sLinode.LinodeProvider             `json:"linode,omitempty"`
	Oblivus      mk8sOblivus.OblivusProvider           `json:"oblivus,omitempty"`
	Lambdalabs   mk8sLambdalabs.LambdalabsProvider     `json:"lambdalabs,omitempty"`
	Paperspace   mk8sPaperspace.PaperspaceProvider     `json:"paperspace,omitempty"`
	Ephemeral    mk8sEphemeral.EphemeralProvider       `json:"ephemeral,omitempty"`
	Triton       mk8sTriton.TritonProvider             `json:"triton,omitempty"`
	Azure        mk8sAzure.AzureProvider               `json:"azure,omitempty"`
	Digitalocean mk8sDigitalOcean.DigitalOceanProvider `json:"digitalocean,omitempty"`
	Gcp          mk8sGcp.GcpProvider                   `json:"gcp,omitempty"`
}

type Mk8SSpecAddOns struct {
	Dashboard             mk8sAddons.NonCustomizableAddonConfig `json:"dashboard,omitempty"`
	Headlamp              mk8sAddons.NonCustomizableAddonConfig `json:"headlamp,omitempty"`
	AzureWorkloadIdentity mk8sAddons.AzureAddonConfig           `json:"azureWorkloadIdentity,omitempty"`
	AwsWorkloadIdentity   mk8sAddons.NonCustomizableAddonConfig `json:"awsWorkloadIdentity,omitempty"`
	LocalPathStorage      mk8sAddons.NonCustomizableAddonConfig `json:"localPathStorage,omitempty"`
	Metrics               mk8sAddons.MetricsAddonConfig         `json:"metrics,omitempty"`
	Logs                  mk8sAddons.LogsAddonConfig            `json:"logs,omitempty"`
	RegistryMirror        mk8sAddons.RegistryMirrorConfig       `json:"registryMirror,omitempty"`
	Nvidia                mk8sAddons.NvidiaAddonConfig          `json:"nvidia,omitempty"`
	AwsEFS                mk8sAddons.AwsEFSAddonConfig          `json:"awsEFS,omitempty"`
	AwsECR                mk8sAddons.AwsECRAddonConfig          `json:"awsECR,omitempty"`
	AwsELB                mk8sAddons.AwsELBAddonConfig          `json:"awsELB,omitempty"`
	AzureACR              mk8sAddons.AzureACRAddonConfig        `json:"azureACR,omitempty"`
	Sysbox                mk8sAddons.NonCustomizableAddonConfig `json:"sysbox,omitempty"`
	Byok                  mk8sAddons.ByokAddonConfig            `json:"byok,omitempty"`
}

type Mk8sSpec struct {
	Version  Mk8SSpecVersion    `json:"version,omitempty"`
	Firewall []Mk8SSpecFirewall `json:"firewall,omitempty"`
	Provider Mk8SSpecProvider   `json:"provider"`
	AddOns   Mk8SSpecAddOns     `json:"addOns,omitempty"`
}

type Mk8SStatusAddOns struct {
	Dashboard           mk8sAddons.DashboardAddonStatus           `json:"dashboard,omitempty"`
	Headlamp            mk8sAddons.DashboardAddonStatus           `json:"headlamp,omitempty"`
	AwsWorkloadIdentity mk8sAddons.AwsWorkloadIdentityAddonStatus `json:"awsWorkloadIdentity,omitempty"`
	Metrics             mk8sAddons.MetricsAddonStatus             `json:"metrics,omitempty"`
	Logs                mk8sAddons.LogsAddonStatus                `json:"logs,omitempty"`
	AwsECR              mk8sAddons.AwsTrustPolicyConfig           `json:"awsECR,omitempty"`
	AwsEFS              mk8sAddons.AwsTrustPolicyConfig           `json:"awsEFS,omitempty"`
	AwsELB              mk8sAddons.AwsTrustPolicyConfig           `json:"awsELB,omitempty"`
}

type Mk8sStatus struct {
	OidcProviderUrl string           `json:"oidcProviderUrl,omitempty"`
	ServerUrl       string           `json:"serverUrl,omitempty"`
	HomeLocation    string           `json:"homeLocation,omitempty"`
	AddOns          Mk8SStatusAddOns `json:"addOns,omitempty"`
}
