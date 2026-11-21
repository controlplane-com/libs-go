/* hand-maintained — canonical source of truth for the `mk8s` resource schema.
 * Formerly emitted by joi-to-go from nodelibs/schema/src/mk8s.ts; now the CANONICAL
 * definition carrying `validate:` struct tags that pkg/validate.FromType reflects into
 * the *validate.Rule tree (replacing pkg/handlers/schema_mk8s.go's Mk8sSchema/
 * Mk8sSpecRule/mk8sProviderRule/firewallRuleRule builders).
 *
 * spec.provider is provider-DISCRIMINATED: exactly one of the twelve provider keys
 * (the Joi .xor). That object-level relation is registered in pkg/handlers
 * (schema_reflect_mk8s.go) via RegisterObjectRules(Mk8SClusterSpecProvider, {Xor}).
 * The twelve provider BODIES, spec.addOns and status are irreducible sub-trees
 * (per-provider complex defaults, pattern-map key/value customs, float bounds — DSL
 * gaps documented in schema_reflect_mk8s.go) so they resolve to the prebuilt
 * pkg/handlers rules via the base-type registry, exactly like base.Name→NameRule.
 * The allow-all firewall default (a content-ful array default a field tag cannot
 * carry) is injected by a Refine on Mk8SClusterSpec. See go-libs/schema/ipSet for
 * the tag DSL grammar.
 */

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

type Mk8SClusterTags = base.Tags

// FirewallSourceCIDR is the strongly-typed replacement for the firewall
// sourceCIDR `custom=mk8sIpV4V6` tag: Joi.string().ip({version:[ipv4,ipv6],
// cidr:optional}).
type FirewallSourceCIDR string

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
	Mk8SClusterSpecVersion1359  Mk8SClusterSpecVersion = "1.35.9"
	Mk8SClusterSpecVersion1365  Mk8SClusterSpecVersion = "1.36.5"
)

// EnumValues resolves the version field to Joi.valid(...upgradePath) via the
// reflection Valuer registry (mk8s.ts upgradePath; keep ascending order).
func (Mk8SClusterSpecVersion) EnumValues() []string {
	return []string{
		"1.26.0", "1.26.4", "1.27.3", "1.28.2", "1.28.4", "1.29.7", "1.30.3",
		"1.31.5", "1.31.13", "1.32.1", "1.32.9", "1.33.5", "1.34.2", "1.35.3", "1.35.9", "1.36.5",
	}
}

type Mk8SClusterSpecFirewall struct {
	SourceCIDR  FirewallSourceCIDR `json:"sourceCIDR" validate:"required"`
	Description string             `json:"description,omitzero" validate:"allow="`
}

type Mk8SClusterSpecProvider struct {
	Generic      *mk8sGeneric.GenericProvider           `json:"generic,omitzero"`
	Hetzner      *mk8sHetzner.HetznerProvider           `json:"hetzner,omitzero"`
	Aws          *mk8sAws.AwsProvider                   `json:"aws,omitzero"`
	Linode       *mk8sLinode.LinodeProvider             `json:"linode,omitzero"`
	Oblivus      *mk8sOblivus.OblivusProvider           `json:"oblivus,omitzero"`
	Lambdalabs   *mk8sLambdalabs.LambdalabsProvider     `json:"lambdalabs,omitzero"`
	Paperspace   *mk8sPaperspace.PaperspaceProvider     `json:"paperspace,omitzero"`
	Ephemeral    *mk8sEphemeral.EphemeralProvider       `json:"ephemeral,omitzero"`
	Triton       *mk8sTriton.TritonProvider             `json:"triton,omitzero"`
	Azure        *mk8sAzure.AzureProvider               `json:"azure,omitzero"`
	Digitalocean *mk8sDigitalOcean.DigitalOceanProvider `json:"digitalocean,omitzero"`
	Gcp          *mk8sGcp.GcpProvider                   `json:"gcp,omitzero"`
}

type Mk8SClusterSpecAddOns struct {
	Dashboard             *mk8sAddons.NonCustomizableAddonConfig `json:"dashboard,omitzero"`
	Headlamp              *mk8sAddons.NonCustomizableAddonConfig `json:"headlamp,omitzero"`
	AzureWorkloadIdentity *mk8sAddons.AzureAddonConfig           `json:"azureWorkloadIdentity,omitzero"`
	AwsWorkloadIdentity   *mk8sAddons.NonCustomizableAddonConfig `json:"awsWorkloadIdentity,omitzero"`
	LocalPathStorage      *mk8sAddons.NonCustomizableAddonConfig `json:"localPathStorage,omitzero"`
	Metrics               *mk8sAddons.MetricsAddonConfig         `json:"metrics,omitzero"`
	Logs                  *mk8sAddons.LogsAddonConfig            `json:"logs,omitzero"`
	RegistryMirror        *mk8sAddons.RegistryMirrorConfig       `json:"registryMirror,omitzero"`
	Nvidia                *mk8sAddons.NvidiaAddonConfig          `json:"nvidia,omitzero"`
	AwsEFS                *mk8sAddons.AwsEFSAddonConfig          `json:"awsEFS,omitzero"`
	AwsECR                *mk8sAddons.AwsECRAddonConfig          `json:"awsECR,omitzero"`
	AwsELB                *mk8sAddons.AwsELBAddonConfig          `json:"awsELB,omitzero"`
	AzureACR              *mk8sAddons.AzureACRAddonConfig        `json:"azureACR,omitzero"`
	Sysbox                *mk8sAddons.NonCustomizableAddonConfig `json:"sysbox,omitzero"`
	Byok                  *mk8sAddons.ByokAddonConfig            `json:"byok,omitzero"`
	Kubevirt              *mk8sAddons.KubevirtAddonConfig        `json:"kubevirt,omitzero"`
	NodeLocalDns          *mk8sAddons.NodeLocalDnsAddonConfig    `json:"nodeLocalDns,omitzero"`
	Kata                  *mk8sAddons.NonCustomizableAddonConfig `json:"kata,omitzero"`
}

type Mk8SClusterSpec struct {
	Version  Mk8SClusterSpecVersion    `json:"version,omitzero" validate:"required"`
	Firewall []Mk8SClusterSpecFirewall `json:"firewall,omitzero"`
	Provider Mk8SClusterSpecProvider   `json:"provider" validate:"required"`
	AddOns   *Mk8SClusterSpecAddOns    `json:"addOns,omitzero"`
}

type Mk8sCluster struct {
	Id           string          `json:"id,omitzero" validate:"uuid,readOnly"`
	Name         base.Name       `json:"name,omitzero"`
	Kind         base.Kind       `json:"kind,omitzero"`
	Version      *float64        `json:"version,omitzero" validate:"positive,integer,readOnly"`
	Description  string          `json:"description,omitzero" validate:"trim,max=250,allow="`
	Tags         Mk8SClusterTags `json:"tags,omitzero" validate:"default={}"`
	Created      string          `json:"created,omitzero" validate:"date,readOnly"`
	LastModified string          `json:"lastModified,omitzero" validate:"date,readOnly"`
	Links        base.Links      `json:"links,omitzero" validate:"readOnly"`
	Spec         Mk8SClusterSpec `json:"spec" validate:"required"`
	Alias        string          `json:"alias,omitzero" validate:"optional,readOnly"`
	Status       *Mk8sStatus     `json:"status,omitzero"`
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
	Mk8SSpecVersion1359  Mk8SSpecVersion = "1.35.9"
	Mk8SSpecVersion1365  Mk8SSpecVersion = "1.36.5"
)

type Mk8SSpecFirewall struct {
	SourceCIDR  string `json:"sourceCIDR"`
	Description string `json:"description,omitzero"`
}

type Mk8SSpecProvider struct {
	Generic      *mk8sGeneric.GenericProvider           `json:"generic,omitzero"`
	Hetzner      *mk8sHetzner.HetznerProvider           `json:"hetzner,omitzero"`
	Aws          *mk8sAws.AwsProvider                   `json:"aws,omitzero"`
	Linode       *mk8sLinode.LinodeProvider             `json:"linode,omitzero"`
	Oblivus      *mk8sOblivus.OblivusProvider           `json:"oblivus,omitzero"`
	Lambdalabs   *mk8sLambdalabs.LambdalabsProvider     `json:"lambdalabs,omitzero"`
	Paperspace   *mk8sPaperspace.PaperspaceProvider     `json:"paperspace,omitzero"`
	Ephemeral    *mk8sEphemeral.EphemeralProvider       `json:"ephemeral,omitzero"`
	Triton       *mk8sTriton.TritonProvider             `json:"triton,omitzero"`
	Azure        *mk8sAzure.AzureProvider               `json:"azure,omitzero"`
	Digitalocean *mk8sDigitalOcean.DigitalOceanProvider `json:"digitalocean,omitzero"`
	Gcp          *mk8sGcp.GcpProvider                   `json:"gcp,omitzero"`
}

type Mk8SSpecAddOns struct {
	Dashboard             *mk8sAddons.NonCustomizableAddonConfig `json:"dashboard,omitzero"`
	Headlamp              *mk8sAddons.NonCustomizableAddonConfig `json:"headlamp,omitzero"`
	AzureWorkloadIdentity *mk8sAddons.AzureAddonConfig           `json:"azureWorkloadIdentity,omitzero"`
	AwsWorkloadIdentity   *mk8sAddons.NonCustomizableAddonConfig `json:"awsWorkloadIdentity,omitzero"`
	LocalPathStorage      *mk8sAddons.NonCustomizableAddonConfig `json:"localPathStorage,omitzero"`
	Metrics               *mk8sAddons.MetricsAddonConfig         `json:"metrics,omitzero"`
	Logs                  *mk8sAddons.LogsAddonConfig            `json:"logs,omitzero"`
	RegistryMirror        *mk8sAddons.RegistryMirrorConfig       `json:"registryMirror,omitzero"`
	Nvidia                *mk8sAddons.NvidiaAddonConfig          `json:"nvidia,omitzero"`
	AwsEFS                *mk8sAddons.AwsEFSAddonConfig          `json:"awsEFS,omitzero"`
	AwsECR                *mk8sAddons.AwsECRAddonConfig          `json:"awsECR,omitzero"`
	AwsELB                *mk8sAddons.AwsELBAddonConfig          `json:"awsELB,omitzero"`
	AzureACR              *mk8sAddons.AzureACRAddonConfig        `json:"azureACR,omitzero"`
	Sysbox                *mk8sAddons.NonCustomizableAddonConfig `json:"sysbox,omitzero"`
	Byok                  *mk8sAddons.ByokAddonConfig            `json:"byok,omitzero"`
	Kubevirt              *mk8sAddons.KubevirtAddonConfig        `json:"kubevirt,omitzero"`
	NodeLocalDns          *mk8sAddons.NodeLocalDnsAddonConfig    `json:"nodeLocalDns,omitzero"`
	Kata                  *mk8sAddons.NonCustomizableAddonConfig `json:"kata,omitzero"`
}

type Mk8sSpec struct {
	Version  Mk8SSpecVersion    `json:"version,omitzero" validate:"required"`
	Firewall []Mk8SSpecFirewall `json:"firewall,omitzero"`
	Provider Mk8SSpecProvider   `json:"provider"`
	AddOns   *Mk8SSpecAddOns    `json:"addOns,omitzero"`
}

type Mk8SStatusAddOns struct {
	Dashboard           *mk8sAddons.DashboardAddonStatus           `json:"dashboard,omitzero"`
	Headlamp            *mk8sAddons.DashboardAddonStatus           `json:"headlamp,omitzero"`
	AwsWorkloadIdentity *mk8sAddons.AwsWorkloadIdentityAddonStatus `json:"awsWorkloadIdentity,omitzero"`
	Metrics             *mk8sAddons.MetricsAddonStatus             `json:"metrics,omitzero"`
	Logs                *mk8sAddons.LogsAddonStatus                `json:"logs,omitzero"`
	AwsECR              *mk8sAddons.AwsTrustPolicyConfig           `json:"awsECR,omitzero"`
	AwsEFS              *mk8sAddons.AwsTrustPolicyConfig           `json:"awsEFS,omitzero"`
	AwsELB              *mk8sAddons.AwsTrustPolicyConfig           `json:"awsELB,omitzero"`
}

type Mk8sStatus struct {
	OidcProviderUrl string            `json:"oidcProviderUrl,omitzero" validate:"uri"`
	ServerUrl       string            `json:"serverUrl,omitzero" validate:"uri"`
	HomeLocation    string            `json:"homeLocation,omitzero"`
	AddOns          *Mk8SStatusAddOns `json:"addOns,omitzero"`
}
