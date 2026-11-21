/* hand-maintained — canonical source of truth for the StatefulSetStatus sub-schema.
 * `validate:` tags drive pkg/validate.FromType reflection. See
 * nodelibs/schema/src/statefulSet.ts: StatefulSetStatus = Joi.object({replicaCount:
 * number.integer.optional, volumeSet: VolumeSetStatus}). The volumeSet field embeds
 * the volumeSet-owned VolumeSetStatus (reflected from its tags). This is a shared SUB
 * schema (not a registered kind). */

package statefulSet

import "github.com/controlplane-com/libs-go/pkg/schema/volumeSet"

type StatefulSetStatus struct {
	ReplicaCount *float64                   `json:"replicaCount,omitzero" validate:"integer"`
	VolumeSet    *volumeSet.VolumeSetStatus `json:"volumeSet,omitzero"`
}
