/* hand-maintained — canonical source of truth for the `valkey` schema.
 * valkey.ts emitted no Go structs (absent from the joi-to-go set), so these are
 * authored directly. `validate:` tags drive pkg/validate.FromType reflection
 * (replaces the schema_valkey.go builder). See nodelibs/schema/src/valkey.ts.
 *
 * Leaf types ValkeyMemory/ValkeyCpu/ValkeyLocationLink/ValkeyReplicas are self-
 * validating via the Validator interface (see validate.go): the workload Memory/
 * Cpu rules, LocalLink+toKind('location'), and the replicas-must-be-odd check
 * respectively. */

package valkey

import "github.com/controlplane-com/libs-go/pkg/schema/base"

// ValkeyMemory / ValkeyCpu are self-validating (see validate.go) to the workload
// Memory / Cpu rules; ValkeyLocationLink to LocalLink.custom(toKind('location'));
// ValkeyReplicas carries the replicas-must-be-odd custom.
type ValkeyMemory string
type ValkeyCpu string
type ValkeyLocationLink string
type ValkeyReplicas float64

// ValkeyOptions ports `ValkeyOptions` (valkey.ts:7).
type ValkeyOptions struct {
	Memory   ValkeyMemory    `json:"memory,omitzero" validate:"default=256Mi"`
	Cpu      ValkeyCpu       `json:"cpu,omitzero" validate:"default=50m"`
	Suspend  *bool           `json:"suspend,omitzero" validate:"default=false"`
	Replicas *ValkeyReplicas `json:"replicas,omitzero" validate:"integer,min=3"`
}

// ValkeyLocalOptions ports `ValkeyOptions.append({location})` (valkey.ts:18).
type ValkeyLocalOptions struct {
	Memory   ValkeyMemory       `json:"memory,omitzero" validate:"default=256Mi"`
	Cpu      ValkeyCpu          `json:"cpu,omitzero" validate:"default=50m"`
	Suspend  *bool              `json:"suspend,omitzero" validate:"default=false"`
	Replicas *ValkeyReplicas    `json:"replicas,omitzero" validate:"integer,min=3"`
	Location ValkeyLocationLink `json:"location,omitzero"`
}

// ValkeyPersistenceOptions ports `ValkeyPersistenceOptions` (valkey.ts:22).
type ValkeyPersistenceOptions struct {
	Enabled  *bool    `json:"enabled,omitzero" validate:"default=false"`
	Capacity *float64 `json:"capacity,omitzero" validate:"required,default=10,integer,min=10"`
}

type ValkeySpecMode string

const (
	ValkeySpecModeSentinel ValkeySpecMode = "sentinel"
	ValkeySpecModeCluster  ValkeySpecMode = "cluster"
)

type ValkeySpecEvictionPolicy string

// ValkeySpec ports `ValkeySpec` (valkey.ts:31).
type ValkeySpec struct {
	Mode                  ValkeySpecMode            `json:"mode,omitzero" validate:"required,valid=sentinel|cluster"`
	Persistence           *ValkeyPersistenceOptions `json:"persistence,omitzero"`
	DefaultOptions        ValkeyOptions             `json:"defaultOptions" validate:"required"`
	LocalOptions          []ValkeyLocalOptions      `json:"localOptions,omitzero"`
	EvictionPolicy        ValkeySpecEvictionPolicy  `json:"evictionPolicy,omitzero" validate:"valid=noeviction|allkey-lru|allkeys-lfu|volatile-lru|volatile-lfu|allkeys-random|volatile-random|volatile-ttl"`
	AdvancedConfiguration string                    `json:"advancedConfiguration,omitzero" validate:"default=,allow="`
}

// ValkeyStatus ports `Joi.object({})`.
type ValkeyStatus struct{}

// Valkey ports `Valkey = Joi.object({name, gvc, spec, status})` (valkey.ts:42).
// NOTE: valkey is a plain object (NOT Base.keys); name is Name.max(49).
type Valkey struct {
	Name   base.Name     `json:"name,omitzero" validate:"max=49"`
	Gvc    any           `json:"gvc,omitzero" validate:"allow="`
	Spec   *ValkeySpec   `json:"spec,omitzero"`
	Status *ValkeyStatus `json:"status,omitzero"`
}
