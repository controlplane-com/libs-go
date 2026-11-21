package auditctx

// validate.go attaches validation to the canonical auditctx types via the
// Validator interface — the go-libs home of what used to be
// pkg/handlers/schema_auditctx.go + the auditctx registrations in
// schema_reflect_orggvc.go. auditctx.ts reflects from AuditContext
// (Base.keys({status, origin})). AuditContextStatus is Joi.object({}).unknown(true).

import (
	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// ValidationRule reflects AuditContext.
func (AuditContext) ValidationRule() *validate.Rule {
	return validate.Reflect(AuditContext{})
}

// ValidationRule is Joi.object({}).unknown(true) — a free-form status object that
// accepts any keys (was RegisterObjectRules{Unknown: true} on this type).
func (AuditContextStatus) ValidationRule() *validate.Rule {
	return validate.Object(nil).Unknown(true)
}
