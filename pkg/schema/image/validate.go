package image

// validate.go attaches validation to the canonical image types via the Validator
// interface — the go-libs home of what used to be pkg/handlers/schema_image.go +
// the eager registration in schema_reflect_image_envoy.go. image.ts reflects from
// the Image struct; the createSchema fork (Image.fork(['name'], required) — name
// required on both create and patch write paths) is layered here so that
// validate.OfType(image.Image{}) is identical to the old registered `image`
// schema.
//
// The base Image struct keeps `name` optional (image.ts base is optional; only the
// createSchema fork makes it required), so the required presence is applied as a
// .Fork on the reflected rule rather than as a struct tag — mirroring the original
// ObjectSchema(FromType(Image).Fork("name", required)) registration.

import (
	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// ValidationRule reflects Image and layers the createSchema `name`-required fork.
// This is the schema the "image" kind validates against (create and patch).
func (Image) ValidationRule() *validate.Rule {
	return validate.Reflect(Image{}).
		Fork("name", func(r *validate.Rule) *validate.Rule { return r.Required() })
}

// ValidationRule reflects ImageSummary (served by the `-summary` sub-route; ported
// for completeness, not a registered write kind — name is not forked to required).
func (ImageSummary) ValidationRule() *validate.Rule {
	return validate.Reflect(ImageSummary{})
}
