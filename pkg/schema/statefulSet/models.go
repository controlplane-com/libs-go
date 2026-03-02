/* auto-generated */

package statefulSet

import "github.com/controlplane-com/libs-go/pkg/schema/volumeSet"

type StatefulSetStatus struct {
	ReplicaCount *float32                  `json:"replicaCount,omitempty"`
	VolumeSet    volumeSet.VolumeSetStatus `json:"volumeSet,omitempty"`
}
