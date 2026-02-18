/* auto-generated */

package containerstatus

type ContainerStatusRestarts struct {
	LastRestartTime string  `json:"lastRestartTime,omitempty"`
	ExitCode        float32 `json:"exitCode"`
	Reason          string  `json:"reason,omitempty"`
	Count           float32 `json:"count"`
}

type ContainerStatus struct {
	Name      string                  `json:"name,omitempty"`
	Image     string                  `json:"image,omitempty"`
	Ready     bool                    `json:"ready,omitempty"`
	Resources DeploymentResources     `json:"resources,omitempty"`
	Message   string                  `json:"message,omitempty"`
	Restarts  ContainerStatusRestarts `json:"restarts,omitempty"`
}

type DeploymentResources struct {
	Replicas      float32 `json:"replicas"`
	ReplicasReady float32 `json:"replicasReady"`
	Cpu           float32 `json:"cpu"`
	Memory        float32 `json:"memory"`
}
