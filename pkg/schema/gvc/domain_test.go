package gvc

import (
	"testing"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// TestGvcSpecDomainChecksTLDs pins gvc.ts:92 `Joi.string().domain({minDomainSegments:2})`.
// Passing ANY options object turns Joi's IANA TLD list on, so this field is the
// TLD-checked form while a bare `.domain()` elsewhere is not. The port had it on the
// bare tag and accepted `foo.notatld` as a GVC domain.
func TestGvcSpecDomainChecksTLDs(t *testing.T) {
	for _, c := range []struct {
		in     string
		reject bool
	}{
		{"example.com", false},
		{"EXAMPLE.COM", false}, // Joi's punycode step lowercases the whole host
		{"föö.com", false},
		{"a.co", false},
		{"foo.notatld", true},
		{"ab.c", true}, // .c is not an IANA TLD
		{"localhost", true},
	} {
		body := `{"name":"g1","spec":{"domain":` + quote(c.in) + `}}`
		_, verr := validate.Decode[Gvc](validate.Bytes(body), nil, validate.Prefs{})
		if c.reject && verr == nil {
			t.Errorf("spec.domain %q accepted, want rejected", c.in)
		}
		if !c.reject && verr != nil {
			t.Errorf("spec.domain %q rejected: %v", c.in, verr)
		}
	}
}

func quote(s string) string {
	b := []byte{'"'}
	for i := 0; i < len(s); i++ {
		if s[i] == '"' || s[i] == '\\' {
			b = append(b, '\\')
		}
		b = append(b, s[i])
	}
	return string(append(b, '"'))
}
