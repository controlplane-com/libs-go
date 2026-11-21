package volumeSet

// StatusLocations is Status.Locations, nil when the volume set has no status yet — a
// freshly created one does not, and Status is a pointer so absent stays distinguishable.
// Read locations through this: ranging over v.Status.Locations compiles and then panics.
func (v *VolumeSet) StatusLocations() []VolumeSetStatusLocation {
	if v == nil || v.Status == nil {
		return nil
	}
	return v.Status.Locations
}
