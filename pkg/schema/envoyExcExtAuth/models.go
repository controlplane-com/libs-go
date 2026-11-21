/* auto-generated */

package envoyExcExtAuth

import "github.com/controlplane-com/libs-go/pkg/schema/envoyCommon"
import "github.com/controlplane-com/libs-go/pkg/schema/port"

type ExcExtAuth struct {
	Match   *envoyCommon.RouteMatch `json:"match,omitzero"`
	Port    *port.Port              `json:"port,omitzero"`
	SvcPort *port.Port              `json:"svcPort,omitzero"`
}

type ExcludedRateLimit struct {
	Match   *envoyCommon.RouteMatch `json:"match,omitzero"`
	Port    *port.Port              `json:"port,omitzero"`
	SvcPort *port.Port              `json:"svcPort,omitzero"`
}
