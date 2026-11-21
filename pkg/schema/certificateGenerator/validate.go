package certificateGenerator

// validate.go attaches validation to the canonical certificateGenerator types via
// the Validator interface, plus the three pure helpers certificateGenerator.ts
// exports for its callers (challenge type, required secret type per secretLink, and
// the spec parts that fix the ACME account). The helpers live here rather than in
// the data-service so the actuator side can share them, as TS does.
//
// TS reference: nodelibs/schema/src/certificateGenerator.ts.

import (
	"github.com/controlplane-com/libs-go/pkg/schema/base"
	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// CertChallengeKind is the ACME challenge type a generator solves with.
type CertChallengeKind string

const (
	CertChallengeKindHttp01 CertChallengeKind = "http01"
	CertChallengeKindDns01  CertChallengeKind = "dns01"
)

// ValidationRule reflects CertificateGenerator (Base.keys({spec: required, status})).
func (CertificateGenerator) ValidationRule() *validate.Rule {
	return validate.Reflect(CertificateGenerator{})
}

// ValidationRule layers `acme.when('provider', {is:'acme', then:required, otherwise:
// forbidden})`. `provider` is declared first, so Joi's when-driven key reordering is a
// no-op here.
func (CertificateGeneratorSpec) ValidationRule() *validate.Rule {
	return validate.Reflect(CertificateGeneratorSpec{}).
		Fork("acme", onlyForProvider(string(CertificateGeneratorProviderAcme)))
}

// ValidationRule layers the `onlyFor(provider, …)` when on each dns01 provider block:
// the block matching `provider` is required and the other three are forbidden.
func (Dns01Challenge) ValidationRule() *validate.Rule {
	r := validate.Reflect(Dns01Challenge{})
	for _, key := range []string{"route53", "cloudDns", "azureDns", "cloudflare"} {
		r = r.Fork(key, onlyForProvider(key))
	}
	return r
}

// onlyForProvider ports certificateGenerator.ts's `onlyFor` helper.
func onlyForProvider(provider string) func(*validate.Rule) *validate.Rule {
	return func(f *validate.Rule) *validate.Rule {
		return f.When("provider", validate.When{
			IsEq:      provider,
			Then:      validate.Any().Required(),
			Otherwise: validate.Any().Forbidden(),
		})
	}
}

// ValidationRule layers .xor('http01','dns01') — exactly one challenge type.
func (AcmeChallenge) ValidationRule() *validate.Rule {
	return validate.Reflect(AcmeChallenge{}).Xor("http01", "dns01")
}

func (SecretLink) ValidationRule() *validate.Rule {
	return base.LocalLinkRule().Custom(base.ToKind("secret"))
}

func (AcmeDirectoryUrl) ValidationRule() *validate.Rule {
	return validate.String().URIScheme("https")
}

// --- shared helpers (certificateGenerator.ts:88-137) ---

// ChallengeType reports which ACME challenge a generator solves with. A spec with no
// dns01 block is http01, matching certificateGeneratorChallengeType.
func ChallengeType(spec *CertificateGeneratorSpec) CertChallengeKind {
	if spec != nil && spec.Acme != nil && spec.Acme.Challenge.Dns01 != nil {
		return CertChallengeKindDns01
	}
	return CertChallengeKindHttp01
}

// AcmeDelegated reports whether the generator's dns01 challenge is delegated to Control
// Plane — the domain CNAMEs _acme-challenge to us and we publish the TXT record. True for
// a dns01 block whose provider is absent (defaults cpln) or 'cpln' (certificateGenerator.ts
// certificateGeneratorAcmeDelegated, master b3c3358ad).
func AcmeDelegated(spec *CertificateGeneratorSpec) bool {
	if spec == nil || spec.Acme == nil || spec.Acme.Challenge.Dns01 == nil {
		return false
	}
	p := spec.Acme.Challenge.Dns01.Provider
	return p == "" || p == Dns01ProviderCpln
}

// SecretRef is one secretLink the spec references and the secret type it must have.
type SecretRef struct {
	Link string
	Type string
}

// SecretLinks ports certificateGeneratorSecretLinks: every secretLink in the spec with
// the secret type the provider requires. The order is the TS object's insertion order
// (externalAccountBinding, then the dns01 provider), which decides which link a caller
// is told about first.
func SecretLinks(spec *CertificateGeneratorSpec) []SecretRef {
	if spec == nil || spec.Acme == nil {
		return nil
	}
	acme := spec.Acme
	var out []SecretRef
	add := func(link SecretLink, typ string) {
		if link != "" {
			out = append(out, SecretRef{Link: string(link), Type: typ})
		}
	}
	if acme.ExternalAccountBinding != nil {
		add(acme.ExternalAccountBinding.SecretLink, "opaque")
	}
	dns01 := acme.Challenge.Dns01
	if dns01 == nil {
		return out
	}
	if dns01.Route53 != nil {
		add(dns01.Route53.SecretLink, "aws")
	}
	if dns01.CloudDns != nil {
		add(dns01.CloudDns.SecretLink, "gcp")
	}
	if dns01.AzureDns != nil {
		add(dns01.AzureDns.SecretLink, "azure-sdk")
	}
	if dns01.Cloudflare != nil {
		add(dns01.Cloudflare.SecretLink, "opaque")
	}
	return out
}

// ImmutableChanged ports certificateGeneratorImmutableChanged: the spec parts that fix
// the ACME account. It returns the field path that changed, or "" when none did.
func ImmutableChanged(before, after *CertificateGeneratorSpec) string {
	if providerOf(before) != providerOf(after) {
		return "spec.provider"
	}
	if ChallengeType(before) != ChallengeType(after) {
		return "spec.acme.challenge"
	}
	if acmeServer(before) != acmeServer(after) {
		return "spec.acme.server"
	}
	return ""
}

func providerOf(spec *CertificateGeneratorSpec) CertificateGeneratorProvider {
	if spec == nil {
		return ""
	}
	return spec.Provider
}

func acmeServer(spec *CertificateGeneratorSpec) AcmeDirectoryUrl {
	if spec == nil || spec.Acme == nil {
		return ""
	}
	return spec.Acme.Server
}
