package domain

import "testing"

// TestParseHostRaisesTldExtractThrow pins the throw the port used to swallow.
// tld-extract (schema/domain.ts:527) raises rather than falling back when the name's
// suffix is not an ICANN key or when the name IS a public suffix; TS lets it out of the
// custom, so it is a 400. Messages captured from the running TS service.
func TestParseHostRaisesTldExtractThrow(t *testing.T) {
	for _, c := range []struct {
		name string
		want string
	}{
		// The ICANN lookup is case-sensitive, so an uppercase TLD is simply absent.
		{"EXAMPLE.COM", `Invalid TLD {"parts":["EXAMPLE","COM"],"tld_level":-1,"allowUnknownTLD":false}`},
		{"co.uk", `Invalid TLD {"parts":["co","uk"],"tld_level":2,"allowUnknownTLD":false}`},
		{"foo.notatld", `Invalid TLD {"parts":["foo","notatld"],"tld_level":-1,"allowUnknownTLD":false}`},
	} {
		_, _, err := ParseHost(c.name)
		if err == nil {
			t.Errorf("ParseHost(%q): no error, want %q", c.name, c.want)
			continue
		}
		if err.Error() != c.want {
			t.Errorf("ParseHost(%q) = %q, want %q", c.name, err.Error(), c.want)
		}
	}
	// Names tld-extract accepts, including a mixed-case label below the TLD and a
	// unicode label, which Joi's own .domain() already let through.
	for _, ok := range []string{"example.com", "sub.EXAMPLE.com", "föö.com", "a.co", "x.co.uk", "my.internal"} {
		if _, _, err := ParseHost(ok); err != nil {
			t.Errorf("ParseHost(%q) = %v, want no error", ok, err)
		}
	}
}

// TestIsValidPublicDomainUsesTheEngine pins the two behaviors the hand-rolled copy got
// wrong: an uppercase TLD is VALID to Joi (its URL step lowercases the whole host), and
// TLD membership comes from the IANA list rather than the Public Suffix List.
func TestIsValidPublicDomainUsesTheEngine(t *testing.T) {
	for _, c := range []struct {
		in   string
		want bool
	}{
		{"EXAMPLE.COM", true},
		{"FOO.com", true},
		{"föö.com", true},
		{"example.com", true},
		{"foo.notatld", false},
		{"ab.c", false},
		{"localhost", false},
		{"1.2.3.4", false},
	} {
		if got := isValidPublicDomain(c.in); got != c.want {
			t.Errorf("isValidPublicDomain(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}
