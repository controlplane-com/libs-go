package domain

import (
	"strings"
	"testing"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// TestRoutePatternSource pins the route pattern messages to the JS source domain.ts:57
// is written with. Go's Regexp.String() renders the same language with different
// escapes (`\-` for `-`, `/` for `\/`), and that text is what the caller reads in
// string.pattern.base — and what the describe document publishes.
func TestRoutePatternSource(t *testing.T) {
	rule := RouteRule()
	_, err := rule.ValidateAt([]string{"spec", "ports", validate.IndexPath(0), "routes", validate.IndexPath(0)},
		map[string]any{"prefix": "no-leading-slash", "workloadLink": "//workload/w"}, nil, validate.Prefs{})
	if err == nil {
		t.Fatal("a prefix with no leading slash accepted")
	}
	d := err.(*validate.Error).Details()[0]
	if d.Type != "string.pattern.base" {
		t.Fatalf("type = %q (%s)", d.Type, d.Message)
	}
	if !strings.Contains(d.Message, `/^\/[0-9a-zA-Z-\._~\/]*$/`) {
		t.Errorf("message = %q, want the TS regex source", d.Message)
	}
}

// TestInternalPortWithoutRoutes pins INTENTIONAL-DIVERGENCES C27: domain.ts:508-512
// reads `port.routes.length` on an internal domain, which throws for a port that
// carries no routes (TS answers 400 with the TypeError text). This port reads the
// length of an absent list as 0 — what the TS author meant — and creates the domain.
func TestInternalPortWithoutRoutes(t *testing.T) {
	body := `{"name":"x.apex.internal","spec":{"ports":[{"number":443}]}}`
	if _, verr := validate.Decode[Domain](validate.Bytes(body),
		&validate.Context{Org: "o1", Method: "POST"}, validate.Prefs{}); verr != nil {
		t.Fatalf("an internal domain port with no routes was rejected: %v", verr)
	}
}
