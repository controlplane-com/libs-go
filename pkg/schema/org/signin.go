package org

import (
	"slices"
	"strings"

	"github.com/controlplane-com/libs-go/pkg/schema/util"
)

// SignInProviders ports nodelibs/schema/src/signInProviders.ts SIGN_IN_PROVIDERS; SAML is allowed only through samlOnly.
var SignInProviders = []string{"email", "github", "google", "microsoft"}

var providerLabels = map[string]string{
	"email":     "Email",
	"github":    "GitHub",
	"google":    "Google",
	"microsoft": "Microsoft",
	"saml":      "SAML",
}

// ProviderFamily maps a Firebase sign_in_provider claim to the provider an org can allow; "" when none.
func ProviderFamily(signInProvider string) string {
	if strings.HasPrefix(signInProvider, "saml.") {
		return "saml"
	}
	switch signInProvider {
	case "google.com":
		return "google"
	case "microsoft.com":
		return "microsoft"
	case "github.com":
		return "github"
	case "password":
		return "email"
	}
	return ""
}

// EffectiveAllowedProviders returns nil when every provider is allowed.
func EffectiveAllowedProviders(ac *AuthConfig) []string {
	if ac == nil {
		return nil
	}
	if util.Bool(ac.SamlOnly) {
		return []string{"saml"}
	}
	if len(ac.AllowedProviders) == 0 {
		return nil
	}
	out := make([]string, len(ac.AllowedProviders))
	for i, p := range ac.AllowedProviders {
		out[i] = string(p)
	}
	return out
}

func IsProviderAllowed(allowed []string, signInProvider string) bool {
	if allowed == nil {
		return true
	}
	family := ProviderFamily(signInProvider)
	return family != "" && slices.Contains(allowed, family)
}

func AllowsSignIn(ac *AuthConfig, signInProvider string) bool {
	return IsProviderAllowed(EffectiveAllowedProviders(ac), signInProvider)
}

func ProviderLabels(providers []string) string {
	out := make([]string, len(providers))
	for i, p := range providers {
		if l, ok := providerLabels[p]; ok {
			out[i] = l
		} else {
			out[i] = p
		}
	}
	return strings.Join(out, ", ")
}

func SignInMethodLabel(signInProvider string) string {
	if l, ok := providerLabels[ProviderFamily(signInProvider)]; ok {
		return l
	}
	return "an unsupported sign-in method"
}

func ProviderDeniedMessage(orgName string, allowed []string, signInProvider string) string {
	return "Org " + orgName + " allows sign-in only with " + ProviderLabels(allowed) + "; you signed in with " + SignInMethodLabel(signInProvider)
}
