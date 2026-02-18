/* auto-generated */

package envoyExcExtAuth

import "github.com/controlplane-com/libs-go/pkg/schema/envoyCommon"
import "github.com/controlplane-com/libs-go/pkg/schema/port"

type ExcExtAuth struct {
	Match   envoyCommon.RouteMatch `json:"match,omitempty"`
	Port    port.Port              `json:"port,omitempty"`
	SvcPort port.Port              `json:"svcPort,omitempty"`
}

type ExcludedRateLimit struct {
	Match   envoyCommon.RouteMatch `json:"match,omitempty"`
	Port    port.Port              `json:"port,omitempty"`
	SvcPort port.Port              `json:"svcPort,omitempty"`
}
