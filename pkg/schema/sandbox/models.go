/* auto-generated */

package sandbox

type SandboxIde string

const (
	SandboxIdeVscode  SandboxIde = "vscode"
	SandboxIdeCursor  SandboxIde = "cursor"
	SandboxIdeSsh     SandboxIde = "ssh"
	SandboxIdeBrowser SandboxIde = "browser"
)

type SandboxSpecVolume struct {
	InitialCapacity *float32 `json:"initialCapacity,omitempty"`
}

type SandboxSpec struct {
	Ide              SandboxIde        `json:"ide,omitempty"`
	Type             SandboxType       `json:"type,omitempty"`
	Password         string            `json:"password,omitempty"`
	Token            string            `json:"token,omitempty"`
	Volume           SandboxSpecVolume `json:"volume,omitempty"`
	AppPort          *float32          `json:"appPort,omitempty"`
	ScaleToZeroDelay *float32          `json:"scaleToZeroDelay,omitempty"`
	Ttl              string            `json:"ttl,omitempty"`
}

type SandboxType string

const (
	SandboxTypeDev    SandboxType = "dev"
	SandboxTypeWorker SandboxType = "worker"
)
