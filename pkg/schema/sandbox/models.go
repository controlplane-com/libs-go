/* auto-generated */

package sandbox

type SandboxIde string

const (
	SandboxIdeVscode  SandboxIde = "vscode"
	SandboxIdeCursor  SandboxIde = "cursor"
	SandboxIdeSsh     SandboxIde = "ssh"
	SandboxIdeBrowser SandboxIde = "browser"
)

type SandboxScaling struct {
	MaxSessions     *float32 `json:"maxSessions,omitempty"`
	MinWorkers      *float32 `json:"minWorkers,omitempty"`
	MaxWorkers      *float32 `json:"maxWorkers,omitempty"`
	IdleDeleteDelay *float32 `json:"idleDeleteDelay,omitempty"`
}

type SandboxSpecVolume struct {
	InitialCapacity *float32 `json:"initialCapacity,omitempty"`
}

type SandboxSpecScaling struct {
	MaxSessions     *float32 `json:"maxSessions,omitempty"`
	MinWorkers      *float32 `json:"minWorkers,omitempty"`
	MaxWorkers      *float32 `json:"maxWorkers,omitempty"`
	IdleDeleteDelay *float32 `json:"idleDeleteDelay,omitempty"`
}

type SandboxSpec struct {
	Ide              SandboxIde          `json:"ide,omitempty"`
	Type             SandboxType         `json:"type,omitempty"`
	Password         string              `json:"password,omitempty"`
	Token            string              `json:"token,omitempty"`
	Volume           *SandboxSpecVolume  `json:"volume,omitempty"`
	AppPort          *float32            `json:"appPort,omitempty"`
	ScaleToZeroDelay *float32            `json:"scaleToZeroDelay,omitempty"`
	Scaling          *SandboxSpecScaling `json:"scaling,omitempty"`
	Ttl              string              `json:"ttl,omitempty"`
}

type SandboxType string

const (
	SandboxTypeDev    SandboxType = "dev"
	SandboxTypeWorker SandboxType = "worker"
)
