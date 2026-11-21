package location

// validate.go attaches validation to the canonical location types (location.ts).

import (
	"github.com/controlplane-com/libs-go/pkg/schema/base"
	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

func (Location) ValidationRule() *validate.Rule {
	return validate.Reflect(Location{})
}

// ValidationRule ports Joi.string().ip({cidr:'required'}): the value must be a
// valid IP in CIDR notation.
func (IpRange) ValidationRule() *validate.Rule {
	return base.IPRule(nil, "required")
}

// ValidationRule ports the `custom=dedupeEmpty` collection custom on
// status.ipRanges: an array of CIDR-required IpRanges with duplicates removed in
// place (makeUnique with no key — NOT an error, unlike Joi array.unique).
func (IpRangeList) ValidationRule() *validate.Rule {
	return validate.Array(IpRange("").ValidationRule()).Dedupe()
}
