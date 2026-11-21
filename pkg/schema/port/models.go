/* hand-maintained — canonical source of truth for the shared `Port` type.
 *
 * Formerly emitted by joi-to-go as a bare `type Port float64`, which carried NO
 * validation at all: the reflection bridge saw a float64 and produced an
 * unconstrained Number(). Every port in the API was therefore unvalidated —
 * container ports, readiness-probe ports, metrics ports and load-balancer
 * container ports all accepted 0, 65536, non-integers, and the reserved Istio
 * sidecar ports (15000/15001/15006/15020/15021) that the TS schema exists to
 * keep out.
 *
 * See nodelibs/schema/src/port.ts for the Joi source.
 */

package port

// Port ports `Joi.number().integer().max(65535).min(80).custom(reserved)`.
// The rule lives in validate.go.
type Port float64
