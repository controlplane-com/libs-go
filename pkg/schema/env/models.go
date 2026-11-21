/* hand-maintained — canonical source of truth for the env value model.
 * `joi:` struct tags drive pkg/validate.FromType. See nodelibs/schema/src/env.ts
 * and the tag DSL on go-libs/schema/ipSet/models.go.
 *
 * EnvCollection (the sort+dedupe array) is resolved via the base-type registry
 * (pkg/handlers registers env.EnvCollection -> Array(EnvVar).custom(sortDedupe));
 * the element customs live as field tags on EnvVar.
 */

package env

import "github.com/controlplane-com/libs-go/pkg/schema/base"

type EnvCollection []EnvVar

type EnvFromCollection []EnvFromSource

type EnvFromSourceSecretRef struct {
	Name base.Name `json:"name" validate:"required"`
}

// EnvFromSource references a dictionary (opaque key/value) secret whose keys are
// injected as environment variables (env.ts EnvFromSource/EnvFromCollection).
type EnvFromSource struct {
	SecretRef EnvFromSourceSecretRef `json:"secretRef" validate:"required"`
}

// EnvName is the env-var name value. Its Validator (validate.go) ports EnvName's
// custom (CPLN_ reserved + name regex) plus the field-level disallow of the
// reserved knative names. trim/max/required stay declarative on the field tag.
type EnvName string //ts:export

// EnvValue is the env-var value. Its Validator (validate.go) chains the three
// internal-reference customs (value ref, manifest ref, secret ref). max/allow
// stay declarative on the field tag.
type EnvValue string

type EnvVar struct {
	Name  EnvName  `json:"name" validate:"trim,max=120,required"`
	Value EnvValue `json:"value,omitzero" validate:"max=4096,allow="`
}
