package memcache

// validate.go attaches validation to the canonical memcache types via the
// Validator interface — the go-libs home of what used to be
// pkg/handlers/schema_memcache.go + the memcache registrations in
// schema_reflect_datanet.go. memcache.ts reflects `MemcacheCluster =
// Base.keys({spec: required, status})`. The `spec.locations` element rule
// (LocalLink.custom(toKind('location')), formerly RegisterType on
// MemcacheLocationLink) is expressed by MemcacheLocationLink.ValidationRule.
// ClusterStatus (Joi.object({}).unknown(true)) is the reflected map[string]any
// pattern object — permissive by construction — and carried no registration.

import (
	"github.com/controlplane-com/libs-go/pkg/schema/base"
	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// ValidationRule reflects MemcacheCluster (Base.keys({spec: required, status})).
func (MemcacheCluster) ValidationRule() *validate.Rule {
	return validate.Reflect(MemcacheCluster{})
}

// ValidationRule makes spec.locations elements self-validating: LocalLink expanded
// then kind-checked to 'location' — LocalLink.custom(toKind('location'))
// (memcache.ts), formerly RegisterType(MemcacheLocationLink, LocalLinkRule+ToKind).
func (MemcacheLocationLink) ValidationRule() *validate.Rule {
	return base.LocalLinkRule().Custom(base.ToKind("location"))
}
