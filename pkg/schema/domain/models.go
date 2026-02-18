/* auto-generated */

package domain

import "github.com/controlplane-com/libs-go/pkg/schema/base"

type CertChallengeType string

const (
	CertChallengeTypeHttp01 CertChallengeType = "http01"
	CertChallengeTypeDns01  CertChallengeType = "dns01"
)

type DnsConfigRecord struct {
	Type  string  `json:"type,omitempty"`
	Ttl   float32 `json:"ttl"`
	Host  string  `json:"host,omitempty"`
	Value string  `json:"value,omitempty"`
}

type DomainTags map[string]any

type Domain struct {
	Id           string       `json:"id,omitempty"`
	Kind         base.Kind    `json:"kind,omitempty"`
	Version      float32      `json:"version"`
	Description  string       `json:"description,omitempty"`
	Tags         DomainTags   `json:"tags,omitempty"`
	Created      string       `json:"created,omitempty"`
	LastModified string       `json:"lastModified,omitempty"`
	Links        base.Links   `json:"links,omitempty"`
	Name         string       `json:"name,omitempty"`
	Spec         DomainSpec   `json:"spec,omitempty"`
	Status       DomainStatus `json:"status,omitempty"`
}

type DomainSpecDnsMode string

const (
	DomainSpecDnsModeCname DomainSpecDnsMode = "cname"
	DomainSpecDnsModeNs    DomainSpecDnsMode = "ns"
)

type DomainSpecCertChallengeType string

const (
	DomainSpecCertChallengeTypeHttp01 DomainSpecCertChallengeType = "http01"
	DomainSpecCertChallengeTypeDns01  DomainSpecCertChallengeType = "dns01"
)

type DomainSpec struct {
	DnsMode             DomainSpecDnsMode           `json:"dnsMode,omitempty"`
	GvcLink             string                      `json:"gvcLink,omitempty"`
	CertChallengeType   DomainSpecCertChallengeType `json:"certChallengeType,omitempty"`
	WorkloadLink        string                      `json:"workloadLink,omitempty"`
	AcceptAllHosts      bool                        `json:"acceptAllHosts,omitempty"`
	AcceptAllSubdomains bool                        `json:"acceptAllSubdomains,omitempty"`
	Ports               []ExternalPort              `json:"ports,omitempty"`
}

type DomainStatusEndpoints struct {
	Url          string         `json:"url"`
	WorkloadLink base.LocalLink `json:"workloadLink,omitempty"`
}

type DomainStatusStatus string

const (
	DomainStatusStatusInitializing       DomainStatusStatus = "initializing"
	DomainStatusStatusReady              DomainStatusStatus = "ready"
	DomainStatusStatusPendingDnsConfig   DomainStatusStatus = "pendingDnsConfig"
	DomainStatusStatusPendingCertificate DomainStatusStatus = "pendingCertificate"
	DomainStatusStatusUsedByGvc          DomainStatusStatus = "usedByGvc"
	DomainStatusStatusWarning            DomainStatusStatus = "warning"
	DomainStatusStatusCreated            DomainStatusStatus = "created"
	DomainStatusStatusUpdated            DomainStatusStatus = "updated"
	DomainStatusStatusDeleted            DomainStatusStatus = "deleted"
	DomainStatusStatusErrored            DomainStatusStatus = "errored"
	DomainStatusStatusIgnored            DomainStatusStatus = "ignored"
)

type DomainStatusLocationsCertificateStatus string

const (
	DomainStatusLocationsCertificateStatusInitializing       DomainStatusLocationsCertificateStatus = "initializing"
	DomainStatusLocationsCertificateStatusReady              DomainStatusLocationsCertificateStatus = "ready"
	DomainStatusLocationsCertificateStatusPendingDnsConfig   DomainStatusLocationsCertificateStatus = "pendingDnsConfig"
	DomainStatusLocationsCertificateStatusPendingCertificate DomainStatusLocationsCertificateStatus = "pendingCertificate"
	DomainStatusLocationsCertificateStatusIgnored            DomainStatusLocationsCertificateStatus = "ignored"
)

type DomainStatusLocations struct {
	Name              string                                 `json:"name"`
	CertificateStatus DomainStatusLocationsCertificateStatus `json:"certificateStatus,omitempty"`
}

type DomainStatus struct {
	Endpoints   []DomainStatusEndpoints `json:"endpoints,omitempty"`
	Status      DomainStatusStatus      `json:"status,omitempty"`
	Warning     string                  `json:"warning,omitempty"`
	Locations   []DomainStatusLocations `json:"locations,omitempty"`
	Fingerprint string                  `json:"fingerprint,omitempty"`
	DnsConfig   []DnsConfigRecord       `json:"dnsConfig,omitempty"`
}

type EnvoyHeaderValue string

type ExternalPortProtocol string

const (
	ExternalPortProtocolHttp  ExternalPortProtocol = "http"
	ExternalPortProtocolHttp2 ExternalPortProtocol = "http2"
	ExternalPortProtocolTcp   ExternalPortProtocol = "tcp"
)

type ExternalPortCorsAllowOrigins struct {
	Exact string `json:"exact,omitempty"`
	Regex string `json:"regex,omitempty"`
}

type ExternalPortCors struct {
	AllowOrigins     []ExternalPortCorsAllowOrigins `json:"allowOrigins,omitempty"`
	AllowMethods     []string                       `json:"allowMethods,omitempty"`
	AllowHeaders     []string                       `json:"allowHeaders,omitempty"`
	ExposeHeaders    []string                       `json:"exposeHeaders,omitempty"`
	MaxAge           string                         `json:"maxAge,omitempty"`
	AllowCredentials bool                           `json:"allowCredentials,omitempty"`
}

type ExternalPortTlsMinProtocolVersion string

const (
	ExternalPortTlsMinProtocolVersionTlsv13 ExternalPortTlsMinProtocolVersion = "TLSV1_3"
	ExternalPortTlsMinProtocolVersionTlsv12 ExternalPortTlsMinProtocolVersion = "TLSV1_2"
	ExternalPortTlsMinProtocolVersionTlsv11 ExternalPortTlsMinProtocolVersion = "TLSV1_1"
	ExternalPortTlsMinProtocolVersionTlsv10 ExternalPortTlsMinProtocolVersion = "TLSV1_0"
)

type ExternalPortTlsCipherSuites string

const (
	ExternalPortTlsCipherSuitesEcdheEcdsaAes256GcmSha384             ExternalPortTlsCipherSuites = "ECDHE-ECDSA-AES256-GCM-SHA384"
	ExternalPortTlsCipherSuitesEcdheEcdsaChacha20Poly1305            ExternalPortTlsCipherSuites = "ECDHE-ECDSA-CHACHA20-POLY1305"
	ExternalPortTlsCipherSuitesEcdheEcdsaAes128GcmSha256             ExternalPortTlsCipherSuites = "ECDHE-ECDSA-AES128-GCM-SHA256"
	ExternalPortTlsCipherSuitesEcdheRsaAes256GcmSha384               ExternalPortTlsCipherSuites = "ECDHE-RSA-AES256-GCM-SHA384"
	ExternalPortTlsCipherSuitesEcdheRsaChacha20Poly1305              ExternalPortTlsCipherSuites = "ECDHE-RSA-CHACHA20-POLY1305"
	ExternalPortTlsCipherSuitesEcdheRsaAes128GcmSha256               ExternalPortTlsCipherSuites = "ECDHE-RSA-AES128-GCM-SHA256"
	ExternalPortTlsCipherSuitesAes256GcmSha384                       ExternalPortTlsCipherSuites = "AES256-GCM-SHA384"
	ExternalPortTlsCipherSuitesAes128GcmSha256                       ExternalPortTlsCipherSuites = "AES128-GCM-SHA256"
	ExternalPortTlsCipherSuitesTlsRsaWithAes256GcmSha384             ExternalPortTlsCipherSuites = "TLS_RSA_WITH_AES_256_GCM_SHA384"
	ExternalPortTlsCipherSuitesTlsRsaWithAes128GcmSha256             ExternalPortTlsCipherSuites = "TLS_RSA_WITH_AES_128_GCM_SHA256"
	ExternalPortTlsCipherSuitesTlsEcdheRsaWithChacha20Poly1305Sha256 ExternalPortTlsCipherSuites = "TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256"
	ExternalPortTlsCipherSuitesTlsEcdheRsaWithAes256GcmSha384        ExternalPortTlsCipherSuites = "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384"
	ExternalPortTlsCipherSuitesTlsEcdheRsaWithAes128GcmSha256        ExternalPortTlsCipherSuites = "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"
	ExternalPortTlsCipherSuitesTlsChacha20Poly1305Sha256             ExternalPortTlsCipherSuites = "TLS_CHACHA20_POLY1305_SHA256"
	ExternalPortTlsCipherSuitesTlsAes256GcmSha384                    ExternalPortTlsCipherSuites = "TLS_AES_256_GCM_SHA384"
	ExternalPortTlsCipherSuitesTlsAes128GcmSha256                    ExternalPortTlsCipherSuites = "TLS_AES_128_GCM_SHA256"
	ExternalPortTlsCipherSuitesDesCbc3Sha                            ExternalPortTlsCipherSuites = "DES-CBC3-SHA"
	ExternalPortTlsCipherSuitesEcdheRsaAes128Sha                     ExternalPortTlsCipherSuites = "ECDHE-RSA-AES128-SHA"
	ExternalPortTlsCipherSuitesEcdheRsaAes256Sha                     ExternalPortTlsCipherSuites = "ECDHE-RSA-AES256-SHA"
	ExternalPortTlsCipherSuitesAes128Sha                             ExternalPortTlsCipherSuites = "AES128-SHA"
	ExternalPortTlsCipherSuitesAes256Sha                             ExternalPortTlsCipherSuites = "AES256-SHA"
)

type ExternalPortTlsClientCertificate struct {
	SecretLink string `json:"secretLink,omitempty"`
}

type ExternalPortTlsServerCertificate struct {
	SecretLink string `json:"secretLink,omitempty"`
}

type ExternalPortTls struct {
	MinProtocolVersion ExternalPortTlsMinProtocolVersion `json:"minProtocolVersion,omitempty"`
	CipherSuites       []ExternalPortTlsCipherSuites     `json:"cipherSuites,omitempty"`
	ClientCertificate  ExternalPortTlsClientCertificate  `json:"clientCertificate,omitempty"`
	ServerCertificate  ExternalPortTlsServerCertificate  `json:"serverCertificate,omitempty"`
}

type ExternalPort struct {
	Number   float32              `json:"number"`
	Protocol ExternalPortProtocol `json:"protocol,omitempty"`
	Routes   []Route              `json:"routes,omitempty"`
	Cors     ExternalPortCors     `json:"cors,omitempty"`
	Tls      ExternalPortTls      `json:"tls,omitempty"`
}

type ExternalPortTLS struct {
	MinProtocolVersion ExternalPortTlsMinProtocolVersion `json:"minProtocolVersion,omitempty"`
	CipherSuites       []ExternalPortTlsCipherSuites     `json:"cipherSuites,omitempty"`
	ClientCertificate  ExternalPortTlsClientCertificate  `json:"clientCertificate,omitempty"`
	ServerCertificate  ExternalPortTlsServerCertificate  `json:"serverCertificate,omitempty"`
}

type HeaderOperationSet map[string]EnvoyHeaderValue

type HeaderOperation struct {
	Set HeaderOperationSet `json:"set,omitempty"`
}

type RouteHeaders struct {
	Request HeaderOperation `json:"request,omitempty"`
}

type Route struct {
	ReplacePrefix string       `json:"replacePrefix,omitempty"`
	Regex         string       `json:"regex,omitempty"`
	Prefix        string       `json:"prefix,omitempty"`
	WorkloadLink  string       `json:"workloadLink"`
	Port          float32      `json:"port"`
	HostPrefix    string       `json:"hostPrefix,omitempty"`
	HostRegex     string       `json:"hostRegex,omitempty"`
	Headers       RouteHeaders `json:"headers,omitempty"`
	Replica       float32      `json:"replica"`
}
