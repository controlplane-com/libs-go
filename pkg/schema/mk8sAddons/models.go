/* auto-generated */

package mk8sAddons

import "github.com/controlplane-com/libs-go/pkg/schema/workload"

type AwsECRAddonConfig struct {
	RoleArn string `json:"roleArn,omitzero"`
}

type AwsEFSAddonConfig struct {
	RoleArn string `json:"roleArn,omitzero"`
}

type AwsELBAddonConfig struct {
	RoleArn string `json:"roleArn,omitzero"`
}

type AwsTrustPolicyConfigTrustPolicy map[string]any

type AwsTrustPolicyConfig struct {
	TrustPolicy AwsTrustPolicyConfigTrustPolicy `json:"trustPolicy,omitzero"`
}

type AwsWorkloadIdentityAddonStatusOidcProviderConfig struct {
	ProviderUrl string `json:"providerUrl,omitzero"`
	Audience    string `json:"audience,omitzero"`
}

type AwsWorkloadIdentityAddonStatusTrustPolicy map[string]any

type AwsWorkloadIdentityAddonStatus struct {
	OidcProviderConfig *AwsWorkloadIdentityAddonStatusOidcProviderConfig `json:"oidcProviderConfig,omitzero"`
	TrustPolicy        AwsWorkloadIdentityAddonStatusTrustPolicy         `json:"trustPolicy,omitzero"`
}

type AzureACRAddonConfig struct {
	ClientId string `json:"clientId"`
}

type AzureAddonConfig struct {
	TenantId string `json:"tenantId,omitzero"`
}

type ByokAddonConfigConfigActuatorLogLevel string

const (
	ByokAddonConfigConfigActuatorLogLevelTrace ByokAddonConfigConfigActuatorLogLevel = "trace"
	ByokAddonConfigConfigActuatorLogLevelInfo  ByokAddonConfigConfigActuatorLogLevel = "info"
	ByokAddonConfigConfigActuatorLogLevelError ByokAddonConfigConfigActuatorLogLevel = "error"
)

type ByokAddonConfigConfigActuator struct {
	MinCpu    workload.Cpu                          `json:"minCpu,omitzero"`
	MaxCpu    workload.Cpu                          `json:"maxCpu,omitzero"`
	MinMemory workload.Memory                       `json:"minMemory,omitzero"`
	MaxMemory workload.Memory                       `json:"maxMemory,omitzero"`
	LogLevel  ByokAddonConfigConfigActuatorLogLevel `json:"logLevel,omitzero"`
	Env       FlexibleAddonConfig                   `json:"env,omitzero"`
}

type ByokAddonConfigConfigMaintainer struct {
	Env FlexibleAddonConfig `json:"env,omitzero"`
}

type ByokAddonConfigConfigJuicefs struct {
	Enabled *bool `json:"enabled,omitzero"`
}

type ByokAddonConfigConfigMiddlebox struct {
	Enabled            *bool    `json:"enabled,omitzero"`
	BandwidthAlertMbps *float64 `json:"bandwidthAlertMbps,omitzero"`
	Ip                 string   `json:"ip,omitzero"`
	IngressReplicas    *float64 `json:"ingressReplicas,omitzero"`
	Port               *float64 `json:"port,omitzero"`
}

type ByokAddonConfigConfigCommonPdb struct {
	MaxUnavailable *float64 `json:"maxUnavailable,omitzero"`
}

type ByokAddonConfigConfigCommon struct {
	DeploymentReplicas *float64                        `json:"deploymentReplicas,omitzero"`
	Pdb                *ByokAddonConfigConfigCommonPdb `json:"pdb,omitzero"`
}

type ByokAddonConfigConfigLonghorn struct {
	NumberOfReplicas *float64 `json:"numberOfReplicas,omitzero"`
	Replicas         *float64 `json:"replicas,omitzero"`
	IsDefault        *bool    `json:"isDefault,omitzero"`
}

type ByokAddonConfigConfigByok struct {
	NoDefaultStorageClasses *bool `json:"noDefaultStorageClasses,omitzero"`
}

type ByokAddonConfigConfigIngress struct {
	Cpu           workload.Cpu    `json:"cpu,omitzero"`
	Memory        workload.Memory `json:"memory,omitzero"`
	TargetPercent *float64        `json:"targetPercent,omitzero"`
}

type ByokAddonConfigConfigIstioIstiod struct {
	Replicas  *float64        `json:"replicas,omitzero"`
	MinCpu    workload.Cpu    `json:"minCpu,omitzero"`
	MaxCpu    workload.Cpu    `json:"maxCpu,omitzero"`
	MinMemory workload.Memory `json:"minMemory,omitzero"`
	MaxMemory workload.Memory `json:"maxMemory,omitzero"`
	Pdb       *float64        `json:"pdb,omitzero"`
}

type ByokAddonConfigConfigIstioIngressgateway struct {
	Replicas  *float64        `json:"replicas,omitzero"`
	MaxCpu    workload.Cpu    `json:"maxCpu,omitzero"`
	MaxMemory workload.Memory `json:"maxMemory,omitzero"`
}

type ByokAddonConfigConfigIstioSidecar struct {
	MinCpu    workload.Cpu    `json:"minCpu,omitzero"`
	MinMemory workload.Memory `json:"minMemory,omitzero"`
}

type ByokAddonConfigConfigIstio struct {
	Istiod         *ByokAddonConfigConfigIstioIstiod         `json:"istiod,omitzero"`
	Ingressgateway *ByokAddonConfigConfigIstioIngressgateway `json:"ingressgateway,omitzero"`
	Sidecar        *ByokAddonConfigConfigIstioSidecar        `json:"sidecar,omitzero"`
}

type ByokAddonConfigConfigLogSplitter struct {
	MinCpu        workload.Cpu    `json:"minCpu,omitzero"`
	MaxCpu        workload.Cpu    `json:"maxCpu,omitzero"`
	MinMemory     workload.Memory `json:"minMemory,omitzero"`
	MaxMemory     workload.Memory `json:"maxMemory,omitzero"`
	MemBufferSize string          `json:"memBufferSize,omitzero"`
	PerPodRate    *float64        `json:"perPodRate,omitzero"`
}

type ByokAddonConfigConfigMonitoringKubeStateMetrics struct {
	MinMemory workload.Memory `json:"minMemory,omitzero"`
}

type ByokAddonConfigConfigMonitoringPrometheusMain struct {
	Storage workload.Memory `json:"storage,omitzero"`
}

type ByokAddonConfigConfigMonitoringPrometheus struct {
	Main *ByokAddonConfigConfigMonitoringPrometheusMain `json:"main,omitzero"`
}

type ByokAddonConfigConfigMonitoringRemoteWriteBasicAuth struct {
	Username      string `json:"username,omitzero"`
	Username_file string `json:"username_file,omitzero"`
	Password      string `json:"password,omitzero"`
	Password_file string `json:"password_file,omitzero"`
}

type ByokAddonConfigConfigMonitoringRemoteWriteAuthorization struct {
	Type             string `json:"type,omitzero"`
	Credentials      string `json:"credentials,omitzero"`
	Credentials_file string `json:"credentials_file,omitzero"`
}

type ByokAddonConfigConfigMonitoringRemoteWriteOauth2 map[string]any

type ByokAddonConfigConfigMonitoringRemoteWriteTlsConfig map[string]any

type ByokAddonConfigConfigMonitoringRemoteWriteProxyConnectHeader map[string][]string

type ByokAddonConfigConfigMonitoringRemoteWriteHttpHeaders map[string]HttpHeaderValue

type ByokAddonConfigConfigMonitoringRemoteWriteWriteRelabelConfigs map[string]any

type ByokAddonConfigConfigMonitoringRemoteWriteSigv4 map[string]any

type ByokAddonConfigConfigMonitoringRemoteWriteAzuread map[string]any

type ByokAddonConfigConfigMonitoringRemoteWriteGoogleIam map[string]any

type ByokAddonConfigConfigMonitoringRemoteWriteQueueConfig map[string]any

type ByokAddonConfigConfigMonitoringRemoteWrite struct {
	Basic_auth             *ByokAddonConfigConfigMonitoringRemoteWriteBasicAuth            `json:"basic_auth,omitzero"`
	Authorization          *ByokAddonConfigConfigMonitoringRemoteWriteAuthorization        `json:"authorization,omitzero"`
	Oauth2                 ByokAddonConfigConfigMonitoringRemoteWriteOauth2                `json:"oauth2,omitzero"`
	Follow_redirects       *bool                                                           `json:"follow_redirects,omitzero"`
	Enable_http2           *bool                                                           `json:"enable_http2,omitzero"`
	Tls_config             ByokAddonConfigConfigMonitoringRemoteWriteTlsConfig             `json:"tls_config,omitzero"`
	Proxy_url              string                                                          `json:"proxy_url,omitzero"`
	No_proxy               string                                                          `json:"no_proxy,omitzero"`
	Proxy_from_environment *bool                                                           `json:"proxy_from_environment,omitzero"`
	Proxy_connect_header   ByokAddonConfigConfigMonitoringRemoteWriteProxyConnectHeader    `json:"proxy_connect_header,omitzero"`
	Http_headers           ByokAddonConfigConfigMonitoringRemoteWriteHttpHeaders           `json:"http_headers,omitzero"`
	Url                    string                                                          `json:"url,omitzero"`
	Remote_timeout         string                                                          `json:"remote_timeout,omitzero"`
	Headers                FlexibleAddonConfig                                             `json:"headers,omitzero"`
	Write_relabel_configs  []ByokAddonConfigConfigMonitoringRemoteWriteWriteRelabelConfigs `json:"write_relabel_configs,omitzero"`
	Name                   string                                                          `json:"name,omitzero"`
	Send_exemplars         *bool                                                           `json:"send_exemplars,omitzero"`
	Send_native_histograms *bool                                                           `json:"send_native_histograms,omitzero"`
	Sigv4                  ByokAddonConfigConfigMonitoringRemoteWriteSigv4                 `json:"sigv4,omitzero"`
	Azuread                ByokAddonConfigConfigMonitoringRemoteWriteAzuread               `json:"azuread,omitzero"`
	Google_iam             ByokAddonConfigConfigMonitoringRemoteWriteGoogleIam             `json:"google_iam,omitzero"`
	Queue_config           ByokAddonConfigConfigMonitoringRemoteWriteQueueConfig           `json:"queue_config,omitzero"`
}

type ByokAddonConfigConfigMonitoring struct {
	MinMemory        workload.Memory                                  `json:"minMemory,omitzero"`
	MaxMemory        workload.Memory                                  `json:"maxMemory,omitzero"`
	KubeStateMetrics *ByokAddonConfigConfigMonitoringKubeStateMetrics `json:"kubeStateMetrics,omitzero"`
	Prometheus       *ByokAddonConfigConfigMonitoringPrometheus       `json:"prometheus,omitzero"`
	RemoteWrite      []ByokAddonConfigConfigMonitoringRemoteWrite     `json:"remoteWrite,omitzero"`
	ExternalLabels   FlexibleAddonConfig                              `json:"externalLabels,omitzero"`
}

type ByokAddonConfigConfigRedis struct {
	MinCpu    workload.Cpu    `json:"minCpu,omitzero"`
	MaxCpu    workload.Cpu    `json:"maxCpu,omitzero"`
	MinMemory workload.Memory `json:"minMemory,omitzero"`
	MaxMemory workload.Memory `json:"maxMemory,omitzero"`
	Storage   workload.Memory `json:"storage,omitzero"`
}

type ByokAddonConfigConfigRedisHa struct {
	MinCpu    workload.Cpu    `json:"minCpu,omitzero"`
	MaxCpu    workload.Cpu    `json:"maxCpu,omitzero"`
	MinMemory workload.Memory `json:"minMemory,omitzero"`
	MaxMemory workload.Memory `json:"maxMemory,omitzero"`
	Storage   *float64        `json:"storage,omitzero"`
}

type ByokAddonConfigConfigRedisSentinel struct {
	MinCpu    workload.Cpu    `json:"minCpu,omitzero"`
	MaxCpu    workload.Cpu    `json:"maxCpu,omitzero"`
	MinMemory workload.Memory `json:"minMemory,omitzero"`
	MaxMemory workload.Memory `json:"maxMemory,omitzero"`
	Storage   *float64        `json:"storage,omitzero"`
}

type ByokAddonConfigConfigTempoAgent struct {
	MinCpu    workload.Cpu    `json:"minCpu,omitzero"`
	MinMemory workload.Memory `json:"minMemory,omitzero"`
}

type ByokAddonConfigConfigInternalDns struct {
	MinCpu    workload.Cpu    `json:"minCpu,omitzero"`
	MaxCpu    workload.Cpu    `json:"maxCpu,omitzero"`
	MinMemory workload.Memory `json:"minMemory,omitzero"`
	MaxMemory workload.Memory `json:"maxMemory,omitzero"`
}

type ByokAddonConfigConfig struct {
	Actuator      *ByokAddonConfigConfigActuator      `json:"actuator,omitzero"`
	Maintainer    *ByokAddonConfigConfigMaintainer    `json:"maintainer,omitzero"`
	Juicefs       *ByokAddonConfigConfigJuicefs       `json:"juicefs,omitzero"`
	Middlebox     *ByokAddonConfigConfigMiddlebox     `json:"middlebox,omitzero"`
	Common        *ByokAddonConfigConfigCommon        `json:"common,omitzero"`
	Longhorn      *ByokAddonConfigConfigLonghorn      `json:"longhorn,omitzero"`
	Byok          *ByokAddonConfigConfigByok          `json:"byok,omitzero"`
	Ingress       *ByokAddonConfigConfigIngress       `json:"ingress,omitzero"`
	Istio         *ByokAddonConfigConfigIstio         `json:"istio,omitzero"`
	LogSplitter   *ByokAddonConfigConfigLogSplitter   `json:"logSplitter,omitzero"`
	Monitoring    *ByokAddonConfigConfigMonitoring    `json:"monitoring,omitzero"`
	Redis         *ByokAddonConfigConfigRedis         `json:"redis,omitzero"`
	RedisHa       *ByokAddonConfigConfigRedisHa       `json:"redisHa,omitzero"`
	RedisSentinel *ByokAddonConfigConfigRedisSentinel `json:"redisSentinel,omitzero"`
	TempoAgent    *ByokAddonConfigConfigTempoAgent    `json:"tempoAgent,omitzero"`
	InternalDns   *ByokAddonConfigConfigInternalDns   `json:"internalDns,omitzero"`
}

type ByokAddonConfig struct {
	IgnoreUpdates *bool                  `json:"ignoreUpdates,omitzero"`
	Location      string                 `json:"location"`
	Config        *ByokAddonConfigConfig `json:"config,omitzero"`
}

type DashboardAddonStatus struct {
	Url string `json:"url,omitzero"`
}

type FlexibleAddonConfig map[string]string

type HttpHeaderValue struct {
	Values  []string `json:"values,omitzero"`
	Secrets []string `json:"secrets,omitzero"`
	Files   []string `json:"files,omitzero"`
}

type KubevirtAddonConfig struct {
	ScratchSpaceStorageClass string `json:"scratchSpaceStorageClass,omitzero"`
}

type LogsAddonConfig struct {
	AuditEnabled      *bool             `json:"auditEnabled,omitzero"`
	IncludeNamespaces RegularExpression `json:"includeNamespaces,omitzero"`
	ExcludeNamespaces RegularExpression `json:"excludeNamespaces,omitzero"`
	Docker            *bool             `json:"docker,omitzero"`
	Kubelet           *bool             `json:"kubelet,omitzero"`
	Kernel            *bool             `json:"kernel,omitzero"`
	Events            *bool             `json:"events,omitzero"`
}

type LogsAddonStatus struct {
	LokiAddress string `json:"lokiAddress,omitzero"`
}

type MetricsAddonConfigScrapeAnnotated struct {
	IntervalSeconds   *float64          `json:"intervalSeconds,omitzero"`
	IncludeNamespaces RegularExpression `json:"includeNamespaces,omitzero"`
	ExcludeNamespaces RegularExpression `json:"excludeNamespaces,omitzero"`
	RetainLabels      RegularExpression `json:"retainLabels,omitzero"`
}

type MetricsAddonConfig struct {
	KubeState    *bool `json:"kubeState,omitzero"`
	CoreDns      *bool `json:"coreDns,omitzero"`
	Kubelet      *bool `json:"kubelet,omitzero"`
	Apiserver    *bool `json:"apiserver,omitzero"`
	NodeExporter *bool `json:"nodeExporter,omitzero"`
	Cadvisor     *bool `json:"cadvisor,omitzero"`
	// Per-collector drop regexes (master 82e90536d): a scraped series whose metric name
	// matches is dropped. Same RE2-validated Regex as the namespace/label fields.
	KubeStateDrop    RegularExpression                  `json:"kubeStateDrop,omitzero"`
	CoreDnsDrop      RegularExpression                  `json:"coreDnsDrop,omitzero"`
	KubeletDrop      RegularExpression                  `json:"kubeletDrop,omitzero"`
	ApiserverDrop    RegularExpression                  `json:"apiserverDrop,omitzero"`
	NodeExporterDrop RegularExpression                  `json:"nodeExporterDrop,omitzero"`
	CadvisorDrop     RegularExpression                  `json:"cadvisorDrop,omitzero"`
	ScrapeAnnotated  *MetricsAddonConfigScrapeAnnotated `json:"scrapeAnnotated,omitzero"`
}

type MetricsAddonStatusRemoteWriteConfig map[string]any

type MetricsAddonStatus struct {
	PrometheusEndpoint string                              `json:"prometheusEndpoint,omitzero"`
	RemoteWriteConfig  MetricsAddonStatusRemoteWriteConfig `json:"remoteWriteConfig,omitzero"`
}

type NodeLocalDnsAddonConfig struct {
}

type NonCustomizableAddonConfig struct {
}

type NvidiaAddonConfig struct {
	TaintGPUNodes *bool `json:"taintGPUNodes,omitzero"`
}

// Registry is a container-registry host: a hostname or "*" (was Joi.alternatives of two strings).
type Registry string

type RegistryMirrorConfigMirrors struct {
	Registry Registry `json:"registry"`
	Mirrors  []string `json:"mirrors,omitzero"`
}

type RegistryMirrorConfig struct {
	Mirrors []RegistryMirrorConfigMirrors `json:"mirrors,omitzero"`
}

type RegularExpression string
