package sandboxImage

// validate.go attaches validation to the canonical sandboxImage types via the
// Validator interface. sandboxImage.ts reflects from SandboxImage
// (Base.keys({spec: required, status})).
//
// TS reference: nodelibs/schema/src/sandboxImage.ts.

import (
	"regexp"

	"github.com/controlplane-com/libs-go/pkg/schema/base"
	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

var imageTagRe = regexp.MustCompile(`^[a-zA-Z0-9_][a-zA-Z0-9_\-.]*$`)

// ValidationRule reflects SandboxImage.
func (SandboxImage) ValidationRule() *validate.Rule {
	return validate.Reflect(SandboxImage{})
}

func (ImageTag) ValidationRule() *validate.Rule {
	return validate.String().Regex(imageTagRe).Max(128)
}

func (OrgImageLink) ValidationRule() *validate.Rule {
	return base.LocalLinkRule().Custom(base.ToKind("image"))
}
