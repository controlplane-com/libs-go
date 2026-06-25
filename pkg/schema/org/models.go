/* auto-generated */

package org

import "github.com/controlplane-com/libs-go/pkg/schema/base"
import "github.com/controlplane-com/libs-go/pkg/schema/orgLogging"
import "github.com/controlplane-com/libs-go/pkg/schema/tracing"

type AuthConfig struct {
	DomainAutoMembers []string `json:"domainAutoMembers,omitempty"`
	SamlOnly          bool     `json:"samlOnly,omitempty"`
}

type ObservabilityConfig struct {
	LogsRetentionDays    *float32 `json:"logsRetentionDays,omitempty"`
	MetricsRetentionDays *float32 `json:"metricsRetentionDays,omitempty"`
	TracesRetentionDays  *float32 `json:"tracesRetentionDays,omitempty"`
	DefaultAlertEmails   []string `json:"defaultAlertEmails,omitempty"`
}

type OrgTags map[string]any

type Org struct {
	Id           string     `json:"id,omitempty"`
	Kind         base.Kind  `json:"kind,omitempty"`
	Version      *float32   `json:"version,omitempty"`
	Description  string     `json:"description,omitempty"`
	Tags         OrgTags    `json:"tags,omitempty"`
	Created      string     `json:"created,omitempty"`
	LastModified string     `json:"lastModified,omitempty"`
	Links        base.Links `json:"links,omitempty"`
	Name         string     `json:"name,omitempty"`
	Spec         OrgSpec    `json:"spec,omitempty"`
	Status       OrgStatus  `json:"status,omitempty"`
}

type OrgConfig struct {
	AwsPrivateLinks    []string        `json:"awsPrivateLinks,omitempty"`
	GcpServiceConnects []string        `json:"gcpServiceConnects,omitempty"`
	QuotaOverrides     []QuotaOverride `json:"quotaOverrides,omitempty"`
}

type OrgSpecLogging struct {
	S3            orgLogging.S3Logging            `json:"s3,omitempty"`
	Coralogix     orgLogging.CoralogixLogging     `json:"coralogix,omitempty"`
	Datadog       orgLogging.DatadogLogging       `json:"datadog,omitempty"`
	Logzio        orgLogging.LogzioLogging        `json:"logzio,omitempty"`
	Elastic       orgLogging.ElasticLogging       `json:"elastic,omitempty"`
	CloudWatch    orgLogging.CloudWatchLogging    `json:"cloudWatch,omitempty"`
	Fluentd       orgLogging.FluentdLogging       `json:"fluentd,omitempty"`
	Stackdriver   orgLogging.StackdriverLogging   `json:"stackdriver,omitempty"`
	Syslog        orgLogging.SyslogLogging        `json:"syslog,omitempty"`
	Opentelemetry orgLogging.OpenTelemetryLogging `json:"opentelemetry,omitempty"`
	Loki          orgLogging.LokiLogging          `json:"loki,omitempty"`
}

type OrgSpecExtraLogging struct {
	S3            orgLogging.S3Logging            `json:"s3,omitempty"`
	Coralogix     orgLogging.CoralogixLogging     `json:"coralogix,omitempty"`
	Datadog       orgLogging.DatadogLogging       `json:"datadog,omitempty"`
	Logzio        orgLogging.LogzioLogging        `json:"logzio,omitempty"`
	Elastic       orgLogging.ElasticLogging       `json:"elastic,omitempty"`
	CloudWatch    orgLogging.CloudWatchLogging    `json:"cloudWatch,omitempty"`
	Fluentd       orgLogging.FluentdLogging       `json:"fluentd,omitempty"`
	Stackdriver   orgLogging.StackdriverLogging   `json:"stackdriver,omitempty"`
	Syslog        orgLogging.SyslogLogging        `json:"syslog,omitempty"`
	Opentelemetry orgLogging.OpenTelemetryLogging `json:"opentelemetry,omitempty"`
	Loki          orgLogging.LokiLogging          `json:"loki,omitempty"`
}

type OrgSpecObservability struct {
	LogsRetentionDays    *float32 `json:"logsRetentionDays,omitempty"`
	MetricsRetentionDays *float32 `json:"metricsRetentionDays,omitempty"`
	TracesRetentionDays  *float32 `json:"tracesRetentionDays,omitempty"`
	DefaultAlertEmails   []string `json:"defaultAlertEmails,omitempty"`
}

type OrgSpecSecurity struct {
	ThreatDetection ThreatDetection `json:"threatDetection,omitempty"`
}

type OrgSpec struct {
	Logging               OrgSpecLogging        `json:"logging,omitempty"`
	ExtraLogging          []OrgSpecExtraLogging `json:"extraLogging,omitempty"`
	Tracing               tracing.Tracing       `json:"tracing,omitempty"`
	SessionTimeoutSeconds *float32              `json:"sessionTimeoutSeconds,omitempty"`
	AuthConfig            AuthConfig            `json:"authConfig,omitempty"`
	Observability         OrgSpecObservability  `json:"observability,omitempty"`
	Security              OrgSpecSecurity       `json:"security,omitempty"`
}

type OrgStatus struct {
	AccountLink    string `json:"accountLink,omitempty"`
	Active         bool   `json:"active,omitempty"`
	EndpointPrefix string `json:"endpointPrefix,omitempty"`
}

type QuotaOverride struct {
	Name string  `json:"name"`
	Max  float32 `json:"max"`
}

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
	Transport ThreatDetectionSyslogTransport `json:"transport,omitempty"`
	Host      string                         `json:"host"`
	Port      float32                        `json:"port"`
}

type ThreatDetection struct {
	Enabled         bool                           `json:"enabled"`
	MinimumSeverity ThreatDetectionMinimumSeverity `json:"minimumSeverity,omitempty"`
	Syslog          ThreatDetectionSyslog          `json:"syslog,omitempty"`
}
