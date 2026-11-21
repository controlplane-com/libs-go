package certificateGenerator

import (
	"encoding/json"
	"testing"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// TestSpecMatchesJoi pins CertificateGeneratorSpec to verdicts CAPTURED from the built
// certificateGenerator.ts (node -e over build/cjs/src/certificateGenerator.js), with
// the org context that expands the secretLinks. Every case below is a recorded
// (type, message) pair, not a reasoned one.
func TestSpecMatchesJoi(t *testing.T) {
	const r53 = `{"provider":"acme","acme":{"server":"https://x/d","challenge":{"dns01":{"provider":"route53","route53":{"region":"r","secretLink":"//secret/s"}}}}}`
	for _, c := range []struct {
		body    string
		typ     string // "" = accepted
		message string
	}{
		{`{}`, "any.required", `"provider" is required`},
		{`{"provider":"acme"}`, "any.required", `"acme" is required`},
		{`{"provider":"venafi"}`, "any.only", `"provider" must be [acme]`},
		{`{"provider":"acme","acme":{}}`, "any.required", `"acme.server" is required`},
		{`{"acme":{"server":"https://x/d","challenge":{"http01":{}}}}`, "any.required", `"provider" is required`},
		{`{"provider":"acme","acme":{"server":"https://x/d","challenge":{}}}`,
			"object.missing", `"acme.challenge" must contain at least one of [http01, dns01]`},
		{`{"provider":"acme","acme":{"server":"https://x/d","challenge":{"http01":{},"dns01":{"provider":"route53","route53":{"region":"r","secretLink":"//secret/s"}}}}}`,
			"object.xor", `"acme.challenge" contains a conflict between exclusive peers [http01, dns01]`},
		{`{"provider":"acme","acme":{"server":"https://x/d","challenge":{"http01":{}}}}`, "", ""},
		{`{"provider":"acme","acme":{"server":"http://x/d","challenge":{"http01":{}}}}`,
			"string.uriCustomScheme", `"acme.server" must be a valid uri with a scheme matching the https pattern`},
		{`{"provider":"acme","acme":{"server":"notaurl","challenge":{"http01":{}}}}`,
			"string.uriCustomScheme", `"acme.server" must be a valid uri with a scheme matching the https pattern`},
		{`{"provider":"acme","acme":{"server":"https://x/d","challenge":{"http01":{"extra":1}}}}`,
			"object.unknown", `"acme.challenge.http01.extra" is not allowed`},
		{`{"provider":"acme","acme":{"server":"https://x/d","challenge":{"dns01":{"provider":"route53"}}}}`,
			"any.required", `"acme.challenge.dns01.route53" is required`},
		{`{"provider":"acme","acme":{"server":"https://x/d","challenge":{"dns01":{"provider":"route53","route53":{"region":"r","secretLink":"//secret/s"},"cloudflare":{"secretLink":"//secret/x"}}}}}`,
			"any.unknown", `"acme.challenge.dns01.cloudflare" is not allowed`},
		// provider now defaults to 'cpln' (master b3c3358ad), so a route53 block without an
		// explicit provider is rejected by onlyFor rather than for a missing provider.
		{`{"provider":"acme","acme":{"server":"https://x/d","challenge":{"dns01":{"route53":{"region":"r","secretLink":"//secret/s"}}}}}`,
			"any.unknown", `"acme.challenge.dns01.route53" is not allowed`},
		{`{"provider":"acme","acme":{"server":"https://x/d","challenge":{"dns01":{"provider":"route53","route53":{"secretLink":"//secret/s"}}}}}`,
			"any.required", `"acme.challenge.dns01.route53.region" is required`},
		{r53, "", ""},
		{`{"provider":"acme","acme":{"server":"https://x/d","challenge":{"dns01":{"provider":"cloudflare","cloudflare":{"secretLink":"//secret/x"}}}}}`, "", ""},
		{`{"provider":"acme","acme":{"server":"https://x/d","challenge":{"dns01":{"provider":"cloudDns","cloudDns":{"project":"p","secretLink":"//secret/s"}}}}}`, "", ""},
		{`{"provider":"acme","acme":{"server":"https://x/d","challenge":{"dns01":{"provider":"azureDns","azureDns":{"resourceGroupName":"rg","secretLink":"//secret/s"}}}}}`, "", ""},
		{`{"provider":"acme","acme":{"server":"https://x/d","email":"nope","challenge":{"http01":{}}}}`,
			"string.email", `"acme.email" must be a valid email`},
		{`{"provider":"acme","acme":{"server":"https://x/d","externalAccountBinding":{"keyId":"k"},"challenge":{"http01":{}}}}`,
			"any.required", `"acme.externalAccountBinding.secretLink" is required`},
		{`{"provider":"acme","acme":{"server":"https://x/d","externalAccountBinding":{"keyId":"k","secretLink":"//secret/s"},"challenge":{"http01":{}}}}`, "", ""},
	} {
		_, verr := validate.Decode[CertificateGeneratorSpec](validate.Bytes(c.body), &validate.Context{Org: "o"}, validate.Prefs{})
		gotTyp, gotMsg := "", ""
		if verr != nil {
			gotTyp, gotMsg = verr.Details()[0].Type, verr.Details()[0].Message
		}
		if gotTyp != c.typ || gotMsg != c.message {
			t.Errorf("%s:\n got  %q %q\n want %q %q", c.body, gotTyp, gotMsg, c.typ, c.message)
		}
	}

	// The secretLink is stored expanded, as Joi returns it.
	row, verr := validate.Decode[CertificateGeneratorSpec](validate.Bytes(r53), &validate.Context{Org: "o"}, validate.Prefs{})
	if verr != nil {
		t.Fatalf("route53 spec: %v", verr)
	}
	if got := row.Acme.Challenge.Dns01.Route53.SecretLink; got != "/org/o/secret/s" {
		t.Errorf("secretLink = %q, want %q", got, "/org/o/secret/s")
	}
}

// TestStatusMatchesJoi pins CertificateGeneratorStatus, including the ISO-millis form
// Joi's date().iso() renders lastUpdated as.
func TestStatusMatchesJoi(t *testing.T) {
	for _, c := range []struct {
		body string
		typ  string
		want string // normalized JSON when accepted
	}{
		{`{}`, "", `{}`},
		{`{"ready":true}`, "", `{"ready":true}`},
		{`{"ready":false}`, "", `{"ready":false}`},
		{`{"ready":"x"}`, "boolean.base", ""},
		{`{"message":"m"}`, "", `{"message":"m"}`},
		{`{"lastUpdated":"2026-01-02T03:04:05Z"}`, "", `{"lastUpdated":"2026-01-02T03:04:05.000Z"}`},
		{`{"lastUpdated":"2026-01-02"}`, "", `{"lastUpdated":"2026-01-02T00:00:00.000Z"}`},
		{`{"lastUpdated":"nope"}`, "date.format", ""},
		{`{"extra":1}`, "object.unknown", ""},
	} {
		row, verr := validate.Decode[CertificateGeneratorStatus](validate.Bytes(c.body), nil, validate.Prefs{})
		got := ""
		if verr != nil {
			got = verr.Details()[0].Type
		}
		if got != c.typ {
			t.Errorf("%s: got %q, want %q", c.body, got, c.typ)
			continue
		}
		if verr != nil {
			continue
		}
		b, _ := json.Marshal(row)
		if string(b) != c.want {
			t.Errorf("%s: normalized to %s, want %s", c.body, b, c.want)
		}
	}
}

// TestSecretLinksAndImmutable pins the two helpers the data-service and the actuator
// share: which secret type each link must have, and which spec edits orphan the ACME
// account.
func TestSecretLinksAndImmutable(t *testing.T) {
	spec := func(body string) *CertificateGeneratorSpec {
		t.Helper()
		var s CertificateGeneratorSpec
		if err := json.Unmarshal([]byte(body), &s); err != nil {
			t.Fatal(err)
		}
		return &s
	}

	got := SecretLinks(spec(`{"provider":"acme","acme":{"server":"https://x/d",
		"externalAccountBinding":{"keyId":"k","secretLink":"/org/o/secret/eab"},
		"challenge":{"dns01":{"provider":"route53","route53":{"region":"r","secretLink":"/org/o/secret/aws"}}}}}`))
	want := []SecretRef{{"/org/o/secret/eab", "opaque"}, {"/org/o/secret/aws", "aws"}}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("SecretLinks = %v, want %v", got, want)
	}
	for _, c := range []struct{ body, typ string }{
		{`{"provider":"acme","acme":{"server":"https://x/d","challenge":{"dns01":{"provider":"cloudDns","cloudDns":{"project":"p","secretLink":"/org/o/secret/s"}}}}}`, "gcp"},
		{`{"provider":"acme","acme":{"server":"https://x/d","challenge":{"dns01":{"provider":"azureDns","azureDns":{"resourceGroupName":"rg","secretLink":"/org/o/secret/s"}}}}}`, "azure-sdk"},
		{`{"provider":"acme","acme":{"server":"https://x/d","challenge":{"dns01":{"provider":"cloudflare","cloudflare":{"secretLink":"/org/o/secret/s"}}}}}`, "opaque"},
	} {
		refs := SecretLinks(spec(c.body))
		if len(refs) != 1 || refs[0].Type != c.typ {
			t.Errorf("%s: SecretLinks = %v, want one %s", c.body, refs, c.typ)
		}
	}
	if refs := SecretLinks(spec(`{"provider":"acme","acme":{"server":"https://x/d","challenge":{"http01":{}}}}`)); len(refs) != 0 {
		t.Errorf("http01 generator references secrets: %v", refs)
	}

	http01 := spec(`{"provider":"acme","acme":{"server":"https://x/d","challenge":{"http01":{}}}}`)
	dns01 := spec(`{"provider":"acme","acme":{"server":"https://x/d","challenge":{"dns01":{"provider":"cloudflare","cloudflare":{"secretLink":"/org/o/secret/s"}}}}}`)
	otherServer := spec(`{"provider":"acme","acme":{"server":"https://y/d","challenge":{"http01":{}}}}`)
	withEmail := spec(`{"provider":"acme","acme":{"server":"https://x/d","email":"a@b.com","challenge":{"http01":{}}}}`)
	if ChallengeType(http01) != CertChallengeKindHttp01 || ChallengeType(dns01) != CertChallengeKindDns01 {
		t.Error("ChallengeType misreports the challenge")
	}
	if ChallengeType(nil) != CertChallengeKindHttp01 {
		t.Error("a nil spec is http01 in TS")
	}
	for _, c := range []struct {
		before, after *CertificateGeneratorSpec
		want          string
	}{
		{http01, http01, ""},
		{http01, withEmail, ""},
		{http01, dns01, "spec.acme.challenge"},
		{http01, otherServer, "spec.acme.server"},
		{http01, spec(`{"provider":"acme"}`), "spec.acme.server"},
		{http01, spec(`{}`), "spec.provider"},
	} {
		if got := ImmutableChanged(c.before, c.after); got != c.want {
			t.Errorf("ImmutableChanged = %q, want %q", got, c.want)
		}
	}
}

// TestResourceMatchesJoi pins the top-level resource. Verdicts captured from the built
// certificateGenerator.ts.
func TestResourceMatchesJoi(t *testing.T) {
	const spec = `"spec":{"provider":"acme","acme":{"server":"https://x/d","challenge":{"http01":{}}}}`
	for _, c := range []struct {
		body    string
		typ     string
		message string
		want    string // normalized JSON when accepted
	}{
		{`{}`, "any.required", `"spec" is required`, ""},
		{`{` + spec + `}`, "", "", `{"tags":{},` + spec + `}`},
		{`{"name":"cg-one",` + spec + `}`, "", "", `{"name":"cg-one","tags":{},` + spec + `}`},
		{`{"name":"cg-one","tags":{"a":"b"},` + spec + `}`, "", "",
			`{"name":"cg-one","tags":{"a":"b"},` + spec + `}`},
		{`{"name":"cg-one","bogus":1,` + spec + `}`, "object.unknown", `"bogus" is not allowed`, ""},
		{`{"name":"cg-one","kind":"certificategenerator",` + spec + `}`, "", "",
			`{"name":"cg-one","kind":"certificategenerator","tags":{},` + spec + `}`},
	} {
		row, verr := validate.Decode[CertificateGenerator](validate.Bytes(c.body),
			&validate.Context{Org: "o"}, validate.Prefs{})
		gotTyp, gotMsg := "", ""
		if verr != nil {
			gotTyp, gotMsg = verr.Details()[0].Type, verr.Details()[0].Message
		}
		if gotTyp != c.typ || gotMsg != c.message {
			t.Errorf("%s:\n got  %q %q\n want %q %q", c.body, gotTyp, gotMsg, c.typ, c.message)
			continue
		}
		if verr != nil {
			continue
		}
		b, _ := json.Marshal(row)
		if string(b) != c.want {
			t.Errorf("%s: normalized to %s, want %s", c.body, b, c.want)
		}
	}
}
