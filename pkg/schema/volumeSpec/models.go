/* hand-maintained — canonical source of truth (nodelibs/schema/src/volumeSpec.ts).
 * `joi:` tags drive pkg/validate.FromType; DSL on go-libs/schema/ipSet/models.go. */

package volumeSpec

// VolumeSpecUri is the strongly-typed replacement for the
// `custom=volUriRef,custom=volUri` tag on VolumeSpec.Uri. The two customs run in
// declaration order: the generic internal-reference check, then the URI scheme
// check.
type VolumeSpecUri string

// VolumeSpecPath is the strongly-typed replacement for the `custom=volPath` tag
// on VolumeSpec.Path.
type VolumeSpecPath string

// VolumeSpecName is the strongly-typed replacement for the `pattern=volName` tag
// on VolumeSpec.Name (the max=63 modifier stays on the field tag).
type VolumeSpecName string

type VolumeSpecRecoveryPolicy string

const (
	VolumeSpecRecoveryPolicyRetain  VolumeSpecRecoveryPolicy = "retain"
	VolumeSpecRecoveryPolicyRecycle VolumeSpecRecoveryPolicy = "recycle"
)

func (VolumeSpecRecoveryPolicy) EnumValues() []string { return []string{"retain", "recycle"} }

type VolumeSpecBus string

const (
	VolumeSpecBusVirtio VolumeSpecBus = "virtio"
	VolumeSpecBusSata   VolumeSpecBus = "sata"
	VolumeSpecBusScsi   VolumeSpecBus = "scsi"
)

func (VolumeSpecBus) EnumValues() []string { return []string{"virtio", "sata", "scsi"} }

type VolumeSpec struct {
	Uri            VolumeSpecUri            `json:"uri" validate:"required"`
	RecoveryPolicy VolumeSpecRecoveryPolicy `json:"recoveryPolicy,omitzero" validate:"default=retain"`
	Path           VolumeSpecPath           `json:"path,omitzero"`
	Name           VolumeSpecName           `json:"name,omitzero" validate:"max=63"`
	Bus            VolumeSpecBus            `json:"bus,omitzero"`
	BootOrder      *float64                 `json:"bootOrder,omitzero" validate:"integer,min=1,max=16"`
}
