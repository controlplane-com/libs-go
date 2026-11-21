package serviceaccount

// validate.go attaches validation to the canonical serviceaccount types via the
// Validator interface — the go-libs home of what used to be
// pkg/handlers/schema_serviceaccount.go + the serviceaccount wiring in
// schema_reflect_iam.go. serviceaccount.ts reflects from ServiceAccount
// (`ServiceAccount = Base.keys({keys, origin})`).

import (
	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// ValidationRule reflects ServiceAccount.
func (ServiceAccount) ValidationRule() *validate.Rule {
	return validate.Reflect(ServiceAccount{})
}
