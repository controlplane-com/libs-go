package dbcluster

// validate.go attaches validation to the canonical dbcluster types via the
// Validator interface — the go-libs home of what used to be
// pkg/handlers/schema_dbcluster.go + the dbcluster registrations in
// schema_reflect_datanet.go. dbcluster.ts reflects `DbCluster = Base.keys({spec,
// status, alias})`; there is no createSchema fork (no description-defaults-to-name
// refine for this kind). The old dbWritable array custom tag on
// ClusterSpecPostgres.Locations is replaced by the strongly-typed
// ClusterSpecPostgresLocationsList field type, and ClusterStatus.unknown(true)
// (formerly a RegisterObjectRules) is layered in ClusterStatus.ValidationRule.

import (
	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// ValidationRule reflects DbCluster. This is the schema the "dbcluster" kind
// validates against (Base.keys({spec, status, alias})).
func (DbCluster) ValidationRule() *validate.Rule {
	return validate.Reflect(DbCluster{})
}

// ValidationRule reflects ClusterStatus and layers .unknown(true) — the
// `Joi.object({parentId, externalEndpoint}).unknown(true)` of dbcluster.ts
// (formerly RegisterObjectRules(ClusterStatus{}, {Unknown: true})).
func (ClusterStatus) ValidationRule() *validate.Rule {
	return validate.Reflect(ClusterStatus{}).Unknown(true)
}

// ValidationRule builds Array(<location>) and attaches the single-writable custom.
func (ClusterSpecPostgresLocationsList) ValidationRule() *validate.Rule {
	return validate.Array(validate.Reflect(ClusterSpecPostgresLocations{})).CustomTyped(validate.CustomT(dbWritable))
}

// dbWritable ports the PostgresConfig.locations single-writable custom
// (dbcluster.ts:51-66): exactly one location may be writable, and at least one
// must be.
func dbWritable(locs *ClusterSpecPostgresLocationsList, h *validate.Helper) error {
	foundWritable := false
	for _, l := range *locs {
		if l.Writable != nil && *l.Writable {
			if foundWritable {
				return h.Errorf("any.custom", "Multiple locations marked as writable")
			}
			foundWritable = true
		}
	}
	if !foundWritable {
		return h.Errorf("any.custom", "At least one location must be marked as writable")
	}
	return nil
}
