/* hand-maintained — canonical source of truth (nodelibs/schema/src/containerstatus.ts).
 * `joi:` tags drive pkg/validate.FromType; DSL on go-libs/schema/ipSet/models.go. */

package containerstatus

type ContainerStatusRestarts struct {
	LastRestartTime string   `json:"lastRestartTime,omitzero" validate:"date"`
	ExitCode        *float64 `json:"exitCode,omitzero"`
	Reason          string   `json:"reason,omitzero"`
	Count           *float64 `json:"count,omitzero"`
}

type ContainerStatus struct {
	Name      string                   `json:"name,omitzero"`
	Image     string                   `json:"image,omitzero"`
	Ready     *bool                    `json:"ready,omitzero"`
	Resources *DeploymentResources     `json:"resources,omitzero"`
	Message   string                   `json:"message,omitzero" validate:"allow="`
	Restarts  *ContainerStatusRestarts `json:"restarts,omitzero"`
}

type DeploymentResources struct {
	Replicas      *float64 `json:"replicas,omitzero"`
	ReplicasReady *float64 `json:"replicasReady,omitzero"`
	Cpu           *float64 `json:"cpu,omitzero"`
	Memory        *float64 `json:"memory,omitzero"`
}
