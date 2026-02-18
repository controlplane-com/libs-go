/* auto-generated */

package locationDns

type LocationDnsMessageOperation string

const (
	LocationDnsMessageOperationUpdate LocationDnsMessageOperation = "update"
	LocationDnsMessageOperationDelete LocationDnsMessageOperation = "delete"
)

type LocationDnsMessage struct {
	Org              string                      `json:"org"`
	Gvc              string                      `json:"gvc"`
	GvcAlias         string                      `json:"gvcAlias"`
	Workload         string                      `json:"workload"`
	Operation        LocationDnsMessageOperation `json:"operation,omitempty"`
	DeleteReason     string                      `json:"deleteReason,omitempty"`
	Name             string                      `json:"name"`
	Target           string                      `json:"target"`
	AvailabilityZone string                      `json:"availabilityZone,omitempty"`
}
