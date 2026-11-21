package mk8sAzure

// validate.go attaches the AzureProvider validation rule (ported from
// pkg/handlers/schema_mk8s.go AzureProviderRule / mk8sAzure.ts).

import (
	"strings"

	"github.com/controlplane-com/libs-go/pkg/schema/base"
	"github.com/controlplane-com/libs-go/pkg/schema/mk8sCommon"
	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

func azureImageRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"recommended": validate.String().Valid("ubuntu/noble-24.04", "ubuntu/jammy-22.04", "ubuntu/focal-20.04", "debian/bookworm-12", "debian/bullseye-11"),
		"reference": validate.Object(map[string]*validate.Rule{
			"publisher": validate.String().Required(),
			"offer":     validate.String().Required(),
			"sku":       validate.String().Required(),
			"version":   validate.String().Required(),
		}),
	}).Xor("recommended", "reference")
}

func azureTagsRule() *validate.Rule {
	// Joi's AzureTags is a pattern-map with .max(10) — up to 10 extra tags.
	return validate.Object(nil).Pattern(
		validate.String().Min(1).Max(512).Custom(azureTagKeyCustom),
		validate.String().Allow("").Max(256),
	).Max(10)
}

func azureTagKeyCustom(value any, h *validate.Helper) (any, error) {
	s, _ := value.(string)
	for _, c := range "<>%&\\?/" {
		if strings.ContainsRune(s, c) {
			return nil, h.Errorf("any.custom", "Azure tag name cannot contain any of <>%%&\\?/")
		}
	}
	for _, prefix := range []string{"microsoft", "azure", "windows"} {
		if strings.HasPrefix(s, prefix) {
			return nil, h.Errorf("any.custom", "Azure tag name cannot start with %s", prefix)
		}
	}
	return s, nil
}

func azurePoolRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"name":          mk8sCommon.NodePoolNameRule().Required(),
		"labels":        mk8sCommon.LabelsRule(),
		"taints":        mk8sCommon.TaintsRule(),
		"maxPods":       mk8sCommon.MaxPodsRule(),
		"size":          validate.String().Required(),
		"subnetId":      validate.String().Required(),
		"zones":         validate.Array(validate.Number().Valid(float64(1), float64(2), float64(3))).Sort().Dedupe().Required(),
		"overrideImage": azureImageRule(),
		"bootDiskSize":  validate.Number().Integer().Min(30).Default(float64(30)).Required(),
		"minSize":       validate.Number().Integer().Min(0).Required(),
		"maxSize":       validate.Number().Integer().Required().MinRef("minSize"),
	})
}

// ValidationRule ports mk8sAzure.ts:130 AzureProvider.
func (AzureProvider) ValidationRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"location":         validate.String().Required().Unmodifiable(),
		"subscriptionId":   validate.String().UUID().Required().Unmodifiable(),
		"sdkSecretLink":    base.LocalLinkRule().Custom(base.ToKind("secret")).Required(),
		"resourceGroup":    validate.String().Required().Unmodifiable(),
		"networking":       mk8sCommon.NetworkingConfigRule(false),
		"preInstallScript": mk8sCommon.PreInstallScriptRule(false),
		"image":            azureImageRule().Required(),
		"sshKeys":          validate.Array(mk8sCommon.SshPublicKeyRule()).Required().Min(1),
		"networkId":        validate.String().Required().Unmodifiable(),
		"tags":             azureTagsRule(),
		"nodePools":        validate.Array(azurePoolRule()).CustomTyped(mk8sCommon.PoolNamesUniqueArrayCustom),
		"autoscaler":       mk8sCommon.AutoscalerConfigRule(),
	})
}
