/* auto-generated */

package orgLogging

type CloudWatchLoggingRegion string

const (
	CloudWatchLoggingRegionUsEast1      CloudWatchLoggingRegion = "us-east-1"
	CloudWatchLoggingRegionUsEast2      CloudWatchLoggingRegion = "us-east-2"
	CloudWatchLoggingRegionUsWest1      CloudWatchLoggingRegion = "us-west-1"
	CloudWatchLoggingRegionUsWest2      CloudWatchLoggingRegion = "us-west-2"
	CloudWatchLoggingRegionApSouth1     CloudWatchLoggingRegion = "ap-south-1"
	CloudWatchLoggingRegionApNortheast2 CloudWatchLoggingRegion = "ap-northeast-2"
	CloudWatchLoggingRegionApSoutheast1 CloudWatchLoggingRegion = "ap-southeast-1"
	CloudWatchLoggingRegionApSoutheast2 CloudWatchLoggingRegion = "ap-southeast-2"
	CloudWatchLoggingRegionApNortheast1 CloudWatchLoggingRegion = "ap-northeast-1"
	CloudWatchLoggingRegionEuCentral1   CloudWatchLoggingRegion = "eu-central-1"
	CloudWatchLoggingRegionEuWest1      CloudWatchLoggingRegion = "eu-west-1"
	CloudWatchLoggingRegionEuWest2      CloudWatchLoggingRegion = "eu-west-2"
	CloudWatchLoggingRegionEuSouth1     CloudWatchLoggingRegion = "eu-south-1"
	CloudWatchLoggingRegionEuWest3      CloudWatchLoggingRegion = "eu-west-3"
	CloudWatchLoggingRegionEuNorth1     CloudWatchLoggingRegion = "eu-north-1"
	CloudWatchLoggingRegionMeSouth1     CloudWatchLoggingRegion = "me-south-1"
	CloudWatchLoggingRegionSaEast1      CloudWatchLoggingRegion = "sa-east-1"
	CloudWatchLoggingRegionAfSouth1     CloudWatchLoggingRegion = "af-south-1"
)

type CloudWatchLoggingExtractFields map[string]string

type CloudWatchLogging struct {
	Region        CloudWatchLoggingRegion        `json:"region,omitempty"`
	Credentials   string                         `json:"credentials"`
	RetentionDays float32                        `json:"retentionDays"`
	GroupName     string                         `json:"groupName"`
	StreamName    string                         `json:"streamName"`
	ExtractFields CloudWatchLoggingExtractFields `json:"extractFields,omitempty"`
}

type CoralogixLoggingCluster string

const (
	CoralogixLoggingClusterCoralogixCom       CoralogixLoggingCluster = "coralogix.com"
	CoralogixLoggingClusterCoralogixUs        CoralogixLoggingCluster = "coralogix.us"
	CoralogixLoggingClusterAppCoralogixIn     CoralogixLoggingCluster = "app.coralogix.in"
	CoralogixLoggingClusterAppEu2CoralogixCom CoralogixLoggingCluster = "app.eu2.coralogix.com"
	CoralogixLoggingClusterAppCoralogixsgCom  CoralogixLoggingCluster = "app.coralogixsg.com"
)

type CoralogixLogging struct {
	Cluster     CoralogixLoggingCluster `json:"cluster,omitempty"`
	Credentials string                  `json:"credentials"`
	App         string                  `json:"app,omitempty"`
	Subsystem   string                  `json:"subsystem,omitempty"`
}

type DatadogLoggingHost string

const (
	DatadogLoggingHostHttpIntakeLogsDatadoghqCom    DatadogLoggingHost = "http-intake.logs.datadoghq.com"
	DatadogLoggingHostHttpIntakeLogsUs3DatadoghqCom DatadogLoggingHost = "http-intake.logs.us3.datadoghq.com"
	DatadogLoggingHostHttpIntakeLogsUs5DatadoghqCom DatadogLoggingHost = "http-intake.logs.us5.datadoghq.com"
	DatadogLoggingHostHttpIntakeLogsDatadoghqEu     DatadogLoggingHost = "http-intake.logs.datadoghq.eu"
)

type DatadogLogging struct {
	Host        DatadogLoggingHost `json:"host,omitempty"`
	Credentials string             `json:"credentials"`
}

type ElasticLoggingAws struct {
	Host        string  `json:"host"`
	Port        float32 `json:"port"`
	Index       string  `json:"index"`
	Type        string  `json:"type"`
	Credentials string  `json:"credentials"`
	Region      string  `json:"region"`
}

type ElasticLoggingElasticCloud struct {
	Index       string `json:"index"`
	Type        string `json:"type"`
	Credentials string `json:"credentials"`
	CloudId     string `json:"cloudId"`
}

type ElasticLoggingGeneric struct {
	Host        string  `json:"host"`
	Port        float32 `json:"port"`
	Path        string  `json:"path,omitempty"`
	Index       string  `json:"index"`
	Type        string  `json:"type"`
	Credentials string  `json:"credentials"`
	Username    string  `json:"username,omitempty"`
	Password    string  `json:"password,omitempty"`
}

type ElasticLogging struct {
	Aws          ElasticLoggingAws          `json:"aws,omitempty"`
	ElasticCloud ElasticLoggingElasticCloud `json:"elasticCloud,omitempty"`
	Generic      ElasticLoggingGeneric      `json:"generic,omitempty"`
}

type FluentdLogging struct {
	Host string  `json:"host"`
	Port float32 `json:"port"`
}

type LogzioLoggingListenerHost string

const (
	LogzioLoggingListenerHostListenerLogzIo   LogzioLoggingListenerHost = "listener.logz.io"
	LogzioLoggingListenerHostListenerNlLogzIo LogzioLoggingListenerHost = "listener-nl.logz.io"
)

type LogzioLogging struct {
	ListenerHost LogzioLoggingListenerHost `json:"listenerHost,omitempty"`
	Credentials  string                    `json:"credentials"`
}

type OpenTelemetryLoggingHeaders map[string]string

type OpenTelemetryLogging struct {
	Endpoint    string                      `json:"endpoint"`
	Headers     OpenTelemetryLoggingHeaders `json:"headers,omitempty"`
	Credentials string                      `json:"credentials,omitempty"`
}

type S3Logging struct {
	Bucket      string `json:"bucket"`
	Region      string `json:"region"`
	Prefix      string `json:"prefix,omitempty"`
	Credentials string `json:"credentials"`
}

type StackdriverLoggingLocation string

const (
	StackdriverLoggingLocationUsEast1                StackdriverLoggingLocation = "us-east1"
	StackdriverLoggingLocationUsEast4                StackdriverLoggingLocation = "us-east4"
	StackdriverLoggingLocationUsCentral1             StackdriverLoggingLocation = "us-central1"
	StackdriverLoggingLocationUsWest1                StackdriverLoggingLocation = "us-west1"
	StackdriverLoggingLocationEuropeWest4            StackdriverLoggingLocation = "europe-west4"
	StackdriverLoggingLocationEuropeWest1            StackdriverLoggingLocation = "europe-west1"
	StackdriverLoggingLocationEuropeWest3            StackdriverLoggingLocation = "europe-west3"
	StackdriverLoggingLocationEuropeWest2            StackdriverLoggingLocation = "europe-west2"
	StackdriverLoggingLocationAsiaEast1              StackdriverLoggingLocation = "asia-east1"
	StackdriverLoggingLocationAsiaSoutheast1         StackdriverLoggingLocation = "asia-southeast1"
	StackdriverLoggingLocationAsiaNortheast1         StackdriverLoggingLocation = "asia-northeast1"
	StackdriverLoggingLocationAsiaSouth1             StackdriverLoggingLocation = "asia-south1"
	StackdriverLoggingLocationAustraliaSoutheast1    StackdriverLoggingLocation = "australia-southeast1"
	StackdriverLoggingLocationSouthamericaEast1      StackdriverLoggingLocation = "southamerica-east1"
	StackdriverLoggingLocationAfricaSouth1           StackdriverLoggingLocation = "africa-south1"
	StackdriverLoggingLocationAsiaEast2              StackdriverLoggingLocation = "asia-east2"
	StackdriverLoggingLocationAsiaNortheast2         StackdriverLoggingLocation = "asia-northeast2"
	StackdriverLoggingLocationAsiaNortheast3         StackdriverLoggingLocation = "asia-northeast3"
	StackdriverLoggingLocationAsiaSouth2             StackdriverLoggingLocation = "asia-south2"
	StackdriverLoggingLocationAsiaSoutheast2         StackdriverLoggingLocation = "asia-southeast2"
	StackdriverLoggingLocationAustraliaSoutheast2    StackdriverLoggingLocation = "australia-southeast2"
	StackdriverLoggingLocationEuropeCentral2         StackdriverLoggingLocation = "europe-central2"
	StackdriverLoggingLocationEuropeNorth1           StackdriverLoggingLocation = "europe-north1"
	StackdriverLoggingLocationEuropeSouthwest1       StackdriverLoggingLocation = "europe-southwest1"
	StackdriverLoggingLocationEuropeWest10           StackdriverLoggingLocation = "europe-west10"
	StackdriverLoggingLocationEuropeWest12           StackdriverLoggingLocation = "europe-west12"
	StackdriverLoggingLocationEuropeWest6            StackdriverLoggingLocation = "europe-west6"
	StackdriverLoggingLocationEuropeWest8            StackdriverLoggingLocation = "europe-west8"
	StackdriverLoggingLocationEuropeWest9            StackdriverLoggingLocation = "europe-west9"
	StackdriverLoggingLocationMeCentral1             StackdriverLoggingLocation = "me-central1"
	StackdriverLoggingLocationMeCentral2             StackdriverLoggingLocation = "me-central2"
	StackdriverLoggingLocationMeWest1                StackdriverLoggingLocation = "me-west1"
	StackdriverLoggingLocationNorthamericaNortheast1 StackdriverLoggingLocation = "northamerica-northeast1"
	StackdriverLoggingLocationNorthamericaNortheast2 StackdriverLoggingLocation = "northamerica-northeast2"
	StackdriverLoggingLocationSouthamericaWest1      StackdriverLoggingLocation = "southamerica-west1"
	StackdriverLoggingLocationUsEast5                StackdriverLoggingLocation = "us-east5"
	StackdriverLoggingLocationUsSouth1               StackdriverLoggingLocation = "us-south1"
	StackdriverLoggingLocationUsWest2                StackdriverLoggingLocation = "us-west2"
	StackdriverLoggingLocationUsWest3                StackdriverLoggingLocation = "us-west3"
	StackdriverLoggingLocationUsWest4                StackdriverLoggingLocation = "us-west4"
)

type StackdriverLogging struct {
	Credentials string                     `json:"credentials"`
	Location    StackdriverLoggingLocation `json:"location,omitempty"`
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
	Host     string              `json:"host"`
	Port     float32             `json:"port"`
	Mode     SyslogLoggingMode   `json:"mode,omitempty"`
	Format   SyslogLoggingFormat `json:"format,omitempty"`
	Severity float32             `json:"severity"`
}
