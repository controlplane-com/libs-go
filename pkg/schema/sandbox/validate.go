package sandbox

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strconv"

	"github.com/controlplane-com/libs-go/pkg/schema/base"
	"github.com/controlplane-com/libs-go/pkg/schema/port"
	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// SandboxReservedPorts are bound by the sandbox runtime in the dev container: sshd (22),
// ttyd (7681), code-server (8443), the ide gateway (8888), devboxd (49983) and the worker
// metrics port (49984, master a7e9e4899).
var SandboxReservedPorts = []int{22, 7681, 8443, 8888, 49983, 49984}

// ValidationRule enforces SandboxScaling's maxWorkers >= minWorkers (sandbox.ts
// SandboxScaling custom, master a7e9e4899); the per-field bounds/defaults are tags.
func (SandboxScaling) ValidationRule() *validate.Rule {
	return validate.RefineT(validate.Reflect(SandboxScaling{}), func(s *SandboxScaling, h *validate.Helper) error {
		if s.MaxWorkers != nil && s.MinWorkers != nil && *s.MaxWorkers < *s.MinWorkers {
			return h.Errorf("any.custom", "spec.sandbox.scaling.maxWorkers must be at least minWorkers")
		}
		return nil
	})
}

var (
	sandboxIdentityNameRe = regexp.MustCompile(`/identity/([^/]+)$`)
	sandboxSessionKeyRe   = regexp.MustCompile(`-([0-9a-f]{8})$`)
)

// SandboxPoolIdentityOf is the pool a worker belongs to (sandbox.ts sandboxPoolIdentityOf):
// the SandboxPoolTag value, or "" when absent/empty.
func SandboxPoolIdentityOf(tags base.Tags) string {
	if tags == nil {
		return ""
	}
	if v, ok := tags[SandboxPoolTag].(string); ok {
		return v
	}
	return ""
}

// SandboxPoolWorkerName is the Nth pool worker's name for an identity (sandboxPoolWorkerName).
func SandboxPoolWorkerName(identity string, n int) string {
	return SandboxChildName(identity, strconv.Itoa(n))
}

// SandboxIdentityNameOf is the identity name an identityLink points at, or "" (sandboxIdentityNameOf).
func SandboxIdentityNameOf(identityLink string) string {
	if m := sandboxIdentityNameRe.FindStringSubmatch(identityLink); m != nil {
		return m[1]
	}
	return ""
}

// SandboxWorkerKey is the stable 8-hex key of a worker name (sandboxWorkerKey).
func SandboxWorkerKey(workerName string) string { return digest8(workerName) }

// SandboxSessionWorkerKey is the worker key a pool session id ends in, or "" (sandboxSessionWorkerKey).
func SandboxSessionWorkerKey(sessionID string) string {
	if m := sandboxSessionKeyRe.FindStringSubmatch(sessionID); m != nil {
		return m[1]
	}
	return ""
}

// SandboxReservedEnvNames are owned by the sandbox runtime in the dev container; the
// overlay derives them and callers may not set them. CODE_SERVER_PASSWORD is retired as
// an injected value but stays reserved so a stored one is still redacted.
var SandboxReservedEnvNames = []string{"APP_PORT", "CODE_SERVER_PASSWORD", "DEVBOXD_TOKEN", "DEVBOXD_MAX_SESSIONS"}

// SandboxDerivedEnvFields maps each derived env entry to the section field it comes from,
// so the refusal can name where the value belongs (sandbox.ts SandboxDerivedEnvFields).
var SandboxDerivedEnvFields = map[string]string{
	"APP_PORT":             "spec.sandbox.appPort",
	"DEVBOXD_TOKEN":        "spec.sandbox.token",
	"DEVBOXD_MAX_SESSIONS": "spec.sandbox.scaling.maxSessions",
}

// SandboxCredentialEnvNames are the reserved names whose values are credentials; only
// these are redacted (sandbox.ts SandboxCredentialEnvNames). DEVBOXD_MAX_SESSIONS is
// reserved but not a credential, so it is derived and canonicalized yet never hidden.
var SandboxCredentialEnvNames = []string{"CODE_SERVER_PASSWORD", "DEVBOXD_TOKEN"}

// ValidationRule = Port.custom(reserved by the sandbox runtime).
func (SandboxPort) ValidationRule() *validate.Rule {
	return port.Port(0).ValidationRule().Custom(port.RejectReserved("by the sandbox runtime", SandboxReservedPorts))
}

// SandboxChildName is the deterministic child-object name, hash-tailed when the join
// would exceed the Name schema's 64 characters so two sandboxes cannot collide.
func SandboxChildName(parts ...string) string {
	full := ""
	for i, p := range parts {
		if i > 0 {
			full += "-"
		}
		full += p
	}
	if len(full) <= 64 {
		return full
	}
	return full[:55] + "-" + digest8(full)
}

func digest8(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])[:8]
}

// SandboxHomeVolumesetName is the sandbox's only child: the home volumeset, created and
// deleted with the workload.
func SandboxHomeVolumesetName(workloadName string) string {
	return SandboxChildName(workloadName, "state")
}

// HomeVolumeset is the home-volumeset body the server creates for a sandbox.
type HomeVolumeset struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Spec        HomeVolumesetSpec `json:"spec"`
}

type HomeVolumesetSpec struct {
	InitialCapacity  int    `json:"initialCapacity"`
	PerformanceClass string `json:"performanceClass"`
	FileSystemType   string `json:"fileSystemType"`
}

// SandboxHomeVolumeset is the shared-filesystem volumeset the server creates and owns.
func SandboxHomeVolumeset(workloadName string, sizeGb int) HomeVolumeset {
	if sizeGb == 0 {
		sizeGb = SandboxHomeVolumeDefaultGb
	}
	return HomeVolumeset{
		Name:        SandboxHomeVolumesetName(workloadName),
		Description: "Home volume for sandbox " + workloadName,
		Spec:        HomeVolumesetSpec{InitialCapacity: sizeGb, PerformanceClass: "shared", FileSystemType: "shared"},
	}
}

// ValidationRule reflects the section; every constraint is on the fields.
func (SandboxSpec) ValidationRule() *validate.Rule {
	// scaling is a worker-pool knob; a non-worker (dev) sandbox is a single box
	// (sandbox.ts, master a7e9e4899).
	return validate.RefineT(validate.Reflect(SandboxSpec{}), func(s *SandboxSpec, h *validate.Helper) error {
		if s.Scaling != nil && !SandboxRunsDaemon(string(s.Type)) {
			return h.Errorf("any.custom", "spec.sandbox.scaling: scaling is a worker feature; a %s sandbox is one box", s.Type)
		}
		return nil
	})
}
