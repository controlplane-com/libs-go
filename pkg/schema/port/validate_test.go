package port

import (
	"testing"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// TestPortRule pins the shared Port rule against nodelibs/schema/src/port.ts.
// Before this rule existed the type reflected to an unconstrained Number(), so
// every one of the reject cases below was accepted and PERSISTED — including the
// Istio sidecar ports.
func TestPortRule(t *testing.T) {
	rule := Port(0).ValidationRule()
	cases := []validate.Case{
		{Name: "in range", Input: float64(8080), Want: float64(8080)},
		{Name: "lower bound", Input: float64(80), Want: float64(80)},
		{Name: "upper bound", Input: float64(65535), Want: float64(65535)},

		{Name: "below min", Input: float64(79), WantErr: "greater than or equal to 80"},
		{Name: "zero", Input: float64(0), WantErr: "greater than or equal to 80"},
		{Name: "above max", Input: float64(65536), WantErr: "less than or equal to 65535"},
		{Name: "not an integer", Input: float64(80.5), WantErr: "integer"},
		{Name: "not a number", Input: "8080x", WantErr: "must be a number"},
	}
	// Every reserved port, so adding one to the list without a rule is caught.
	for _, p := range ReservedPorts {
		cases = append(cases, validate.Case{
			Name:    "reserved",
			Input:   float64(p),
			WantErr: "is reserved for system use",
		})
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			if err := c.Run(rule); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// TestReservedPortMessage pins the exact wording and the JS Array.toString
// rendering of the reserved list (comma-joined, no spaces), inside Joi's custom
// wrapper — verified against the TS service, which answers
// `"spec.containers[0].ports[0].number" failed custom validation because The port
// 15000 is reserved...`. Standalone the label is `value`.
func TestReservedPortMessage(t *testing.T) {
	_, err := validate.Validate(Port(0).ValidationRule(), float64(15000), nil, validate.Prefs{})
	if err == nil {
		t.Fatal("expected 15000 to be rejected")
	}
	const want = `"value" failed custom validation because ` +
		"The port 15000 is reserved for system use. Please choose a different port. " +
		"Reserved ports include: 8012,8022,9090,9091,15000,15001,15006,15020,15021,15090,41000"
	if err.Message != want {
		t.Fatalf("message mismatch:\n got: %s\nwant: %s", err.Message, want)
	}
}
