/* auto-generated */

package agent

import "github.com/controlplane-com/libs-go/pkg/schema/base"

type AgentTags map[string]any

type Agent struct {
	Id           string      `json:"id,omitempty"`
	Name         base.Name   `json:"name,omitempty"`
	Kind         base.Kind   `json:"kind,omitempty"`
	Version      float32     `json:"version"`
	Description  string      `json:"description,omitempty"`
	Tags         AgentTags   `json:"tags,omitempty"`
	Created      string      `json:"created,omitempty"`
	LastModified string      `json:"lastModified,omitempty"`
	Links        base.Links  `json:"links,omitempty"`
	Status       AgentStatus `json:"status,omitempty"`
}

type AgentInfoProtocolVersion string

const (
	AgentInfoProtocolVersionV1 AgentInfoProtocolVersion = "v1"
	AgentInfoProtocolVersionV2 AgentInfoProtocolVersion = "v2"
)

type AgentInfoEnv map[string]any

type AgentInfo struct {
	ProtocolVersion AgentInfoProtocolVersion `json:"protocolVersion,omitempty"`
	InstanceId      string                   `json:"instanceId,omitempty"`
	LastActive      string                   `json:"lastActive,omitempty"`
	Env             AgentInfoEnv             `json:"env,omitempty"`
	PeerCount       float32                  `json:"peerCount"`
	ServiceCount    float32                  `json:"serviceCount"`
}

type AgentStatusProtocolVersion string

const (
	AgentStatusProtocolVersionV1 AgentStatusProtocolVersion = "v1"
	AgentStatusProtocolVersionV2 AgentStatusProtocolVersion = "v2"
)

type AgentStatus struct {
	BootstrapConfig BootstrapConfig            `json:"bootstrapConfig,omitempty"`
	ProtocolVersion AgentStatusProtocolVersion `json:"protocolVersion,omitempty"`
}

type BootstrapConfigProtocolVersion string

const (
	BootstrapConfigProtocolVersionV1 BootstrapConfigProtocolVersion = "v1"
	BootstrapConfigProtocolVersionV2 BootstrapConfigProtocolVersion = "v2"
)

type BootstrapConfig struct {
	RegistrationToken string                         `json:"registrationToken"`
	AgentId           string                         `json:"agentId"`
	AgentLink         string                         `json:"agentLink"`
	HubEndpoint       string                         `json:"hubEndpoint"`
	ProtocolVersion   BootstrapConfigProtocolVersion `json:"protocolVersion,omitempty"`
}
