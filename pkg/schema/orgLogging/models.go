/* hand-maintained — canonical source of truth for the org logging providers.
 *
 * `validate:` struct tags drive pkg/validate.FromType. See
 * nodelibs/schema/src/orgLogging.ts and the tag DSL on
 * go-libs/schema/ipSet/models.go. The ElasticLogging .xor and the region-enum leaf
 * types (CloudWatchLoggingRegion, StackdriverLoggingLocation) are attached in
 * validate.go.
 */

package orgLogging

// CloudWatchLoggingRegion is orgLogging.ts:148 `Joi.string().valid(...AWS_REGIONS)`;
// EnumValues (below) carries the AWS_REGIONS list in TS declaration order.
type CloudWatchLoggingRegion string //ts:'us-east-1' | 'us-east-2' | 'us-west-1' | 'us-west-2' | 'ap-south-1' | 'ap-northeast-2' | 'ap-southeast-1' | 'ap-southeast-2' | 'ap-northeast-1' | 'eu-central-1' | 'eu-west-1' | 'eu-west-2' | 'eu-south-1' | 'eu-west-3' | 'eu-north-1' | 'me-south-1' | 'sa-east-1' | 'af-south-1'

type CloudWatchLoggingExtractFields map[string]string

type CloudWatchLogging struct {
	Region        CloudWatchLoggingRegion        `json:"region,omitzero" validate:"required"`
	Credentials   SecretLink                     `json:"credentials" validate:"required"`
	RetentionDays *float64                       `json:"retentionDays,omitzero" validate:"integer,valid=1|3|5|7|14|30|60|90|120|150|180|365|400|545|731|1827|3653"`
	GroupName     FluentBitTemplate              `json:"groupName" validate:"required"`
	StreamName    FluentBitTemplate              `json:"streamName" validate:"required"`
	ExtractFields CloudWatchLoggingExtractFields `json:"extractFields,omitzero"`
}

type CoralogixLoggingCluster string

type CoralogixLogging struct {
	Cluster     CoralogixLoggingCluster `json:"cluster,omitzero" validate:"valid=coralogix.com|coralogix.us|app.coralogix.in|app.eu2.coralogix.com|app.coralogixsg.com,required"`
	Credentials SecretLink              `json:"credentials" validate:"required"`
	App         LogPattern              `json:"app,omitzero"`
	Subsystem   LogPattern              `json:"subsystem,omitzero"`
}

type DatadogLoggingHost string

type DatadogLogging struct {
	Host        DatadogLoggingHost `json:"host,omitzero" validate:"required,valid=http-intake.logs.datadoghq.com|http-intake.logs.us3.datadoghq.com|http-intake.logs.us5.datadoghq.com|http-intake.logs.datadoghq.eu"`
	Credentials SecretLink         `json:"credentials" validate:"required"`
}

// ElasticLoggingAws ports orgLogging.ts AwsProvider.
type ElasticLoggingAws struct {
	Host        EsAmazonHost `json:"host" validate:"required"`
	Port        float64      `json:"port" validate:"positive,default=443,integer,required"`
	Index       string       `json:"index" validate:"required"`
	Type        string       `json:"type" validate:"required"`
	Credentials SecretLink   `json:"credentials" validate:"required"`
	Region      string       `json:"region" validate:"required"`
}

// ElasticLoggingElasticCloud ports orgLogging.ts ElastiCloudProvider.
type ElasticLoggingElasticCloud struct {
	Index       string     `json:"index" validate:"required"`
	Type        string     `json:"type" validate:"required"`
	Credentials SecretLink `json:"credentials" validate:"required"`
	CloudId     string     `json:"cloudId" validate:"required"`
}

// ElasticLoggingGeneric ports orgLogging.ts GenericESProvider.
type ElasticLoggingGeneric struct {
	Host        EsHost     `json:"host" validate:"required"`
	Port        *float64   `json:"port,omitzero" validate:"positive,default=443,integer"`
	Path        EsPath     `json:"path,omitzero"`
	Index       string     `json:"index" validate:"required"`
	Type        string     `json:"type" validate:"required"`
	Credentials SecretLink `json:"credentials" validate:"required"`
	Username    string     `json:"username,omitzero" validate:"optional"`
	Password    string     `json:"password,omitzero" validate:"optional"`
}

// ElasticLogging ports orgLogging.ts:224. .xor(elasticCloud, aws, generic) is
// attached by ValidationRule in validate.go.
type ElasticLogging struct {
	Aws          *ElasticLoggingAws          `json:"aws,omitzero"`
	ElasticCloud *ElasticLoggingElasticCloud `json:"elasticCloud,omitzero"`
	Generic      *ElasticLoggingGeneric      `json:"generic,omitzero"`
}

type FluentdLogging struct {
	Host string       `json:"host" validate:"required"`
	Port *FluentdPort `json:"port,omitzero" validate:"min=1,max=65535,integer,default=24224"`
}

type LogzioLoggingListenerHost string

type LogzioLogging struct {
	ListenerHost LogzioLoggingListenerHost `json:"listenerHost,omitzero" validate:"valid=listener.logz.io|listener-nl.logz.io,required"`
	Credentials  SecretLink                `json:"credentials" validate:"required"`
}

type LokiLogging struct {
	Endpoint    string     `json:"endpoint" validate:"required"`
	Credentials SecretLink `json:"credentials,omitzero" validate:"optional"`
	TenantId    string     `json:"tenantId,omitzero" validate:"optional"`
}

type OpenTelemetryLoggingHeaders map[string]string

type OpenTelemetryLogging struct {
	Endpoint    string                      `json:"endpoint" validate:"required"`
	Headers     OpenTelemetryLoggingHeaders `json:"headers,omitzero" validate:"optional"`
	Credentials SecretLink                  `json:"credentials,omitzero" validate:"optional"`
}

type S3Logging struct {
	Bucket      string     `json:"bucket" validate:"required"`
	Region      string     `json:"region" validate:"required"`
	Prefix      string     `json:"prefix,omitzero" validate:"default=/"`
	Credentials SecretLink `json:"credentials" validate:"required"`
}

// StackdriverLoggingLocation is orgLogging.ts:232 `Joi.string().valid(...GCP_REGIONS)`;
// EnumValues (below) carries the GCP_REGIONS list in TS declaration order.
type StackdriverLoggingLocation string //ts:'us-east1' | 'us-east4' | 'us-central1' | 'us-west1' | 'europe-west4' | 'europe-west1' | 'europe-west3' | 'europe-west2' | 'asia-east1' | 'asia-southeast1' | 'asia-northeast1' | 'asia-south1' | 'australia-southeast1' | 'southamerica-east1' | 'africa-south1' | 'asia-east2' | 'asia-northeast2' | 'asia-northeast3' | 'asia-south2' | 'asia-southeast2' | 'australia-southeast2' | 'europe-central2' | 'europe-north1' | 'europe-southwest1' | 'europe-west10' | 'europe-west12' | 'europe-west6' | 'europe-west8' | 'europe-west9' | 'me-central1' | 'me-central2' | 'me-west1' | 'northamerica-northeast1' | 'northamerica-northeast2' | 'southamerica-west1' | 'us-east5' | 'us-south1' | 'us-west2' | 'us-west3' | 'us-west4'

type StackdriverLogging struct {
	Credentials SecretLink                 `json:"credentials" validate:"required"`
	Location    StackdriverLoggingLocation `json:"location,omitzero" validate:"required"`
}

type SyslogLoggingMode string

const (
	SyslogLoggingModeTcp SyslogLoggingMode = "tcp"
	SyslogLoggingModeUdp SyslogLoggingMode = "udp"
	SyslogLoggingModeTls SyslogLoggingMode = "tls"
)

type SyslogLoggingFormat string

const (
	SyslogLoggingFormatRfc3164 SyslogLoggingFormat = "rfc3164"
	SyslogLoggingFormatRfc5424 SyslogLoggingFormat = "rfc5424"
)

type SyslogLogging struct {
	Host     string              `json:"host" validate:"required"`
	Port     float64             `json:"port" validate:"required,integer,positive"`
	Mode     SyslogLoggingMode   `json:"mode,omitzero" validate:"required,valid=tcp|udp|tls,default=tcp"`
	Format   SyslogLoggingFormat `json:"format,omitzero" validate:"valid=rfc3164|rfc5424,default=rfc5424,required"`
	Severity float64             `json:"severity" validate:"valid=0|1|2|3|4|5|6|7,default=6,required"`
}

// --- validation leaf types (ValidationRule methods live in validate.go) ---

// SecretLink is a localLink string pointing to a resource of kind secret. Fields
// keep their required/optional modifier tag.
type SecretLink string

// FluentBitTemplate is a fluent-bit template string: every $...$ pair must contain
// a '.' or ',' separator.
type FluentBitTemplate string

// LogPattern is a log pattern string: only {org}/{gvc}/{workload}/{location}
// template vars are allowed.
type LogPattern string

// EsHost is an RFC-1123-ish hostname (Joi.string().hostname()).
type EsHost string

// EsAmazonHost is a hostname that must additionally end with es.amazonaws.com
// (validHostname then esAmazonHost, in that order).
type EsAmazonHost string

// EsPath is an elasticsearch path that must start with '/'.
type EsPath string

// FluentdPort is the fluentd port. The field's min/max/integer/default modifier
// tags layer on top of the required-with-default custom.
type FluentdPort float64
