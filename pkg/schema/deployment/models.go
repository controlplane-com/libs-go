/* auto-generated */

package deployment

import "github.com/controlplane-com/libs-go/pkg/schema/base"
import "github.com/controlplane-com/libs-go/pkg/schema/cronjob"
import "github.com/controlplane-com/libs-go/pkg/schema/containerstatus"

type Deployment struct {
	Name         string           `json:"name"`
	Kind         base.Kind        `json:"kind,omitempty"`
	Links        base.Links       `json:"links,omitempty"`
	LastModified string           `json:"lastModified,omitempty"`
	Status       DeploymentStatus `json:"status,omitempty"`
}

type DeploymentIdentity struct {
	Link      string   `json:"link,omitempty"`
	Version   *float32 `json:"version,omitempty"`
	AppliedAt string   `json:"appliedAt,omitempty"`
}

type DeploymentStatus struct {
	Endpoint                  string                       `json:"endpoint,omitempty"`
	Remote                    string                       `json:"remote,omitempty"`
	LastProcessedVersion      *float32                     `json:"lastProcessedVersion,omitempty"`
	ExpectedDeploymentVersion *float32                     `json:"expectedDeploymentVersion,omitempty"`
	Identity                  DeploymentIdentity           `json:"identity,omitempty"`
	Internal                  any                          `json:"internal,omitempty"`
	Ready                     bool                         `json:"ready,omitempty"`
	Deploying                 bool                         `json:"deploying,omitempty"`
	Versions                  []DeploymentVersion          `json:"versions,omitempty"`
	JobExecutions             []cronjob.JobExecutionStatus `json:"jobExecutions,omitempty"`
	Message                   string                       `json:"message,omitempty"`
	DesiredScale              *float32                     `json:"desiredScale,omitempty"`
}

type DeploymentVersionContainers map[string]containerstatus.ContainerStatus

type DeploymentVersion struct {
	Name         string                      `json:"name,omitempty"`
	Created      string                      `json:"created,omitempty"`
	Workload     *float32                    `json:"workload,omitempty"`
	Gvc          *float32                    `json:"gvc,omitempty"`
	Containers   DeploymentVersionContainers `json:"containers,omitempty"`
	Ready        bool                        `json:"ready,omitempty"`
	Message      string                      `json:"message,omitempty"`
	Zone         string                      `json:"zone,omitempty"`
	DesiredScale *float32                    `json:"desiredScale,omitempty"`
}
