package cloudaccount

// validate.go attaches validation to the canonical cloudaccount types via the
// Validator interface — the go-libs home of what used to be
// pkg/handlers/schema_cloudaccount.go (the CloudAccount*Rule FromType shims) plus
// pkg/handlers/schema_reflect_cloudaccount.go (the RegisterPattern awsRoleArn /
// gcpProjectId, the base.CloudProvider mapping, and the cloudAccountRefine).
//
// cloudaccount.ts reflects from CloudAccount. The `data` field is a
// Joi.alternatives().conditional('provider', …); a field tag cannot carry a
// cross-field conditional, so Data stays `any` and the provider→credential-body
// selection is expressed by the object-level .Refine below (cloudAccountRefine),
// validating `data` against the selected config type and writing the normalized
// output (the expanded azure/ngs secretLink) back — exactly as the old registered
// schema did.
//
// The old `pattern=`/`localLink=` tags are replaced by strongly-typed field
// types: AwsRoleArn, GcpProjectId, and SecretLink (a LocalLink kind-checked to
// "secret").

import (
	"regexp"

	"github.com/controlplane-com/libs-go/pkg/schema/base"
	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// ValidationRule reflects CloudAccount and layers the `data`
// alternatives().conditional('provider', …). This is the schema the
// "cloudaccount" kind validates against.
func (CloudAccount) ValidationRule() *validate.Rule {
	return validate.Reflect(CloudAccount{}).CustomTyped(cloudAccountRefine)
}

// cloudAccountRefine ports `data: Joi.alternatives().conditional('provider', [...])`
// (cloudaccount.ts:32):
// the provider scalar selects the credential body (gcp→GcpConfig, aws→AwsConfig,
// azure→AzureConfig, ngs→NgsConfig); an unmatched provider leaves `data`
// unconstrained (Joi's conditional with no matching branch and no otherwise). The
// selected shape is validated against `data`, and its normalized output (e.g. the
// expanded azure/ngs secretLink) is written back — matching the old builder.
func cloudAccountRefine(ca any, h *validate.Helper) (any, error) {
	var rule *validate.Rule
	switch validate.Str(ca, "provider") {
	case "gcp":
		rule = GcpConfig{}.ValidationRule()
	case "aws":
		rule = AwsConfig{}.ValidationRule()
	case "azure":
		rule = AzureConfig{}.ValidationRule()
	case "ngs":
		rule = NgsConfig{}.ValidationRule()
	default:
		return ca, nil // no matching branch → data unconstrained
	}

	data := validate.Get(ca, "data")
	if data == nil {
		if !dataPresent(ca, h) {
			return ca, nil
		}
		path := append(append([]string(nil), h.Path...), "data")
		err := &validate.Error{Type: "object.base", Message: `"data" must be of type object`, Path: path}
		return nil, h.WithCtx(err, "type", "object", "value", nil)
	}

	// ValidateAt, not Validate: validating from a fresh root reported an unknown key
	// inside `data` at path ["x"] with label "x" where TS says ["data","x"] / "data.x".
	out, verr := rule.ValidateAt(append(append([]string(nil), h.Path...), "data"), data, h.Ctx, validate.Prefs{})
	if verr != nil {
		return nil, verr
	}
	validate.Set(ca, "data", out)
	return ca, nil
}

// dataPresent tells a `data: null` body from one without data.
func dataPresent(ca any, h *validate.Helper) bool {
	if m, ok := ca.(map[string]any); ok {
		_, present := m["data"]
		return present
	}
	return h.Present("data")
}

// --- nested credential bodies (reflected; no object-level rules of their own) ---

// ValidationRule reflects AwsConfig (was CloudAccountAwsConfigRule / Joi AwsConfig).
func (AwsConfig) ValidationRule() *validate.Rule { return validate.Reflect(AwsConfig{}) }

// ValidationRule reflects GcpConfig (was CloudAccountGcpConfigRule / Joi GcpConfig).
func (GcpConfig) ValidationRule() *validate.Rule { return validate.Reflect(GcpConfig{}) }

// ValidationRule reflects AzureConfig (was CloudAccountAzureConfigRule / Joi AzureConfig).
func (AzureConfig) ValidationRule() *validate.Rule { return validate.Reflect(AzureConfig{}) }

// ValidationRule reflects NgsConfig (was CloudAccountNgsConfigRule / Joi NgsConfig).
func (NgsConfig) ValidationRule() *validate.Rule { return validate.Reflect(NgsConfig{}) }

// --- nominal leaf types (replace the old pattern=/localLink= tags) ---

// awsRoleArnRe ports AwsConfig.roleArn's regex (nodelibs/schema/src/cloudaccount.ts:6-8),
// formerly RegisterPattern("cloudAccountAwsRoleArn", …).
var awsRoleArnRe = regexp.MustCompile(`^arn:(aws|aws-us-gov|aws-cn):iam::[0-9]+:role/.+`)

func (AwsRoleArn) ValidationRule() *validate.Rule { return validate.String().Regex(awsRoleArnRe) }

// gcpProjectIDRe ports GcpConfig.projectId's regex (cloudaccount.ts:14-18),
// formerly RegisterPattern("gcpProjectId", …). Joi applies min/max separately, so
// those stay declarative tags on the field.
var gcpProjectIDRe = regexp.MustCompile(`[a-z]([a-z]|-|[0-9])+`)

func (GcpProjectId) ValidationRule() *validate.Rule { return validate.String().Regex(gcpProjectIDRe) }

func (SecretLink) ValidationRule() *validate.Rule {
	return base.LocalLinkRule().Custom(base.ToKind("secret"))
}
