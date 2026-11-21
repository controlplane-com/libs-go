package sandbox

import (
	"regexp"
	"testing"

	"github.com/controlplane-com/libs-go/pkg/schema/base"
	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// Mirrors nodelibs/schema/src/sandbox.test.ts sandbox-pool helpers (master a7e9e4899).
func TestSandboxPoolHelpers(t *testing.T) {
	if got := SandboxPoolWorkerName("agents", 3); got != "agents-3" {
		t.Errorf("SandboxPoolWorkerName = %q, want agents-3", got)
	}
	longName := SandboxPoolWorkerName(repeat("a", 70), 12)
	if len(longName) > 64 || !regexp.MustCompile(`-[0-9a-f]{8}$`).MatchString(longName) {
		t.Errorf("long worker name = %q, want <=64 with -8hex tail", longName)
	}

	k3 := SandboxWorkerKey("agents-3")
	if !regexp.MustCompile(`^[0-9a-f]{8}$`).MatchString(k3) {
		t.Errorf("SandboxWorkerKey = %q, want 8 hex", k3)
	}
	if k3 == SandboxWorkerKey("agents-4") {
		t.Error("worker keys for different names collided")
	}

	if got := SandboxSessionWorkerKey("run-7-" + k3); got != k3 {
		t.Errorf("SandboxSessionWorkerKey = %q, want %q", got, k3)
	}
	for _, s := range []string{"run-7", "run-7-zzzzzzzz"} {
		if got := SandboxSessionWorkerKey(s); got != "" {
			t.Errorf("SandboxSessionWorkerKey(%q) = %q, want empty", s, got)
		}
	}

	if got := SandboxPoolIdentityOf(base.Tags{SandboxPoolTag: "agents"}); got != "agents" {
		t.Errorf("SandboxPoolIdentityOf = %q, want agents", got)
	}
	if got := SandboxPoolIdentityOf(base.Tags{}); got != "" {
		t.Errorf("SandboxPoolIdentityOf({}) = %q, want empty", got)
	}

	if got := SandboxIdentityNameOf("/org/o/gvc/g/identity/bot"); got != "bot" {
		t.Errorf("SandboxIdentityNameOf = %q, want bot", got)
	}
	for _, l := range []string{"", "/org/o/gvc/g/workload/bot"} {
		if got := SandboxIdentityNameOf(l); got != "" {
			t.Errorf("SandboxIdentityNameOf(%q) = %q, want empty", l, got)
		}
	}

	if SandboxPoolSessionIdMax != 55 {
		t.Errorf("SandboxPoolSessionIdMax = %d, want 55", SandboxPoolSessionIdMax)
	}
}

func TestSandboxScalingValidation(t *testing.T) {
	norm, err := validate.Decode[SandboxScaling](validate.Bytes("{}"), &validate.Context{}, validate.Prefs{})
	if err != nil {
		t.Fatalf("empty scaling: %v", err)
	}
	if norm.MaxSessions == nil || *norm.MaxSessions != 20 ||
		norm.MinWorkers == nil || *norm.MinWorkers != 1 ||
		norm.MaxWorkers == nil || *norm.MaxWorkers != 10 ||
		norm.IdleDeleteDelay == nil || *norm.IdleDeleteDelay != 3600 {
		t.Errorf("defaults = %+v, want 20/1/10/3600", norm)
	}

	for _, body := range []string{
		`{"minWorkers":3,"maxWorkers":2}`, // maxWorkers < minWorkers
		`{"idleDeleteDelay":10}`,          // below min 300
		`{"maxSessions":0}`,               // below min 1
	} {
		if _, err := validate.Decode[SandboxScaling](validate.Bytes(body), &validate.Context{}, validate.Prefs{}); err == nil {
			t.Errorf("%s accepted, want rejected", body)
		}
	}
}

func TestSandboxScalingIsWorkerOnly(t *testing.T) {
	// a worker may carry scaling; a dev sandbox may not.
	worker := `{"type":"worker","token":"tokentokentokentok","scaling":{}}`
	if _, err := validate.Decode[SandboxSpec](validate.Bytes(worker), &validate.Context{}, validate.Prefs{}); err != nil {
		t.Errorf("worker with scaling rejected: %v", err)
	}
	dev := `{"type":"dev","scaling":{}}`
	if _, err := validate.Decode[SandboxSpec](validate.Bytes(dev), &validate.Context{}, validate.Prefs{}); err == nil {
		t.Error("dev sandbox with scaling accepted, want rejected")
	}
}

func repeat(s string, n int) string {
	out := ""
	for i := 0; i < n; i++ {
		out += s
	}
	return out
}
