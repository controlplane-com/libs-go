/* hand-maintained — canonical source of truth (nodelibs/schema/src/deployment.ts).
 * `joi:` tags drive pkg/validate.FromType; DSL on go-libs/schema/ipSet/models.go.
 * Registered as the `deployment` kind via reflection in schema_reflect_workload.go. */

package deployment

import "github.com/controlplane-com/libs-go/pkg/schema/base"
import "github.com/controlplane-com/libs-go/pkg/schema/cronjob"
import "github.com/controlplane-com/libs-go/pkg/schema/containerstatus"

type Deployment struct {
	Name         base.Name         `json:"name" validate:"required"`
	Kind         base.Kind         `json:"kind,omitzero"`
	Links        base.Links        `json:"links,omitzero"`
	LastModified string            `json:"lastModified,omitzero" validate:"date,iso"`
	Status       *DeploymentStatus `json:"status,omitzero"`
}

type DeploymentStatus struct {
	Endpoint                  string                       `json:"endpoint,omitzero" validate:"uri"`
	Remote                    string                       `json:"remote,omitzero" validate:"uri"`
	LastProcessedVersion      *float64                     `json:"lastProcessedVersion,omitzero" validate:"integer"`
	ExpectedDeploymentVersion *float64                     `json:"expectedDeploymentVersion,omitzero" validate:"integer"`
	Identity                  *DeploymentIdentity          `json:"identity,omitzero"`
	Internal                  any                          `json:"internal,omitzero"`
	Ready                     *bool                        `json:"ready,omitzero"`
	Deploying                 *bool                        `json:"deploying,omitzero"`
	Versions                  []DeploymentVersion          `json:"versions,omitzero"`
	JobExecutions             []cronjob.JobExecutionStatus `json:"jobExecutions,omitzero"`
	Message                   string                       `json:"message,omitzero" validate:"default=,allow="`
	DesiredScale              *float64                     `json:"desiredScale,omitzero" validate:"integer"`
}

// DeploymentIdentity is the last identity version fully applied in this location
// (deployment.ts DeploymentIdentity, master 621e7da48).
type DeploymentIdentity struct {
	Link      string   `json:"link,omitzero"`
	Version   *float64 `json:"version,omitzero" validate:"integer"`
	AppliedAt string   `json:"appliedAt,omitzero" validate:"date"`
}

type DeploymentVersionContainers map[string]containerstatus.ContainerStatus

type DeploymentVersion struct {
	Name         string                      `json:"name,omitzero"`
	Created      string                      `json:"created,omitzero" validate:"date"`
	Workload     *float64                    `json:"workload,omitzero"`
	Gvc          *float64                    `json:"gvc,omitzero"`
	Containers   DeploymentVersionContainers `json:"containers,omitzero"`
	Ready        *bool                       `json:"ready,omitzero"`
	Message      string                      `json:"message,omitzero" validate:"allow="`
	Zone         string                      `json:"zone,omitzero" validate:"allow="`
	DesiredScale *float64                    `json:"desiredScale,omitzero" validate:"integer"`
}
