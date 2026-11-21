/* hand-maintained — canonical source of truth for the `certificategenerator`
 * resource schema. Ported from nodelibs/schema/src/certificateGenerator.ts
 * (master 0c0ece779); the joi-to-go emission it replaces carried one inlined copy
 * of every nested object per parent (Dns01ChallengeRoute53 == Route53Dns01 == …),
 * which the TS file itself names once.
 *
 * `validate:` struct tags carry the rules validate.Reflect can express; the xor /
 * when clauses live in validate.go. See go-libs/schema/ipSet/models.go for the
 * tag DSL.
 */

package certificateGenerator

import "github.com/controlplane-com/libs-go/pkg/schema/base"

// SecretLink ports `LocalLink.custom(toKind('secret'))`: the four dns01 provider
// credentials and the externalAccountBinding HMAC key all point at a secret.
type SecretLink string

// AcmeDirectoryUrl ports `Joi.string().uri({ scheme: ['https'] })` — the CA's ACME
// directory URL. https only, so a plaintext directory cannot be configured.
type AcmeDirectoryUrl string

type Dns01Provider string

const (
	Dns01ProviderCpln       Dns01Provider = "cpln"
	Dns01ProviderRoute53    Dns01Provider = "route53"
	Dns01ProviderCloudDns   Dns01Provider = "cloudDns"
	Dns01ProviderAzureDns   Dns01Provider = "azureDns"
	Dns01ProviderCloudflare Dns01Provider = "cloudflare"
)

func (Dns01Provider) EnumValues() []string {
	return []string{
		string(Dns01ProviderCpln), string(Dns01ProviderRoute53), string(Dns01ProviderCloudDns),
		string(Dns01ProviderAzureDns), string(Dns01ProviderCloudflare),
	}
}

type Route53Dns01 struct {
	Region       string     `json:"region" validate:"required"`
	HostedZoneId string     `json:"hostedZoneId,omitzero" validate:"optional"`
	SecretLink   SecretLink `json:"secretLink" validate:"required"`
}

type CloudDnsDns01 struct {
	Project    string     `json:"project" validate:"required"`
	SecretLink SecretLink `json:"secretLink" validate:"required"`
}

type AzureDnsDns01 struct {
	ResourceGroupName string     `json:"resourceGroupName" validate:"required"`
	HostedZoneName    string     `json:"hostedZoneName,omitzero" validate:"optional"`
	SecretLink        SecretLink `json:"secretLink" validate:"required"`
}

type CloudflareDns01 struct {
	SecretLink SecretLink `json:"secretLink" validate:"required"`
}

// Http01Challenge is `Joi.object({})`: present-but-empty is the whole signal, so any
// key is rejected.
type Http01Challenge struct{}

// Dns01Challenge is the discriminated union over `provider`: the matching provider
// block is required and the other three are forbidden (the .when()s in validate.go).
type Dns01Challenge struct {
	Provider   Dns01Provider    `json:"provider,omitzero" validate:"default=cpln"`
	Route53    *Route53Dns01    `json:"route53,omitzero"`
	CloudDns   *CloudDnsDns01   `json:"cloudDns,omitzero"`
	AzureDns   *AzureDnsDns01   `json:"azureDns,omitzero"`
	Cloudflare *CloudflareDns01 `json:"cloudflare,omitzero"`
}

// AcmeChallenge carries exactly one challenge type (the .xor in validate.go).
type AcmeChallenge struct {
	Http01 *Http01Challenge `json:"http01,omitzero"`
	Dns01  *Dns01Challenge  `json:"dns01,omitzero"`
}

type ExternalAccountBinding struct {
	KeyId      string     `json:"keyId" validate:"required"`
	SecretLink SecretLink `json:"secretLink" validate:"required"`
}

type AcmeGenerator struct {
	Server                 AcmeDirectoryUrl        `json:"server" validate:"required"`
	Email                  string                  `json:"email,omitzero" validate:"email,optional"`
	ExternalAccountBinding *ExternalAccountBinding `json:"externalAccountBinding,omitzero" validate:"optional"`
	Challenge              AcmeChallenge           `json:"challenge" validate:"required"`
}

type CertificateGeneratorProvider string

const (
	CertificateGeneratorProviderAcme CertificateGeneratorProvider = "acme"
)

func (CertificateGeneratorProvider) EnumValues() []string {
	return []string{string(CertificateGeneratorProviderAcme)}
}

// CertificateGeneratorSpec is the discriminated union over `provider`; `acme` is
// required for provider acme and forbidden otherwise (the .when in validate.go).
type CertificateGeneratorSpec struct {
	Provider CertificateGeneratorProvider `json:"provider,omitzero" validate:"required"`
	Acme     *AcmeGenerator               `json:"acme,omitzero"`
}

type CertificateGeneratorStatus struct {
	Ready       *bool  `json:"ready,omitzero"`
	Message     string `json:"message,omitzero"`
	LastUpdated string `json:"lastUpdated,omitzero" validate:"date,iso"`
}

// CertificateGeneratorTags aliases base.Tags so the reflection base-type registry
// resolves it to the shared TagsRule.
type CertificateGeneratorTags = base.Tags

type CertificateGenerator struct {
	Id           string                      `json:"id,omitzero" validate:"uuid,readOnly"`
	Name         base.Name                   `json:"name,omitzero"`
	Kind         base.Kind                   `json:"kind,omitzero"`
	Version      *float64                    `json:"version,omitzero" validate:"positive,integer,readOnly"`
	Description  string                      `json:"description,omitzero" validate:"trim,max=250,allow="`
	Tags         CertificateGeneratorTags    `json:"tags,omitzero" validate:"default={}"`
	Created      string                      `json:"created,omitzero" validate:"date,readOnly"`
	LastModified string                      `json:"lastModified,omitzero" validate:"date,readOnly"`
	Links        base.Links                  `json:"links,omitzero" validate:"readOnly"`
	Spec         CertificateGeneratorSpec    `json:"spec" validate:"required"`
	Status       *CertificateGeneratorStatus `json:"status,omitzero"`
}
