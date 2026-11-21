package mk8sGeneric

// validate.go attaches the GenericProvider validation rule (ported from
// pkg/handlers/schema_mk8s.go GenericProviderRule / mk8sGeneric.ts).

import (
	"github.com/controlplane-com/libs-go/pkg/schema/mk8sCommon"
	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

func genericPoolRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"name":    mk8sCommon.NodePoolNameRule().Required(),
		"labels":  mk8sCommon.LabelsRule(),
		"taints":  mk8sCommon.TaintsRule(),
		"maxPods": mk8sCommon.MaxPodsRule(),
	})
}

// ValidationRule ports mk8sGeneric.ts:52 GenericProvider.
func (GenericProvider) ValidationRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"location":   validate.String().Valid(mk8sCommon.CplnHostLocations...).Required().Unmodifiable(),
		"networking": mk8sCommon.NetworkingConfigRule(false),
		"nodePools":  validate.Array(genericPoolRule()),
	}).CustomTyped(mk8sCommon.PoolNamesUniqueObjectCustom("nodePools"))
}
