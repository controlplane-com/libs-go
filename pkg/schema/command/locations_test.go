package command

import (
	"testing"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// TestDeleteVolumeSetLocationsMinOne pins command.ts:405
// `locations: Joi.array().items(Joi.string().required())` — a REQUIRED item means the
// array must contain at least one element, so an explicit [] is rejected. The port
// dropped the rule and persisted a deleteVolumeSet command with no locations.
func TestDeleteVolumeSetLocationsMinOne(t *testing.T) {
	// The typed path (the nested VolumeSet carries typed refines of its own).
	const vs = `"volumeSet":{"name":"v1","spec":{"initialCapacity":10,"performanceClass":"general-purpose-ssd"}}`
	dec := func(body string) *validate.Error {
		_, verr := validate.Decode[DeleteVolumeSetSpec](validate.Bytes(body), nil, validate.Prefs{})
		return verr
	}
	verr := dec(`{` + vs + `,"locations":[]}`)
	if verr == nil {
		t.Fatal("empty locations accepted")
	}
	// Joi blames the required ITEM schema, not a min-length rule the schema never had.
	if got := verr.Details()[0].Type; got != "array.includesRequiredUnknowns" {
		t.Errorf("empty locations: type %q, want array.includesRequiredUnknowns", got)
	}
	if got, want := verr.Details()[0].Message, `"locations" does not contain 1 required value(s)`; got != want {
		t.Errorf("empty locations: message %q, want %q", got, want)
	}
	if verr := dec(`{` + vs + `,"locations":["aws-us-east-1"]}`); verr != nil {
		t.Errorf("one location rejected: %v", verr)
	}
	// Absent is still fine — the array itself is optional.
	if verr := dec(`{` + vs + `}`); verr != nil {
		t.Errorf("absent locations rejected: %v", verr)
	}
}
