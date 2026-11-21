package sandbox

import "testing"

// SandboxChildName keeps a hash tail so two sandboxes cannot collide after truncation,
// and every child name fits the Name schema's 64 characters (sandbox.test.ts).
func TestSandboxChildName(t *testing.T) {
	long := ""
	for i := 0; i < 80; i++ {
		long += "x"
	}
	a := SandboxChildName(long, "state")
	b := SandboxChildName(long+"y", "state")
	if len(a) > 64 || len(b) > 64 {
		t.Errorf("names exceed 64: %d, %d", len(a), len(b))
	}
	if a == b {
		t.Error("two different sandboxes truncated to the same child name")
	}
}

// The home volumeset is the sandbox's only child: <workload>-state, a shared-filesystem
// volumeset the server creates and owns (sandbox.ts).
func TestSandboxHomeVolumeset(t *testing.T) {
	if got := SandboxHomeVolumesetName("example"); got != "example-state" {
		t.Errorf("home volumeset name = %q, want example-state", got)
	}
	vs := SandboxHomeVolumeset("example", 0)
	if vs.Name != "example-state" || vs.Spec.InitialCapacity != SandboxHomeVolumeDefaultGb ||
		vs.Spec.PerformanceClass != "shared" || vs.Spec.FileSystemType != "shared" {
		t.Errorf("home volumeset = %+v", vs)
	}
	if vs := SandboxHomeVolumeset("example", 50); vs.Spec.InitialCapacity != 50 {
		t.Errorf("sizeGb not honored: %+v", vs)
	}
}
