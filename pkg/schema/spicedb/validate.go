package spicedb

// validate.go attaches validation to the canonical spicedb types via the
// Validator interface — the go-libs home of what used to be the
// pkg/handlers/schema_spicedb.go FromType shim + the spicedbcluster bindings in
// schema_reflect_datanet.go. spicedb.ts reflects `Base.keys({spec, alias,
// status})` from SpicedbCluster; the only non-declarative pieces are:
//   - spec.locations element = LocalLink.custom(toKind('location')) — the
//     SpicedbLocationLink leaf type below (was RegisterType).
//   - status = Joi.object({externalEndpoint}).unknown(true) — ClusterStatus
//     .Unknown(true) below (was RegisterObjectRules Unknown:true).

import (
	"github.com/controlplane-com/libs-go/pkg/schema/base"
	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// ValidationRule reflects SpicedbCluster. This is the schema the "spicedbcluster"
// kind validates against (formerly RegisterReflect("spicedbcluster", ...)).
func (SpicedbCluster) ValidationRule() *validate.Rule {
	return validate.Reflect(SpicedbCluster{})
}

// ValidationRule for spec.locations elements: LocalLink constrained to
// kind 'location' (was locationLinkRule = LocalLinkRule().Custom(ToKind("location"))).
func (SpicedbLocationLink) ValidationRule() *validate.Rule {
	return base.LocalLinkRule().Custom(base.ToKind("location"))
}

// ValidationRule ports `Joi.object({externalEndpoint}).unknown(true)` — the
// reflected status shape with unknown fields allowed (was RegisterObjectRules
// Unknown:true on ClusterStatus).
func (ClusterStatus) ValidationRule() *validate.Rule {
	return validate.Reflect(ClusterStatus{}).Unknown(true)
}
