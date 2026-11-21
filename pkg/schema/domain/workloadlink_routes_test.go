package domain

import (
	"testing"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// Verdicts captured from the built domain.ts: with workloadLink set, a port without routes fails the for..of.
func TestWorkloadLinkPortRoutes(t *testing.T) {
	const w = `"//gvc/g/workload/w"`
	for _, c := range []struct {
		name, ports, message string // message "" = accepted
	}{
		{"routes absent", `{"number":443,"protocol":"http2"}`,
			`"value" failed custom validation because port.routes is not iterable`},
		{"routes empty", `{"number":443,"protocol":"http2","routes":[]}`,
			`"value" failed custom validation because when workloadLink is configured, every port must have exactly ONE route.`},
		{"one route", `{"number":443,"protocol":"http2","routes":[{"prefix":"/","workloadLink":` + w + `}]}`, ""},
	} {
		body := `{"workloadLink":` + w + `,"ports":[` + c.ports + `]}`
		_, verr := validate.Decode[DomainSpec](validate.Bytes(body), &validate.Context{Org: "o", Gvc: "g"}, validate.Prefs{})
		if c.message == "" {
			if verr != nil {
				t.Errorf("%s: rejected: %v", c.name, verr.Details()[0].Message)
			}
			continue
		}
		if verr == nil {
			t.Errorf("%s: accepted", c.name)
			continue
		}
		if got := verr.Details()[0].Message; got != c.message {
			t.Errorf("%s:\n got  %s\n want %s", c.name, got, c.message)
		}
	}
}
