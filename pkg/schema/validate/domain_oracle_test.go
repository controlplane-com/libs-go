package validate

import "testing"

// domainOracle is Joi 17.6.0's verdict for Joi.string().domain() (bare, no TLD list)
// and Joi.string().domain({minDomainSegments:2}) / ({tlds:{allow:true}}) -- the two
// option forms the schemas use -- captured from the running library, not inferred.
var domainOracle = []struct {
	in               string
	bare, min2, tlds bool
}{
	{"föö.com", true, true, true},
	{"FÖÖ.COM", true, true, true},
	{"xn--f-1gaa.com", true, true, true},
	{"日本.jp", true, true, true},
	{"日本.com", true, true, true},
	{"a_b.com", false, false, false},
	{"a..com", false, false, false},
	{"-a.com", false, false, false},
	{"a-.com", false, false, false},
	{"a.b-", false, false, false},
	{"ab.c", true, false, false},
	{"A.COM", true, true, true},
	{"1.com", true, true, true},
	{"a.1", false, false, false},
	{"x.internal", true, false, false},
	{"example.com", true, true, true},
	{"example.invalidtld", true, false, false},
	{"foo.zzz", true, false, false},
	{"a.co", true, true, true},
	{"example.xn--p1ai", true, true, true},
	{"localhost", false, false, false},
	{"café.fr", true, true, true},
	{"mü.de", true, true, true},
	{"ümlaut.example.com", true, true, true},
	{"a.com.", false, false, false},
	{"％.com", false, false, false},
	{"100%.com", false, false, false},
	{"ex%20ample.com", false, false, false},
	{"straße.de", true, true, true},
	{"ＦＵＬＬ.com", true, true, true},
	{"xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx.com", false, false, false},
	{"a.bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", false, false, false},
	{"\u200b.com", false, false, false},
	{"é.com", true, true, true},
	{"é.com", true, true, true},
	{"emoji😀.com", true, true, true},
	{".com", false, false, false},
	{"com.", false, false, false},
	{"a b.com", false, false, false},
	{"a\tb.com", false, false, false},
	{"[a].com", false, false, false},
	{"a:b.com", false, false, false},
	{"xn--.com", false, false, false},
	{"xn--a.com", true, true, true},
	{"1.2.3.4", false, false, false},
	{"::1", false, false, false},
	{"a.xn--p1ai", true, true, true},
}

func TestIsDomainMatchesJoi(t *testing.T) {
	for _, c := range domainOracle {
		// bare .domain() and .domain({minDomainSegments:2}) differ ONLY in the TLD list:
		// 2 is already the library default.
		if got := isDomain(c.in, false); got != c.bare {
			t.Errorf("isDomain(%q, tlds=false) = %v, want %v", c.in, got, c.bare)
		}
		if got := isDomain(c.in, true); got != c.min2 {
			t.Errorf("isDomain(%q, tlds=true) = %v, want %v", c.in, got, c.min2)
		}
		if c.min2 != c.tlds {
			t.Errorf("%q: Joi's two option forms disagree (%v vs %v)", c.in, c.min2, c.tlds)
		}
	}
}
