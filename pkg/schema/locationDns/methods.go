package locationDns

import (
	"fmt"
)

func (r LocationDnsMessage) WorkloadLink() string {
	return fmt.Sprintf("/org/%s/gvc/%s/workload/%s", r.Org, r.Gvc, r.Workload)
}

/*
func (r LocationDnsMessage) ParseTimestamp() time.Time {
	t, err := time.Parse(time.RFC3339, r.Timestamp)
	if err != nil {
		return time.Time{}
	}
	return t
}
*/
