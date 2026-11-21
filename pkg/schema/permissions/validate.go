package permissions

// validate.go attaches validation to the canonical permissions types via the
// Validator interface — the go-libs home of what used to be the RegisterPattern/
// RegisterCustom wiring in pkg/handlers/schema_reflect_iam.go (backed by
// schema_permissions.go). permissions is a read-only served resource; these are the
// irreducible Joi parts a struct tag cannot carry:
//   - the name regex (was pattern=permissionName) as the PermissionName leaf type;
//   - the items sort({by:'name'}) custom (was custom=permSortByName) as the
//     PermissionItems collection type. The field's min/required/uniqueBy tags apply
//     on top.

import (
	"regexp"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// permissionNameRe backs what was the `pattern=permissionName` tag (permissions.ts:8).
var permissionNameRe = regexp.MustCompile(`^[a-z][a-zA-Z0-9]*`)

// ValidationRule ports the `pattern=permissionName` name check.
func (PermissionName) ValidationRule() *validate.Rule {
	return validate.String().Regex(permissionNameRe)
}

// ValidationRule reflects each Permission item and applies the sort({by:'name'})
// custom to the array (was custom=permSortByName). The field's min/required/uniqueBy
// declarative tokens apply on top.
func (PermissionItems) ValidationRule() *validate.Rule {
	return validate.Array(validate.Reflect(Permission{})).SortBy("name")
}
