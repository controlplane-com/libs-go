package validate

import "testing"

// The tables below extend the smaller hand-written set in engine_parity_test.go.
// emailOracle / uriOracle / guidOracle are Joi 17.6.0's verdicts for
// Joi.string().email(), .uri() and .guid(), captured from the running library rather
// than inferred -- the hand-rolled regexes these replaced accepted 12 of the invalid
// addresses and 5 of the invalid URIs below.
var emailOracle = []struct {
	in string
	ok bool
}{
	{"a@b.c", false},
	{"a@b.co", true},
	{"kyle@controlplane.com", true},
	{"a..b@c.com", false},
	{".a@b.co", false},
	{"a.@b.co", false},
	{"a@b..co", false},
	{"a@-b.co", false},
	{"a@b-.co", false},
	{"a@b.co.", false},
	{"a@127.0.0.1", false},
	{"a@[127.0.0.1]", false},
	{"a@b_c.co", false},
	{"<a@b.co>", false},
	{"a b@c.com", false},
	{"a@b c.com", false},
	{"", false},
	{"@b.com", false},
	{"a@", false},
	{"a@@b.com", false},
	{"a@b", false},
	{"first.last@sub.domain.example.com", true},
	{"user+tag@gmail.com", true},
	{"user_name@example.io", true},
	{"UPPER@EXAMPLE.COM", true},
	{"a@xn--fiq228c.com", true},
	{"a@f\u00f6\u00f6.com", true},
	{"f\u00f6\u00f6@bar.com", true},
	{"a@b.museum", true},
	{"a@b.notatld", false},
	{"a@b.c.d.e.f.g", false},
	{"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa@b.com", false},
	{"a@bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb.com", false},
	{"a@b.co\n", false},
	{"\ta@b.co", false},
	{"a@b.com ", false},
	{"o'brien@x.com", true},
	{"a!#$%&'*+-/=?^_`{|}~@x.com", true},
	{"a@x.com.", false},
	{"a@.x.com", false},
	{"a@x..com", false},
	{"a@[IPv6:::1]", false},
	{"a@xn--p1ai.xn--p1ai", true},
	{"admin@localhost", false},
	{"a@b.xyz", true},
	{"a@b.app", true},
	{"a@b.dev", true},
	{"a@b.zz", false},
	{"a@b.aaa", true},
	{"test@[192.168.1.1]", false},
	{"a@b.c-d.com", true},
	{"a@-.com", false},
	{"a@b-c.com", true},
	{"a@1.com", true},
	{"a@1b.com", true},
	{"a@b1.com", true},
	{"noreply@cpln.io", true},
	{"billing@controlplane.com", true},
	{"a@sub-domain.example.co.uk", true},
}

var uriOracle = []struct {
	in string
	ok bool
}{
	{"zz:", false},
	{"zz://", false},
	{"http://", false},
	{"http://x.com/a b", false},
	{"http://x.com/\\", false},
	{"http://example.com", true},
	{"https://example.com/", true},
	{"https://example.com/a/b?c=d#e", true},
	{"ftp://ftp.example.com/x", true},
	{"mailto:a@b.com", true},
	{"urn:isbn:0451450523", true},
	{"//example.com", false},
	{"example.com", false},
	{"http:/x.com", false},
	{"http:x.com", false},
	{"http:///x", false},
	{"https://user:pw@host:8443/p?q#f", true},
	{"git+ssh://git@github.com/o/r.git", true},
	{"s3://bucket/key", true},
	{"nats://localhost:4222", true},
	{"redis://:pass@127.0.0.1:6379/0", true},
	{"http://[::1]:8080/", true},
	{"http://256.1.1.1/", true},
	{"http://1.2.3.4/", true},
	{"http://\u4f8b\u3048.\u30c6\u30b9\u30c8", false},
	{"http://xn--r8jz45g.xn--zckzah", true},
	{"", false},
	{"   ", false},
	{"http://x.com/%20", true},
	{"http://x.com/%zz", true},
	{"postgres://u:p@h:5432/db?sslmode=require", true},
	{"file:///etc/passwd", true},
	{"data:text/plain;base64,aGk=", true},
	{"tel:+15551212", true},
	{"HTTP://X.COM", true},
	{"Http://x.com/A", true},
	{"http://x.com:80", true},
	{"http://x.com:", true},
	{"http://x.com?", true},
	{"http://x.com#", true},
	{"wss://x.com/socket", true},
	{"docker://us-docker.pkg.dev/controlplane/internal/x:tag", true},
	{"https://api.cpln.io/org/x/gvc/y", true},
	{"/relative/path", false},
	{"./rel", false},
	{"?q=1", false},
	{"#frag", false},
	{"javascript:alert(1)", true},
	{"vm+bios:x", true},
	{"a-b.c+d:foo", true},
}

var guidOracle = []struct {
	in string
	ok bool
}{
	{"2c9f6a3e-1b2c-4d3e-8f4a-5b6c7d8e9f01", true},
	{"{2c9f6a3e-1b2c-4d3e-8f4a-5b6c7d8e9f01}", true},
	{"(2c9f6a3e-1b2c-4d3e-8f4a-5b6c7d8e9f01)", true},
	{"2c9f6a3e1b2c4d3e8f4a5b6c7d8e9f01", true},
	{"[2c9f6a3e-1b2c-4d3e-8f4a-5b6c7d8e9f01]", true},
	{"{2c9f6a3e-1b2c-4d3e-8f4a-5b6c7d8e9f01)", false},
	{"2C9F6A3E-1B2C-4D3E-8F4A-5B6C7D8E9F01", true},
	{"2c9f6a3e:1b2c:4d3e:8f4a:5b6c7d8e9f01", true},
	{"not-a-guid", false},
	{"", false},
}

func TestStringFormatsMatchJoiOracle(t *testing.T) {
	for _, c := range []struct {
		name  string
		rule  *Rule
		cases []struct {
			in string
			ok bool
		}
	}{
		{"email", String().Email(), emailOracle},
		{"uri", String().URI(), uriOracle},
		{"guid", String().UUID(), guidOracle},
	} {
		for _, k := range c.cases {
			_, err := c.rule.ValidateAt(nil, k.in, nil, Prefs{})
			if (err == nil) != k.ok {
				t.Errorf("%s(%q) = ok:%v, Joi says ok:%v", c.name, k.in, err == nil, k.ok)
			}
		}
	}
}
