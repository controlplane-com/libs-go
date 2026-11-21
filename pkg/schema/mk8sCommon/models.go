/* auto-generated */

package mk8sCommon

type AutoscalerConfigExpander string

const (
	AutoscalerConfigExpanderRandom     AutoscalerConfigExpander = "random"
	AutoscalerConfigExpanderMostPods   AutoscalerConfigExpander = "most-pods"
	AutoscalerConfigExpanderLeastWaste AutoscalerConfigExpander = "least-waste"
)

type AutoscalerConfig struct {
	Expander             []AutoscalerConfigExpander `json:"expander,omitzero"`
	UnneededTime         string                     `json:"unneededTime,omitzero"`
	UnreadyTime          string                     `json:"unreadyTime,omitzero"`
	UtilizationThreshold *float64                   `json:"utilizationThreshold,omitzero"`
}

type CacertsRes struct {
	Cacerts string `json:"cacerts,omitzero"`
}

type GoDuration string

type InitBody map[string]any

type InitRes struct {
	Script string `json:"script,omitzero"`
}

type JoinBody map[string]any

type JoinRes struct {
	Script string `json:"script,omitzero"`
}

type KubeConfigBody map[string]any

type KubeConfigRes struct {
	Kubeconfig string `json:"kubeconfig,omitzero"`
	FileName   string `json:"fileName,omitzero"`
}

type Labels map[string]string

type MaxPods float32

type NodePoolName string

type PreInstallScript string

type ReadyRes struct {
	Ready   *bool  `json:"ready,omitzero"`
	Message string `json:"message,omitzero"`
}

type SshPublicKey string

type TaintEffect string

const (
	TaintEffectNoSchedule       TaintEffect = "NoSchedule"
	TaintEffectPreferNoSchedule TaintEffect = "PreferNoSchedule"
	TaintEffectNoExecute        TaintEffect = "NoExecute"
)

type Taint struct {
	Key    SshPublicKey `json:"key,omitzero"`
	Value  string       `json:"value,omitzero"`
	Effect TaintEffect  `json:"effect,omitzero"`
}

type Taints []Taint

type UnmanagedPool struct {
	Name   string `json:"name"`
	Labels Labels `json:"labels,omitzero"`
	Taints Taints `json:"taints,omitzero"`
}
