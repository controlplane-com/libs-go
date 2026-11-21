package task

// validate.go attaches validation to the canonical task types via the Validator
// interface — the go-libs home of what used to be the pkg/handlers/schema_task.go
// FromType shim + the RegisterReflect("task", …) registration in
// schema_reflect_orggvc.go. task.ts is a plain Joi.object reflected field-by-field
// from Task; there are no createSchema fork, object-level refines, xor/oxor rules,
// or custom=/pattern= tags to layer, so ValidationRule is a straight reflect of the
// tagged struct (the links field resolves to base.Links' own ValidationRule).

import (
	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// ValidationRule reflects Task — the schema the "task" kind validates against.
func (Task) ValidationRule() *validate.Rule {
	return validate.Reflect(Task{})
}
