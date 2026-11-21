package mk8sTriton

// validate.go attaches the TritonProvider validation rule (ported from
// pkg/handlers/schema_mk8s.go TritonProviderRule / mk8sTriton.ts).

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/controlplane-com/libs-go/pkg/schema/base"
	"github.com/controlplane-com/libs-go/pkg/schema/mk8sCommon"
	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

var tritonTagKeyRe = regexp.MustCompile(`^[:a-zA-Z0-9_.-]+$`)

func tritonTagsRule() *validate.Rule {
	return validate.Object(nil).Pattern(
		validate.String().Regex(tritonTagKeyRe).Min(1).Max(63),
		validate.String().Max(255).Min(1),
	)
}

func tritonMetadataRule() *validate.Rule {
	// mk8sTriton.ts:33 — the metadata VALUE is `Joi.string().allow('')` with no max
	// (the tags value has max 255, metadata does not); the extra .Max(255) here was a
	// Go-only rejection.
	return validate.Object(nil).Pattern(
		validate.String().Regex(tritonTagKeyRe).Min(1).Max(63),
		validate.String().Allow(""),
	)
}

func tritonPoolRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"name":              mk8sCommon.NodePoolNameRule().Required(),
		"labels":            mk8sCommon.LabelsRule(),
		"taints":            mk8sCommon.TaintsRule(),
		"maxPods":           mk8sCommon.MaxPodsRule(),
		"packageId":         validate.String().UUID().Required(),
		"overrideImageId":   validate.String().UUID(),
		"publicNetworkId":   validate.String().UUID(),
		"privateNetworkIds": validate.Array(validate.String().UUID()).Sort().Dedupe(),
		"tritonTags":        tritonTagsRule(),
		"minSize":           validate.Number().Integer().Min(0).Required(),
		"maxSize":           validate.Number().Integer().Required().MinRef("minSize"),
	})
}

// hostPortRule ports mk8sTriton.ts:51 HostPortSchema.
func hostPortRule() *validate.Rule {
	return validate.String().Trim().Custom(func(value any, h *validate.Helper) (any, error) {
		s, _ := value.(string)
		parts := strings.Split(s, ":")
		if len(parts) != 2 {
			return nil, h.Errorf("any.custom", "Must be in format host:port")
		}
		host, portStr := parts[0], parts[1]
		if strings.TrimSpace(host) == "" {
			return nil, h.Errorf("any.custom", "Host cannot be empty")
		}
		port, err := strconv.Atoi(portStr)
		if err != nil || strconv.Itoa(port) != portStr {
			return nil, h.Errorf("any.custom", "Port must be a valid number")
		}
		if port < 1 || port > 65535 {
			return nil, h.Errorf("any.custom", "Port must be between 1 and 65535")
		}
		if strings.Contains(host, " ") || strings.HasPrefix(host, ".") || strings.HasSuffix(host, ".") {
			return nil, h.Errorf("any.custom", "Invalid host format")
		}
		return s, nil
	})
}

func tritonManualRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"packageId":         validate.String().UUID().Required(),
		"imageId":           validate.String().UUID().Required(),
		"publicNetworkId":   validate.String().UUID().Required(),
		"privateNetworkIds": validate.Array(validate.String().UUID()).Sort().Dedupe(),
		"metadata":          tritonMetadataRule(),
		"tags":              tritonTagsRule(),
		"logging": validate.Object(map[string]*validate.Rule{
			"nodePort":       validate.Number().Integer().Min(30000).Max(65535),
			"externalSyslog": hostPortRule(),
		}),
		"count":             validate.Number().Integer().Min(1).Max(3).Default(float64(1)).Required(),
		"cnsInternalDomain": validate.String().Domain().Required(),
		"cnsPublicDomain":   validate.String().Domain().Required(),
	})
}

func tritonLoadBalancerRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"manual":  tritonManualRule(),
		"none":    validate.Object(nil),
		"gateway": validate.Object(nil),
	}).Xor("manual", "gateway", "none").Default(func() any { return map[string]any{"gateway": map[string]any{}} })
}

func tritonSdcConnectionRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"url":                  mk8sCommon.UriStringRule("https").Required(),
		"account":              validate.String().Required(),
		"user":                 validate.String(),
		"privateKeySecretLink": base.LocalLinkRule().Custom(base.ToKind("secret")).Required(),
	})
}

// ValidationRule ports mk8sTriton.ts:167 TritonProvider.
func (TritonProvider) ValidationRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"connection":       tritonSdcConnectionRule().Required(),
		"networking":       mk8sCommon.NetworkingConfigRule(false),
		"preInstallScript": mk8sCommon.PreInstallScriptRule(false),
		"location":         validate.String().Valid(mk8sCommon.CplnHostLocations...).Required().Unmodifiable(),
		"loadBalancer":     tritonLoadBalancerRule(),
		"privateNetworkId": validate.String().UUID().Required().Unmodifiable(),
		"firewallEnabled":  validate.Boolean(),
		"nodePools":        validate.Array(tritonPoolRule()),
		"imageId":          validate.String().Required(),
		"sshKeys":          validate.Array(mk8sCommon.SshPublicKeyRule()),
		"autoscaler":       mk8sCommon.AutoscalerConfigRule(),
	}).CustomTyped(mk8sCommon.PoolNamesUniqueObjectCustom("nodePools", "dedicatedServerNodePools"))
}
