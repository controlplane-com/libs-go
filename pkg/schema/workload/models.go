/* hand-maintained — canonical source of truth (nodelibs/schema/src/workload.ts).
 * `validate:` tags drive validate.Reflect; DSL on go-libs/schema/ipSet/models.go.
 *
 * Reached-from-Workload types are tagged; the joi-to-go duplicate copies (FirewallSpec,
 * JobSpec, RolloutOptions, LoadBalancerSpec, GpuResource, HealthCheckSpec,
 * ReadinessProbe/LivenessProbe, WorkloadSpecVm*, …) are left untagged — WorkloadSpec
 * references the WorkloadSpec*-prefixed inline copies (or the shared vm.VmSpec).
 *
 * Validation (the object-level relations/customs, the field patterns/customs and the
 * leaf/collection nominal types) is attached in validate.go via the Validator
 * interface (ValidationRule) — the go-libs home of what used to be wired from
 * pkg/handlers (schema_workload*.go / schema_reflect_workload.go). Object-level:
 *   - Workload (root)  : Refine (container cross-validation + rollout grace + -headless)
 *   - WorkloadSpec     : Refine (type .when/customs, options validator, ports, gpu, …)
 *   - WorkloadStatus   : Unknown(true)
 *   - ContainerSpecReadinessProbe/LivenessProbe : Xor(exec,grpc,tcpSocket,httpGet)
 *   - ContainerSpecGpu : Refine (xor nvidia|custom)
 *   - HeaderFilter     : Nand(allowedValues,blockedValues)
 *   - WorkloadSpecLoadBalancer : Refine (geoLocation headers)
 *   - WorkloadSpecRequestRetryPolicy : Refine (retryOn default list)
 *
 * Named nominal types (Cpu/Memory/ScheduleType/GpuResourceName/ContainerName/
 * WorkingDir/IntOrPercent/WorkloadName scalars; CIDR/DnsName/WorkloadLink/
 * RetryCondition leaves; CIDRList/WorkloadLinkList/HeaderFilterList/VolumeList/
 * *HttpHeaderList collections) each implement ValidationRule — slice elements and
 * shared types carry their rule on the type, not a field tag. */

package workload

import "github.com/controlplane-com/libs-go/pkg/schema/env"
import "github.com/controlplane-com/libs-go/pkg/schema/base"
import "github.com/controlplane-com/libs-go/pkg/schema/volumeSpec"
import "github.com/controlplane-com/libs-go/pkg/schema/workloadOptions"
import "github.com/controlplane-com/libs-go/pkg/schema/vm"
import "github.com/controlplane-com/libs-go/pkg/schema/port"
import "github.com/controlplane-com/libs-go/pkg/schema/sandbox"
import "github.com/controlplane-com/libs-go/pkg/schema/envoy"

// --- named leaf types resolved via the base-type registry (pkg/handlers) ---

// CIDR resolves to ipCidrRule (ipv4/ipv6 with optional cidr).
type CIDR string

// DnsName resolves to dnsWildcardRule (DnsWildcardPrefix).
type DnsName string

// WorkloadLink resolves to LocalLink+ToKind("!workload","internal","agent").
type WorkloadLink string

// IpSetLink resolves to LocalLink+ToKind("ipset"): a localLink to a kind-ipset resource.
type IpSetLink string

// IdentityLink resolves to LocalLink+ToKind("identity"): a localLink to a kind-identity resource.
type IdentityLink string

// RetryCondition resolves to String().Max(32).Custom(retryOn).
type RetryCondition string

type ContainerOverride struct {
	Name    string              `json:"name" validate:"required"`
	Env     env.EnvCollection   `json:"env,omitzero"`
	EnvFrom []env.EnvFromSource `json:"envFrom,omitzero"`
	Command string              `json:"command,omitzero" validate:"max=256"`
	Args    []string            `json:"args,omitzero"`
	Memory  Memory              `json:"memory,omitzero"`
	Cpu     Cpu                 `json:"cpu,omitzero"`
	Image   ContainerImage      `json:"image,omitzero"`
}

// CronWorkloadContainerOverride is the CLOSED containerOverride schema for the ORG-level
// executeCronWorkload command (orgCommands.ts:12-21 cronWorkloadContainerOverrides). It is
// deliberately NOT the generic ContainerOverride: it carries NO memory/cpu/image keys (a
// submitted one is rejected as unknown) and NO command length cap (TS command has no max).
type CronWorkloadContainerOverride struct {
	Name    string            `json:"name" validate:"required"`
	Env     env.EnvCollection `json:"env,omitzero"`
	Command string            `json:"command,omitzero"`
	Args    []string          `json:"args,omitzero"`
}

type ContainerSpecMetrics struct {
	Port        port.Port    `json:"port" validate:"required"`
	Path        string       `json:"path" validate:"required,default=/metrics,max=128"`
	DropMetrics []base.Regex `json:"dropMetrics,omitzero"`
}

type ContainerSpecPortsProtocol string

const (
	ContainerSpecPortsProtocolHttp  ContainerSpecPortsProtocol = "http"
	ContainerSpecPortsProtocolHttp2 ContainerSpecPortsProtocol = "http2"
	ContainerSpecPortsProtocolGrpc  ContainerSpecPortsProtocol = "grpc"
	ContainerSpecPortsProtocolTcp   ContainerSpecPortsProtocol = "tcp"
)

func (ContainerSpecPortsProtocol) EnumValues() []string {
	return []string{"http", "http2", "grpc", "tcp"}
}

type ContainerSpecPorts struct {
	Protocol ContainerSpecPortsProtocol `json:"protocol,omitzero" validate:"default=http"`
	Number   port.Port                  `json:"number" validate:"required"`
}

type ContainerSpecReadinessProbeExec struct {
	Command []string `json:"command,omitzero" validate:"min=1"`
}

type ContainerSpecReadinessProbeGrpc struct {
	Port *port.Port `json:"port,omitzero"`
}

type ContainerSpecReadinessProbeTcpSocket struct {
	Port *port.Port `json:"port,omitzero"`
}

type ContainerSpecReadinessProbeHttpGetHttpHeaders struct {
	Name  string `json:"name" validate:"max=128,required"`
	Value string `json:"value" validate:"max=128,required"`
}

type ContainerSpecReadinessProbeHttpGetScheme string

const (
	ContainerSpecReadinessProbeHttpGetSchemeHttp  ContainerSpecReadinessProbeHttpGetScheme = "HTTP"
	ContainerSpecReadinessProbeHttpGetSchemeHttps ContainerSpecReadinessProbeHttpGetScheme = "HTTPS"
)

func (ContainerSpecReadinessProbeHttpGetScheme) EnumValues() []string {
	return []string{"HTTP", "HTTPS"}
}

type ContainerSpecReadinessProbeHttpGet struct {
	Path        string                                   `json:"path,omitzero" validate:"default=/,max=256"`
	Port        *port.Port                               `json:"port,omitzero"`
	HttpHeaders ReadinessHttpHeaderList                  `json:"httpHeaders,omitzero"`
	Scheme      ContainerSpecReadinessProbeHttpGetScheme `json:"scheme,omitzero" validate:"default=HTTP"`
}

type ContainerSpecReadinessProbe struct {
	Exec                *ContainerSpecReadinessProbeExec      `json:"exec,omitzero"`
	Grpc                *ContainerSpecReadinessProbeGrpc      `json:"grpc,omitzero"`
	TcpSocket           *ContainerSpecReadinessProbeTcpSocket `json:"tcpSocket,omitzero"`
	HttpGet             *ContainerSpecReadinessProbeHttpGet   `json:"httpGet,omitzero"`
	InitialDelaySeconds *float64                              `json:"initialDelaySeconds,omitzero" validate:"integer,min=0,max=600,default=10"`
	PeriodSeconds       *float64                              `json:"periodSeconds,omitzero" validate:"integer,min=1,max=600,default=10"`
	TimeoutSeconds      *float64                              `json:"timeoutSeconds,omitzero" validate:"integer,min=1,max=600,default=1"`
	SuccessThreshold    *float64                              `json:"successThreshold,omitzero" validate:"integer,min=1,max=20,default=1"`
	FailureThreshold    *float64                              `json:"failureThreshold,omitzero" validate:"integer,min=1,max=20,default=3"`
}

type ContainerSpecLivenessProbeExec struct {
	Command []string `json:"command,omitzero" validate:"min=1"`
}

type ContainerSpecLivenessProbeGrpc struct {
	Port *port.Port `json:"port,omitzero"`
}

type ContainerSpecLivenessProbeTcpSocket struct {
	Port *port.Port `json:"port,omitzero"`
}

type ContainerSpecLivenessProbeHttpGetHttpHeaders struct {
	Name  string `json:"name" validate:"max=128,required"`
	Value string `json:"value" validate:"max=128,required"`
}

type ContainerSpecLivenessProbeHttpGetScheme string

const (
	ContainerSpecLivenessProbeHttpGetSchemeHttp  ContainerSpecLivenessProbeHttpGetScheme = "HTTP"
	ContainerSpecLivenessProbeHttpGetSchemeHttps ContainerSpecLivenessProbeHttpGetScheme = "HTTPS"
)

func (ContainerSpecLivenessProbeHttpGetScheme) EnumValues() []string {
	return []string{"HTTP", "HTTPS"}
}

type ContainerSpecLivenessProbeHttpGet struct {
	Path        string                                  `json:"path,omitzero" validate:"default=/,max=256"`
	Port        *port.Port                              `json:"port,omitzero"`
	HttpHeaders LivenessHttpHeaderList                  `json:"httpHeaders,omitzero"`
	Scheme      ContainerSpecLivenessProbeHttpGetScheme `json:"scheme,omitzero" validate:"default=HTTP"`
}

type ContainerSpecLivenessProbe struct {
	Exec                *ContainerSpecLivenessProbeExec      `json:"exec,omitzero"`
	Grpc                *ContainerSpecLivenessProbeGrpc      `json:"grpc,omitzero"`
	TcpSocket           *ContainerSpecLivenessProbeTcpSocket `json:"tcpSocket,omitzero"`
	HttpGet             *ContainerSpecLivenessProbeHttpGet   `json:"httpGet,omitzero"`
	PeriodSeconds       *float64                             `json:"periodSeconds,omitzero" validate:"integer,min=1,max=600,default=10"`
	TimeoutSeconds      *float64                             `json:"timeoutSeconds,omitzero" validate:"integer,min=1,max=600,default=1"`
	SuccessThreshold    *float64                             `json:"successThreshold,omitzero" validate:"integer,min=1,max=20,default=1"`
	FailureThreshold    *float64                             `json:"failureThreshold,omitzero" validate:"integer,min=1,max=20,default=3"`
	InitialDelaySeconds *float64                             `json:"initialDelaySeconds,omitzero" validate:"integer,min=0,max=600,default=60"`
}

type ContainerSpecGpuNvidia struct {
	Model    any      `json:"model,omitzero"`
	Quantity *float64 `json:"quantity,omitzero" validate:"min=0,max=4,default=1"`
}

type ContainerSpecGpuCustom struct {
	Resource     GpuResourceName     `json:"resource" validate:"max=64,required"`
	RuntimeClass GpuRuntimeClassName `json:"runtimeClass,omitzero" validate:"max=64"`
	Quantity     *float64            `json:"quantity,omitzero" validate:"min=0,max=8,default=1"`
}

type ContainerSpecGpu struct {
	Nvidia *ContainerSpecGpuNvidia `json:"nvidia,omitzero"`
	Custom *ContainerSpecGpuCustom `json:"custom,omitzero"`
}

type ContainerSpecLifecyclePostStartExec struct {
	Command []string `json:"command,omitzero" validate:"max=256"`
}

type ContainerSpecLifecyclePostStart struct {
	Exec *ContainerSpecLifecyclePostStartExec `json:"exec,omitzero"`
}

type ContainerSpecLifecyclePreStopExec struct {
	Command []string `json:"command,omitzero" validate:"max=256"`
}

type ContainerSpecLifecyclePreStop struct {
	Exec *ContainerSpecLifecyclePreStopExec `json:"exec,omitzero"`
}

type ContainerSpecLifecycle struct {
	PostStart *ContainerSpecLifecyclePostStart `json:"postStart,omitzero"`
	PreStop   *ContainerSpecLifecyclePreStop   `json:"preStop,omitzero"`
}

type ContainerSpec struct {
	Name           ContainerName                `json:"name,omitzero" validate:"max=64,unmodifiable"`
	Image          ContainerImage               `json:"image,omitzero"`
	WorkingDir     WorkingDir                   `json:"workingDir,omitzero" validate:"max=128"`
	Metrics        *ContainerSpecMetrics        `json:"metrics,omitzero"`
	Port           *port.Port                   `json:"port,omitzero"`
	Ports          []ContainerSpecPorts         `json:"ports,omitzero"`
	Memory         Memory                       `json:"memory,omitzero" validate:"default=128Mi"`
	ReadinessProbe *ContainerSpecReadinessProbe `json:"readinessProbe,omitzero"`
	LivenessProbe  *ContainerSpecLivenessProbe  `json:"livenessProbe,omitzero"`
	Cpu            Cpu                          `json:"cpu,omitzero" validate:"default=50m"`
	MinCpu         Cpu                          `json:"minCpu,omitzero"`
	MinMemory      Memory                       `json:"minMemory,omitzero"`
	Env            env.EnvCollection            `json:"env,omitzero"`
	EnvFrom        []env.EnvFromSource          `json:"envFrom,omitzero"`
	Gpu            *ContainerSpecGpu            `json:"gpu,omitzero"`
	InheritEnv     *bool                        `json:"inheritEnv,omitzero" validate:"default=false"`
	Command        string                       `json:"command,omitzero" validate:"max=256"`
	Args           []string                     `json:"args,omitzero"`
	Lifecycle      *ContainerSpecLifecycle      `json:"lifecycle,omitzero"`
	Volumes        VolumeList                   `json:"volumes,omitzero" validate:"max=15"`
}

type Cpu string

// SchedulerName is the pod schedulerName routing pods to a custom scheduler
// (workload.ts Extras, 2e07953ae). A DNS-1123 subdomain — what k8s accepts here.
type SchedulerName string

type ExtrasMetadataLabels map[string]string

type ExtrasMetadata struct {
	Labels ExtrasMetadataLabels `json:"labels,omitzero"`
}

type Extras struct {
	Affinity                  any             `json:"affinity,omitzero"`
	Tolerations               []any           `json:"tolerations,omitzero"`
	TopologySpreadConstraints []any           `json:"topologySpreadConstraints,omitzero"`
	SchedulerName             SchedulerName   `json:"schedulerName,omitzero"`
	Metadata                  *ExtrasMetadata `json:"metadata,omitzero"`
}

type FirewallSpecExternalOutboundAllowPortProtocol string

const (
	FirewallSpecExternalOutboundAllowPortProtocolHttp  FirewallSpecExternalOutboundAllowPortProtocol = "http"
	FirewallSpecExternalOutboundAllowPortProtocolHttps FirewallSpecExternalOutboundAllowPortProtocol = "https"
	FirewallSpecExternalOutboundAllowPortProtocolTcp   FirewallSpecExternalOutboundAllowPortProtocol = "tcp"
)

type FirewallSpecExternalOutboundAllowPort struct {
	Protocol FirewallSpecExternalOutboundAllowPortProtocol `json:"protocol,omitzero" validate:"required"`
	Number   float64                                       `json:"number"`
}

type FirewallSpecExternalHttp struct {
	InboundHeaderFilter []HeaderFilter `json:"inboundHeaderFilter,omitzero"`
}

type FirewallSpecExternal struct {
	InboundAllowCIDR      []string                                `json:"inboundAllowCIDR,omitzero"`
	InboundBlockedCIDR    []string                                `json:"inboundBlockedCIDR,omitzero"`
	OutboundAllowHostname []string                                `json:"outboundAllowHostname,omitzero"`
	OutboundAllowPort     []FirewallSpecExternalOutboundAllowPort `json:"outboundAllowPort,omitzero"`
	OutboundAllowCIDR     []string                                `json:"outboundAllowCIDR,omitzero"`
	OutboundBlockedCIDR   []string                                `json:"outboundBlockedCIDR,omitzero"`
	Http                  *FirewallSpecExternalHttp               `json:"http,omitzero"`
}

type FirewallSpecInternalInboundAllowType string

const (
	FirewallSpecInternalInboundAllowTypeNone         FirewallSpecInternalInboundAllowType = "none"
	FirewallSpecInternalInboundAllowTypeSameGvc      FirewallSpecInternalInboundAllowType = "same-gvc"
	FirewallSpecInternalInboundAllowTypeSameOrg      FirewallSpecInternalInboundAllowType = "same-org"
	FirewallSpecInternalInboundAllowTypeWorkloadList FirewallSpecInternalInboundAllowType = "workload-list"
)

type FirewallSpecInternal struct {
	InboundAllowType     FirewallSpecInternalInboundAllowType `json:"inboundAllowType,omitzero"`
	InboundAllowWorkload []string                             `json:"inboundAllowWorkload,omitzero"`
}

type FirewallSpec struct {
	External *FirewallSpecExternal `json:"external,omitzero"`
	Internal *FirewallSpecInternal `json:"internal,omitzero"`
}

type GpuResourceNvidia struct {
	Model    any      `json:"model,omitzero"`
	Quantity *float64 `json:"quantity,omitzero"`
}

type GpuResourceCustom struct {
	Resource     string   `json:"resource"`
	RuntimeClass string   `json:"runtimeClass,omitzero"`
	Quantity     *float64 `json:"quantity,omitzero"`
}

type GpuResource struct {
	Nvidia *GpuResourceNvidia `json:"nvidia,omitzero"`
	Custom *GpuResourceCustom `json:"custom,omitzero"`
}

type HeaderFilter struct {
	Key           string       `json:"key" validate:"max=128,required"`
	AllowedValues []base.Regex `json:"allowedValues,omitzero"`
	BlockedValues []base.Regex `json:"blockedValues,omitzero"`
}

type HealthCheckSpecExec struct {
	Command []string `json:"command,omitzero"`
}

type HealthCheckSpecGrpc struct {
	Port *float64 `json:"port,omitzero"`
}

type HealthCheckSpecTcpSocket struct {
	Port *float64 `json:"port,omitzero"`
}

type HealthCheckSpecHttpGetHttpHeaders struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type HealthCheckSpecHttpGetScheme string

const (
	HealthCheckSpecHttpGetSchemeHttp  HealthCheckSpecHttpGetScheme = "HTTP"
	HealthCheckSpecHttpGetSchemeHttps HealthCheckSpecHttpGetScheme = "HTTPS"
)

type HealthCheckSpecHttpGet struct {
	Path        string                              `json:"path,omitzero"`
	Port        *float64                            `json:"port,omitzero"`
	HttpHeaders []HealthCheckSpecHttpGetHttpHeaders `json:"httpHeaders,omitzero"`
	Scheme      HealthCheckSpecHttpGetScheme        `json:"scheme,omitzero"`
}

type HealthCheckSpec struct {
	Exec                *HealthCheckSpecExec      `json:"exec,omitzero"`
	Grpc                *HealthCheckSpecGrpc      `json:"grpc,omitzero"`
	TcpSocket           *HealthCheckSpecTcpSocket `json:"tcpSocket,omitzero"`
	HttpGet             *HealthCheckSpecHttpGet   `json:"httpGet,omitzero"`
	InitialDelaySeconds *float64                  `json:"initialDelaySeconds,omitzero"`
	PeriodSeconds       *float64                  `json:"periodSeconds,omitzero"`
	TimeoutSeconds      *float64                  `json:"timeoutSeconds,omitzero"`
	SuccessThreshold    *float64                  `json:"successThreshold,omitzero"`
	FailureThreshold    *float64                  `json:"failureThreshold,omitzero"`
}

type HealthCheckStatus struct {
	Active      bool     `json:"active" validate:"required"`
	Success     *bool    `json:"success,omitzero"`
	Code        *float64 `json:"code,omitzero" validate:"integer,min=200,max=599"`
	Message     string   `json:"message,omitzero"`
	Failures    *float64 `json:"failures,omitzero" validate:"integer,min=0"`
	Successes   *float64 `json:"successes,omitzero" validate:"integer,min=0"`
	LastChecked string   `json:"lastChecked,omitzero" validate:"date"`
}

type JobSpecConcurrencyPolicy string

const (
	JobSpecConcurrencyPolicyForbid  JobSpecConcurrencyPolicy = "Forbid"
	JobSpecConcurrencyPolicyReplace JobSpecConcurrencyPolicy = "Replace"
	JobSpecConcurrencyPolicyAllow   JobSpecConcurrencyPolicy = "Allow"
)

type JobSpecRestartPolicy string

const (
	JobSpecRestartPolicyOnFailure JobSpecRestartPolicy = "OnFailure"
	JobSpecRestartPolicyNever     JobSpecRestartPolicy = "Never"
)

type JobSpec struct {
	Schedule              string                   `json:"schedule"`
	ConcurrencyPolicy     JobSpecConcurrencyPolicy `json:"concurrencyPolicy,omitzero"`
	HistoryLimit          *float64                 `json:"historyLimit,omitzero"`
	RestartPolicy         JobSpecRestartPolicy     `json:"restartPolicy,omitzero"`
	ActiveDeadlineSeconds *float64                 `json:"activeDeadlineSeconds,omitzero"`
}

type LivenessProbeExec struct {
	Command []string `json:"command,omitzero"`
}

type LivenessProbeGrpc struct {
	Port *float64 `json:"port,omitzero"`
}

type LivenessProbeTcpSocket struct {
	Port *float64 `json:"port,omitzero"`
}

type LivenessProbeHttpGetHttpHeaders struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type LivenessProbeHttpGetScheme string

const (
	LivenessProbeHttpGetSchemeHttp  LivenessProbeHttpGetScheme = "HTTP"
	LivenessProbeHttpGetSchemeHttps LivenessProbeHttpGetScheme = "HTTPS"
)

type LivenessProbeHttpGet struct {
	Path        string                            `json:"path,omitzero"`
	Port        *float64                          `json:"port,omitzero"`
	HttpHeaders []LivenessProbeHttpGetHttpHeaders `json:"httpHeaders,omitzero"`
	Scheme      LivenessProbeHttpGetScheme        `json:"scheme,omitzero"`
}

type LivenessProbe struct {
	Exec                *LivenessProbeExec      `json:"exec,omitzero"`
	Grpc                *LivenessProbeGrpc      `json:"grpc,omitzero"`
	TcpSocket           *LivenessProbeTcpSocket `json:"tcpSocket,omitzero"`
	HttpGet             *LivenessProbeHttpGet   `json:"httpGet,omitzero"`
	PeriodSeconds       *float64                `json:"periodSeconds,omitzero"`
	TimeoutSeconds      *float64                `json:"timeoutSeconds,omitzero"`
	SuccessThreshold    *float64                `json:"successThreshold,omitzero"`
	FailureThreshold    *float64                `json:"failureThreshold,omitzero"`
	InitialDelaySeconds *float64                `json:"initialDelaySeconds,omitzero"`
}

type LoadBalancerPortProtocol string

const (
	LoadBalancerPortProtocolTcp LoadBalancerPortProtocol = "TCP"
	LoadBalancerPortProtocolUdp LoadBalancerPortProtocol = "UDP"
)

func (LoadBalancerPortProtocol) EnumValues() []string { return []string{"TCP", "UDP"} }

type LoadBalancerPortScheme string

const (
	LoadBalancerPortSchemeHttp  LoadBalancerPortScheme = "http"
	LoadBalancerPortSchemeTcp   LoadBalancerPortScheme = "tcp"
	LoadBalancerPortSchemeHttps LoadBalancerPortScheme = "https"
	LoadBalancerPortSchemeWs    LoadBalancerPortScheme = "ws"
	LoadBalancerPortSchemeWss   LoadBalancerPortScheme = "wss"
)

func (LoadBalancerPortScheme) EnumValues() []string {
	return []string{"http", "tcp", "https", "ws", "wss"}
}

type LoadBalancerPort struct {
	ExternalPort  float64                  `json:"externalPort" validate:"min=22,max=32768,required"`
	Protocol      LoadBalancerPortProtocol `json:"protocol,omitzero" validate:"required"`
	Scheme        LoadBalancerPortScheme   `json:"scheme,omitzero"`
	ContainerPort *port.Port               `json:"containerPort,omitzero"`
}

type LoadBalancerSpecDirect struct {
	Enabled bool               `json:"enabled"`
	Ports   []LoadBalancerPort `json:"ports,omitzero"`
	IpSet   *string            `json:"ipSet,omitzero"`
}

type LoadBalancerSpecGeoLocationHeaders struct {
	Asn     string `json:"asn,omitzero"`
	City    string `json:"city,omitzero"`
	Country string `json:"country,omitzero"`
	Region  string `json:"region,omitzero"`
}

type LoadBalancerSpecGeoLocation struct {
	Enabled *bool                               `json:"enabled,omitzero"`
	Headers *LoadBalancerSpecGeoLocationHeaders `json:"headers,omitzero"`
}

type LoadBalancerSpec struct {
	Direct        *LoadBalancerSpecDirect      `json:"direct,omitzero"`
	GeoLocation   *LoadBalancerSpecGeoLocation `json:"geoLocation,omitzero"`
	ReplicaDirect *bool                        `json:"replicaDirect,omitzero"`
}

type LoadBalancerStatus struct {
	Origin string `json:"origin,omitzero"`
	Url    string `json:"url,omitzero" validate:"uri"`
}

type Memory string

type ReadinessProbeExec struct {
	Command []string `json:"command,omitzero"`
}

type ReadinessProbeGrpc struct {
	Port *float64 `json:"port,omitzero"`
}

type ReadinessProbeTcpSocket struct {
	Port *float64 `json:"port,omitzero"`
}

type ReadinessProbeHttpGetHttpHeaders struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type ReadinessProbeHttpGetScheme string

const (
	ReadinessProbeHttpGetSchemeHttp  ReadinessProbeHttpGetScheme = "HTTP"
	ReadinessProbeHttpGetSchemeHttps ReadinessProbeHttpGetScheme = "HTTPS"
)

type ReadinessProbeHttpGet struct {
	Path        string                             `json:"path,omitzero"`
	Port        *float64                           `json:"port,omitzero"`
	HttpHeaders []ReadinessProbeHttpGetHttpHeaders `json:"httpHeaders,omitzero"`
	Scheme      ReadinessProbeHttpGetScheme        `json:"scheme,omitzero"`
}

type ReadinessProbe struct {
	Exec                *ReadinessProbeExec      `json:"exec,omitzero"`
	Grpc                *ReadinessProbeGrpc      `json:"grpc,omitzero"`
	TcpSocket           *ReadinessProbeTcpSocket `json:"tcpSocket,omitzero"`
	HttpGet             *ReadinessProbeHttpGet   `json:"httpGet,omitzero"`
	InitialDelaySeconds *float64                 `json:"initialDelaySeconds,omitzero"`
	PeriodSeconds       *float64                 `json:"periodSeconds,omitzero"`
	TimeoutSeconds      *float64                 `json:"timeoutSeconds,omitzero"`
	SuccessThreshold    *float64                 `json:"successThreshold,omitzero"`
	FailureThreshold    *float64                 `json:"failureThreshold,omitzero"`
}

type RequestRetryPolicy struct {
	Attempts *float64 `json:"attempts,omitzero"`
	RetryOn  []string `json:"retryOn,omitzero"`
}

type ResolvedImageManifestsPlatform map[string]string

type ResolvedImageManifests struct {
	Image     string                         `json:"image" validate:"required"`
	MediaType string                         `json:"mediaType" validate:"required"`
	Digest    string                         `json:"digest" validate:"required"`
	Platform  ResolvedImageManifestsPlatform `json:"platform,omitzero"`
}

type ResolvedImage struct {
	Digest    string                   `json:"digest,omitzero"`
	Manifests []ResolvedImageManifests `json:"manifests,omitzero"`
}

type ResolvedImages struct {
	ResolvedForVersion *float64        `json:"resolvedForVersion,omitzero" validate:"integer"`
	ResolvedAt         string          `json:"resolvedAt,omitzero" validate:"date"`
	ErrorMessages      []string        `json:"errorMessages,omitzero"`
	NextRetryAt        string          `json:"nextRetryAt,omitzero" validate:"date"`
	Images             []ResolvedImage `json:"images,omitzero"`
}

type RolloutOptionsScalingPolicy string

const (
	RolloutOptionsScalingPolicyOrderedReady RolloutOptionsScalingPolicy = "OrderedReady"
	RolloutOptionsScalingPolicyParallel     RolloutOptionsScalingPolicy = "Parallel"
)

type RolloutOptions struct {
	MinReadySeconds               *float64                    `json:"minReadySeconds,omitzero"`
	MaxUnavailableReplicas        string                      `json:"maxUnavailableReplicas,omitzero"`
	MaxSurgeReplicas              string                      `json:"maxSurgeReplicas,omitzero"`
	ScalingPolicy                 RolloutOptionsScalingPolicy `json:"scalingPolicy,omitzero"`
	TerminationGracePeriodSeconds *float64                    `json:"terminationGracePeriodSeconds,omitzero"`
}

type ScheduleType string //ts:export

type SecurityOptions struct {
	FilesystemGroupId *float64 `json:"filesystemGroupId,omitzero"`
	RunAsUser         *float64 `json:"runAsUser,omitzero"`
}

type WorkloadTags = base.Tags

type WorkloadSpecFirewallConfigExternalOutboundAllowPortProtocol string

const (
	WorkloadSpecFirewallConfigExternalOutboundAllowPortProtocolHttp  WorkloadSpecFirewallConfigExternalOutboundAllowPortProtocol = "http"
	WorkloadSpecFirewallConfigExternalOutboundAllowPortProtocolHttps WorkloadSpecFirewallConfigExternalOutboundAllowPortProtocol = "https"
	WorkloadSpecFirewallConfigExternalOutboundAllowPortProtocolTcp   WorkloadSpecFirewallConfigExternalOutboundAllowPortProtocol = "tcp"
)

func (WorkloadSpecFirewallConfigExternalOutboundAllowPortProtocol) EnumValues() []string {
	return []string{"http", "https", "tcp"}
}

type WorkloadSpecFirewallConfigExternalOutboundAllowPort struct {
	Protocol WorkloadSpecFirewallConfigExternalOutboundAllowPortProtocol `json:"protocol,omitzero" validate:"default=tcp,required"`
	Number   port.Port                                                   `json:"number" validate:"positive,max=65000,required"`
}

type WorkloadSpecFirewallConfigExternalHttp struct {
	InboundHeaderFilter HeaderFilterList `json:"inboundHeaderFilter,omitzero"`
}

type WorkloadSpecFirewallConfigExternal struct {
	InboundAllowCIDR      CIDRList                                              `json:"inboundAllowCIDR,omitzero" validate:"max=250,sort,default=[]"`
	InboundBlockedCIDR    CIDRList                                              `json:"inboundBlockedCIDR,omitzero" validate:"sort,default=[]"`
	OutboundAllowHostname []DnsName                                             `json:"outboundAllowHostname,omitzero" validate:"default=[]"`
	OutboundAllowPort     []WorkloadSpecFirewallConfigExternalOutboundAllowPort `json:"outboundAllowPort,omitzero"`
	OutboundAllowCIDR     CIDRList                                              `json:"outboundAllowCIDR,omitzero" validate:"default=[],sort"`
	OutboundBlockedCIDR   CIDRList                                              `json:"outboundBlockedCIDR,omitzero" validate:"default=[],sort"`
	Http                  *WorkloadSpecFirewallConfigExternalHttp               `json:"http,omitzero"`
}

type WorkloadSpecFirewallConfigInternalInboundAllowType string

const (
	WorkloadSpecFirewallConfigInternalInboundAllowTypeNone         WorkloadSpecFirewallConfigInternalInboundAllowType = "none"
	WorkloadSpecFirewallConfigInternalInboundAllowTypeSameGvc      WorkloadSpecFirewallConfigInternalInboundAllowType = "same-gvc"
	WorkloadSpecFirewallConfigInternalInboundAllowTypeSameOrg      WorkloadSpecFirewallConfigInternalInboundAllowType = "same-org"
	WorkloadSpecFirewallConfigInternalInboundAllowTypeWorkloadList WorkloadSpecFirewallConfigInternalInboundAllowType = "workload-list"
)

func (WorkloadSpecFirewallConfigInternalInboundAllowType) EnumValues() []string {
	return []string{"none", "same-gvc", "same-org", "workload-list"}
}

type WorkloadSpecFirewallConfigInternal struct {
	InboundAllowType     WorkloadSpecFirewallConfigInternalInboundAllowType `json:"inboundAllowType,omitzero" validate:"default=none"`
	InboundAllowWorkload WorkloadLinkList                                   `json:"inboundAllowWorkload,omitzero" validate:"default=[],sort"`
}

type WorkloadSpecFirewallConfig struct {
	External *WorkloadSpecFirewallConfigExternal `json:"external,omitzero"`
	Internal *WorkloadSpecFirewallConfigInternal `json:"internal,omitzero"`
}

type WorkloadSpecJobConcurrencyPolicy string

const (
	WorkloadSpecJobConcurrencyPolicyForbid  WorkloadSpecJobConcurrencyPolicy = "Forbid"
	WorkloadSpecJobConcurrencyPolicyReplace WorkloadSpecJobConcurrencyPolicy = "Replace"
	WorkloadSpecJobConcurrencyPolicyAllow   WorkloadSpecJobConcurrencyPolicy = "Allow"
)

func (WorkloadSpecJobConcurrencyPolicy) EnumValues() []string {
	return []string{"Forbid", "Replace", "Allow"}
}

type WorkloadSpecJobRestartPolicy string

const (
	WorkloadSpecJobRestartPolicyOnFailure WorkloadSpecJobRestartPolicy = "OnFailure"
	WorkloadSpecJobRestartPolicyNever     WorkloadSpecJobRestartPolicy = "Never"
)

func (WorkloadSpecJobRestartPolicy) EnumValues() []string { return []string{"OnFailure", "Never"} }

type WorkloadSpecJob struct {
	Schedule              ScheduleType                     `json:"schedule" validate:"required"`
	ConcurrencyPolicy     WorkloadSpecJobConcurrencyPolicy `json:"concurrencyPolicy,omitzero" validate:"default=Forbid"`
	HistoryLimit          *float64                         `json:"historyLimit,omitzero" validate:"integer,default=5,min=1,max=10"`
	RestartPolicy         WorkloadSpecJobRestartPolicy     `json:"restartPolicy,omitzero"`
	ActiveDeadlineSeconds *float64                         `json:"activeDeadlineSeconds,omitzero" validate:"min=1,max=86400"`
}

// WorkloadSpecSidecarEnvoy is the SAME shape gvc's sidecar carries. It used to be a
// structural duplicate, which meant every object-level rule on envoy.EnvoyFilters
// (the root refine, the xor groups) reached gvc and silently missed workload.
type WorkloadSpecSidecarEnvoy = envoy.EnvoyFilters

type WorkloadSpecSidecar struct {
	Envoy *WorkloadSpecSidecarEnvoy `json:"envoy,omitzero"`
}

type WorkloadSpecRolloutOptionsScalingPolicy string

const (
	WorkloadSpecRolloutOptionsScalingPolicyOrderedReady WorkloadSpecRolloutOptionsScalingPolicy = "OrderedReady"
	WorkloadSpecRolloutOptionsScalingPolicyParallel     WorkloadSpecRolloutOptionsScalingPolicy = "Parallel"
)

func (WorkloadSpecRolloutOptionsScalingPolicy) EnumValues() []string {
	return []string{"OrderedReady", "Parallel"}
}

type WorkloadSpecRolloutOptions struct {
	MinReadySeconds               *float64                                `json:"minReadySeconds,omitzero" validate:"integer,default=0"`
	MaxUnavailableReplicas        IntOrPercent                            `json:"maxUnavailableReplicas,omitzero" validate:"max=5"`
	MaxSurgeReplicas              IntOrPercent                            `json:"maxSurgeReplicas,omitzero" validate:"max=5"`
	ScalingPolicy                 WorkloadSpecRolloutOptionsScalingPolicy `json:"scalingPolicy,omitzero" validate:"default=OrderedReady"`
	TerminationGracePeriodSeconds *float64                                `json:"terminationGracePeriodSeconds,omitzero" validate:"min=0,max=3600,default=90"`
	RolloutOnReferenceChange      *bool                                   `json:"rolloutOnReferenceChange,omitzero"`
}

type WorkloadSpecSecurityOptions struct {
	FilesystemGroupId *float64 `json:"filesystemGroupId,omitzero" validate:"integer,min=1,max=65534"`
	RunAsUser         *float64 `json:"runAsUser,omitzero" validate:"integer,min=1,max=65534"`
}

type WorkloadSpecLoadBalancerDirect struct {
	Enabled bool               `json:"enabled" validate:"required"`
	Ports   []LoadBalancerPort `json:"ports,omitzero"`
	IpSet   *IpSetLink         `json:"ipSet,omitzero"`
}

type WorkloadSpecLoadBalancerGeoLocationHeaders struct {
	Asn     string `json:"asn,omitzero" validate:"max=128"`
	City    string `json:"city,omitzero" validate:"max=128"`
	Country string `json:"country,omitzero" validate:"max=128"`
	Region  string `json:"region,omitzero" validate:"max=128"`
}

type WorkloadSpecLoadBalancerGeoLocation struct {
	Enabled *bool                                       `json:"enabled,omitzero" validate:"default=false"`
	Headers *WorkloadSpecLoadBalancerGeoLocationHeaders `json:"headers,omitzero"`
}

type WorkloadSpecLoadBalancer struct {
	Direct        *WorkloadSpecLoadBalancerDirect      `json:"direct,omitzero"`
	GeoLocation   *WorkloadSpecLoadBalancerGeoLocation `json:"geoLocation,omitzero"`
	ReplicaDirect *bool                                `json:"replicaDirect,omitzero" validate:"default=false"`
}

type WorkloadSpecNetworkingInternalRouting string

const (
	WorkloadSpecNetworkingInternalRoutingLocationEndpoint WorkloadSpecNetworkingInternalRouting = "locationEndpoint"
	WorkloadSpecNetworkingInternalRoutingGvcEndpoint      WorkloadSpecNetworkingInternalRouting = "gvcEndpoint"
)

func (WorkloadSpecNetworkingInternalRouting) EnumValues() []string {
	return []string{"locationEndpoint", "gvcEndpoint"}
}

type WorkloadSpecNetworkingInternal struct {
	Routing WorkloadSpecNetworkingInternalRouting `json:"routing,omitzero"`
}

type WorkloadSpecNetworking struct {
	Internal *WorkloadSpecNetworkingInternal `json:"internal,omitzero"`
}

type WorkloadSpecExtrasMetadataLabels map[string]string

type WorkloadSpecExtrasMetadata struct {
	Labels WorkloadSpecExtrasMetadataLabels `json:"labels,omitzero"`
}

type WorkloadSpecExtras struct {
	Affinity                  any                         `json:"affinity,omitzero"`
	Tolerations               []any                       `json:"tolerations,omitzero"`
	TopologySpreadConstraints []any                       `json:"topologySpreadConstraints,omitzero"`
	SchedulerName             SchedulerName               `json:"schedulerName,omitzero"`
	Metadata                  *WorkloadSpecExtrasMetadata `json:"metadata,omitzero"`
}

type WorkloadSpecRequestRetryPolicy struct {
	Attempts *float64         `json:"attempts,omitzero" validate:"integer,min=0,max=4294967295,default=2"`
	RetryOn  []RetryCondition `json:"retryOn,omitzero"`
}

type WorkloadSpec struct {
	Type               WorkloadType                    `json:"type,omitzero" validate:"default=serverless"`
	IdentityLink       *IdentityLink                   `json:"identityLink,omitzero"`
	Containers         []ContainerSpec                 `json:"containers" validate:"min=1,max=8,required"`
	FirewallConfig     *WorkloadSpecFirewallConfig     `json:"firewallConfig,omitzero"`
	DefaultOptions     *workloadOptions.DefaultOptions `json:"defaultOptions,omitzero"`
	LocalOptions       workloadOptions.LocalOptions    `json:"localOptions,omitzero"`
	Job                *WorkloadSpecJob                `json:"job,omitzero"`
	Sidecar            *WorkloadSpecSidecar            `json:"sidecar,omitzero"`
	SupportDynamicTags *bool                           `json:"supportDynamicTags,omitzero" validate:"default=false"`
	RolloutOptions     *WorkloadSpecRolloutOptions     `json:"rolloutOptions,omitzero"`
	DisruptionBudget   *DisruptionBudget               `json:"disruptionBudget,omitzero"`
	SecurityOptions    *WorkloadSpecSecurityOptions    `json:"securityOptions,omitzero"`
	LoadBalancer       *WorkloadSpecLoadBalancer       `json:"loadBalancer,omitzero"`
	Networking         *WorkloadSpecNetworking         `json:"networking,omitzero"`
	Extras             *WorkloadSpecExtras             `json:"extras,omitzero"`
	RequestRetryPolicy *WorkloadSpecRequestRetryPolicy `json:"requestRetryPolicy,omitzero"`
	Vm                 *vm.VmSpec                      `json:"vm,omitzero"`
	Sandbox            *sandbox.SandboxSpec            `json:"sandbox,omitzero" validate:"optional"`
}

type WorkloadStatusHealthCheck struct {
	Active      bool     `json:"active" validate:"required"`
	Success     *bool    `json:"success,omitzero"`
	Code        *float64 `json:"code,omitzero" validate:"integer,min=200,max=599"`
	Message     string   `json:"message,omitzero"`
	Failures    *float64 `json:"failures,omitzero" validate:"integer,min=0"`
	Successes   *float64 `json:"successes,omitzero" validate:"integer,min=0"`
	LastChecked string   `json:"lastChecked,omitzero" validate:"date"`
}

type WorkloadStatusResolvedImages struct {
	ResolvedForVersion *float64        `json:"resolvedForVersion,omitzero" validate:"integer"`
	ResolvedAt         string          `json:"resolvedAt,omitzero" validate:"date"`
	ErrorMessages      []string        `json:"errorMessages,omitzero"`
	NextRetryAt        string          `json:"nextRetryAt,omitzero" validate:"date"`
	Images             []ResolvedImage `json:"images,omitzero"`
}

type WorkloadStatus struct {
	ParentId             string                        `json:"parentId,omitzero" validate:"uuid"`
	CanonicalEndpoint    string                        `json:"canonicalEndpoint,omitzero" validate:"uri"`
	Endpoint             string                        `json:"endpoint,omitzero" validate:"uri"`
	InternalName         string                        `json:"internalName,omitzero"`
	ReplicaInternalNames []string                      `json:"replicaInternalNames,omitzero"`
	HealthCheck          *WorkloadStatusHealthCheck    `json:"healthCheck,omitzero"`
	CurrentReplicaCount  *float64                      `json:"currentReplicaCount,omitzero"`
	ResolvedImages       *WorkloadStatusResolvedImages `json:"resolvedImages,omitzero"`
	LoadBalancer         []LoadBalancerStatus          `json:"loadBalancer,omitzero"`
	SuspendedStatus      string                        `json:"suspendedStatus,omitzero"`
	// WorkloadStatus is .unknown(true) (validate.go:401): properties it does not
	// name round-trip untouched.
	Extras base.Extras `json:"-"`
}

type workloadStatusAlias WorkloadStatus

func (s WorkloadStatus) MarshalJSON() ([]byte, error) {
	return base.MarshalExtras(workloadStatusAlias(s), s.Extras)
}

func (s *WorkloadStatus) UnmarshalJSON(b []byte) error {
	return base.UnmarshalExtras(b, (*workloadStatusAlias)(s), &s.Extras)
}

type Workload struct {
	Id           string          `json:"id,omitzero" validate:"uuid,readOnly"`
	Kind         base.Kind       `json:"kind,omitzero"`
	Version      *float64        `json:"version,omitzero" validate:"positive,integer,readOnly"`
	Description  string          `json:"description,omitzero" validate:"trim,max=250,allow="`
	Tags         WorkloadTags    `json:"tags,omitzero" validate:"default={}"`
	Created      string          `json:"created,omitzero" validate:"date,readOnly"`
	LastModified string          `json:"lastModified,omitzero" validate:"date,readOnly"`
	Links        base.Links      `json:"links,omitzero" validate:"readOnly"`
	Name         WorkloadName    `json:"name,omitzero" validate:"max=49,unmodifiable"`
	Gvc          string          `json:"gvc,omitzero" validate:"allow=,max=64"`
	Health       *WorkloadHealth `json:"health,omitzero"`
	Spec         WorkloadSpec    `json:"spec" validate:"required"`
	Status       *WorkloadStatus `json:"status,omitzero"`
}

type WorkloadHealth struct {
	Readiness      string   `json:"readiness,omitzero"`
	SyncFailed     *bool    `json:"syncFailed,omitzero"`
	ReadyLocations *float64 `json:"readyLocations,omitzero" validate:"integer,min=0"`
	TotalLocations *float64 `json:"totalLocations,omitzero" validate:"integer,min=0"`
	ReadyReplicas  *float64 `json:"readyReplicas,omitzero" validate:"integer,min=0"`
	TotalReplicas  *float64 `json:"totalReplicas,omitzero" validate:"integer,min=0"`
} //ts:export

type WorkloadType string

const (
	WorkloadTypeServerless WorkloadType = "serverless"
	WorkloadTypeStandard   WorkloadType = "standard"
	WorkloadTypeCron       WorkloadType = "cron"
	WorkloadTypeStateful   WorkloadType = "stateful"
	WorkloadTypeVm         WorkloadType = "vm"
)

func (WorkloadType) EnumValues() []string {
	return []string{"serverless", "standard", "cron", "stateful", "vm"}
}

// --- named nominal types (their ValidationRule rules live in validate.go) ---

// GpuResourceName ports gpu.custom.resource `regex(/^[a-zA-Z0-9./_-]*$/)`
// (workload.ts:274-278) — a k8s extended-resource name, which may carry `_`/`-`.
type GpuResourceName string

// GpuRuntimeClassName ports gpu.custom.runtimeClass `regex(/^[a-zA-Z0-9./]*$/)`
// (workload.ts:279-283) — the NARROWER of the two; sharing one type with Resource
// rejected every `_`/`-` resource name TS accepts.
type GpuRuntimeClassName string

// ContainerName ports ContainerSpec.name: `regex(name)` + the excluded-names
// disallow + the cpln-/debugger- prefix check (max/unmodifiable stay field tags).
type ContainerName string

// WorkingDir ports ContainerSpec.workingDir (absolute-path custom).
type WorkingDir string

// IntOrPercent ports rolloutOptions maxUnavailable/maxSurge (integer or percent).
type IntOrPercent string

// DisruptionBudget ports workload.ts DisruptionBudget (master f94fb6d78): set either
// minAvailable or maxUnavailable (oxor), each a whole number or whole percentage.
type DisruptionBudget struct {
	MinAvailable   string `json:"minAvailable,omitzero"`
	MaxUnavailable string `json:"maxUnavailable,omitzero"`
}

// WorkloadName ports Workload.name `regex(name)` (max/unmodifiable stay field tags).
type WorkloadName string

// CIDRList ports the firewall CIDR arrays' makeUnique(empty) dedupe. (sort / max /
// default=[] remain declarative field tags.)
type CIDRList []CIDR

// WorkloadLinkList ports firewall inboundAllowWorkload makeUnique(empty).
type WorkloadLinkList []WorkloadLink

// HeaderFilterList ports firewall inboundHeaderFilter makeUnique('key').
type HeaderFilterList []HeaderFilter

// ReadinessHttpHeaderList ports readinessProbe.httpGet.httpHeaders makeUnique('name').
type ReadinessHttpHeaderList []ContainerSpecReadinessProbeHttpGetHttpHeaders

// LivenessHttpHeaderList ports livenessProbe.httpGet.httpHeaders makeUnique('name').
type LivenessHttpHeaderList []ContainerSpecLivenessProbeHttpGetHttpHeaders

// VolumeList ports container volumes .unique() + mount-path parent/dup checks.
type VolumeList []volumeSpec.VolumeSpec

// ContainerImage is a workload container image reference (a literal image string
// or an org/gvc image link) — the strongly-typed replacement for the `imageLink`
// tag. Its rule is base.ImageLink (format check for literals, expand+toKind for links).
type ContainerImage string
