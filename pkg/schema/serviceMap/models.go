/* auto-generated */

package serviceMap

type ServiceWorkloadType string

const (
	ServiceWorkloadTypeServerless ServiceWorkloadType = "serverless"
	ServiceWorkloadTypeStandard   ServiceWorkloadType = "standard"
	ServiceWorkloadTypeCron       ServiceWorkloadType = "cron"
	ServiceWorkloadTypeStateful   ServiceWorkloadType = "stateful"
	ServiceWorkloadTypeVm         ServiceWorkloadType = "vm"
)

type ServiceWorkload struct {
	Name              string              `json:"name"`
	CanonicalEndpoint string              `json:"canonicalEndpoint"`
	PortMap           []ServicePortMap    `json:"portMap"`
	Type              ServiceWorkloadType `json:"type,omitempty"`
	Replicas          []ServiceReplica    `json:"replicas,omitempty"`
	AgentAccessible   bool                `json:"agentAccessible,omitempty"`
	Kata              bool                `json:"kata,omitempty"`
	ScaleToZero       bool                `json:"scaleToZero,omitempty"`
}

type ServiceGvc struct {
	Name  string `json:"name"`
	Alias string `json:"alias"`
}

type Service struct {
	Workload ServiceWorkload `json:"workload"`
	Gvc      ServiceGvc      `json:"gvc"`
}

type ServiceMap struct {
	Outbound      []Service `json:"outbound"`
	LastUpdate    string    `json:"lastUpdate"`
	OrgGvcAliases []string  `json:"orgGvcAliases,omitempty"`
}

type ServicePortMap struct {
	ServicePort   float32 `json:"servicePort"`
	ContainerPort float32 `json:"containerPort"`
	Protocol      string  `json:"protocol,omitempty"`
}

type ServiceReplica struct {
	Location string  `json:"location"`
	Replicas float32 `json:"replicas"`
}
