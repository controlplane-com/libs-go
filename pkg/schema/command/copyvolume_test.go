package command

import (
	"testing"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// TestCopyVolumeSpecCustom pins command.ts CopyVolumeSpec's custom: timeoutSeconds
// defaults to 600, the ends may share a location only when the copy goes to another
// volume set, and `locations` is always derived rather than trusted.
func TestCopyVolumeSpecCustom(t *testing.T) {
	decode := func(body string) (*CopyVolumeSpec, *validate.Error) {
		return validate.Decode[CopyVolumeSpec](validate.Bytes(body), nil, validate.Prefs{})
	}

	spec, verr := decode(`{"location":"l1","targetLocation":"l2","volumeIndex":0,"locations":["bogus"]}`)
	if verr != nil {
		t.Fatal(verr)
	}
	if spec.TimeoutSeconds == nil || *spec.TimeoutSeconds != 600 {
		t.Errorf("timeoutSeconds = %v, want 600", spec.TimeoutSeconds)
	}
	if got := spec.Locations; len(got) != 2 || got[0] != "l1" || got[1] != "l2" {
		t.Errorf("locations = %v, want [l1 l2] (derived, the caller's value ignored)", got)
	}
	if spec.AcknowledgeDestinationVolumeDeletion == nil || *spec.AcknowledgeDestinationVolumeDeletion {
		t.Errorf("acknowledgeDestinationVolumeDeletion should default to false")
	}

	// same location, no other volume set: the two ends would be the same volume
	if _, verr := decode(`{"location":"l1","targetLocation":"l1","volumeIndex":0}`); verr == nil {
		t.Error("same-location copy within one volume set was accepted")
	} else if got := verr.Details()[0].Type; got != "any.custom" {
		t.Errorf("type %q, want any.custom", got)
	}

	// same location INTO another volume set is a legitimate in-cluster copy
	spec, verr = decode(`{"location":"l1","targetLocation":"l1","volumeIndex":0,"targetVolumeSet":"//volumeset/other"}`)
	if verr != nil {
		t.Fatal(verr)
	}
	if got := spec.Locations; len(got) != 1 || got[0] != "l1" {
		t.Errorf("locations = %v, want [l1]", got)
	}
}
