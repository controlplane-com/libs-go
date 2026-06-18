/* auto-generated */

package volumeSpec

type VolumeSpecRecoveryPolicy string

const (
	VolumeSpecRecoveryPolicyRetain  VolumeSpecRecoveryPolicy = "retain"
	VolumeSpecRecoveryPolicyRecycle VolumeSpecRecoveryPolicy = "recycle"
)

type VolumeSpecBus string

const (
	VolumeSpecBusVirtio VolumeSpecBus = "virtio"
	VolumeSpecBusSata   VolumeSpecBus = "sata"
	VolumeSpecBusScsi   VolumeSpecBus = "scsi"
)

type VolumeSpec struct {
	Uri            string                   `json:"uri"`
	RecoveryPolicy VolumeSpecRecoveryPolicy `json:"recoveryPolicy,omitempty"`
	Path           string                   `json:"path,omitempty"`
	Name           string                   `json:"name,omitempty"`
	Bus            VolumeSpecBus            `json:"bus,omitempty"`
	BootOrder      *float32                 `json:"bootOrder,omitempty"`
}
