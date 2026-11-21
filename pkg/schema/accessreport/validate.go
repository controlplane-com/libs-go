package accessreport

// validate.go attaches validation to the canonical accessreport types via the
// Validator interface — the go-libs home of what used to be
// pkg/handlers/schema_accessreport.go's AccessReportSchema (FromType) plus the
// RegisterReflect("accessreport", …) wiring in schema_reflect_iam.go. accessreport.ts
// is a plain reflected object: no customs, patterns, object-relations or refines, so
// the rule is a straight Reflect over the tagged struct (kind enum, permissions array
// of GrantedPermission → PrincipalBinding, created date, base links).

import "github.com/controlplane-com/libs-go/pkg/schema/validate"

// ValidationRule reflects AccessReport — the schema the "accessreport" kind validates
// against.
func (AccessReport) ValidationRule() *validate.Rule { return validate.Reflect(AccessReport{}) }
