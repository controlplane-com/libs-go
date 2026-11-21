/* hand-maintained — canonical source of truth for the `event` resource schema.
 * `validate:` tags drive pkg/validate.FromType. See nodelibs/schema/src/event.ts.
 * NOT Base-derived (plain object). EventContext is .unknown(true) (registered on
 * that type). Registered as the `event` kind.
 */

package event

import "github.com/controlplane-com/libs-go/pkg/schema/base"

type EventContextComponent string

const (
	EventContextComponentActuator      EventContextComponent = "actuator"
	EventContextComponentDnsUpdater    EventContextComponent = "dns-updater"
	EventContextComponentScheduler     EventContextComponent = "scheduler"
	EventContextComponentIamBroker     EventContextComponent = "iam-broker"
	EventContextComponentMetadataProxy EventContextComponent = "metadata-proxy"
	EventContextComponentDataService   EventContextComponent = "data-service"
	EventContextComponentMk8S          EventContextComponent = "mk8s"
	EventContextComponentOther         EventContextComponent = "other"
)

// EventContext ports event.ts:3. .unknown(true) is registered on this type so
// arbitrary properties pass through.
type EventContext struct {
	Category      string                `json:"category,omitzero"`
	Component     EventContextComponent `json:"component,omitzero" validate:"valid=actuator|dns-updater|scheduler|iam-broker|metadata-proxy|data-service|mk8s|other"`
	CloudProvider string                `json:"cloudProvider,omitzero"`
	Cluster       string                `json:"cluster,omitzero"`
	PrincipalLink string                `json:"principalLink,omitzero"`
	Fingerprint   string                `json:"fingerprint,omitzero" validate:"max=64"`
	// EventContext is .unknown(true) (validate.go:22): properties it does not name
	// round-trip untouched.
	Extras base.Extras `json:"-"`
} //ts:export

type eventContextAlias EventContext

func (c EventContext) MarshalJSON() ([]byte, error) {
	return base.MarshalExtras(eventContextAlias(c), c.Extras)
}

func (c *EventContext) UnmarshalJSON(b []byte) error {
	return base.UnmarshalExtras(b, (*eventContextAlias)(c), &c.Extras)
}

type Event struct {
	Id            string        `json:"id,omitzero"`
	Created       string        `json:"created,omitzero" validate:"date"`
	Kind          string        `json:"kind,omitzero"`
	LastEventTime string        `json:"lastEventTime,omitzero" validate:"date"`
	EventCount    *float64      `json:"eventCount,omitzero" validate:"integer"`
	Status        string        `json:"status" validate:"required"`
	Pinned        *bool         `json:"pinned,omitzero"`
	Context       *EventContext `json:"context,omitzero" validate:"default={},optional"`
}
