/* hand-maintained — canonical source of truth for the LocationDnsMessage schema.
 * `validate:` tags drive pkg/validate.FromType. See nodelibs/schema/src/locationDns.ts.
 * A NATS/DNS-updater message shape (SUB — not a registered kind).
 */

package locationDns

type LocationDnsMessageOperation string

const (
	LocationDnsMessageOperationUpdate LocationDnsMessageOperation = "update"
	LocationDnsMessageOperationDelete LocationDnsMessageOperation = "delete"
)

// EnumValues carries locationDns.ts's `valid('update','delete')` on the TYPE. The
// field's own `valid=` tag says the same thing, but a second field of this type
// declared without that tag would otherwise have no enum at all.
func (LocationDnsMessageOperation) EnumValues() []string {
	return []string{"update", "delete"}
}

type LocationDnsMessage struct {
	Org                string                      `json:"org" validate:"required"`
	Gvc                string                      `json:"gvc" validate:"required"`
	GvcAlias           string                      `json:"gvcAlias" validate:"required"`
	Workload           string                      `json:"workload" validate:"required"`
	WorkloadId         string                      `json:"workloadId,omitzero"`
	Operation          LocationDnsMessageOperation `json:"operation,omitzero" validate:"valid=update|delete,required"`
	DeleteReason       string                      `json:"deleteReason,omitzero"`
	Name               string                      `json:"name" validate:"required"`
	Target             string                      `json:"target" validate:"required"`
	AvailabilityZone   string                      `json:"availabilityZone,omitzero"`
	RoutingTier        *float64                    `json:"routingTier,omitzero" validate:"integer,min=0"`
	LatencyOffsetMs    *float64                    `json:"latencyOffsetMs,omitzero" validate:"integer"`
	LatencyToleranceMs *float64                    `json:"latencyToleranceMs,omitzero" validate:"integer,min=0"`
}
