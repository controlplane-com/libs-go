package orgLogging

// validate.go attaches validation to the canonical org-logging leaf types via the
// Validator interface — the go-libs home of the irreducible Joi validators that
// used to live in pkg/handlers/schema_orgLogging.go and were wired to `custom=`/
// `pattern=`/`localLink=` tags in schema_reflect_orggvc.go.
//
// Ported here as strongly-typed nominal field types (each field points at one of
// these, so the engine dispatches by Go type; the field's declarative modifier
// tags — required/optional/min/max/integer/default — layer on top):
//   - SecretLink        (was localLink=secret): a localLink pointing to a secret;
//   - FluentBitTemplate (was custom=validFluentBitTemplate): $...$ separator check;
//   - LogPattern        (was custom=validLogPattern): only {org}/{gvc}/{workload}/
//     {location} template vars allowed;
//   - EsHost            (was custom=validHostname): RFC-1123-ish hostname;
//   - EsAmazonHost      (was custom=validHostname,custom=esAmazonHost): hostname
//     that must also end with es.amazonaws.com;
//   - EsPath            (was pattern=esPath): must start with '/';
//   - FluentdPort       (was custom=requiredWithDefault): required-with-default.
//
// The region-enum leaf types (CloudWatchLoggingRegion, StackdriverLoggingLocation)
// carry their valid-set via EnumValues, and the ElasticLogging .xor is a
// ValidationRule — both at the bottom of this file.

import (
	"regexp"
	"strings"

	"github.com/controlplane-com/libs-go/pkg/schema/base"
	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// errString is a tiny error wrapper (avoids importing errors for one-liners).
type errString string

func (e errString) Error() string { return string(e) }

// label2 ports handlers.label2: the last path segment, or "value" at the root.
func label2(h *validate.Helper) string {
	if len(h.Path) == 0 {
		return "value"
	}
	return h.Path[len(h.Path)-1]
}

// --- SecretLink (was localLink=secret) ---

func (SecretLink) ValidationRule() *validate.Rule {
	return base.LocalLinkRule().Custom(base.ToKind("secret"))
}

// --- FluentBitTemplate (was custom=validFluentBitTemplate) ---

var dollarPairRe = regexp.MustCompile(`\$[^\$]*\$`)
var fluentBitSepRe = regexp.MustCompile(`[\.\,]+`)

func (FluentBitTemplate) ValidationRule() *validate.Rule {
	return validate.String().Custom(validFluentBitTemplateCustom)
}

// findAllDollarPairs ports orgLogging.ts:248: overlapping (stride-1) matches of
// `$...$` — mirrors the JS lastIndex = match.index + 1 loop.
func findAllDollarPairs(value string) []string {
	var matches []string
	lastIndex := 0
	for lastIndex <= len(value) {
		loc := dollarPairRe.FindStringIndex(value[lastIndex:])
		if loc == nil {
			break
		}
		start := lastIndex + loc[0]
		end := lastIndex + loc[1]
		matches = append(matches, value[start:end])
		lastIndex = start + 1
	}
	return matches
}

// validFluentBitTemplate ports orgLogging.ts:263: every $...$ pair must contain a
// '.' or ',' separator.
func validFluentBitTemplate(value string) error {
	for _, pair := range findAllDollarPairs(value) {
		if !fluentBitSepRe.MatchString(pair) {
			return errString("the pattern " + pair + " in " + value + " does not pass fluent bit template string validation. The only characters that can separate template variables are '.' and ','. You must also include at least one template variable")
		}
	}
	return nil
}

func validFluentBitTemplateCustom(value any, h *validate.Helper) (any, error) {
	v, _ := value.(string)
	if err := validFluentBitTemplate(v); err != nil {
		return nil, h.Errorf("any.custom", "%s", err.Error())
	}
	return v, nil
}

// --- LogPattern (was custom=validLogPattern) ---

var logTemplateVarRe = regexp.MustCompile(`\{.+?\}`)

func (LogPattern) ValidationRule() *validate.Rule {
	return validate.String().Custom(validPatternCustom)
}

// validLogPattern ports orgLogging.ts:237 validPattern: only {org}/{gvc}/{workload}/
// {location} template vars are allowed.
func validLogPattern(value string) error {
	allowed := map[string]bool{"{org}": true, "{gvc}": true, "{workload}": true, "{location}": true}
	for _, m := range logTemplateVarRe.FindAllString(value, -1) {
		if !allowed[m] {
			return errString("the pattern " + value + " contains unsupported var " + m)
		}
	}
	return nil
}

func validPatternCustom(value any, h *validate.Helper) (any, error) {
	v, _ := value.(string)
	if err := validLogPattern(v); err != nil {
		return nil, h.Errorf("any.custom", "%s", err.Error())
	}
	return v, nil
}

// --- EsHost / EsAmazonHost (was custom=validHostname[,custom=esAmazonHost]) ---

func (EsHost) ValidationRule() *validate.Rule {
	return base.HostnameRule()
}

func (EsAmazonHost) ValidationRule() *validate.Rule {
	return base.HostnameRule().Custom(esAmazonHostCustom)
}

// esAmazonHostCustom ports orgLogging.ts AwsProvider.host: the value must end with
// es.amazonaws.com (chained after the hostname() check).
func esAmazonHostCustom(value any, h *validate.Helper) (any, error) {
	v, _ := value.(string)
	if !strings.HasSuffix(v, "es.amazonaws.com") {
		return nil, h.Errorf("any.custom", "The host %s doesnt look like a valid AWS ElasticSearch host", v)
	}
	return v, nil
}

// --- EsPath (was pattern=esPath) ---

var esPathRe = regexp.MustCompile(`^/`)

func (EsPath) ValidationRule() *validate.Rule {
	return validate.String().Regex(esPathRe)
}

// --- FluentdPort (was custom=requiredWithDefault) ---

func (FluentdPort) ValidationRule() *validate.Rule {
	return validate.Number().Custom(requiredWithDefaultCustom)
}

// requiredWithDefaultCustom ports base.ts:542 requiredWithDefault(): error if the
// value is absent (only reachable when a .default() was forgotten; here inert).
func requiredWithDefaultCustom(value any, h *validate.Helper) (any, error) {
	if value == nil {
		return nil, h.Errorf("any.required", "%q is required", label2(h))
	}
	return value, nil
}

// --- region enums (orgLogging.ts:5-66 GCP_REGIONS / AWS_REGIONS) ---

// awsRegions is orgLogging.ts:48's AWS_REGIONS, in declaration order: Joi prints the
// valids in that order in the any.only message, so the order is load-bearing.
var awsRegions = []string{
	"us-east-1", "us-east-2", "us-west-1", "us-west-2", "ap-south-1",
	"ap-northeast-2", "ap-southeast-1", "ap-southeast-2", "ap-northeast-1",
	"eu-central-1", "eu-west-1", "eu-west-2", "eu-south-1", "eu-west-3",
	"eu-north-1", "me-south-1", "sa-east-1", "af-south-1",
}

// gcpRegions is orgLogging.ts:5's GCP_REGIONS, in declaration order.
var gcpRegions = []string{
	"us-east1", "us-east4", "us-central1", "us-west1", "europe-west4",
	"europe-west1", "europe-west3", "europe-west2", "asia-east1",
	"asia-southeast1", "asia-northeast1", "asia-south1", "australia-southeast1",
	"southamerica-east1", "africa-south1", "asia-east2", "asia-northeast2",
	"asia-northeast3", "asia-south2", "asia-southeast2", "australia-southeast2",
	"europe-central2", "europe-north1", "europe-southwest1", "europe-west10",
	"europe-west12", "europe-west6", "europe-west8", "europe-west9",
	"me-central1", "me-central2", "me-west1", "northamerica-northeast1",
	"northamerica-northeast2", "southamerica-west1", "us-east5", "us-south1",
	"us-west2", "us-west3", "us-west4",
}

// EnumValues ports CloudWatchLogging.region's .valid(...AWS_REGIONS). Without it the
// field was an unconstrained string and an org could be pointed at a region that does
// not exist (the failure then surfaced in the log pipeline, not at the API).
func (CloudWatchLoggingRegion) EnumValues() []string { return awsRegions }

// EnumValues ports StackdriverLogging.location's .valid(...GCP_REGIONS).
func (StackdriverLoggingLocation) EnumValues() []string { return gcpRegions }

// --- ElasticLogging (orgLogging.ts:224-228) ---

// ValidationRule layers the .xor('elasticCloud','aws','generic') that only a models.go
// comment recorded. Peer order is the .xor() argument order: Joi renders the peers in
// that order in both the object.xor and the object.missing message.
func (ElasticLogging) ValidationRule() *validate.Rule {
	return validate.Reflect(ElasticLogging{}).Xor("elasticCloud", "aws", "generic")
}
