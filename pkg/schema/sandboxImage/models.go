/* hand-maintained — canonical source of truth for the `sandboximage` resource
 * schema. Ported from nodelibs/schema/src/sandboxImage.ts (master 48c49467c),
 * replacing the joi-to-go emission.
 *
 * `validate:` struct tags carry the rules validate.Reflect can express; the
 * ImageTag pattern, the two link types and the phase enum live in validate.go.
 * See go-libs/schema/ipSet/models.go for the tag DSL.
 */

package sandboxImage

import "github.com/controlplane-com/libs-go/pkg/schema/base"

// ImageTag ports `Joi.string().regex(/^[a-zA-Z0-9_][a-zA-Z0-9_\-.]*$/).max(128)` —
// the tag of the built image, an OCI tag without the leading separator classes.
type ImageTag string

// OrgImageLink ports `LocalLink.custom(toKind('image'))`: the org image the latest
// successful build produced.
type OrgImageLink string

// SandboxImageTags aliases base.Tags so the reflection base-type registry resolves
// it to the shared TagsRule.
type SandboxImageTags = base.Tags

// SandboxImageSpecBase carries the sole tenant-controlled build input. `base` is
// itself optional — omitting it supplements the standard sandbox base.
type SandboxImageSpecBase struct {
	Image base.ImageLink `json:"image,omitzero" validate:"optional"`
}

type SandboxImageSpec struct {
	Base *SandboxImageSpecBase `json:"base,omitzero" validate:"optional"`
	Tag  ImageTag              `json:"tag,omitzero" validate:"default=latest"`
	Type SandboxImageSpecType  `json:"type,omitzero" validate:"default=dev"`
}

// SandboxImageSpecType: dev builds the base plus the sandbox layer; session copies the base alone, for sessions.
type SandboxImageSpecType string

func (SandboxImageSpecType) EnumValues() []string { return []string{"dev", "session"} }

type SandboxImageStatusPhase string

const (
	SandboxImageStatusPhasePending  SandboxImageStatusPhase = "pending"
	SandboxImageStatusPhaseQueued   SandboxImageStatusPhase = "queued"
	SandboxImageStatusPhaseBuilding SandboxImageStatusPhase = "building"
	SandboxImageStatusPhaseReady    SandboxImageStatusPhase = "ready"
	SandboxImageStatusPhaseFailed   SandboxImageStatusPhase = "failed"
)

func (SandboxImageStatusPhase) EnumValues() []string {
	return []string{
		string(SandboxImageStatusPhasePending), string(SandboxImageStatusPhaseQueued),
		string(SandboxImageStatusPhaseBuilding), string(SandboxImageStatusPhaseReady),
		string(SandboxImageStatusPhaseFailed),
	}
}

type SandboxImageStatus struct {
	Phase            SandboxImageStatusPhase `json:"phase,omitzero"`
	ImageLink        OrgImageLink            `json:"imageLink,omitzero" validate:"optional"`
	ImageRef         string                  `json:"imageRef,omitzero" validate:"optional"`
	BuildId          string                  `json:"buildId,omitzero" validate:"optional"`
	BuiltFingerprint string                  `json:"builtFingerprint,omitzero" validate:"optional"`
	Error            string                  `json:"error,omitzero" validate:"allow=,optional"`
	BuildDuration    *float64                `json:"buildDuration,omitzero" validate:"optional"`
	LastBuildAt      string                  `json:"lastBuildAt,omitzero" validate:"date,optional"`
}

type SandboxImage struct {
	Id           string              `json:"id,omitzero" validate:"uuid,readOnly"`
	Name         base.Name           `json:"name,omitzero"`
	Kind         base.Kind           `json:"kind,omitzero"`
	Version      *float64            `json:"version,omitzero" validate:"positive,integer,readOnly"`
	Description  string              `json:"description,omitzero" validate:"trim,max=250,allow="`
	Tags         SandboxImageTags    `json:"tags,omitzero" validate:"default={}"`
	Created      string              `json:"created,omitzero" validate:"date,readOnly"`
	LastModified string              `json:"lastModified,omitzero" validate:"date,readOnly"`
	Links        base.Links          `json:"links,omitzero" validate:"readOnly"`
	Spec         SandboxImageSpec    `json:"spec" validate:"required"`
	Status       *SandboxImageStatus `json:"status,omitzero"`
}
