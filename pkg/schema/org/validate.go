package org

// validate.go attaches validation to the canonical org types via the Validator
// interface — the go-libs home of what used to be pkg/handlers/schema_org.go +
// the org registrations in schema_reflect_orggvc.go. org.ts reflects from Org; the
// object-level "complex parts of Joi" that a field tag cannot carry are layered
// here as .Xor / .Refine on the affected types' ValidationRule methods:
//   - OrgSpecLogging: the logCfg .xor over the 11 log providers;
//   - ThreatDetection: .xor('syslog');
//   - OrgSpec: the observability whole-object default + the extraLogging
//     .when('logging') guard (orgSpecRefine).
// The former `custom=`/`pattern=` tags are replaced by strongly-typed nominal
// field types (OrgName, OrgAccountLink, SyslogHostname, AlertEmail, DomainName,
// and the DomainNameList / StringDedupeList / QuotaOverrideList collection types).

import (
	"github.com/controlplane-com/libs-go/pkg/schema/util"
	"regexp"
	"strings"

	"github.com/controlplane-com/libs-go/pkg/schema/base"
	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// --- top-level + nested object rules ---

// ValidationRule reflects Org. This is the schema the "org" kind validates against.
func (Org) ValidationRule() *validate.Rule { return validate.Reflect(Org{}) }

// ValidationRule reflects OrgSpec and layers the two object-level rules a field tag
// cannot carry: the extraLogging .when('logging') guard and the observability
// whole-object default (orgSpecRefine).
func (OrgSpec) ValidationRule() *validate.Rule {
	// extraLogging is `.when('logging', {not: Joi.exist(), then: Joi.forbidden()})`
	// (org.ts:80). Expressed as the .when() Joi spells rather than as a refine, so the
	// rejection is raised AT spec.extraLogging (with the value in context) instead of
	// at the parent, and so describe() shows the branch.
	return validate.RefineT(validate.Reflect(OrgSpec{}).Fork("extraLogging", func(r *validate.Rule) *validate.Rule {
		return r.When("logging", validate.When{
			Not:  validate.Any().Required(),
			Then: validate.Any().Forbidden(),
		})
	}), orgSpecRefine)
}

// ValidationRule reflects OrgSpecLogging and layers the logCfg .xor: exactly one of
// the 11 log providers may be present.
func (OrgSpecLogging) ValidationRule() *validate.Rule {
	return validate.Reflect(OrgSpecLogging{}).Xor(
		"s3", "coralogix", "datadog", "logzio", "elastic", "cloudWatch",
		"fluentd", "stackdriver", "syslog", "opentelemetry", "loki",
	)
}

// ValidationRule reflects ThreatDetection and layers .xor('syslog') (org.ts:58).
func (ThreatDetection) ValidationRule() *validate.Rule {
	return validate.Reflect(ThreatDetection{}).Xor("syslog")
}

// --- leaf nominal types ---

// ValidationRule ports the AlertEmail leaf: Joi.string().email().
func (AlertEmail) ValidationRule() *validate.Rule { return validate.String().Email() }

// ValidationRule ports the DomainName leaf: Joi.string().domain().
func (DomainName) ValidationRule() *validate.Rule { return validate.String().Domain() }

func (OrgName) ValidationRule() *validate.Rule {
	return base.Name("").ValidationRule().Custom(orgNameCustom)
}

// orgNameCustom ports org.ts:101 Name.custom: reject xserve- prefix and "index".
func orgNameCustom(value any, h *validate.Helper) (any, error) {
	s, _ := value.(string)
	if strings.HasPrefix(s, "xserve-") {
		return nil, h.Errorf("any.custom", "org name cannot start with xserve-")
	}
	if s == "index" {
		return nil, h.Errorf("any.custom", `org name cannot be "index"`)
	}
	return s, nil
}

var orgAccountLinkRe = regexp.MustCompile(`^/account/[a-f0-9-]+$`)

func (OrgAccountLink) ValidationRule() *validate.Rule {
	return validate.String().Regex(orgAccountLinkRe)
}

// syslogHostRe is the RE2 equivalent of org.ts:71's pattern, whose (?!-)/(?<!-)
// lookarounds RE2 cannot express: dot-separated labels of [A-Za-z0-9-]{1,63} that
// neither start nor end with '-'. syslogHostJS is what the MESSAGE and the describe
// document must show — the pattern the schema was written with.
var (
	syslogHostRe = regexp.MustCompile(`^[A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?(\.[A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?)*$`)
	syslogHostJS = `^(?!-)[A-Za-z0-9-]{1,63}(?<!-)(\.(?!-)[A-Za-z0-9-]{1,63}(?<!-))*$`
)

// ValidationRule ports org.ts:71's syslog host pattern. It was a hand-rolled custom
// raising a bare string.pattern.base that quoted the VALUE and carried neither the
// regex nor the value in context; RegexAs gives Joi's message and context for free.
func (SyslogHostname) ValidationRule() *validate.Rule {
	return validate.String().RegexAs(syslogHostRe, syslogHostJS)
}

// --- collection nominal types (dedupe/sort over arrays) ---

func (DomainNameList) ValidationRule() *validate.Rule {
	return validate.Array(DomainName("").ValidationRule()).Sort().Dedupe()
}

func (StringDedupeList) ValidationRule() *validate.Rule {
	return validate.Array(validate.String()).Dedupe()
}

func (QuotaOverrideList) ValidationRule() *validate.Rule {
	return validate.Array(validate.Reflect(QuotaOverride{})).Dedupe().SortBy("name")
}

// --- OrgSpec object-level refine ---

// observabilityDefault is the whole-object default for OrgSpec.observability
// (org.ts:88): a fully-populated config so an absent observability still gets 30s.
// It is content-ful, so a field `default=` tag cannot carry it — orgSpecRefine
// injects it.
func observabilityDefault() *ObservabilityConfig {
	return &ObservabilityConfig{
		LogsRetentionDays:    util.Ptr(30.0),
		MetricsRetentionDays: util.Ptr(30.0),
		TracesRetentionDays:  util.Ptr(30.0),
		DefaultAlertEmails:   []AlertEmail{},
	}
}

// orgSpecRefine carries the OrgSpec object-level rule a field tag cannot express: the
// observability whole-object default (org.ts:88). The extraLogging .when() guard is
// declared on the field itself (see OrgSpec.ValidationRule).
func orgSpecRefine(spec *OrgSpec, h *validate.Helper) error {
	if !h.Present("observability") {
		spec.Observability = observabilityDefault()
	}
	return nil
}
