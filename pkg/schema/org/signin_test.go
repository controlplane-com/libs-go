package org

import (
	"reflect"
	"strings"
	"testing"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

func TestProviderFamily(t *testing.T) {
	cases := map[string]string{
		"saml.okta.com": "saml",
		"google.com":    "google",
		"microsoft.com": "microsoft",
		"github.com":    "github",
		"password":      "email",
		"custom":        "",
		"test":          "",
		"":              "",
	}
	for in, want := range cases {
		if got := ProviderFamily(in); got != want {
			t.Errorf("ProviderFamily(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestEffectiveAllowedProviders(t *testing.T) {
	yes, no := true, false
	cases := []struct {
		ac   *AuthConfig
		want []string
	}{
		{nil, nil},
		{&AuthConfig{}, nil},
		{&AuthConfig{AllowedProviders: SignInProviderList{}}, nil},
		{&AuthConfig{SamlOnly: &no}, nil},
		{&AuthConfig{SamlOnly: &yes}, []string{"saml"}},
		{&AuthConfig{SamlOnly: &yes, AllowedProviders: SignInProviderList{}}, []string{"saml"}},
		{&AuthConfig{AllowedProviders: SignInProviderList{"github", "google"}}, []string{"github", "google"}},
	}
	for i, c := range cases {
		if got := EffectiveAllowedProviders(c.ac); !reflect.DeepEqual(got, c.want) {
			t.Errorf("case %d: got %v, want %v", i, got, c.want)
		}
	}
}

func TestIsProviderAllowed(t *testing.T) {
	cases := []struct {
		allowed  []string
		provider string
		want     bool
	}{
		{nil, "", true},
		{nil, "test", true},
		{[]string{"saml"}, "saml.okta.com", true},
		{[]string{"saml"}, "google.com", false},
		{[]string{"google", "microsoft"}, "microsoft.com", true},
		{[]string{"google", "microsoft"}, "github.com", false},
		{[]string{"google", "microsoft"}, "saml.okta.com", false},
		{[]string{"email"}, "password", true},
		{[]string{"google"}, "", false},
		{[]string{"google"}, "test", false},
	}
	for _, c := range cases {
		if got := IsProviderAllowed(c.allowed, c.provider); got != c.want {
			t.Errorf("IsProviderAllowed(%v, %q) = %v, want %v", c.allowed, c.provider, got, c.want)
		}
	}
}

func TestProviderDeniedMessage(t *testing.T) {
	if got := ProviderDeniedMessage("acme", []string{"google", "microsoft"}, "github.com"); got != "Org acme allows sign-in only with Google, Microsoft; you signed in with GitHub" {
		t.Errorf("got %q", got)
	}
	if got := ProviderDeniedMessage("acme", []string{"saml"}, ""); got != "Org acme allows sign-in only with SAML; you signed in with an unsupported sign-in method" {
		t.Errorf("got %q", got)
	}
}

func decodeAuthConfig(t *testing.T, authConfig string) (*Org, *validate.Error) {
	t.Helper()
	return validate.Decode[Org](validate.Bytes(`{"name":"acme","spec":{"authConfig":`+authConfig+`}}`), octx(), validate.Prefs{})
}

func TestAllowedProvidersSortedDeduped(t *testing.T) {
	o, verr := decodeAuthConfig(t, `{"allowedProviders":["microsoft","google","microsoft"]}`)
	if verr != nil {
		t.Fatalf("unexpected error: %v", verr)
	}
	if got := o.Spec.AuthConfig.AllowedProviders; !reflect.DeepEqual(got, SignInProviderList{"google", "microsoft"}) {
		t.Fatalf("allowedProviders = %v, want [google microsoft]", got)
	}
}

func TestAllowedProvidersValidation(t *testing.T) {
	cases := []struct {
		authConfig string
		wantErr    string
	}{
		{`{"allowedProviders":["email","github","google","microsoft"]}`, ""},
		{`{"allowedProviders":[]}`, ""},
		{`{"allowedProviders":["saml"]}`, `must be one of [email, github, google, microsoft]`},
		{`{"allowedProviders":["facebook"]}`, `must be one of [email, github, google, microsoft]`},
		{`{"samlOnly":true,"allowedProviders":[]}`, ""},
		{`{"samlOnly":false,"allowedProviders":["google"]}`, ""},
		{`{"samlOnly":true,"allowedProviders":["google"]}`,
			`"spec.authConfig" failed custom validation because samlOnly and allowedProviders are mutually exclusive`},
		{`{"domainAutoMembers":["acme.com"],"samlOnly":true}`, ""},
		{`{"domainAutoMembers":["acme.com"],"allowedProviders":[]}`, ""},
		{`{"domainAutoMembers":["acme.com"],"allowedProviders":["google"]}`,
			`"spec.authConfig" failed custom validation because domainAutoMembers adds members only on SAML sign-in, which allowedProviders does not allow`},
	}
	for _, c := range cases {
		t.Run(c.authConfig, func(t *testing.T) {
			_, verr := decodeAuthConfig(t, c.authConfig)
			if c.wantErr == "" {
				if verr != nil {
					t.Fatalf("unexpected error: %v", verr)
				}
				return
			}
			if verr == nil {
				t.Fatalf("want error %q", c.wantErr)
			}
			if msg := verr.Details()[0].Message; !strings.Contains(msg, c.wantErr) {
				t.Fatalf("message = %q, want it to contain %q", msg, c.wantErr)
			}
		})
	}
}

func TestAllowsSignIn(t *testing.T) {
	yes := true
	cases := []struct {
		ac       *AuthConfig
		provider string
		want     bool
	}{
		{nil, "github.com", true},
		{&AuthConfig{SamlOnly: &yes}, "saml.okta.com", true},
		{&AuthConfig{SamlOnly: &yes}, "google.com", false},
		{&AuthConfig{AllowedProviders: SignInProviderList{"github"}}, "github.com", true},
		{&AuthConfig{AllowedProviders: SignInProviderList{"github"}}, "saml.okta.com", false},
	}
	for i, c := range cases {
		if got := AllowsSignIn(c.ac, c.provider); got != c.want {
			t.Errorf("case %d: AllowsSignIn(%q) = %v, want %v", i, c.provider, got, c.want)
		}
	}
}
