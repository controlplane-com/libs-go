/* hand-maintained — canonical source of truth for the `ipset` resource schema.
 *
 * These structs (formerly emitted by joi-to-go from nodelibs/schema/src/ipSet.ts)
 * are now the CANONICAL definition. `joi:` struct tags carry the validation rules;
 * pkg/validate.FromType reflects over these types to build the *validate.Rule tree
 * that the engine runs — replacing the hand-written pkg/handlers/schema_ipset.go
 * builder. See nodelibs/schema/src/ipSet.ts for the original Joi source.
 *
 * Tag DSL (pkg/validate/reflect.go):
 *   presence : required | optional | forbidden          (default: optional)
 *   core     : default=<lit> allow=<v> valid=a|b invalid=a|b
 *   string   : trim lowercase uppercase min=n max=n length=n email uri uuid domain pattern=<name>
 *   number   : min=n max=n integer positive negative port precision=n minRef=f maxRef=f
 *   array    : min=n max=n length=n sort unique uniqueBy=f single
 *   type     : date  (declare a Go string/number field as a Joi.date())
 *   links    : localLink=kinds  imageLink=kinds   (kinds separated by '|', e.g. !workload|gvc)
 *   custom   : custom=<name>    (named CustomFn registered in pkg/handlers)
 *   meta     : readOnly unmodifiable
 *
 * Base fields resolve via the base-type registry (base.Name→NameRule, base.Kind→
 * KindRule, base.Tags→TagsRule, base.Links→LinksRule); enum typed-strings resolve
 * via the field `valid=` tag OR an EnumValues() method (Valuer registry).
 */

package ipSet

import "github.com/controlplane-com/libs-go/pkg/schema/base"

// LocationLink ports `LocalLink.custom(toKind('location'))`: a local link that must
// resolve to a resource of kind `location`. Replaces the `localLink=location` tag
// on IpAddress.Name and IpSetLocation.Name.
type LocationLink string

// WorkloadOrGvcLink ports `LocalLink.custom(toKind('!workload', 'gvc'))`: the
// spec.link target, which must point to a workload (via `!workload`) or a gvc.
// Replaces the `localLink=!workload|gvc` tag on IpSetSpec.Link.
type WorkloadOrGvcLink string

// Ipv4NoCidr ports `Joi.string().ip({ version: 'ipv4', cidr: 'forbidden' })`: a
// plain IPv4 dotted-quad with no CIDR suffix. The error message contains
// "ip address" to match Joi (ipSet.test.ts asserts `.toContain('ip address')`).
// Replaces the `custom=ipv4NoCidr` tag on IpAddress.Ip.
type Ipv4NoCidr string

type IpAddressState string

const (
	IpAddressStateBound   IpAddressState = "bound"
	IpAddressStateUnbound IpAddressState = "unbound"
)

// EnumValues makes IpAddressState resolve to Joi.string().valid('bound','unbound')
// via the reflection Valuer registry (demonstrates the EnumValues() path; the
// sibling IpSetLocation.RetentionPolicy uses the `valid=` tag path instead).
func (IpAddressState) EnumValues() []string {
	return []string{string(IpAddressStateBound), string(IpAddressStateUnbound)}
}

type IpAddress struct {
	Name    LocationLink   `json:"name" validate:"required"`
	Ip      Ipv4NoCidr     `json:"ip" validate:"required"`
	Id      string         `json:"id" validate:"required"`
	State   IpAddressState `json:"state,omitzero" validate:"required"`
	Created string         `json:"created" validate:"date,required"`
}

// IpSetTags is an alias for base.Tags so the reflection base-type registry
// (keyed on base.Tags) resolves it to the shared TagsRule.
type IpSetTags = base.Tags

type IpSetSpec struct {
	Link      WorkloadOrGvcLink `json:"link,omitzero" validate:"optional"`
	Locations []IpSetLocation   `json:"locations,omitzero"`
}

type IpSet struct {
	Id           string       `json:"id,omitzero" validate:"uuid,readOnly"`
	Name         base.Name    `json:"name,omitzero"`
	Kind         base.Kind    `json:"kind,omitzero"`
	Version      *float64     `json:"version,omitzero" validate:"positive,integer,readOnly"`
	Description  string       `json:"description,omitzero" validate:"trim,max=250,allow="`
	Tags         IpSetTags    `json:"tags,omitzero" validate:"default={}"`
	Created      string       `json:"created,omitzero" validate:"date,readOnly"`
	LastModified string       `json:"lastModified,omitzero" validate:"date,readOnly"`
	Links        base.Links   `json:"links,omitzero" validate:"readOnly"`
	Spec         IpSetSpec    `json:"spec" validate:"required"`
	Status       *IpSetStatus `json:"status,omitzero"`
}

type IpSetLocationRetentionPolicy string

const (
	IpSetLocationRetentionPolicyKeep IpSetLocationRetentionPolicy = "keep"
	IpSetLocationRetentionPolicyFree IpSetLocationRetentionPolicy = "free"
)

type IpSetLocation struct {
	Name            LocationLink                 `json:"name" validate:"required"`
	RetentionPolicy IpSetLocationRetentionPolicy `json:"retentionPolicy,omitzero" validate:"valid=keep|free,required"`
}

type IpSetStatus struct {
	IpAddresses []IpAddress `json:"ipAddresses,omitzero"`
	Error       string      `json:"error,omitzero" validate:"optional"`
	Warning     string      `json:"warning,omitzero" validate:"optional"`
}
