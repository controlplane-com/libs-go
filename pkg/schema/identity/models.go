/* auto-generated */

package identity

import "github.com/controlplane-com/libs-go/pkg/schema/base"

type AwsIdentity struct {
	CloudAccountLink string            `json:"cloudAccountLink"`
	PolicyRefs       []PolicyRef       `json:"policyRefs,omitempty"`
	TrustPolicy      AwsPolicyDocument `json:"trustPolicy,omitempty"`
	RoleName         AwsRoleName       `json:"roleName,omitempty"`
}

type AwsPolicyDocumentStatement map[string]any

type AwsPolicyDocument struct {
	Version   string                       `json:"version,omitempty"`
	Statement []AwsPolicyDocumentStatement `json:"statement,omitempty"`
}

type AwsRoleName string

type AzureIdentity struct {
	CloudAccountLink string                `json:"cloudAccountLink"`
	RoleAssignments  []AzureRoleAssignment `json:"roleAssignments,omitempty"`
}

type AzureRoleAssignment struct {
	Scope string   `json:"scope,omitempty"`
	Roles []string `json:"roles,omitempty"`
}

type GcpIdentityBindings struct {
	Resource string        `json:"resource,omitempty"`
	Roles    []GcpRoleName `json:"roles,omitempty"`
}

type GcpIdentity struct {
	CloudAccountLink string                `json:"cloudAccountLink"`
	Scopes           []string              `json:"scopes,omitempty"`
	ServiceAccount   string                `json:"serviceAccount,omitempty"`
	Bindings         []GcpIdentityBindings `json:"bindings,omitempty"`
}

type GcpRoleName string

type IdentityTags map[string]any

type Identity struct {
	Id                     string                  `json:"id,omitempty"`
	Name                   base.Name               `json:"name,omitempty"`
	Kind                   base.Kind               `json:"kind,omitempty"`
	Version                *float32                `json:"version,omitempty"`
	Description            string                  `json:"description,omitempty"`
	Tags                   IdentityTags            `json:"tags,omitempty"`
	Created                string                  `json:"created,omitempty"`
	LastModified           string                  `json:"lastModified,omitempty"`
	Links                  base.Links              `json:"links,omitempty"`
	Aws                    *AwsIdentity            `json:"aws,omitempty"`
	Gcp                    *GcpIdentity            `json:"gcp,omitempty"`
	Azure                  *AzureIdentity          `json:"azure,omitempty"`
	Ngs                    *NgsIdentity            `json:"ngs,omitempty"`
	NetworkResources       []NetworkResource       `json:"networkResources,omitempty"`
	NativeNetworkResources []NativeNetworkResource `json:"nativeNetworkResources,omitempty"`
	MemcacheAccess         []MemcacheAccess        `json:"memcacheAccess,omitempty"`
	SpicedbAccess          []SpicedbAccess         `json:"spicedbAccess,omitempty"`
	Status                 Status                  `json:"status,omitempty"`
	Gvc                    string                  `json:"gvc,omitempty"`
}

type MemcacheAccessRole string

const (
	MemcacheAccessRoleReadWrite MemcacheAccessRole = "readWrite"
)

type MemcacheAccess struct {
	ClusterLink string             `json:"clusterLink"`
	Role        MemcacheAccessRole `json:"role,omitempty"`
}

type NativeNetworkResourceAwsPrivateLink struct {
	EndpointServiceName string `json:"endpointServiceName"`
}

type NativeNetworkResourceGcpServiceConnect struct {
	TargetService string `json:"targetService"`
}

type NativeNetworkResource struct {
	Name              any/* TODO: [object Object]*/ `json:"name"`
	FQDN              string                                 `json:"fQDN,omitempty"`
	Ports             []float32                              `json:"ports"`
	AwsPrivateLink    NativeNetworkResourceAwsPrivateLink    `json:"awsPrivateLink,omitempty"`
	GcpServiceConnect NativeNetworkResourceGcpServiceConnect `json:"gcpServiceConnect,omitempty"`
}

type NetworkResource struct {
	Name       any/* TODO: [object Object]*/ `json:"name"`
	AgentLink  string    `json:"agentLink,omitempty"`
	IPs        []string  `json:"iPs,omitempty"`
	FQDN       string    `json:"fQDN,omitempty"`
	ResolverIP string    `json:"resolverIP,omitempty"`
	Ports      []float32 `json:"ports"`
}

type NgsIdentityPub struct {
	Allow []string `json:"allow,omitempty"`
	Deny  []string `json:"deny,omitempty"`
}

type NgsIdentitySub struct {
	Allow []string `json:"allow,omitempty"`
	Deny  []string `json:"deny,omitempty"`
}

type NgsIdentityResp struct {
	Max *float32 `json:"max,omitempty"`
	Ttl string   `json:"ttl,omitempty"`
}

type NgsIdentity struct {
	CloudAccountLink string          `json:"cloudAccountLink"`
	Pub              NgsIdentityPub  `json:"pub,omitempty"`
	Sub              NgsIdentitySub  `json:"sub,omitempty"`
	Resp             NgsIdentityResp `json:"resp,omitempty"`
	Subs             *float32        `json:"subs,omitempty"`
	Data             *float32        `json:"data,omitempty"`
	Payload          *float32        `json:"payload,omitempty"`
}

type PolicyRef string

type SpicedbAccessRole string

const (
	SpicedbAccessRoleCheckPermission SpicedbAccessRole = "checkPermission"
	SpicedbAccessRoleRead            SpicedbAccessRole = "read"
	SpicedbAccessRoleWrite           SpicedbAccessRole = "write"
)

type SpicedbAccess struct {
	ClusterLink string            `json:"clusterLink"`
	Role        SpicedbAccessRole `json:"role,omitempty"`
}

type StatusAws struct {
	LastError string `json:"lastError,omitempty"`
	Usable    bool   `json:"usable,omitempty"`
}

type StatusGcp struct {
	LastError string `json:"lastError,omitempty"`
	Usable    bool   `json:"usable,omitempty"`
}

type StatusAzure struct {
	LastError string `json:"lastError,omitempty"`
	Usable    bool   `json:"usable,omitempty"`
}

type Status struct {
	ObjectName string      `json:"objectName,omitempty"`
	Aws        StatusAws   `json:"aws,omitempty"`
	Gcp        StatusGcp   `json:"gcp,omitempty"`
	Azure      StatusAzure `json:"azure,omitempty"`
}
