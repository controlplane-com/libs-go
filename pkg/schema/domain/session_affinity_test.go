package domain

import (
	"strings"
	"testing"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

func TestRouteSessionAffinity(t *testing.T) {
	route := func(prefix, cookie, extra string) string {
		return `{"prefix":"` + prefix + `","workloadLink":"//gvc/g/workload/w","sessionAffinity":{"canary":{"cookie":"` + cookie + `"}}` + extra + `}`
	}
	port := func(number string, routes ...string) string {
		return `{"number":` + number + `,"protocol":"http2","routes":[` + strings.Join(routes, ",") + `]}`
	}
	for _, c := range []struct {
		name, ports, message string // message "" = accepted
	}{
		{"distinct cookies", port("443", route("/", "a", ""), route("/api", "b", "")), ""},
		{"same cookie on one port", port("443", route("/", "a", ""), route("/api", "a", "")), ""},
		{"same cookie across ports", port("443", route("/", "a", "")) + "," + port("80", route("/", "a", "")), ""},
		{"distinct canary", port("443", route("/", "a", `,"canaries":[{"workloadLink":"//gvc/g/workload/c","weight":10}]`)), ""},
		{"canary repeats the primary", port("443", route("/", "a", `,"canaries":[{"workloadLink":"//gvc/g/workload/w","port":8081,"weight":10}]`)), ""},
		{"invalid cookie name", port("443", route("/", "a;b", "")), "pattern"},
		{"cookie name too long", port("443", route("/", strings.Repeat("x", 65), "")), "64"},
		{"canary required", port("443", `{"prefix":"/","workloadLink":"//gvc/g/workload/w","sessionAffinity":{}}`), "required"},
		{"cookie required", port("443", `{"prefix":"/","workloadLink":"//gvc/g/workload/w","sessionAffinity":{"canary":{}}}`), "required"},
	} {
		body := `{"acceptAllHosts":true,"ports":[` + c.ports + `]}`
		_, verr := validate.Decode[DomainSpec](validate.Bytes(body), &validate.Context{Org: "o", Gvc: "g"}, validate.Prefs{})
		if c.message == "" {
			if verr != nil {
				t.Errorf("%s: rejected: %v", c.name, verr.Details()[0].Message)
			}
			continue
		}
		if verr == nil {
			t.Errorf("%s: accepted", c.name)
		} else if m := verr.Details()[0].Message; !strings.Contains(m, c.message) {
			t.Errorf("%s: got %q, want it to contain %q", c.name, m, c.message)
		}
	}
}
