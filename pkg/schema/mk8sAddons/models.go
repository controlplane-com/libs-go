/* auto-generated */

package mk8sAddons

import "github.com/controlplane-com/libs-go/pkg/schema/workload"

type AwsECRAddonConfig struct {
	RoleArn string `json:"roleArn,omitempty"`
}

type AwsEFSAddonConfig struct {
	RoleArn string `json:"roleArn"`
}

type AwsELBAddonConfig struct {
	RoleArn string `json:"roleArn,omitempty"`
}

type AwsTrustPolicyConfigTrustPolicy map[string]any

type AwsTrustPolicyConfig struct {
	TrustPolicy AwsTrustPolicyConfigTrustPolicy `json:"trustPolicy,omitempty"`
}

type AwsWorkloadIdentityAddonStatusOidcProviderConfig struct {
	ProviderUrl string `json:"providerUrl,omitempty"`
	Audience    string `json:"audience,omitempty"`
}

type AwsWorkloadIdentityAddonStatusTrustPolicy map[string]any

type AwsWorkloadIdentityAddonStatus struct {
	OidcProviderConfig AwsWorkloadIdentityAddonStatusOidcProviderConfig `json:"oidcProviderConfig,omitempty"`
	TrustPolicy        AwsWorkloadIdentityAddonStatusTrustPolicy        `json:"trustPolicy,omitempty"`
}

type AzureACRAddonConfig struct {
	ClientId string `json:"clientId"`
}

type AzureAddonConfig struct {
	TenantId string `json:"tenantId,omitempty"`
}

type ByokAddonConfigConfigActuatorLogLevel string

const (
	ByokAddonConfigConfigActuatorLogLevelTrace ByokAddonConfigConfigActuatorLogLevel = "trace"
	ByokAddonConfigConfigActuatorLogLevelInfo  ByokAddonConfigConfigActuatorLogLevel = "info"
	ByokAddonConfigConfigActuatorLogLevelError ByokAddonConfigConfigActuatorLogLevel = "error"
)

type ByokAddonConfigConfigActuator struct {
	MinCpu    workload.Cpu                          `json:"minCpu,omitempty"`
	MaxCpu    workload.Cpu                          `json:"maxCpu,omitempty"`
	MinMemory workload.Memory                       `json:"minMemory,omitempty"`
	MaxMemory workload.Memory                       `json:"maxMemory,omitempty"`
	LogLevel  ByokAddonConfigConfigActuatorLogLevel `json:"logLevel,omitempty"`
	Env       FlexibleAddonConfig                   `json:"env,omitempty"`
}

type ByokAddonConfigConfigJuicefs struct {
	Enabled bool `json:"enabled,omitempty"`
}

type ByokAddonConfigConfigMiddlebox struct {
	Enabled            bool     `json:"enabled,omitempty"`
	BandwidthAlertMbps *float32 `json:"bandwidthAlertMbps,omitempty"`
	Ip                 string   `json:"ip,omitempty"`
	IngressReplicas    *float32 `json:"ingressReplicas,omitempty"`
	Port               *float32 `json:"port,omitempty"`
}

type ByokAddonConfigConfigCommonPdb struct {
	MaxUnavailable *float32 `json:"maxUnavailable,omitempty"`
}

type ByokAddonConfigConfigCommon struct {
	DeploymentReplicas *float32                       `json:"deploymentReplicas,omitempty"`
	Pdb                ByokAddonConfigConfigCommonPdb `json:"pdb,omitempty"`
}

type ByokAddonConfigConfigLonghorn struct {
	NumberOfReplicas *float32 `json:"numberOfReplicas,omitempty"`
	Replicas         *float32 `json:"replicas,omitempty"`
	IsDefault        bool     `json:"isDefault,omitempty"`
}

type ByokAddonConfigConfigByok struct {
	NoDefaultStorageClasses bool `json:"noDefaultStorageClasses,omitempty"`
}

type ByokAddonConfigConfigIngress struct {
	Cpu           workload.Cpu    `json:"cpu,omitempty"`
	Memory        workload.Memory `json:"memory,omitempty"`
	TargetPercent *float32        `json:"targetPercent,omitempty"`
}

type ByokAddonConfigConfigIstioIstiod struct {
	Replicas  *float32        `json:"replicas,omitempty"`
	MinCpu    workload.Cpu    `json:"minCpu,omitempty"`
	MaxCpu    workload.Cpu    `json:"maxCpu,omitempty"`
	MinMemory workload.Memory `json:"minMemory,omitempty"`
	MaxMemory workload.Memory `json:"maxMemory,omitempty"`
	Pdb       *float32        `json:"pdb,omitempty"`
}

type ByokAddonConfigConfigIstioIngressgateway struct {
	Replicas  *float32        `json:"replicas,omitempty"`
	MaxCpu    workload.Cpu    `json:"maxCpu,omitempty"`
	MaxMemory workload.Memory `json:"maxMemory,omitempty"`
}

type ByokAddonConfigConfigIstioSidecar struct {
	MinCpu    workload.Cpu    `json:"minCpu,omitempty"`
	MinMemory workload.Memory `json:"minMemory,omitempty"`
}

type ByokAddonConfigConfigIstio struct {
	Istiod         ByokAddonConfigConfigIstioIstiod         `json:"istiod,omitempty"`
	Ingressgateway ByokAddonConfigConfigIstioIngressgateway `json:"ingressgateway,omitempty"`
	Sidecar        ByokAddonConfigConfigIstioSidecar        `json:"sidecar,omitempty"`
}

type ByokAddonConfigConfigLogSplitter struct {
	MinCpu        workload.Cpu    `json:"minCpu,omitempty"`
	MaxCpu        workload.Cpu    `json:"maxCpu,omitempty"`
	MinMemory     workload.Memory `json:"minMemory,omitempty"`
	MaxMemory     workload.Memory `json:"maxMemory,omitempty"`
	MemBufferSize string          `json:"memBufferSize,omitempty"`
	PerPodRate    *float32        `json:"perPodRate,omitempty"`
}

type ByokAddonConfigConfigMonitoringKubeStateMetrics struct {
	MinMemory workload.Memory `json:"minMemory,omitempty"`
}

type ByokAddonConfigConfigMonitoringPrometheusMain struct {
	Storage workload.Memory `json:"storage,omitempty"`
}

type ByokAddonConfigConfigMonitoringPrometheus struct {
	Main ByokAddonConfigConfigMonitoringPrometheusMain `json:"main,omitempty"`
}

type ByokAddonConfigConfigMonitoringRemoteWriteBasicAuth struct {
	Username      string `json:"username,omitempty"`
	Username_file string `json:"username_file,omitempty"`
	Password      string `json:"password,omitempty"`
	Password_file string `json:"password_file,omitempty"`
}

type ByokAddonConfigConfigMonitoringRemoteWriteAuthorization struct {
	Type             string `json:"type,omitempty"`
	Credentials      string `json:"credentials,omitempty"`
	Credentials_file string `json:"credentials_file,omitempty"`
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
	Basic_auth             ByokAddonConfigConfigMonitoringRemoteWriteBasicAuth             `json:"basic_auth,omitempty"`
	Authorization          ByokAddonConfigConfigMonitoringRemoteWriteAuthorization         `json:"authorization,omitempty"`
	Oauth2                 ByokAddonConfigConfigMonitoringRemoteWriteOauth2                `json:"oauth2,omitempty"`
	Follow_redirects       bool                                                            `json:"follow_redirects,omitempty"`
	Enable_http2           bool                                                            `json:"enable_http2,omitempty"`
	Tls_config             ByokAddonConfigConfigMonitoringRemoteWriteTlsConfig             `json:"tls_config,omitempty"`
	Proxy_url              string                                                          `json:"proxy_url,omitempty"`
	No_proxy               string                                                          `json:"no_proxy,omitempty"`
	Proxy_from_environment bool                                                            `json:"proxy_from_environment,omitempty"`
	Proxy_connect_header   ByokAddonConfigConfigMonitoringRemoteWriteProxyConnectHeader    `json:"proxy_connect_header,omitempty"`
	Http_headers           ByokAddonConfigConfigMonitoringRemoteWriteHttpHeaders           `json:"http_headers,omitempty"`
	Url                    string                                                          `json:"url,omitempty"`
	Remote_timeout         string                                                          `json:"remote_timeout,omitempty"`
	Headers                FlexibleAddonConfig                                             `json:"headers,omitempty"`
	Write_relabel_configs  []ByokAddonConfigConfigMonitoringRemoteWriteWriteRelabelConfigs `json:"write_relabel_configs,omitempty"`
	Name                   string                                                          `json:"name,omitempty"`
	Send_exemplars         bool                                                            `json:"send_exemplars,omitempty"`
	Send_native_histograms bool                                                            `json:"send_native_histograms,omitempty"`
	Sigv4                  ByokAddonConfigConfigMonitoringRemoteWriteSigv4                 `json:"sigv4,omitempty"`
	Azuread                ByokAddonConfigConfigMonitoringRemoteWriteAzuread               `json:"azuread,omitempty"`
	Google_iam             ByokAddonConfigConfigMonitoringRemoteWriteGoogleIam             `json:"google_iam,omitempty"`
	Queue_config           ByokAddonConfigConfigMonitoringRemoteWriteQueueConfig           `json:"queue_config,omitempty"`
}

type ByokAddonConfigConfigMonitoring struct {
	MinMemory        workload.Memory                                 `json:"minMemory,omitempty"`
	MaxMemory        workload.Memory                                 `json:"maxMemory,omitempty"`
	KubeStateMetrics ByokAddonConfigConfigMonitoringKubeStateMetrics `json:"kubeStateMetrics,omitempty"`
	Prometheus       ByokAddonConfigConfigMonitoringPrometheus       `json:"prometheus,omitempty"`
	RemoteWrite      []ByokAddonConfigConfigMonitoringRemoteWrite    `json:"remoteWrite,omitempty"`
	ExternalLabels   FlexibleAddonConfig                             `json:"externalLabels,omitempty"`
}

type ByokAddonConfigConfigRedis struct {
	MinCpu    workload.Cpu    `json:"minCpu,omitempty"`
	MaxCpu    workload.Cpu    `json:"maxCpu,omitempty"`
	MinMemory workload.Memory `json:"minMemory,omitempty"`
	MaxMemory workload.Memory `json:"maxMemory,omitempty"`
	Storage   workload.Memory `json:"storage,omitempty"`
}

type ByokAddonConfigConfigRedisHa struct {
	MinCpu    workload.Cpu    `json:"minCpu,omitempty"`
	MaxCpu    workload.Cpu    `json:"maxCpu,omitempty"`
	MinMemory workload.Memory `json:"minMemory,omitempty"`
	MaxMemory workload.Memory `json:"maxMemory,omitempty"`
	Storage   *float32        `json:"storage,omitempty"`
}

type ByokAddonConfigConfigRedisSentinel struct {
	MinCpu    workload.Cpu    `json:"minCpu,omitempty"`
	MaxCpu    workload.Cpu    `json:"maxCpu,omitempty"`
	MinMemory workload.Memory `json:"minMemory,omitempty"`
	MaxMemory workload.Memory `json:"maxMemory,omitempty"`
	Storage   *float32        `json:"storage,omitempty"`
}

type ByokAddonConfigConfigTempoAgent struct {
	MinCpu    workload.Cpu    `json:"minCpu,omitempty"`
	MinMemory workload.Memory `json:"minMemory,omitempty"`
}

type ByokAddonConfigConfigInternalDns struct {
	MinCpu    workload.Cpu    `json:"minCpu,omitempty"`
	MaxCpu    workload.Cpu    `json:"maxCpu,omitempty"`
	MinMemory workload.Memory `json:"minMemory,omitempty"`
	MaxMemory workload.Memory `json:"maxMemory,omitempty"`
}

type ByokAddonConfigConfig struct {
	Actuator      ByokAddonConfigConfigActuator      `json:"actuator,omitempty"`
	Juicefs       ByokAddonConfigConfigJuicefs       `json:"juicefs,omitempty"`
	Middlebox     ByokAddonConfigConfigMiddlebox     `json:"middlebox,omitempty"`
	Common        ByokAddonConfigConfigCommon        `json:"common,omitempty"`
	Longhorn      ByokAddonConfigConfigLonghorn      `json:"longhorn,omitempty"`
	Byok          ByokAddonConfigConfigByok          `json:"byok,omitempty"`
	Ingress       ByokAddonConfigConfigIngress       `json:"ingress,omitempty"`
	Istio         ByokAddonConfigConfigIstio         `json:"istio,omitempty"`
	LogSplitter   ByokAddonConfigConfigLogSplitter   `json:"logSplitter,omitempty"`
	Monitoring    ByokAddonConfigConfigMonitoring    `json:"monitoring,omitempty"`
	Redis         ByokAddonConfigConfigRedis         `json:"redis,omitempty"`
	RedisHa       ByokAddonConfigConfigRedisHa       `json:"redisHa,omitempty"`
	RedisSentinel ByokAddonConfigConfigRedisSentinel `json:"redisSentinel,omitempty"`
	TempoAgent    ByokAddonConfigConfigTempoAgent    `json:"tempoAgent,omitempty"`
	InternalDns   ByokAddonConfigConfigInternalDns   `json:"internalDns,omitempty"`
}

type ByokAddonConfig struct {
	IgnoreUpdates bool                  `json:"ignoreUpdates,omitempty"`
	Location      string                `json:"location"`
	Config        ByokAddonConfigConfig `json:"config,omitempty"`
}

type DashboardAddonStatus struct {
	Url string `json:"url,omitempty"`
}

type FlexibleAddonConfig map[string]string

type HttpHeaderValue struct {
	Values  []string `json:"values,omitempty"`
	Secrets []string `json:"secrets,omitempty"`
	Files   []string `json:"files,omitempty"`
}

type LogsAddonConfig struct {
	AuditEnabled      bool              `json:"auditEnabled,omitempty"`
	IncludeNamespaces RegularExpression `json:"includeNamespaces,omitempty"`
	ExcludeNamespaces RegularExpression `json:"excludeNamespaces,omitempty"`
	Docker            bool              `json:"docker,omitempty"`
	Kubelet           bool              `json:"kubelet,omitempty"`
	Kernel            bool              `json:"kernel,omitempty"`
	Events            bool              `json:"events,omitempty"`
}

type LogsAddonStatus struct {
	LokiAddress string `json:"lokiAddress,omitempty"`
}

type MetricsAddonConfigScrapeAnnotated struct {
	IntervalSeconds   *float32          `json:"intervalSeconds,omitempty"`
	IncludeNamespaces RegularExpression `json:"includeNamespaces,omitempty"`
	ExcludeNamespaces RegularExpression `json:"excludeNamespaces,omitempty"`
	RetainLabels      RegularExpression `json:"retainLabels,omitempty"`
}

type MetricsAddonConfig struct {
	KubeState       bool                              `json:"kubeState,omitempty"`
	CoreDns         bool                              `json:"coreDns,omitempty"`
	Kubelet         bool                              `json:"kubelet,omitempty"`
	Apiserver       bool                              `json:"apiserver,omitempty"`
	NodeExporter    bool                              `json:"nodeExporter,omitempty"`
	Cadvisor        bool                              `json:"cadvisor,omitempty"`
	ScrapeAnnotated MetricsAddonConfigScrapeAnnotated `json:"scrapeAnnotated,omitempty"`
}

type MetricsAddonStatusRemoteWriteConfig map[string]any

type MetricsAddonStatus struct {
	PrometheusEndpoint string                              `json:"prometheusEndpoint,omitempty"`
	RemoteWriteConfig  MetricsAddonStatusRemoteWriteConfig `json:"remoteWriteConfig,omitempty"`
}

type NonCustomizableAddonConfig struct {
}

type NvidiaAddonConfig struct {
	TaintGPUNodes bool `json:"taintGPUNodes,omitempty"`
}

type RegistryMirrorConfigMirrors struct {
	Registry any/* TODO: [object Object]*/ `json:"registry"`
	Mirrors  []string `json:"mirrors,omitempty"`
}

type RegistryMirrorConfig struct {
	Mirrors []RegistryMirrorConfigMirrors `json:"mirrors,omitempty"`
}

type RegularExpression string
