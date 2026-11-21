package volumeSet

import "testing"

func TestStatusLocationsIsNilSafe(t *testing.T) {
	var nilSet *VolumeSet
	if got := nilSet.StatusLocations(); got != nil {
		t.Errorf("nil volume set: %v", got)
	}
	if got := (&VolumeSet{}).StatusLocations(); got != nil {
		t.Errorf("no status: %v", got)
	}
	vs := &VolumeSet{Status: &VolumeSetStatus{Locations: []VolumeSetStatusLocation{{Name: "l1"}}}}
	if got := vs.StatusLocations(); len(got) != 1 || got[0].Name != "l1" {
		t.Errorf("with status: %v", got)
	}
}
