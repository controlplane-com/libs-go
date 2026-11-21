package tenant

// validate.go attaches validation to the canonical tenant types via the Validator
// interface — the go-libs home of what used to be pkg/handlers/schema_tenant.go +
// the tenant registration in schema_reflect_orggvc.go. Tenant reflects from Tenant
// (tenant.ts: Base.keys({spec: TenantSpec.required()})). The former
// `sort,custom=dedupeEmpty` tags on TenantSpec.orgs/clusters are ported onto the
// strongly-typed NameList collection field type below (Joi.array().items(Name)
// .sort().custom(makeUnique())).

import (
	"github.com/controlplane-com/libs-go/pkg/schema/base"
	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// ValidationRule reflects Tenant. This is the schema the "tenant" kind validates
// against.
func (Tenant) ValidationRule() *validate.Rule { return validate.Reflect(Tenant{}) }

// ValidationRule reflects TenantSpec (tenant.ts:5).
func (TenantSpec) ValidationRule() *validate.Rule { return validate.Reflect(TenantSpec{}) }

func (NameList) ValidationRule() *validate.Rule {
	return validate.Array(base.Name("").ValidationRule()).Sort().Dedupe()
}
