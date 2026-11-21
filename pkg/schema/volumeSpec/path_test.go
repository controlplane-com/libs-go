package volumeSpec

import (
	"testing"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// TestVolumePathRootNormalizesEmpty pins volumeSpec.ts:85-97: a path that trims to the
// empty string (e.g. "/") is returned as "" — the workload custom then rejects it as
// "volumes[].path is required", so a mount over the container root is refused. Go used
// to remap "" back to "/" and accept it.
func TestVolumePathRootNormalizesEmpty(t *testing.T) {
	r := VolumeSpecPath("").ValidationRule()
	for _, in := range []string{"/", "//", "/."} {
		got, err := r.ValidateAt([]string{"path"}, in, nil, validate.Prefs{})
		if err != nil {
			t.Fatalf("%q errored: %v", in, err)
		}
		if got != "" {
			t.Fatalf("%q normalized to %q, want \"\"", in, got)
		}
	}
	// A reserved path is still rejected; a real path still passes through cleaned.
	if _, err := r.ValidateAt([]string{"path"}, "/tmp/", nil, validate.Prefs{}); err == nil {
		t.Fatal("/tmp (reserved) was accepted")
	}
	if got, err := r.ValidateAt([]string{"path"}, "/data/", nil, validate.Prefs{}); err != nil || got != "/data" {
		t.Fatalf("/data/ => %q, %v; want \"/data\", nil", got, err)
	}
}
