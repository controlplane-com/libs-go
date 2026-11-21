/* hand-maintained — canonical source of truth for the `sandbox` section carried at
 * workload.spec.sandbox. A sandbox is NOT a kind: it is a stateful workload whose
 * presence of this section makes it one. The write-time overlay that normalizes the
 * section into the workload (gateway/devboxd port, boot env, home mount, marker tags)
 * needs the workload types, so it lives beside them in the workload package; this
 * package holds the section, its rules and the names the overlay and data-service share.
 *
 * See nodelibs/schema/src/sandbox.ts for the Joi source. */

package sandbox

type SandboxIde string

const (
	SandboxIdeVscode  SandboxIde = "vscode"
	SandboxIdeCursor  SandboxIde = "cursor"
	SandboxIdeSsh     SandboxIde = "ssh"
	SandboxIdeBrowser SandboxIde = "browser"
)

var SandboxIdes = []string{"vscode", "cursor", "ssh", "browser"}

func (SandboxIde) EnumValues() []string { return SandboxIdes }

// SandboxType selects the sandbox profile: dev (human devbox, browser IDE behind a
// Control Plane sign-in) or worker (agent box, the devboxd API gated by the token).
type SandboxType string

const (
	SandboxTypeDev    SandboxType = "dev"
	SandboxTypeWorker SandboxType = "worker"
)

var SandboxTypes = []string{"dev", "worker"}

func (SandboxType) EnumValues() []string { return SandboxTypes }

// SandboxRunsDaemon reports whether the profile runs devboxd as its only listener
// (worker), secured by the token.
func SandboxRunsDaemon(sandboxType string) bool { return sandboxType == "worker" }

// SandboxPort is Port minus the ports the sandbox runtime binds inside the dev
// container; the rule lives in validate.go.
type SandboxPort float64

// SandboxVolume is the home volumeset's sole knob. initialCapacity is the GB size the
// server creates it at; immutable after create (resize the volumeset itself).
type SandboxVolume struct {
	InitialCapacity *float64 `json:"initialCapacity,omitzero" validate:"integer,min=10,default=20"`
}

type SandboxSpec struct {
	Ide  SandboxIde  `json:"ide,omitzero" validate:"default=browser"`
	Type SandboxType `json:"type,omitzero" validate:"default=dev"`
	// retired; stripped not rejected, so legacy rows and older clients stay writable
	Password         string          `json:"password,omitzero" validate:"strip"`
	Token            string          `json:"token,omitzero" validate:"min=16,max=512"`
	Volume           *SandboxVolume  `json:"volume,omitzero" validate:"default={}"`
	AppPort          *SandboxPort    `json:"appPort,omitzero" validate:"default=8080"`
	ScaleToZeroDelay *float64        `json:"scaleToZeroDelay,omitzero" validate:"min=30,max=3600,default=300"`
	Scaling          *SandboxScaling `json:"scaling,omitzero"`
	// retired; stripped not rejected
	Ttl string `json:"ttl,omitzero" validate:"strip"`
}

const (
	// SandboxGatewayPort is the ide gateway (dev type), the reserved port the overlay
	// declares on the dev container.
	SandboxGatewayPort = 8888
	// SandboxDevboxdPort is the devboxd HTTP API, the worker's only listener.
	SandboxDevboxdPort = 49983

	// SandboxStandardImage is the dev container image injected when a dev sandbox is
	// created without containers; SandboxWorkerStandardImage is the worker equivalent
	// (devboxd as PID 1).
	SandboxStandardImage       = "build.cpln.io/public/sandbox-base:v1"
	SandboxWorkerStandardImage = "build.cpln.io/public/agentbox-base:v1"
	// The injected container carries its own resources: the workload defaults
	// (50m/128Mi) cannot run the IDE. The worker is sized for the sessions it holds,
	// not one IDE (sandbox.ts SandboxWorkerStandardContainer, master 115ed3912).
	SandboxStandardContainerName      = "dev"
	SandboxStandardContainerCpu       = "1000m"
	SandboxStandardContainerMem       = "2Gi"
	SandboxWorkerStandardContainerCpu = "4000m"
	SandboxWorkerStandardContainerMem = "8Gi"

	SandboxHomeVolumeDefaultGb     = 20
	SandboxScaleToZeroDelayDefault = 300

	// SandboxReservedMountPath is where the overlay mounts the home volumeset.
	SandboxReservedMountPath = "/home/devuser"

	// Tags the overlay maintains. SandboxMarkerTag is present iff spec.sandbox is, so
	// sandbox list queries never see stale rows.
	SandboxMarkerTag      = "cpln/sandbox"
	SandboxDevTypeTag     = "cpln/dev-type"
	SandboxDevTypeValue   = "devbox"
	SandboxIdePasswordTag = "cpln/dev-ide-password"

	// E2B sandbox pools (master a7e9e4899). SandboxPoolTag marks a worker as part of an
	// identity's (or the default) pool; the metrics port is the worker's per-session gauge.
	SandboxPoolTag = "cpln/sandbox-pool"
	// SandboxFirewallTag is the fingerprint of the session firewall a server-added worker was created for.
	SandboxFirewallTag           = "cpln/sandbox-firewall"
	SandboxMeshInboundExcludeTag = "cpln/disableServiceMeshInboundPort"
	SandboxMetricsPort           = 49984
	SandboxWorkerMetricsPort     = 49984
	// SandboxPoolSessionIdMax leaves room for the "-<8 hex worker key>" suffix in a 64-char id.
	SandboxPoolSessionIdMax = 64 - 9

	SandboxScalingMaxSessionsDefault     = 20
	SandboxScalingMinWorkersDefault      = 1
	SandboxScalingMaxWorkersDefault      = 10
	SandboxScalingIdleDeleteDelayDefault = 3600

	// The GVC every sandbox is created in (sandbox.ts, master 115ed3912). The fallback
	// name serves while the first is held by a GVC placed elsewhere; the location is
	// where kata runs (the default, SANDBOX_GVC_LOCATION may override it); the tag marks
	// the org's chosen GVC so a later GVC under the other name does not take over.
	SandboxGvcName         = "sandboxes"
	SandboxGvcFallbackName = "cpln-sandboxes"
	SandboxGvcLocation     = "oci-us-ashburn-1"
	SandboxGvcTag          = "cpln/sandboxGvc"
)

// SandboxScaling is the worker-pool policy for a worker sandbox's identity (sandbox.ts
// SandboxScaling, master a7e9e4899). maxWorkers must be >= minWorkers (validate.go).
type SandboxScaling struct {
	MaxSessions     *float64 `json:"maxSessions,omitzero" validate:"integer,min=1,max=500,default=20"`
	MinWorkers      *float64 `json:"minWorkers,omitzero" validate:"integer,min=1,max=100,default=1"`
	MaxWorkers      *float64 `json:"maxWorkers,omitzero" validate:"integer,min=1,max=100,default=10"`
	IdleDeleteDelay *float64 `json:"idleDeleteDelay,omitzero" validate:"integer,min=300,max=604800,default=3600"`
}
