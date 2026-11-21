package port

import (
	"strconv"
	"strings"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// ReservedPorts are the ports the platform binds for itself — the Istio sidecar
// (15000/15001/15006/15020/15021/15090), the envoy admin/metrics pair, and the
// agent tunnel. A workload bound to one of these is an operational hazard, which
// is why port.ts rejects them rather than letting the actuator discover the
// clash. Order matters: it is rendered into the error message as-is.
var ReservedPorts = []int{8012, 8022, 9090, 9091, 15000, 15001, 15006, 15020, 15021, 15090, 41000}

// ValidationRule ports `Joi.number().integer().max(65535).min(80).custom(...)`
// (nodelibs/schema/src/port.ts), in Joi's declared order so the first failing
// constraint is the one reported.
func (Port) ValidationRule() *validate.Rule {
	return validate.Number().Integer().NumMax(65535).NumMin(80).Custom(RejectReserved("for system use", ReservedPorts))
}

// RejectReserved ports the `.custom()` at port.ts:6-12, and serves the sandbox
// runtime's own reserved list too (sandbox.ts SandboxPort) — same check, same message,
// a different owner. The list renders with JS's Array.toString (comma-joined).
func RejectReserved(reason string, reserved []int) validate.CustomFn {
	return func(value any, h *validate.Helper) (any, error) {
		n, ok := toInt(value)
		if !ok {
			return value, nil // a non-numeric value already failed the base rule
		}
		for _, p := range reserved {
			if p == n {
				return nil, h.Errorf("any.custom",
					"The port %d is reserved %s. Please choose a different port. Reserved ports include: %s",
					n, reason, joinInts(reserved))
			}
		}
		return value, nil
	}
}

func toInt(v any) (int, bool) {
	switch t := v.(type) {
	case float64:
		return int(t), t == float64(int(t))
	case float32:
		return int(t), float64(t) == float64(int(t))
	case int:
		return t, true
	case int64:
		return int(t), true
	}
	return 0, false
}

func joinInts(ns []int) string {
	parts := make([]string, len(ns))
	for i, n := range ns {
		parts[i] = strconv.Itoa(n)
	}
	return strings.Join(parts, ",")
}
