package domain

import (
	"testing"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

const certGenLink = "/org/o/certificategenerator/cg"

// TestCertificateGeneratorLinkExclusions pins the DomainSpec custom's
// certificateGeneratorLink rules (domain.ts:339-347). Verdicts captured from the built
// domain.ts.
func TestCertificateGeneratorLinkExclusions(t *testing.T) {
	for _, c := range []struct {
		name    string
		body    string
		message string // "" = accepted
	}{
		{"link alone", `{"dnsMode":"cname","certificateGeneratorLink":"` + certGenLink + `"}`, ""},
		{"link with dnsMode ns", `{"dnsMode":"ns","certificateGeneratorLink":"` + certGenLink + `"}`, ""},
		{"link with certChallengeType",
			`{"dnsMode":"cname","certChallengeType":"dns01","certificateGeneratorLink":"` + certGenLink + `"}`,
			`"value" failed custom validation because certChallengeType must not be set when certificateGeneratorLink is configured; the generator defines the challenge type.`},
		{"link with serverCertificate",
			`{"dnsMode":"cname","certificateGeneratorLink":"` + certGenLink + `",` +
				`"ports":[{"number":443,"protocol":"http","tls":{"serverCertificate":{"secretLink":"/org/o/secret/s"}}}]}`,
			`"value" failed custom validation because certificateGeneratorLink cannot be combined with tls.serverCertificate.secretLink.`},
	} {
		_, verr := validate.Decode[DomainSpec](validate.Bytes(c.body), &validate.Context{Org: "o"}, validate.Prefs{})
		got := ""
		if verr != nil {
			got = verr.Details()[0].Message
		}
		if got != c.message {
			t.Errorf("%s:\n got  %q\n want %q", c.name, got, c.message)
		}
	}
}

// TestCertificateGeneratorLinkInternalDomain pins the .internal rejection
// (domain.ts:524-526): an internal domain has no ACME challenge to solve.
func TestCertificateGeneratorLinkInternalDomain(t *testing.T) {
	const want = `"value" failed custom validation because certificateGeneratorLink is not supported for .internal domains; provide tls.serverCertificate.secretLink instead.`
	_, verr := validate.Decode[Domain](
		validate.Bytes(`{"name":"svc.internal","spec":{"certificateGeneratorLink":"`+certGenLink+`"}}`),
		&validate.Context{Org: "o"}, validate.Prefs{})
	if verr == nil {
		t.Fatal("an .internal domain with a certificateGeneratorLink must be rejected")
	}
	if got := verr.Details()[0].Message; got != want {
		t.Errorf("message = %q, want %q", got, want)
	}
	// A public domain with the same spec is accepted.
	if _, verr := validate.Decode[Domain](
		validate.Bytes(`{"name":"foo.example.com","spec":{"dnsMode":"cname","certificateGeneratorLink":"`+certGenLink+`"}}`),
		&validate.Context{Org: "o"}, validate.Prefs{}); verr != nil {
		t.Errorf("public domain rejected: %v", verr)
	}
}
