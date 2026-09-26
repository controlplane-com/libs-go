/* auto-generated */

package sandbox

type SandboxIde string

const (
	SandboxIdeVscode  SandboxIde = "vscode"
	SandboxIdeCursor  SandboxIde = "cursor"
	SandboxIdeSsh     SandboxIde = "ssh"
	SandboxIdeBrowser SandboxIde = "browser"
)

type SandboxSpecVolumePerformanceClass string

const (
	SandboxSpecVolumePerformanceClassGeneralPurposeSsd SandboxSpecVolumePerformanceClass = "general-purpose-ssd"
	SandboxSpecVolumePerformanceClassHighThroughputSsd SandboxSpecVolumePerformanceClass = "high-throughput-ssd"
)

type SandboxSpecVolume struct {
	Size             *float32                          `json:"size,omitempty"`
	PerformanceClass SandboxSpecVolumePerformanceClass `json:"performanceClass,omitempty"`
}

type SandboxSpec struct {
	Ide              SandboxIde         `json:"ide,omitempty"`
	AppPort          *float32           `json:"appPort,omitempty"`
	Volume           *SandboxSpecVolume `json:"volume,omitempty"`
	ScaleToZeroDelay *float32           `json:"scaleToZeroDelay,omitempty"`
	Ttl              string             `json:"ttl,omitempty"`
}

type SandboxVolumePerformanceClass string

const (
	SandboxVolumePerformanceClassGeneralPurposeSsd SandboxVolumePerformanceClass = "general-purpose-ssd"
	SandboxVolumePerformanceClassHighThroughputSsd SandboxVolumePerformanceClass = "high-throughput-ssd"
)

type SandboxVolume struct {
	Size             *float32                      `json:"size,omitempty"`
	PerformanceClass SandboxVolumePerformanceClass `json:"performanceClass,omitempty"`
}
