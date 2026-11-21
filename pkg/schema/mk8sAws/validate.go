package mk8sAws

// validate.go attaches the AwsProvider validation rule (ported from
// pkg/handlers/schema_mk8s.go AwsProviderRule / mk8sAws.ts). The provider body is a
// hand-built Object tree (its deep constructs exceed the tag DSL), resolved through
// the Validator interface when mk8s.Mk8SClusterSpecProvider reflects its `aws` field.

import (
	"regexp"

	"github.com/controlplane-com/libs-go/pkg/schema/mk8sCommon"
	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

var (
	awsRoleArnMk8sRe   = regexp.MustCompile(`^arn:(aws):iam::[0-9]+:role/.+$`)
	awsPolicyArnRe     = regexp.MustCompile(`^arn:(aws):iam::([0-9]+|aws):policy/.+$`)
	awsSessionPrefixRe = regexp.MustCompile(`[\w+=,.@-]*`)
	awsAmiExactRe      = regexp.MustCompile(`^(ami-[0-9a-f]+)$|^(resolve:ssm:.+)$`)
	awsSecurityGroupRe = regexp.MustCompile(`^(sg-[a-f0-9]+)$|^(resolve:ssm:.+)$`)
	awsSubnetRe        = regexp.MustCompile(`^(subnet-[a-f0-9]+)$|^(resolve:ssm:.+)$`)
	awsVpcIDRe         = regexp.MustCompile(`^(vpc-[a-f0-9]+)$|^(resolve:ssm:.+)$`)
	awsKmsArnRe        = regexp.MustCompile(`^(arn:aws:kms:.+)$|^(resolve:ssm:.+)$`)
	awsTagCharRe       = regexp.MustCompile(`^[ a-zA-Z0-9=_+:./@-]+$`)
)

var mk8sAwsRegions = []any{
	"af-south-1", "ap-east-1", "ap-northeast-1", "ap-northeast-2", "ap-northeast-3",
	"ap-south-1", "ap-south-2", "ap-southeast-1", "ap-southeast-2", "ap-southeast-3",
	"ap-southeast-4", "ca-central-1", "eu-central-1", "eu-central-2", "eu-north-1",
	"eu-south-1", "eu-south-2", "eu-west-1", "eu-west-2", "eu-west-3", "me-central-1",
	"me-south-1", "sa-east-1", "us-east-1", "us-east-2", "us-west-1", "us-west-2",
}

// policyArnRule ports mk8sAws.ts:19 PolicyArn.
func policyArnRule() *validate.Rule { return validate.String().Regex(awsPolicyArnRe) }

func awsAmiRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"recommended": validate.String().Valid(
			"ubuntu/jammy-22.04", "ubuntu/jammy-22.04+nvidia", "ubuntu/focal-20.04",
			"ubuntu/focal-20.04+nvidia", "ubuntu/noble-24.04", "ubuntu/noble-24.04+nvidia",
			"ubuntu/bionic-18.04", "amazon/amzn2", "amazon/al2023",
		),
		"exact": validate.String().Regex(awsAmiExactRe),
	}).Xor("recommended", "exact")
}

func awsAssumeRoleLinkRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"roleArn":           validate.String().Regex(awsRoleArnMk8sRe).Required(),
		"externalId":        validate.String(),
		"sessionNamePrefix": validate.String().Regex(awsSessionPrefixRe).Max(50),
	})
}

func awsTagsRule() *validate.Rule {
	// Joi's AwsTags is a pattern-map with .max(10) — up to 10 extra tags.
	return validate.Object(nil).Pattern(
		validate.String().Regex(awsTagCharRe).Min(1).Max(128),
		validate.String().Regex(awsTagCharRe).Allow("").Max(256),
	).Max(10)
}

func awsPoolRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"name":                                mk8sCommon.NodePoolNameRule().Required(),
		"labels":                              mk8sCommon.LabelsRule(),
		"taints":                              mk8sCommon.TaintsRule(),
		"maxPods":                             mk8sCommon.MaxPodsRule(),
		"instanceTypes":                       validate.Array(validate.String()).Min(1).Required().Dedupe(),
		"overrideImage":                       awsAmiRule(),
		"bootDiskSize":                        validate.Number().Integer().Min(20).Default(float64(20)).Required(),
		"minSize":                             validate.Number().Integer().Min(0).Required(),
		"maxSize":                             validate.Number().Integer().Required().MinRef("minSize"),
		"onDemandBaseCapacity":                validate.Number().Integer().Min(0).MaxRef("maxSize").Default(float64(0)).Required(),
		"onDemandPercentageAboveBaseCapacity": validate.Number().Integer().Min(0).Max(100).Default(float64(0)).Required(),
		"spotAllocationStrategy":              validate.String().Valid("lowest-price", "capacity-optimized", "capacity-optimized-prioritized", "price-capacity-optimized").Default("lowest-price"),
		"subnetIds":                           validate.Array(validate.String().Regex(awsSubnetRe)).Min(1).Required().Dedupe(),
		"extraSecurityGroupIds":               validate.Array(validate.String().Regex(awsSecurityGroupRe)).Dedupe(),
		"cpuOptions": validate.Object(map[string]*validate.Rule{
			"nestedVirtualization": validate.Boolean(),
		}),
	})
}

// ValidationRule ports mk8sAws.ts:168 AwsProvider.
func (AwsProvider) ValidationRule() *validate.Rule {
	return validate.Object(map[string]*validate.Rule{
		"region":               validate.String().Valid(mk8sAwsRegions...).Required().Unmodifiable(),
		"awsTags":              awsTagsRule(),
		"skipCreateRoles":      validate.Boolean().Default(false),
		"networking":           mk8sCommon.NetworkingConfigRule(true),
		"preInstallScript":     mk8sCommon.PreInstallScriptRule(true),
		"image":                awsAmiRule().Required(),
		"deployRoleArn":        validate.String().Regex(awsRoleArnMk8sRe).Required(),
		"deployRoleChain":      validate.Array(awsAssumeRoleLinkRule()),
		"vpcId":                validate.String().Regex(awsVpcIDRe).Required().Unmodifiable(),
		"keyPair":              validate.String(),
		"diskEncryptionKeyArn": validate.String().Regex(awsKmsArnRe),
		"securityGroupIds":     validate.Array(validate.String().Regex(awsSecurityGroupRe)).Dedupe(),
		"extraNodePolicies":    validate.Array(policyArnRule()),
		"nodePools":            validate.Array(awsPoolRule()).CustomTyped(mk8sCommon.PoolNamesUniqueArrayCustom),
		"autoscaler":           mk8sCommon.AutoscalerConfigRule(),
	})
}
