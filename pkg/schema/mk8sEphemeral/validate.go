package mk8sEphemeral

// validate.go attaches the EphemeralProvider validation rule (ported from
// pkg/handlers/schema_mk8s.go EphemeralProviderRule / mk8sEphemeral.ts).

import (
	"github.com/controlplane-com/libs-go/pkg/schema/mk8sCommon"
	"github.com/controlplane-com/libs-go/pkg/schema/validate"
	"github.com/controlplane-com/libs-go/pkg/schema/workload"
)

func ephemeralPoolRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"name":    mk8sCommon.NodePoolNameRule().Required(),
		"labels":  mk8sCommon.LabelsRule(),
		"taints":  mk8sCommon.TaintsRule(),
		"maxPods": mk8sCommon.MaxPodsRule(),
		"count":   validate.Number().Integer().Min(0).Max(4).Required(),
		"arch":    validate.String().Valid("any", "arm64", "amd64").Required(),
		"flavor":  validate.String().Valid("debian", "ubuntu", "fedora").Default("debian").Required(),
		"cpu":     workload.Cpu("").ValidationRule().Required(),
		"memory":  workload.Memory("").ValidationRule().Required(),
	})
}

// ValidationRule ports mk8sEphemeral.ts:31 EphemeralProvider.
func (EphemeralProvider) ValidationRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"location":  validate.String().Valid(mk8sCommon.CplnHostLocations...).Required().Unmodifiable(),
		"nodePools": validate.Array(ephemeralPoolRule()).CustomTyped(mk8sCommon.PoolNamesUniqueArrayCustom),
	})
}
