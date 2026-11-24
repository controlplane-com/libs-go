/* hand-maintained — canonical source of truth for the `org` resource schema.
 *
 * `validate:` struct tags drive pkg/validate.FromType (reflection) to build the
 * *validate.Rule tree, replacing the hand-written pkg/handlers/schema_org.go
 * builder. See nodelibs/schema/src/org.ts for the original Joi source and the tag
 * DSL documented on go-libs/schema/ipSet/models.go.
 *
 * Validation is attached via the Validator interface in this package's
 * validate.go: the top Org reflects via validate.Reflect, and the object-level
 * "complex parts of Joi" that a field tag cannot carry — the logCfg .xor over 11
 * providers, ThreatDetection .xor('syslog'), the OrgSpec observability whole-object
 * default, and the extraLogging .when('logging') guard — are layered as .Xor /
 * .Refine on those types' ValidationRule methods. The former registry-backed
 * tags are replaced by strongly-typed nominal field types (also in validate.go).
 */

package org

import (
	"github.com/controlplane-com/libs-go/pkg/schema/base"
	"github.com/controlplane-com/libs-go/pkg/schema/orgLogging"
	"github.com/controlplane-com/libs-go/pkg/schema/tracing"
)

// AlertEmail is an email-constrained leaf (ValidationRule -> String().Email()).
type AlertEmail string

// DomainName is a domain-constrained leaf (ValidationRule -> String().Domain()).
type DomainName string

// OrgTags aliases base.Tags so the reflection base-type registry resolves it to
// the shared TagsRule.
type OrgTags = base.Tags

type OrgStatus struct {
	AccountLink    OrgAccountLink `json:"accountLink,omitzero"`
	Active         *bool          `json:"active,omitzero"`
	EndpointPrefix string         `json:"endpointPrefix,omitzero"`
}

type AuthConfig struct {
	DomainAutoMembers DomainNameList     `json:"domainAutoMembers,omitzero"`
	SamlOnly          *bool              `json:"samlOnly,omitzero"`
	AllowedProviders  SignInProviderList `json:"allowedProviders,omitzero"`
}

// SignInProvider is a sign-in method an org can allow (see signin.go).
type SignInProvider string

// SignInProviderList ports AuthConfig.allowedProviders: sorted, duplicates removed.
type SignInProviderList []SignInProvider

// ObservabilityConfig ports org.ts:46. The retention fields default to 30; the
// whole-object default (org.ts:88) is injected by orgSpecRefine when absent.
type ObservabilityConfig struct {
	LogsRetentionDays    *float64     `json:"logsRetentionDays,omitzero" validate:"integer,min=0,max=3650,default=30"`
	MetricsRetentionDays *float64     `json:"metricsRetentionDays,omitzero" validate:"integer,min=0,max=3650,default=30"`
	TracesRetentionDays  *float64     `json:"tracesRetentionDays,omitzero" validate:"integer,min=0,max=3650,default=30"`
	DefaultAlertEmails   []AlertEmail `json:"defaultAlertEmails,omitzero" validate:"default=[]"`
}

// OrgSpecObservability aliases ObservabilityConfig (org.ts uses one type).
type OrgSpecObservability = ObservabilityConfig

type ThreatDetectionMinimumSeverity string

const (
	ThreatDetectionMinimumSeverityWarning  ThreatDetectionMinimumSeverity = "warning"
	ThreatDetectionMinimumSeverityError    ThreatDetectionMinimumSeverity = "error"
	ThreatDetectionMinimumSeverityCritical ThreatDetectionMinimumSeverity = "critical"
)

type ThreatDetectionSyslogTransport string

const (
	ThreatDetectionSyslogTransportTcp ThreatDetectionSyslogTransport = "tcp"
	ThreatDetectionSyslogTransportUdp ThreatDetectionSyslogTransport = "udp"
)

type ThreatDetectionSyslog struct {
	Transport ThreatDetectionSyslogTransport `json:"transport,omitzero" validate:"valid=tcp|udp,default=tcp"`
	Host      SyslogHostname                 `json:"host" validate:"required"`
	Port      float64                        `json:"port" validate:"integer,min=1,max=100000,required"`
}

// ThreatDetection ports org.ts:58. .xor('syslog') is registered on this type.
type ThreatDetection struct {
	Enabled         bool                           `json:"enabled" validate:"required"`
	MinimumSeverity ThreatDetectionMinimumSeverity `json:"minimumSeverity,omitzero" validate:"optional,valid=warning|error|critical"`
	Syslog          *ThreatDetectionSyslog         `json:"syslog,omitzero"`
}

// OrgSpecLogging ports the logCfg object (org.ts:32) — exactly one of the 11
// providers (.xor registered on this type).
type OrgSpecLogging struct {
	S3            *orgLogging.S3Logging            `json:"s3,omitzero"`
	Coralogix     *orgLogging.CoralogixLogging     `json:"coralogix,omitzero"`
	Datadog       *orgLogging.DatadogLogging       `json:"datadog,omitzero"`
	Logzio        *orgLogging.LogzioLogging        `json:"logzio,omitzero"`
	Elastic       *orgLogging.ElasticLogging       `json:"elastic,omitzero"`
	CloudWatch    *orgLogging.CloudWatchLogging    `json:"cloudWatch,omitzero"`
	Fluentd       *orgLogging.FluentdLogging       `json:"fluentd,omitzero"`
	Stackdriver   *orgLogging.StackdriverLogging   `json:"stackdriver,omitzero"`
	Syslog        *orgLogging.SyslogLogging        `json:"syslog,omitzero"`
	Opentelemetry *orgLogging.OpenTelemetryLogging `json:"opentelemetry,omitzero"`
	Loki          *orgLogging.LokiLogging          `json:"loki,omitzero"`
}

// OrgSpecExtraLogging aliases OrgSpecLogging (extraLogging items share logCfg).
type OrgSpecExtraLogging = OrgSpecLogging

type OrgSpecSecurity struct {
	ThreatDetection *ThreatDetection `json:"threatDetection,omitzero"`
}

// OrgSpec ports org.ts:79. The observability whole-object default and the
// extraLogging .when('logging') guard live in orgSpecRefine.
type OrgSpec struct {
	Logging               *OrgSpecLogging       `json:"logging,omitzero"`
	ExtraLogging          []OrgSpecExtraLogging `json:"extraLogging,omitzero" validate:"max=3,unique"`
	Tracing               *tracing.Tracing      `json:"tracing,omitzero"`
	SessionTimeoutSeconds *float64              `json:"sessionTimeoutSeconds,omitzero" validate:"min=900"`
	AuthConfig            *AuthConfig           `json:"authConfig,omitzero"`
	Observability         *OrgSpecObservability `json:"observability,omitzero"`
	Security              *OrgSpecSecurity      `json:"security,omitzero"`
}

// Org ports org.ts:100 Org = Base.keys({name, spec, status}). name carries the
// xserve-/index guard (OrgName nominal type). Joi's .keys() APPENDS, so a key Base
// already declares is re-declared at the END — which is why `name` sits between
// `links` and `spec` here and not where Base puts it. The order is load-bearing:
// abortEarly reports the first failing key in declaration order.
type Org struct {
	Id           string     `json:"id,omitzero" validate:"uuid,readOnly"`
	Kind         base.Kind  `json:"kind,omitzero"`
	Version      *float64   `json:"version,omitzero" validate:"positive,integer,readOnly"`
	Description  string     `json:"description,omitzero" validate:"trim,max=250,allow="`
	Tags         OrgTags    `json:"tags,omitzero" validate:"default={}"`
	Created      string     `json:"created,omitzero" validate:"date,readOnly"`
	LastModified string     `json:"lastModified,omitzero" validate:"date,readOnly"`
	Links        base.Links `json:"links,omitzero" validate:"readOnly"`
	Name         OrgName    `json:"name,omitzero"`
	Spec         *OrgSpec   `json:"spec,omitzero"`
	Status       *OrgStatus `json:"status,omitzero"`
}

type QuotaOverride struct {
	Name string  `json:"name" validate:"required"`
	Max  float64 `json:"max" validate:"required,positive"`
}

type OrgConfig struct {
	AwsPrivateLinks    StringDedupeList  `json:"awsPrivateLinks,omitzero"`
	GcpServiceConnects StringDedupeList  `json:"gcpServiceConnects,omitzero"`
	QuotaOverrides     QuotaOverrideList `json:"quotaOverrides,omitzero"`
}

// --- validation leaf & collection types (ValidationRule methods live in validate.go) ---

// OrgName is the org name leaf: base Name (regex/max-64/unmodifiable) plus the
// xserve-/index guard — the strongly-typed replacement for the `custom=orgName`
// tag on Org.name.
type OrgName string

// OrgAccountLink is the OrgStatus.accountLink leaf — the strongly-typed replacement
// for the `pattern=orgAccountLink` tag.
type OrgAccountLink string

// SyslogHostname is the ThreatDetectionSyslog.host leaf — the strongly-typed
// replacement for the `custom=syslogHostname` tag.
type SyslogHostname string

// DomainNameList ports AuthConfig.domainAutoMembers `sort,custom=dedupeEmpty`: an
// array of domain leaves, sorted ascending, with duplicates removed in place.
type DomainNameList []DomainName

// StringDedupeList ports the OrgConfig aws/gcp `custom=dedupeEmpty` string lists:
// an array of strings with duplicates removed in place.
type StringDedupeList []string

// QuotaOverrideList ports OrgConfig.quotaOverrides `custom=dedupeEmpty,custom=
// sortByName`: dedupe (makeUnique with no key) then sort by the "name" field.
type QuotaOverrideList []QuotaOverride
