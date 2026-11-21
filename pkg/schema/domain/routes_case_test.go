package domain

import (
	"fmt"
	"strings"
	"testing"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// TestRouteCaseInsensitiveAndUniqueness pins the routes rules against verdicts captured
// from the built domain.ts: caseInsensitive (master a7f9298a2) and the order uniqueRoutes
// checks in — duplicates over the whole array first, then the first group by appearance.
func TestRouteCaseInsensitiveAndUniqueness(t *testing.T) {
	const w = `"workloadLink":"//gvc/g/workload/w"`
	spec := func(routes ...string) string {
		return `{"ports":[{"number":443,"protocol":"http2","routes":[` + strings.Join(routes, ",") + `]}]}`
	}
	for _, c := range []struct {
		name, body, typ, path, message string // typ "" = accepted
	}{
		{"caseInsensitive with a regex", spec(`{"regex":"/x","caseInsensitive":true,` + w + `}`),
			"any.unknown", "ports.0.routes.0.caseInsensitive",
			"caseInsensitive applies to prefix only; use the (?i) flag in the regex instead"},
		{"even false is forbidden with a regex", spec(`{"regex":"/x","caseInsensitive":false,` + w + `}`),
			"any.unknown", "ports.0.routes.0.caseInsensitive",
			"caseInsensitive applies to prefix only; use the (?i) flag in the regex instead"},
		{"prefixes differing only by case, one insensitive",
			spec(`{"prefix":"/Api",`+w+`}`, `{"prefix":"/api","caseInsensitive":true,`+w+`}`),
			"any.custom", "ports.0.routes",
			`"ports[0].routes" failed custom validation because The case insensitive route /api overlaps with another route that differs only by case`},
		{"prefixes differing only by case, both sensitive",
			spec(`{"prefix":"/Api",`+w+`}`, `{"prefix":"/api",`+w+`}`), "", "", ""},
		{"the first shared group by appearance is the one reported",
			spec(`{"prefix":"/a",`+w+`}`, `{"prefix":"/b",`+w+`,"port":8080}`, `{"prefix":"/b",`+w+`,"port":8081}`, `{"prefix":"/a",`+w+`,"port":8082}`),
			"any.custom", "ports.0.routes",
			`"ports[0].routes" failed custom validation because There are more than one routes for the prefix /a`},
		{"an exact duplicate anywhere wins over a shared prefix",
			spec(`{"prefix":"/x",`+w+`,"port":8080}`, `{"prefix":"/x",`+w+`,"port":8081}`, `{"prefix":"/c",`+w+`}`, `{"prefix":"/c",`+w+`}`),
			"any.custom", "ports.0.routes",
			`"ports[0].routes" failed custom validation because Routes must be unique`},
	} {
		_, verr := validate.Decode[DomainSpec](validate.Bytes(c.body), &validate.Context{Org: "o", Gvc: "g"}, validate.Prefs{})
		if c.typ == "" {
			if verr != nil {
				t.Errorf("%s: rejected: %v", c.name, verr.Details()[0].Message)
			}
			continue
		}
		if verr == nil {
			t.Errorf("%s: accepted", c.name)
			continue
		}
		d := verr.Details()[0]
		path := make([]string, len(d.Path))
		for i, p := range d.Path {
			path[i] = fmt.Sprint(p)
		}
		if d.Type != c.typ || strings.Join(path, ".") != c.path || d.Message != c.message {
			t.Errorf("%s:\n got  %s @%s | %s\n want %s @%s | %s", c.name, d.Type, strings.Join(path, "."), d.Message, c.typ, c.path, c.message)
		}
	}
}
