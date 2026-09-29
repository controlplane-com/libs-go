/* auto-generated */

package certificateGenerator

import "github.com/controlplane-com/libs-go/pkg/schema/base"

type AcmeChallenge struct {
	Http01 Http01Challenge `json:"http01,omitempty"`
	Dns01  Dns01Challenge  `json:"dns01,omitempty"`
}

type AcmeGeneratorExternalAccountBinding struct {
	KeyId      string `json:"keyId"`
	SecretLink string `json:"secretLink"`
}

type AcmeGeneratorChallenge struct {
	Http01 Http01Challenge `json:"http01,omitempty"`
	Dns01  Dns01Challenge  `json:"dns01,omitempty"`
}

type AcmeGenerator struct {
	Server                 string                              `json:"server"`
	Email                  string                              `json:"email,omitempty"`
	ExternalAccountBinding AcmeGeneratorExternalAccountBinding `json:"externalAccountBinding,omitempty"`
	Challenge              AcmeGeneratorChallenge              `json:"challenge"`
}

type AzureDnsDns01 struct {
	ResourceGroupName string `json:"resourceGroupName"`
	HostedZoneName    string `json:"hostedZoneName,omitempty"`
	SecretLink        string `json:"secretLink"`
}

type CertificateGeneratorTags map[string]any

type CertificateGeneratorSpecProvider string

const (
	CertificateGeneratorSpecProviderAcme CertificateGeneratorSpecProvider = "acme"
)

type CertificateGeneratorSpecAcmeExternalAccountBinding struct {
	KeyId      string `json:"keyId"`
	SecretLink string `json:"secretLink"`
}

type CertificateGeneratorSpecAcmeChallenge struct {
	Http01 Http01Challenge `json:"http01,omitempty"`
	Dns01  Dns01Challenge  `json:"dns01,omitempty"`
}

type CertificateGeneratorSpecAcme struct {
	Server                 string                                             `json:"server"`
	Email                  string                                             `json:"email,omitempty"`
	ExternalAccountBinding CertificateGeneratorSpecAcmeExternalAccountBinding `json:"externalAccountBinding,omitempty"`
	Challenge              CertificateGeneratorSpecAcmeChallenge              `json:"challenge"`
}

type CertificateGeneratorSpec struct {
	Provider CertificateGeneratorSpecProvider `json:"provider,omitempty"`
	Acme     CertificateGeneratorSpecAcme     `json:"acme,omitempty"`
}

type CertificateGenerator struct {
	Id           string                     `json:"id,omitempty"`
	Name         base.Name                  `json:"name,omitempty"`
	Kind         base.Kind                  `json:"kind,omitempty"`
	Version      *float32                   `json:"version,omitempty"`
	Description  string                     `json:"description,omitempty"`
	Tags         CertificateGeneratorTags   `json:"tags,omitempty"`
	Created      string                     `json:"created,omitempty"`
	LastModified string                     `json:"lastModified,omitempty"`
	Links        base.Links                 `json:"links,omitempty"`
	Spec         CertificateGeneratorSpec   `json:"spec"`
	Status       CertificateGeneratorStatus `json:"status,omitempty"`
}

type CertificateGeneratorProvider string

const (
	CertificateGeneratorProviderAcme CertificateGeneratorProvider = "acme"
)

type CertificateGeneratorStatus struct {
	Ready       bool   `json:"ready,omitempty"`
	Message     string `json:"message,omitempty"`
	LastUpdated string `json:"lastUpdated,omitempty"`
}

type CloudDnsDns01 struct {
	Project    string `json:"project"`
	SecretLink string `json:"secretLink"`
}

type CloudflareDns01 struct {
	SecretLink string `json:"secretLink"`
}

type Dns01ChallengeProvider string

const (
	Dns01ChallengeProviderRoute53    Dns01ChallengeProvider = "route53"
	Dns01ChallengeProviderCloudDns   Dns01ChallengeProvider = "cloudDns"
	Dns01ChallengeProviderAzureDns   Dns01ChallengeProvider = "azureDns"
	Dns01ChallengeProviderCloudflare Dns01ChallengeProvider = "cloudflare"
)

type Dns01ChallengeRoute53 struct {
	Region       string `json:"region"`
	HostedZoneId string `json:"hostedZoneId,omitempty"`
	SecretLink   string `json:"secretLink"`
}

type Dns01ChallengeCloudDns struct {
	Project    string `json:"project"`
	SecretLink string `json:"secretLink"`
}

type Dns01ChallengeAzureDns struct {
	ResourceGroupName string `json:"resourceGroupName"`
	HostedZoneName    string `json:"hostedZoneName,omitempty"`
	SecretLink        string `json:"secretLink"`
}

type Dns01ChallengeCloudflare struct {
	SecretLink string `json:"secretLink"`
}

type Dns01Challenge struct {
	Provider   Dns01ChallengeProvider   `json:"provider,omitempty"`
	Route53    Dns01ChallengeRoute53    `json:"route53,omitempty"`
	CloudDns   Dns01ChallengeCloudDns   `json:"cloudDns,omitempty"`
	AzureDns   Dns01ChallengeAzureDns   `json:"azureDns,omitempty"`
	Cloudflare Dns01ChallengeCloudflare `json:"cloudflare,omitempty"`
}

type Dns01Provider string

const (
	Dns01ProviderRoute53    Dns01Provider = "route53"
	Dns01ProviderCloudDns   Dns01Provider = "cloudDns"
	Dns01ProviderAzureDns   Dns01Provider = "azureDns"
	Dns01ProviderCloudflare Dns01Provider = "cloudflare"
)

type ExternalAccountBinding struct {
	KeyId      string `json:"keyId"`
	SecretLink string `json:"secretLink"`
}

type Http01Challenge struct {
}

type Route53Dns01 struct {
	Region       string `json:"region"`
	HostedZoneId string `json:"hostedZoneId,omitempty"`
	SecretLink   string `json:"secretLink"`
}
