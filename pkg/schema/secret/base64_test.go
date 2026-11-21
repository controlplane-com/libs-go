package secret

import (
	"testing"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// TestOpaqueBase64RejectsNewlines pins `Joi.string().base64()`, whose regex anchors the
// WHOLE string. The port validated by attempting a decode instead, and Go's
// base64.StdEncoding ignores \r and \n by design — so a payload with a trailing newline
// was accepted and stored where TS answers 400.
func TestOpaqueBase64RejectsNewlines(t *testing.T) {
	dec := func(payload string) *validate.Error {
		body := `{"payload":` + quoteJSON(payload) + `,"encoding":"base64"}`
		_, verr := validate.Decode[Opaque](validate.Bytes(body), nil, validate.Prefs{})
		return verr
	}
	for _, bad := range []string{"aGk=\n", "\naGk=", "aG\nk=", "aGk=\r\n", "aG k=", "aGk", "a"} {
		verr := dec(bad)
		if verr == nil {
			t.Errorf("payload %q accepted, want rejected", bad)
			continue
		}
		if got := verr.Details()[0].Type; got != "string.base64" {
			t.Errorf("payload %q: type %q, want string.base64", bad, got)
		}
	}
	// "" is rejected by string.empty, which is a separate rule.
	for _, ok := range []string{"aGk=", "YWJjZA=="} {
		if verr := dec(ok); verr != nil {
			t.Errorf("payload %q rejected: %v", ok, verr)
		}
	}
}

// TestOpaqueBase64OnlyWhenEncodingSaysSo pins the conditional itself. secret.ts:8 spells
// a LITERAL `is: 'base64'`, which Joi compiles to `valid('base64').required()` — so an
// ABSENT encoding takes the otherwise branch and any string is fine. Verdicts captured
// from joi@17.6.0.
func TestOpaqueBase64OnlyWhenEncodingSaysSo(t *testing.T) {
	for _, c := range []struct {
		body   string
		reject bool
	}{
		{`{"payload":"x"}`, false},
		{`{"payload":"x","encoding":"plain"}`, false},
		{`{"payload":"x","encoding":"base64"}`, true},
		{`{"payload":"aGk=","encoding":"base64"}`, false},
	} {
		_, verr := validate.Decode[Opaque](validate.Bytes(c.body), nil, validate.Prefs{})
		if c.reject && verr == nil {
			t.Errorf("%s accepted, want rejected", c.body)
		}
		if !c.reject && verr != nil {
			t.Errorf("%s rejected: %v", c.body, verr)
		}
	}
}

func quoteJSON(s string) string {
	b := []byte{'"'}
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '"', '\\':
			b = append(b, '\\', s[i])
		case '\n':
			b = append(b, '\\', 'n')
		case '\r':
			b = append(b, '\\', 'r')
		default:
			b = append(b, s[i])
		}
	}
	return string(append(b, '"'))
}
