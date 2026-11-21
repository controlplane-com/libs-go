package quota

// validate.go attaches validation to the canonical quota type via the Validator
// interface — the go-libs home of what used to be pkg/handlers/schema_quota.go
// (QuotaSchema, a FromType shim) + its RegisterReflect("quota", ...) registration.
// quota.ts is a plain Joi.object with no cross-field refines and no `.unknown(true)`,
// so the schema is a straight reflection of the tagged Quota struct with default
// unknown-field rejection.

import "github.com/controlplane-com/libs-go/pkg/schema/validate"

// ValidationRule reflects Quota — the schema the "quota" kind validates against.
func (Quota) ValidationRule() *validate.Rule { return validate.Reflect(Quota{}) }
