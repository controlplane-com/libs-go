/* auto-generated */

package sandboxImage

import "github.com/controlplane-com/libs-go/pkg/schema/base"

type SandboxImageTags map[string]any

type SandboxImage struct {
	Id           string             `json:"id,omitempty"`
	Name         base.Name          `json:"name,omitempty"`
	Kind         base.Kind          `json:"kind,omitempty"`
	Version      *float32           `json:"version,omitempty"`
	Description  string             `json:"description,omitempty"`
	Tags         SandboxImageTags   `json:"tags,omitempty"`
	Created      string             `json:"created,omitempty"`
	LastModified string             `json:"lastModified,omitempty"`
	Links        base.Links         `json:"links,omitempty"`
	Spec         SandboxImageSpec   `json:"spec"`
	Status       SandboxImageStatus `json:"status,omitempty"`
}

type SandboxImageSpecBase struct {
	Image string `json:"image,omitempty"`
}

type SandboxImageSpec struct {
	Base *SandboxImageSpecBase `json:"base,omitempty"`
	Tag  string                `json:"tag,omitempty"`
}

type SandboxImageStatusPhase string

const (
	SandboxImageStatusPhasePending  SandboxImageStatusPhase = "pending"
	SandboxImageStatusPhaseQueued   SandboxImageStatusPhase = "queued"
	SandboxImageStatusPhaseBuilding SandboxImageStatusPhase = "building"
	SandboxImageStatusPhaseReady    SandboxImageStatusPhase = "ready"
	SandboxImageStatusPhaseFailed   SandboxImageStatusPhase = "failed"
)

type SandboxImageStatus struct {
	Phase            SandboxImageStatusPhase `json:"phase,omitempty"`
	ImageLink        string                  `json:"imageLink,omitempty"`
	ImageRef         string                  `json:"imageRef,omitempty"`
	BuildId          string                  `json:"buildId,omitempty"`
	BuiltFingerprint string                  `json:"builtFingerprint,omitempty"`
	Error            string                  `json:"error,omitempty"`
	BuildDuration    *float32                `json:"buildDuration,omitempty"`
	LastBuildAt      string                  `json:"lastBuildAt,omitempty"`
}
