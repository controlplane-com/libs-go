package deployment

// validate.go attaches validation to the canonical deployment types via the
// Validator interface — the go-libs home of what pkg/handlers/schema_deployment.go
// (DeploymentSchema/DeploymentStatusRule/DeploymentVersionRule) + the eager
// reflection registration in schema_reflect_workload.go used to provide. The
// `deployment` kind is a pure reflection of deployment.ts (Deployment): no
// object-level refine / xor / unknown policy, so ValidationRule just reflects the
// tagged struct. The old `pattern=name` tag on Name is now carried by the strongly-
// typed base.Name field type.

import "github.com/controlplane-com/libs-go/pkg/schema/validate"

// ValidationRule reflects Deployment — the schema the "deployment" kind validates
// against (deployment.ts: Deployment). base.Name (the name field type) supplies the
// name regex/max/unmodifiable; the field's `required` tag layers on top.
func (Deployment) ValidationRule() *validate.Rule { return validate.Reflect(Deployment{}) }

// ValidationRule reflects DeploymentStatus (deployment.ts: DeploymentStatus).
func (DeploymentStatus) ValidationRule() *validate.Rule {
	return validate.Reflect(DeploymentStatus{})
}

// ValidationRule reflects DeploymentVersion (deployment.ts: DeploymentVersion).
func (DeploymentVersion) ValidationRule() *validate.Rule {
	return validate.Reflect(DeploymentVersion{})
}
