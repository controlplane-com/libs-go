/* hand-maintained — canonical source of truth for the `agent` resource schema.
 *
 * These structs (formerly emitted by joi-to-go from nodelibs/schema/src/agent.ts)
 * are now the CANONICAL definition. `validate:` struct tags carry the validation rules;
 * pkg/validate.FromType reflects over these types to build the *validate.Rule tree
 * that the engine runs — replacing the hand-written pkg/handlers/schema_agent.go
 * builder. See nodelibs/schema/src/agent.ts for the original Joi source, and
 * go-libs/schema/ipSet/models.go for the tag DSL grammar.
 *
 * The createSchema fork (name required, description defaults to the name) is
 * expressed as: `required` on Name (a field tag) + a cross-field Refine registered
 * in pkg/handlers (schema_reflect_agent.go) for the description-defaults-to-name
 * part (a Joi.ref default a field tag cannot carry).
 */

package agent

import "github.com/controlplane-com/libs-go/pkg/schema/base"

// AgentTags aliases base.Tags so the reflection base-type registry resolves it to
// the shared TagsRule.
type AgentTags = base.Tags

// HttpHttpsUri is a URI restricted to the http/https schemes — the strongly-typed
// replacement for the `uri,custom=uriHttpHttps` tag (Joi.string().uri({scheme:
// ['http','https']})).
type HttpHttpsUri string

type Agent struct {
	Id           string       `json:"id,omitzero" validate:"uuid,readOnly"`
	Name         base.Name    `json:"name,omitzero"`
	Kind         base.Kind    `json:"kind,omitzero"`
	Version      *float64     `json:"version,omitzero" validate:"positive,integer,readOnly"`
	Description  string       `json:"description,omitzero" validate:"trim,max=250,allow="`
	Tags         AgentTags    `json:"tags,omitzero" validate:"default={}"`
	Created      string       `json:"created,omitzero" validate:"date,readOnly"`
	LastModified string       `json:"lastModified,omitzero" validate:"date,readOnly"`
	Links        base.Links   `json:"links,omitzero" validate:"readOnly"`
	Status       *AgentStatus `json:"status,omitzero"`
}

type AgentInfoProtocolVersion string

const (
	AgentInfoProtocolVersionV1 AgentInfoProtocolVersion = "v1"
	AgentInfoProtocolVersionV2 AgentInfoProtocolVersion = "v2"
)

func (AgentInfoProtocolVersion) EnumValues() []string {
	return []string{string(AgentInfoProtocolVersionV1), string(AgentInfoProtocolVersionV2)}
}

type AgentInfoEnv map[string]any

type AgentInfo struct {
	ProtocolVersion AgentInfoProtocolVersion `json:"protocolVersion,omitzero"`
	InstanceId      string                   `json:"instanceId,omitzero"`
	LastActive      string                   `json:"lastActive,omitzero" validate:"date,iso"`
	Env             AgentInfoEnv             `json:"env,omitzero"`
	PeerCount       *float64                 `json:"peerCount,omitzero" validate:"positive,integer"`
	ServiceCount    *float64                 `json:"serviceCount,omitzero" validate:"positive,integer"`
}

type AgentStatusProtocolVersion string

const (
	AgentStatusProtocolVersionV1 AgentStatusProtocolVersion = "v1"
	AgentStatusProtocolVersionV2 AgentStatusProtocolVersion = "v2"
)

func (AgentStatusProtocolVersion) EnumValues() []string {
	return []string{string(AgentStatusProtocolVersionV1), string(AgentStatusProtocolVersionV2)}
}

type AgentStatus struct {
	BootstrapConfig *BootstrapConfig           `json:"bootstrapConfig,omitzero"`
	ProtocolVersion AgentStatusProtocolVersion `json:"protocolVersion,omitzero"`
}

type BootstrapConfigProtocolVersion string

const (
	BootstrapConfigProtocolVersionV1 BootstrapConfigProtocolVersion = "v1"
	BootstrapConfigProtocolVersionV2 BootstrapConfigProtocolVersion = "v2"
)

func (BootstrapConfigProtocolVersion) EnumValues() []string {
	return []string{string(BootstrapConfigProtocolVersionV1), string(BootstrapConfigProtocolVersionV2)}
}

type BootstrapConfig struct {
	RegistrationToken string                         `json:"registrationToken" validate:"required"`
	AgentId           string                         `json:"agentId" validate:"required"`
	AgentLink         string                         `json:"agentLink" validate:"required"`
	HubEndpoint       HttpHttpsUri                   `json:"hubEndpoint" validate:"required"`
	ProtocolVersion   BootstrapConfigProtocolVersion `json:"protocolVersion,omitzero"`
}
