/* auto-generated */

package event

type EventContextComponent string

const (
	EventContextComponentActuator      EventContextComponent = "actuator"
	EventContextComponentDnsUpdater    EventContextComponent = "dns-updater"
	EventContextComponentScheduler     EventContextComponent = "scheduler"
	EventContextComponentIamBroker     EventContextComponent = "iam-broker"
	EventContextComponentMetadataProxy EventContextComponent = "metadata-proxy"
	EventContextComponentDataService   EventContextComponent = "data-service"
)

type EventContextCloudProvider string

const (
	EventContextCloudProviderAws    EventContextCloudProvider = "aws"
	EventContextCloudProviderGcp    EventContextCloudProvider = "gcp"
	EventContextCloudProviderAzure  EventContextCloudProvider = "azure"
	EventContextCloudProviderLinode EventContextCloudProvider = "linode"
	EventContextCloudProviderByok   EventContextCloudProvider = "byok"
)

type EventContext struct {
	Category      string                    `json:"category,omitempty"`
	Component     EventContextComponent     `json:"component,omitempty"`
	CloudProvider EventContextCloudProvider `json:"cloudProvider,omitempty"`
	Cluster       string                    `json:"cluster,omitempty"`
	PrincipalLink string                    `json:"principalLink,omitempty"`
	Fingerprint   string                    `json:"fingerprint,omitempty"`

	/* WARNING!! Arbitrary properties are being ignored! */
}

type Event struct {
	Id            string       `json:"id,omitempty"`
	Created       string       `json:"created,omitempty"`
	Kind          string       `json:"kind,omitempty"`
	LastEventTime string       `json:"lastEventTime,omitempty"`
	EventCount    float32      `json:"eventCount"`
	Status        string       `json:"status"`
	Pinned        bool         `json:"pinned,omitempty"`
	Context       EventContext `json:"context,omitempty"`
}
