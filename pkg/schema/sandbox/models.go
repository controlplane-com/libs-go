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
	Size *float32 `json:"size,omitempty"`
}

type SandboxSpec struct {
	Ide              SandboxIde         `json:"ide,omitempty"`
	AppPort          *float32           `json:"appPort,omitempty"`
	Volume           *SandboxSpecVolume `json:"volume,omitempty"`
	ScaleToZeroDelay *float32           `json:"scaleToZeroDelay,omitempty"`
	Ttl              string             `json:"ttl,omitempty"`
}

type SandboxVolume struct {
	Size *float32 `json:"size,omitempty"`
}
