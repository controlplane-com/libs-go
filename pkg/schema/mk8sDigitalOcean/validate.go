package mk8sDigitalOcean

// validate.go attaches the DigitalOceanProvider validation rule (ported from
// pkg/handlers/schema_mk8s.go DigitalOceanProviderRule / mk8sDigitalOcean.ts).

import (
	"regexp"

	"github.com/controlplane-com/libs-go/pkg/schema/base"
	"github.com/controlplane-com/libs-go/pkg/schema/mk8sCommon"
	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

var doValidTagRe = regexp.MustCompile(`^[a-zA-Z0-9:_-]+$`)

// doValidTagRule ports mk8sDigitalOcean.ts:17 ValidTag.
func doValidTagRule() *validate.Rule { return validate.String().Regex(doValidTagRe).Max(254) }

func doPoolRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"name":          mk8sCommon.NodePoolNameRule().Required(),
		"labels":        mk8sCommon.LabelsRule(),
		"taints":        mk8sCommon.TaintsRule(),
		"maxPods":       mk8sCommon.MaxPodsRule(),
		"dropletSize":   validate.String().Required(),
		"overrideImage": validate.String(),
		"minSize":       validate.Number().Integer().Min(0).Required(),
		"maxSize":       validate.Number().Integer().Required().MinRef("minSize"),
	})
}

// ValidationRule ports mk8sDigitalOcean.ts:77 DigitalOceanProvider.
func (DigitalOceanProvider) ValidationRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"region":           validate.String().Valid("ams3", "blr1", "fra1", "lon1", "nyc1", "nyc2", "nyc3", "sfo2", "sfo3", "sgp1", "syd1", "tor1").Required().Unmodifiable(),
		"digitalOceanTags": validate.Array(doValidTagRule()).Sort().Dedupe(),
		"networking":       mk8sCommon.NetworkingConfigRule(false),
		"preInstallScript": mk8sCommon.PreInstallScriptRule(false),
		"tokenSecretLink":  base.LocalLinkRule().Custom(base.ToKind("secret")).Required(),
		"vpcId":            validate.String().UUID().Required().Unmodifiable(),
		"nodePools":        validate.Array(doPoolRule()),
		"image":            validate.String().Required(),
		"sshKeys":          validate.Array(validate.String()).Min(1).Required(),
		"extraSshKeys":     validate.Array(mk8sCommon.SshPublicKeyRule()),
		"autoscaler":       mk8sCommon.AutoscalerConfigRule(),
		"reservedIPs":      validate.Array(mk8sCommon.IpStringRule([]string{"ipv4"}, "forbidden")).Sort().Dedupe(),
	}).CustomTyped(mk8sCommon.PoolNamesUniqueObjectCustom("nodePools"))
}
