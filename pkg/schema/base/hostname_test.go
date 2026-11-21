package base

import (
	"testing"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// TestHostnameRule pins Joi's string.hostname, which is Domain.isValid(v,
// {minDomainSegments:1}) OR an IP literal — not an RFC-1123 regex. The differences
// that bite: a single label is a hostname, the LAST label must start with a letter,
// unicode is punycoded (and accepted), and an all-digit last label is not a hostname.
func TestHostnameRule(t *testing.T) {
	cases := map[string]bool{
		"a.example.com":     true,
		"localhost":         true,
		"FOO.COM":           true, // analyze lowercases via the URL parser
		"f\u00f6\u00f6.com": true, // allowUnicode defaults to true
		"xn--fsq.com":       true,
		"1.2.3.4":           true, // an IP is a hostname
		"::1":               true,
		"-bad-":             false,
		"a..b":              false,
		"a b.com":           false,
		"":                  false,
		"a.example.com/x":   false,
		"x.1b":              false, // the TLD segment must start with a letter
		"x.b1":              true,
	}
	for v, want := range cases {
		_, err := HostnameRule().ValidateAt(nil, v, nil, validate.Prefs{})
		if (err == nil) != want {
			t.Errorf("hostname %q: want ok=%v, err=%v", v, want, err)
		}
	}
}

// TestHostnameError pins the error shape: Joi's string.hostname quoting the LABEL,
// with the value in context (it used to be an any.custom quoting the value).
func TestHostnameError(t *testing.T) {
	_, err := HostnameRule().ValidateAt([]string{"elastic", "host"}, "-bad-", nil, validate.Prefs{})
	d := err.(*validate.Error).Details()[0]
	if d.Type != "string.hostname" {
		t.Errorf("type = %q, want string.hostname", d.Type)
	}
	if d.Message != `"elastic.host" must be a valid hostname` {
		t.Errorf("message = %q", d.Message)
	}
	if d.Context["value"] != "-bad-" {
		t.Errorf("context.value = %v", d.Context["value"])
	}
}
