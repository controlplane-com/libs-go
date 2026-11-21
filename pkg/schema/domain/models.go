/* hand-maintained — canonical source of truth for the `domain` schema.
 * `validate:` tags drive pkg/validate.FromType reflection (replaces most of the
 * schema_domain.go builder). See nodelibs/schema/src/domain.ts.
 *
 * Route (+ RouteMirror/RouteCanary/RouteHeaders/HeaderOperation) is NOT tagged: its
 * `prefix.when('regex', {not:exist, then:default '/'})` is a field-level default
 * that must apply BEFORE the object .xor('prefix','regex') — a hook (which runs AFTER
 * xor) cannot express it — so Route stays a prebuilt rule registered via RegisterType
 * (RouteRule() in schema_domain.go), exactly like the mk8s provider bodies. The
 * remaining structs reflect from these tags; the many .custom()s (externalPortCustom,
 * domainSpecCustom, maxRoutes, domainName, uniqueRoute) are Refines / array customs and
 * the .xor/.nand are RegisterObjectRules (see schema_reflect_datanet.go). */

package domain

import "github.com/controlplane-com/libs-go/pkg/schema/base"

// DomainTags aliases base.Tags for the reflection base-type registry.
type DomainTags = base.Tags

// LowerHeader is the constrained element type for cors allow/expose headers
// (String().lowercase()); registered in schema_reflect_datanet.go.
type LowerHeader string

type CertChallengeType string

const (
	CertChallengeTypeHttp01 CertChallengeType = "http01"
	CertChallengeTypeDns01  CertChallengeType = "dns01"
)

type DnsConfigRecord struct {
	Type  string   `json:"type,omitzero"`
	Ttl   *float64 `json:"ttl,omitzero" validate:"integer,positive"`
	Host  string   `json:"host,omitzero"`
	Value string   `json:"value,omitzero"`
}

type Domain struct {
	Id           string        `json:"id,omitzero" validate:"uuid,readOnly"`
	Kind         base.Kind     `json:"kind,omitzero"`
	Version      *float64      `json:"version,omitzero" validate:"positive,integer,readOnly"`
	Description  string        `json:"description,omitzero" validate:"trim,max=250,allow="`
	Tags         DomainTags    `json:"tags,omitzero" validate:"default={}"`
	Created      string        `json:"created,omitzero" validate:"date,readOnly"`
	LastModified string        `json:"lastModified,omitzero" validate:"date,readOnly"`
	Links        base.Links    `json:"links,omitzero" validate:"readOnly"`
	Name         string        `json:"name,omitzero" validate:"max=253"`
	Spec         *DomainSpec   `json:"spec,omitzero"`
	Status       *DomainStatus `json:"status,omitzero"`
}

type DomainSpecDnsMode string

const (
	DomainSpecDnsModeCname DomainSpecDnsMode = "cname"
	DomainSpecDnsModeNs    DomainSpecDnsMode = "ns"
)

func (DomainSpecDnsMode) EnumValues() []string { return []string{"cname", "ns"} }

type DomainSpecCertChallengeType string

const (
	DomainSpecCertChallengeTypeHttp01 DomainSpecCertChallengeType = "http01"
	DomainSpecCertChallengeTypeDns01  DomainSpecCertChallengeType = "dns01"
)

func (DomainSpecCertChallengeType) EnumValues() []string { return []string{"http01", "dns01"} }

// DomainSpec ports `DomainSpec` (domain.ts:297). nand(gvcLink, workloadLink) is a
// RegisterObjectRules; the big cross-field .custom() is the domainSpecCustom Refine;
// ports' content-ful default ([getDefaultExternalPort()]) is applied by that Refine.
type DomainSpec struct {
	DnsMode                  DomainSpecDnsMode           `json:"dnsMode,omitzero"`
	GvcLink                  GvcLink                     `json:"gvcLink,omitzero"`
	CertChallengeType        DomainSpecCertChallengeType `json:"certChallengeType,omitzero"`
	CertificateGeneratorLink CertificateGeneratorLink    `json:"certificateGeneratorLink,omitzero" validate:"optional"`
	WorkloadLink             WorkloadLink                `json:"workloadLink,omitzero"`
	AcceptAllHosts           *bool                       `json:"acceptAllHosts,omitzero" validate:"default=false"`
	AcceptAllSubdomains      *bool                       `json:"acceptAllSubdomains,omitzero" validate:"default=false"`
	Ports                    []ExternalPort              `json:"ports,omitzero" validate:"min=0,max=10"`
}

type DomainStatusEndpoints struct {
	Url          string         `json:"url" validate:"required"`
	WorkloadLink base.LocalLink `json:"workloadLink,omitzero"`
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

func (DomainStatusStatus) EnumValues() []string {
	return []string{"initializing", "ready", "pendingDnsConfig", "pendingCertificate",
		"usedByGvc", "warning", "created", "updated", "deleted", "errored", "ignored"}
}

type DomainStatusLocationsCertificateStatus string

const (
	DomainStatusLocationsCertificateStatusInitializing       DomainStatusLocationsCertificateStatus = "initializing"
	DomainStatusLocationsCertificateStatusReady              DomainStatusLocationsCertificateStatus = "ready"
	DomainStatusLocationsCertificateStatusPendingDnsConfig   DomainStatusLocationsCertificateStatus = "pendingDnsConfig"
	DomainStatusLocationsCertificateStatusPendingCertificate DomainStatusLocationsCertificateStatus = "pendingCertificate"
	DomainStatusLocationsCertificateStatusIgnored            DomainStatusLocationsCertificateStatus = "ignored"
	DomainStatusLocationsCertificateStatusErrored            DomainStatusLocationsCertificateStatus = "errored"
)

func (DomainStatusLocationsCertificateStatus) EnumValues() []string {
	return []string{"initializing", "ready", "pendingDnsConfig", "pendingCertificate", "ignored", "errored"}
}

type DomainStatusLocations struct {
	Name              string                                 `json:"name" validate:"required"`
	CertificateStatus DomainStatusLocationsCertificateStatus `json:"certificateStatus,omitzero"`
}

type DomainStatus struct {
	Endpoints     []DomainStatusEndpoints `json:"endpoints,omitzero"`
	Status        DomainStatusStatus      `json:"status,omitzero"`
	Warning       string                  `json:"warning,omitzero" validate:"allow="`
	StatusUpdated string                  `json:"statusUpdated,omitzero" validate:"date"`
	Locations     []DomainStatusLocations `json:"locations,omitzero"`
	Fingerprint   string                  `json:"fingerprint,omitzero"`
	DnsConfig     []DnsConfigRecord       `json:"dnsConfig,omitzero"`
}

type EnvoyHeaderValue string //ts:export

type ExternalPortProtocol string

const (
	ExternalPortProtocolHttp  ExternalPortProtocol = "http"
	ExternalPortProtocolHttp2 ExternalPortProtocol = "http2"
	ExternalPortProtocolTcp   ExternalPortProtocol = "tcp"
)

func (ExternalPortProtocol) EnumValues() []string { return []string{"http", "http2", "tcp"} }

type ExternalPortCorsAllowOrigins struct {
	Exact string     `json:"exact,omitzero"`
	Regex base.Regex `json:"regex,omitzero"`
}

type ExternalPortCors struct {
	AllowOrigins     []ExternalPortCorsAllowOrigins `json:"allowOrigins,omitzero"`
	AllowMethods     []string                       `json:"allowMethods,omitzero"`
	AllowHeaders     []LowerHeader                  `json:"allowHeaders,omitzero"`
	ExposeHeaders    []LowerHeader                  `json:"exposeHeaders,omitzero"`
	MaxAge           CorsMaxAge                     `json:"maxAge,omitzero"`
	AllowCredentials *bool                          `json:"allowCredentials,omitzero"`
}

type ExternalPortTlsMinProtocolVersion string

const (
	ExternalPortTlsMinProtocolVersionTlsv13 ExternalPortTlsMinProtocolVersion = "TLSV1_3"
	ExternalPortTlsMinProtocolVersionTlsv12 ExternalPortTlsMinProtocolVersion = "TLSV1_2"
	ExternalPortTlsMinProtocolVersionTlsv11 ExternalPortTlsMinProtocolVersion = "TLSV1_1"
	ExternalPortTlsMinProtocolVersionTlsv10 ExternalPortTlsMinProtocolVersion = "TLSV1_0"
)

func (ExternalPortTlsMinProtocolVersion) EnumValues() []string {
	return []string{"TLSV1_3", "TLSV1_2", "TLSV1_1", "TLSV1_0"}
}

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

// EnumValues = allowedCiphers (domain.ts:135): the defaultCiphers plus the legacy
// names the cipher-migration custom rewrites.
func (ExternalPortTlsCipherSuites) EnumValues() []string {
	return []string{
		"ECDHE-ECDSA-AES256-GCM-SHA384", "ECDHE-ECDSA-CHACHA20-POLY1305", "ECDHE-ECDSA-AES128-GCM-SHA256",
		"ECDHE-RSA-AES256-GCM-SHA384", "ECDHE-RSA-CHACHA20-POLY1305", "ECDHE-RSA-AES128-GCM-SHA256",
		"AES256-GCM-SHA384", "AES128-GCM-SHA256",
		"TLS_RSA_WITH_AES_256_GCM_SHA384", "TLS_RSA_WITH_AES_128_GCM_SHA256",
		"TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256", "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
		"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256", "TLS_CHACHA20_POLY1305_SHA256",
		"TLS_AES_256_GCM_SHA384", "TLS_AES_128_GCM_SHA256",
		"DES-CBC3-SHA", "ECDHE-RSA-AES128-SHA", "ECDHE-RSA-AES256-SHA", "AES128-SHA", "AES256-SHA",
	}
}

type ExternalPortTlsClientCertificate struct {
	SecretLink SecretLink `json:"secretLink,omitzero"`
}

type ExternalPortTlsServerCertificate struct {
	SecretLink SecretLink `json:"secretLink,omitzero"`
}

type ExternalPortTls struct {
	MinProtocolVersion ExternalPortTlsMinProtocolVersion `json:"minProtocolVersion,omitzero" validate:"default=TLSV1_2"`
	CipherSuites       []ExternalPortTlsCipherSuites     `json:"cipherSuites,omitzero"`
	ClientCertificate  *ExternalPortTlsClientCertificate `json:"clientCertificate,omitzero"`
	ServerCertificate  *ExternalPortTlsServerCertificate `json:"serverCertificate,omitzero"`
}

// ExternalPort ports `ExternalPort` (domain.ts:208). routes uses the prebuilt Route
// rule (RegisterType) + the uniqueRoute array custom; the port-443 TLS-default /
// route-sort / tcp-canary .custom() is the externalPortCustom Refine.
type ExternalPort struct {
	Number   *float64             `json:"number,omitzero" validate:"default=443"`
	Protocol ExternalPortProtocol `json:"protocol,omitzero" validate:"default=http2"`
	Routes   Routes               `json:"routes,omitzero" validate:"min=0"`
	Cors     *ExternalPortCors    `json:"cors,omitzero"`
	Tls      *ExternalPortTls     `json:"tls,omitzero"`
}

type ExternalPortTLS struct {
	MinProtocolVersion ExternalPortTlsMinProtocolVersion `json:"minProtocolVersion,omitzero"`
	CipherSuites       []ExternalPortTlsCipherSuites     `json:"cipherSuites,omitzero"`
	ClientCertificate  *ExternalPortTlsClientCertificate `json:"clientCertificate,omitzero"`
	ServerCertificate  *ExternalPortTlsServerCertificate `json:"serverCertificate,omitzero"`
}

type HeaderOperationSet map[string]EnvoyHeaderValue

type HeaderOperation struct {
	Set HeaderOperationSet `json:"set,omitzero"`
}

type RouteHeaders struct {
	Request *HeaderOperation `json:"request,omitzero"`
}

type RouteMirror struct {
	WorkloadLink string   `json:"workloadLink"`
	Port         *float64 `json:"port,omitzero"`
	Percent      float64  `json:"percent"`
} //ts:export

type Route struct {
	ReplacePrefix   string        `json:"replacePrefix,omitzero"`
	Regex           string        `json:"regex,omitzero"`
	Prefix          string        `json:"prefix,omitzero"`
	CaseInsensitive *bool         `json:"caseInsensitive,omitzero"`
	WorkloadLink    string        `json:"workloadLink"`
	Port            *float64      `json:"port,omitzero"`
	HostPrefix      string        `json:"hostPrefix,omitzero"`
	HostRegex       string        `json:"hostRegex,omitzero"`
	Headers         *RouteHeaders `json:"headers,omitzero"`
	Replica         *float64      `json:"replica,omitzero"`
	Mirror          []RouteMirror `json:"mirror,omitzero"`
	Canaries        []RouteCanary `json:"canaries,omitzero"`
}

type RouteCanary struct {
	WorkloadLink string   `json:"workloadLink"`
	Port         *float64 `json:"port,omitzero"`
	Weight       float64  `json:"weight"`
}

// --- validation leaf & collection types (ValidationRule methods live in validate.go) ---

// Routes is ExternalPort.routes: an array of the prebuilt Route rule plus the
// uniqueRoute array custom (was custom=uniqueRoute on the field). Replaces the
// []Route field type.
type Routes []Route

// CorsMaxAge is ExternalPortCors.maxAge restricted to the corsMaxAge pattern
// (was pattern=corsMaxAge). Replaces the string field type.
type CorsMaxAge string

// GvcLink is an org-local link constrained to a "gvc" — DomainSpec.gvcLink
// (domain.ts). Replaces the `localLink=gvc` tag.
type GvcLink string

// WorkloadLink is an org-local link constrained to a workload via the `!workload`
// expansion — DomainSpec.workloadLink (domain.ts). Replaces the
// `localLink=!workload` tag.
type WorkloadLink string

// CertificateGeneratorLink is an org-local link to a certificategenerator —
// DomainSpec.certificateGeneratorLink (domain.ts:304, master 0c0ece779).
type CertificateGeneratorLink string

// SecretLink is an org-local link constrained to a "secret" — the TLS client/server
// certificate secretLink (domain.ts). Replaces the `localLink=secret` tag.
type SecretLink string
