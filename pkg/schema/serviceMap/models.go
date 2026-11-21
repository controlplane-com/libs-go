/* hand-maintained — canonical source of truth for the `serviceMap` schema.
 * ServiceMap is the materialized per-workload mTLS service-discovery map
 * (data-service/pkg/handlers/serviceMap.go); it is computed output returned
 * to the scheduler/actuator, never request-validated against an HTTP body.
 * The Joi rules are carried as struct tags anyway (every serviceMap.ts field is
 * `.required()`, and `type` is an enum) so that the day this shape IS validated —
 * or reused in a request schema — the constraints are already there rather than
 * silently absent. `validate:"date"` on ServiceMap.LastUpdate additionally forces
 * the schema-ts generator to emit `Date` (mirrors gvc.Created / location.Created).
 */

package serviceMap

// ServicePortMap is a service-port to container-port mapping for a reachable
// workload (serviceMap.ts:3).
type ServicePortMap struct {
	ServicePort   float64 `json:"servicePort" validate:"required"`
	ContainerPort float64 `json:"containerPort" validate:"required"`
	Protocol      string  `json:"protocol,omitzero" validate:"optional"`
}

// ServiceReplica is the per-location replica count for a replica-direct
// (stateful) workload (serviceMap.ts:9).
type ServiceReplica struct {
	Location string  `json:"location" validate:"required"`
	Replicas float64 `json:"replicas" validate:"required"`
}

// ServiceWorkloadType is Service.workload.type's enum (serviceMap.ts:19).
type ServiceWorkloadType string

const (
	ServiceWorkloadTypeServerless ServiceWorkloadType = "serverless"
	ServiceWorkloadTypeStandard   ServiceWorkloadType = "standard"
	ServiceWorkloadTypeCron       ServiceWorkloadType = "cron"
	ServiceWorkloadTypeStateful   ServiceWorkloadType = "stateful"
	ServiceWorkloadTypeVm         ServiceWorkloadType = "vm"
)

// EnumValues ports serviceMap.ts:19 `.valid('serverless','standard','cron','stateful','vm')`.
// Without it the constants were decoration and any string passed.
func (ServiceWorkloadType) EnumValues() []string {
	return []string{"serverless", "standard", "cron", "stateful", "vm"}
}

type ServiceWorkloadRouting string

const (
	ServiceWorkloadRoutingGvcEndpoint ServiceWorkloadRouting = "gvcEndpoint"
)

func (ServiceWorkloadRouting) EnumValues() []string { return []string{"gvcEndpoint"} }

// ServiceWorkload is the inline `workload` shape of Service (serviceMap.ts:15).
type ServiceWorkload struct {
	Name              string              `json:"name" validate:"required"`
	CanonicalEndpoint string              `json:"canonicalEndpoint" validate:"required"`
	PortMap           []ServicePortMap    `json:"portMap" validate:"required"`
	Type              ServiceWorkloadType `json:"type" validate:"required"`
	Replicas          []ServiceReplica    `json:"replicas,omitzero"`
	AgentAccessible   *bool               `json:"agentAccessible,omitzero"`
	// kata / scaleToZero landed on master while this port was branched
	// (serviceMap.ts:22-27): kata serverless has no knative queue-proxy, and
	// scale-to-zero ranks live pods above the standing activator entry.
	Kata             *bool `json:"kata,omitzero"`
	ScaleToZero      *bool `json:"scaleToZero,omitzero"`
	Routing          ServiceWorkloadRouting `json:"routing,omitzero"`
}

// ServiceGvc is the inline `gvc` shape of Service (serviceMap.ts:23).
type ServiceGvc struct {
	Name  string `json:"name" validate:"required"`
	Alias string `json:"alias" validate:"required"`
}

// Service is a single reachable workload in a service map (serviceMap.ts:14).
type Service struct {
	Workload ServiceWorkload `json:"workload" validate:"required"`
	Gvc      ServiceGvc      `json:"gvc" validate:"required"`
}

// ServiceMap is the mTLS service-discovery map for a workload (serviceMap.ts:29).
type ServiceMap struct {
	Outbound      []Service `json:"outbound" validate:"required"`
	LastUpdate    string    `json:"lastUpdate" validate:"date,required"`
	OrgGvcAliases []string  `json:"orgGvcAliases,omitzero"`
}
