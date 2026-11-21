package tracing

// validate.go attaches validation to the canonical tracing types via the
// Validator interface. The LightstepTracing.credentials localLink leaf rule
// (was a localLink=secret struct tag) lives here as a strongly-typed field type.

import (
	"github.com/controlplane-com/libs-go/pkg/schema/base"
	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// ValidationRule implements the Validator interface for SecretLink.
func (SecretLink) ValidationRule() *validate.Rule {
	return base.LocalLinkRule().Custom(base.ToKind("secret"))
}

// ValidationRule ports tracing.ts:29-33 `provider: Joi.object({...}).xor('lightstep',
// 'otel','controlplane')`. The xor was recorded only in a models.go comment pointing at
// a handler file that no longer exists, so a provider block with NONE or SEVERAL
// providers was accepted and persisted.
func (TracingProvider) ValidationRule() *validate.Rule {
	return validate.Reflect(TracingProvider{}).Xor("lightstep", "otel", "controlplane")
}

// ValidationRule ports tracing.ts:34's top-level custom. Its mirroring half (moving a
// deprecated top-level `lightstep` under `provider`) runs in the handler after
// validation, and since master 314afc47f the custom reads `value.provider?.lightstep`,
// so a block carrying neither `lightstep` nor `provider` is accepted rather than
// crashing into a 400. Nothing here rejects.
func (Tracing) ValidationRule() *validate.Rule {
	return validate.Reflect(Tracing{})
}
