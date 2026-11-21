/* hand-maintained — canonical source of truth for the `identity` resource schema.
 * Was joi-to-go-generated from nodelibs/schema/src/identity.ts; now the canonical
 * definition. `validate:` struct tags carry the rules; pkg/validate.FromType reflects
 * over these types. Complex parts (xor/oxor, per-array Name/FQDN uniqueness, the gcp
 * scopes content-default) and the nominal leaf types (NetworkName/Ipv4Addr/
 * NetworkPort/NatsSubject and the pattern=/custom= tag replacements) carry their own
 * rules via the Validator interface in validate.go. See identity.ts for the Joi source.
 */

package identity

import "github.com/controlplane-com/libs-go/pkg/schema/base"

// --- named leaf types (element-level rules; see validate.go for their rules) ---

// AwsRoleName ports `AwsRoleName` (identity.ts:6).
type AwsRoleName string //ts:export

// PolicyRef ports `PolicyRef` (identity.ts:11).
type PolicyRef string //ts:export

// GcpRoleName ports `GcpRoleName` (identity.ts:15).
type GcpRoleName string //ts:export

// NetworkName ports `Joi.alternatives(Name, Joi.string().domain())` — a network
// resource name (a Name OR a domain).
type NetworkName string

// Ipv4Addr ports `Joi.string().ip({ version: 'ipv4' })` (optional CIDR).
type Ipv4Addr string

// NetworkPort ports `Joi.number().integer().min(0).max(65535)`.
type NetworkPort float64

// NatsSubject ports `Joi.string().trim().custom(validSubject)` — a valid NATS subject.
type NatsSubject string

// --- pattern=/custom= tag sites, now strongly-typed nominal types (rules in validate.go) ---

// GserviceAccount is the gcp service-account email (identity.ts:44) — the
// strongly-typed replacement for the `pattern=gserviceaccount` tag; the field keeps
// its trim/email tags.
type GserviceAccount string

// NgsTtl is the ngs response TTL (identity.ts:106) — replaces the `pattern=ngsTtl`
// tag.
type NgsTtl string

// GcpTargetService is a gcp serviceAttachment target (identity.ts:177) — replaces
// the `pattern=gcpServiceConnect,custom=gcpSvcConnTrim` tags; the leading/trailing
// slashes are trimmed after the pattern check. The field keeps its `required` tag.
type GcpTargetService string

// --- nominal link types (the `localLink=<kind>` tag sites; rules in validate.go) ---

// CloudAccountLink is an org-local link constrained to a "cloudaccount" — the
// aws/azure/gcp/ngs cloudAccountLink (identity.ts). The field keeps its `required`
// tag.
type CloudAccountLink string

// AgentLink is an org-local link constrained to an "agent" — networkResources[].
// agentLink (identity.ts).
type AgentLink string

// MemcacheClusterLink is an org-local link constrained to a "memcachecluster" —
// memcacheAccess[].clusterLink (identity.ts). The field keeps its `required` tag.
type MemcacheClusterLink string

// SpicedbClusterLink is an org-local link constrained to a "spicedbcluster" —
// spicedbAccess[].clusterLink (identity.ts). The field keeps its `required` tag.
type SpicedbClusterLink string

// --- collection nominal slice types (the `custom=dedupeEmpty` array sites; rules in validate.go) ---

// NatsSubjectList is `Joi.array().items(NatsSubject).custom(makeUnique())`.
type NatsSubjectList []NatsSubject

// Ipv4AddrList is `Joi.array().items(Ipv4Addr).custom(makeUnique())`; the field
// keeps its min/max tags.
type Ipv4AddrList []Ipv4Addr

// NetworkPortList is `Joi.array().items(NetworkPort).custom(makeUnique())`; the
// field keeps its min/max/sort/required tags.
type NetworkPortList []NetworkPort

// --- aws ---

type AwsIdentity struct {
	CloudAccountLink CloudAccountLink   `json:"cloudAccountLink" validate:"required"`
	PolicyRefs       []PolicyRef        `json:"policyRefs,omitzero"`
	TrustPolicy      *AwsPolicyDocument `json:"trustPolicy,omitzero"`
	RoleName         AwsRoleName        `json:"roleName,omitzero"`
}

type AwsPolicyDocumentStatement map[string]any

// AwsPolicyDocument keeps the AWS-canonical capitalized keys `Version`/`Statement`
// (identity.ts:18) rather than the joi-to-go lowercased forms.
type AwsPolicyDocument struct {
	Version   string                       `json:"Version,omitzero" validate:"default=2012-10-17,optional"`
	Statement []AwsPolicyDocumentStatement `json:"Statement,omitzero"`
}

// --- azure ---

type AzureIdentity struct {
	CloudAccountLink CloudAccountLink      `json:"cloudAccountLink" validate:"required"`
	RoleAssignments  []AzureRoleAssignment `json:"roleAssignments,omitzero"`
}

type AzureRoleAssignment struct {
	Scope string   `json:"scope,omitzero"`
	Roles []string `json:"roles,omitzero" validate:"min=1"`
}

// --- gcp ---

type GcpCondition struct {
	Title       string `json:"title"`
	Description string `json:"description,omitzero"`
	Expression  string `json:"expression"`
}

type GcpIdentityBindings struct {
	Resource string        `json:"resource,omitzero"`
	Roles    []GcpRoleName `json:"roles,omitzero" validate:"min=1"`
	Condition *GcpCondition `json:"condition,omitzero"`
}

type GcpIdentity struct {
	CloudAccountLink CloudAccountLink      `json:"cloudAccountLink" validate:"required"`
	Scopes           []string              `json:"scopes,omitzero"`
	ServiceAccount   GserviceAccount       `json:"serviceAccount,omitzero" validate:"trim,email"`
	Bindings         []GcpIdentityBindings `json:"bindings,omitzero"`
}

// --- ngs ---

type NgsIdentityPub struct {
	Allow NatsSubjectList `json:"allow,omitzero"`
	Deny  NatsSubjectList `json:"deny,omitzero"`
}

type NgsIdentitySub struct {
	Allow NatsSubjectList `json:"allow,omitzero"`
	Deny  NatsSubjectList `json:"deny,omitzero"`
}

type NgsIdentityResp struct {
	Max *float64 `json:"max,omitzero" validate:"integer,min=-1,default=1"`
	Ttl NgsTtl   `json:"ttl,omitzero"`
}

type NgsIdentity struct {
	CloudAccountLink CloudAccountLink `json:"cloudAccountLink" validate:"required"`
	Pub              *NgsIdentityPub  `json:"pub,omitzero"`
	Sub              *NgsIdentitySub  `json:"sub,omitzero"`
	Resp             *NgsIdentityResp `json:"resp,omitzero"`
	Subs             *float64         `json:"subs,omitzero" validate:"integer,min=-1,default=-1"`
	Data             *float64         `json:"data,omitzero" validate:"integer,min=-1,default=-1"`
	Payload          *float64         `json:"payload,omitzero" validate:"integer,min=-1,default=-1"`
}

// --- network resources ---

type NetworkResource struct {
	Name       NetworkName     `json:"name" validate:"required"`
	AgentLink  AgentLink       `json:"agentLink,omitzero"`
	IPs        Ipv4AddrList    `json:"IPs,omitzero" validate:"min=1,max=5"`
	FQDN       string          `json:"FQDN,omitzero" validate:"domain,lowercase"`
	ResolverIP Ipv4Addr        `json:"resolverIP,omitzero"`
	Ports      NetworkPortList `json:"ports" validate:"min=1,max=10,sort,required"`
}

type NativeNetworkResourceAwsPrivateLink struct {
	EndpointServiceName string `json:"endpointServiceName" validate:"required"`
}

type NativeNetworkResourceGcpServiceConnect struct {
	TargetService GcpTargetService `json:"targetService" validate:"required"`
}

type NativeNetworkResource struct {
	Name              NetworkName                             `json:"name" validate:"required"`
	FQDN              string                                  `json:"FQDN,omitzero" validate:"domain,lowercase"`
	Ports             NetworkPortList                         `json:"ports" validate:"min=1,max=10,sort,required"`
	AwsPrivateLink    *NativeNetworkResourceAwsPrivateLink    `json:"awsPrivateLink,omitzero"`
	GcpServiceConnect *NativeNetworkResourceGcpServiceConnect `json:"gcpServiceConnect,omitzero"`
}

// --- access ---

type MemcacheAccessRole string

const (
	MemcacheAccessRoleReadWrite MemcacheAccessRole = "readWrite"
)

func (MemcacheAccessRole) EnumValues() []string {
	return []string{string(MemcacheAccessRoleReadWrite)}
}

type MemcacheAccess struct {
	ClusterLink MemcacheClusterLink `json:"clusterLink" validate:"required"`
	Role        MemcacheAccessRole  `json:"role,omitzero" validate:"required"`
}

type SpicedbAccessRole string

const (
	SpicedbAccessRoleCheckPermission SpicedbAccessRole = "checkPermission"
	SpicedbAccessRoleRead            SpicedbAccessRole = "read"
	SpicedbAccessRoleWrite           SpicedbAccessRole = "write"
)

func (SpicedbAccessRole) EnumValues() []string {
	return []string{
		string(SpicedbAccessRoleCheckPermission),
		string(SpicedbAccessRoleRead),
		string(SpicedbAccessRoleWrite),
	}
}

type SpicedbAccess struct {
	ClusterLink SpicedbClusterLink `json:"clusterLink" validate:"required"`
	Role        SpicedbAccessRole  `json:"role,omitzero" validate:"required"`
}

// --- status (server-populated) ---

type StatusAws struct {
	LastError string `json:"lastError,omitzero"`
	Usable    *bool  `json:"usable,omitzero"`
}

type StatusGcp struct {
	LastError string `json:"lastError,omitzero"`
	Usable    *bool  `json:"usable,omitzero"`
}

type StatusAzure struct {
	LastError string `json:"lastError,omitzero"`
	Usable    *bool  `json:"usable,omitzero"`
}

type Status struct {
	ObjectName string       `json:"objectName,omitzero"`
	Aws        *StatusAws   `json:"aws,omitzero"`
	Gcp        *StatusGcp   `json:"gcp,omitzero"`
	Azure      *StatusAzure `json:"azure,omitzero"`
}

// --- top-level Identity (Base.keys) ---

// IdentityTags aliases base.Tags so the reflection base-type registry resolves it
// to the shared TagsRule.
type IdentityTags = base.Tags

type Identity struct {
	Id                     string                  `json:"id,omitzero" validate:"uuid,readOnly"`
	Name                   base.Name               `json:"name,omitzero"`
	Kind                   base.Kind               `json:"kind,omitzero"`
	Version                *float64                `json:"version,omitzero" validate:"positive,integer,readOnly"`
	Description            string                  `json:"description,omitzero" validate:"trim,max=250,allow="`
	Tags                   IdentityTags            `json:"tags,omitzero" validate:"default={}"`
	Created                string                  `json:"created,omitzero" validate:"date,readOnly"`
	LastModified           string                  `json:"lastModified,omitzero" validate:"date,readOnly"`
	Links                  base.Links              `json:"links,omitzero" validate:"readOnly"`
	Aws                    *AwsIdentity            `json:"aws,omitzero"`
	Gcp                    *GcpIdentity            `json:"gcp,omitzero"`
	Azure                  *AzureIdentity          `json:"azure,omitzero"`
	Ngs                    *NgsIdentity            `json:"ngs,omitzero"`
	NetworkResources       []NetworkResource       `json:"networkResources,omitzero" validate:"uniqueBy=name,max=50"`
	NativeNetworkResources []NativeNetworkResource `json:"nativeNetworkResources,omitzero" validate:"uniqueBy=name,max=50"`
	MemcacheAccess         []MemcacheAccess        `json:"memcacheAccess,omitzero" validate:"uniqueBy=clusterLink,max=5"`
	SpicedbAccess          []SpicedbAccess         `json:"spicedbAccess,omitzero" validate:"uniqueBy=clusterLink,max=5"`
	Status                 *Status                 `json:"status,omitzero"`
	Gvc                    string                  `json:"gvc,omitzero" validate:"optional,allow="`
}
