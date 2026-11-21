/* hand-maintained — canonical source of truth for the `image` resource schema.
 *
 * Ported from nodelibs/schema/src/image.ts. `validate:` struct tags drive
 * pkg/validate.FromType, which reflects these types into the *validate.Rule tree
 * (replacing the hand-written pkg/handlers/schema_image.go builder). The
 * createSchema fork (name required on create/patch) is applied at the registration
 * call site (pkg/handlers/schema_reflect_image_envoy.go), not on the base struct,
 * so the base Image keeps `name` optional. Tag DSL: see go-libs/schema/ipSet/models.go.
 *
 * NOTE: the container image-STRING format validation (repository/tag/digest of an
 * actual image reference — checkImageAndTag) lives in pkg/handlers/schema_base.go
 * and is NOT part of this resource schema; here `name` is a plain string (image
 * names may contain '/' and ':') exactly as image.ts.
 */

package image

import "github.com/controlplane-com/libs-go/pkg/schema/base"

// ImageTags aliases base.Tags so the reflection base-type registry (keyed on
// base.Tags) resolves it to the shared TagsRule.
type ImageTags = base.Tags

// ImageManifest is a free-form pattern object (Joi.object().pattern(string, any)).
type ImageManifest map[string]any

type Image struct {
	Id           string        `json:"id,omitzero" validate:"uuid"`
	Name         string        `json:"name,omitzero"`
	Kind         base.Kind     `json:"kind,omitzero"`
	Version      *float64      `json:"version,omitzero" validate:"positive"`
	Created      string        `json:"created,omitzero" validate:"date"`
	LastModified string        `json:"lastModified,omitzero" validate:"date"`
	Tags         ImageTags     `json:"tags,omitzero" validate:"default={}"`
	Links        base.Links    `json:"links,omitzero"`
	Repository   string        `json:"repository,omitzero" validate:"allow=,readOnly"`
	Tag          string        `json:"tag,omitzero" validate:"readOnly"`
	Manifest     ImageManifest `json:"manifest,omitzero" validate:"readOnly"`
	Digest       string        `json:"digest,omitzero" validate:"readOnly"`
}

type ImageSummaryRepositories struct {
	Name      string   `json:"name,omitzero"`
	TagsCount *float64 `json:"tagsCount,omitzero"`
}

type ImageSummary struct {
	Kind         base.Kind                  `json:"kind,omitzero"`
	Links        base.Links                 `json:"links,omitzero"`
	Repositories []ImageSummaryRepositories `json:"repositories,omitzero"`
}
