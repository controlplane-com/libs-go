package api

import (
	"net/http"
	"time"
)

const (
	// DefaultBaseURL is the default Control Plane API base URL.
	DefaultBaseURL = "https://api.cpln.io"

	// DefaultBillingURL is the default billing service URL.
	DefaultBillingURL = "https://billing-ng.cpln.io"

	// DefaultTimeout is the default HTTP client timeout.
	DefaultTimeout = 30 * time.Second
)

// Client is the Control Plane API client.
type Client struct {
	baseURL    string
	billingURL string
	token      string
	org        string
	userAgent  string
	httpClient *http.Client
	headers    http.Header

	// Services for each resource type (lazily initialized)
	agents          *AgentService
	auditContexts   *AuditContextService
	billing         *BillingService
	cloudAccounts   *CloudAccountService
	deployments     *DeploymentService
	domains         *DomainService
	groups          *GroupService
	gvcs            *GVCService
	identities      *IdentityService
	images          *ImageService
	ipSets          *IPSetService
	locations       *LocationService
	mk8s            *Mk8sService
	orgs            *OrgService
	policies        *PolicyService
	quotas          *QuotaService
	secrets         *SecretService
	serviceAccounts *ServiceAccountService
	tasks           *TaskService
	users           *UserService
	volumeSets      *VolumeSetService
	workloads       *WorkloadService
}

// New creates a new Control Plane API client.
func New(token string, opts ...Option) *Client {
	c := &Client{
		baseURL:    DefaultBaseURL,
		billingURL: DefaultBillingURL,
		token:      token,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// NewWithBaseURL creates a new Control Plane API client with a custom base URL.
func NewWithBaseURL(baseURL, token string, opts ...Option) *Client {
	c := New(token, opts...)
	c.baseURL = baseURL
	return c
}

// SetToken updates the authentication token.
func (c *Client) SetToken(token string) {
	c.token = token
}

// SetOrg updates the default organization.
func (c *Client) SetOrg(org string) {
	c.org = org
}

// Org returns the default organization.
func (c *Client) Org() string {
	return c.org
}

// resolveOrg returns the provided org if non-empty, otherwise the default org.
func (c *Client) resolveOrg(org string) string {
	if org != "" {
		return org
	}
	return c.org
}

// Agents returns the AgentService for managing agents.
func (c *Client) Agents() *AgentService {
	if c.agents == nil {
		c.agents = &AgentService{client: c}
	}
	return c.agents
}

// AuditContexts returns the AuditContextService for managing audit contexts.
func (c *Client) AuditContexts() *AuditContextService {
	if c.auditContexts == nil {
		c.auditContexts = &AuditContextService{client: c}
	}
	return c.auditContexts
}

// Billing returns the BillingService for billing operations via billing-ng.
func (c *Client) Billing() *BillingService {
	if c.billing == nil {
		c.billing = &BillingService{client: c}
	}
	return c.billing
}

// CloudAccounts returns the CloudAccountService for managing cloud accounts.
func (c *Client) CloudAccounts() *CloudAccountService {
	if c.cloudAccounts == nil {
		c.cloudAccounts = &CloudAccountService{client: c}
	}
	return c.cloudAccounts
}

// Deployments returns the DeploymentService for managing deployments.
func (c *Client) Deployments() *DeploymentService {
	if c.deployments == nil {
		c.deployments = &DeploymentService{client: c}
	}
	return c.deployments
}

// Domains returns the DomainService for managing domains.
func (c *Client) Domains() *DomainService {
	if c.domains == nil {
		c.domains = &DomainService{client: c}
	}
	return c.domains
}

// Groups returns the GroupService for managing groups.
func (c *Client) Groups() *GroupService {
	if c.groups == nil {
		c.groups = &GroupService{client: c}
	}
	return c.groups
}

// GVCs returns the GVCService for managing GVCs.
func (c *Client) GVCs() *GVCService {
	if c.gvcs == nil {
		c.gvcs = &GVCService{client: c}
	}
	return c.gvcs
}

// Identities returns the IdentityService for managing identities.
func (c *Client) Identities() *IdentityService {
	if c.identities == nil {
		c.identities = &IdentityService{client: c}
	}
	return c.identities
}

// Images returns the ImageService for managing images.
func (c *Client) Images() *ImageService {
	if c.images == nil {
		c.images = &ImageService{client: c}
	}
	return c.images
}

// IPSets returns the IPSetService for managing IP sets.
func (c *Client) IPSets() *IPSetService {
	if c.ipSets == nil {
		c.ipSets = &IPSetService{client: c}
	}
	return c.ipSets
}

// Locations returns the LocationService for managing locations.
func (c *Client) Locations() *LocationService {
	if c.locations == nil {
		c.locations = &LocationService{client: c}
	}
	return c.locations
}

// Mk8s returns the Mk8sService for managing managed Kubernetes clusters.
func (c *Client) Mk8s() *Mk8sService {
	if c.mk8s == nil {
		c.mk8s = &Mk8sService{client: c}
	}
	return c.mk8s
}

// Orgs returns the OrgService for managing organizations.
func (c *Client) Orgs() *OrgService {
	if c.orgs == nil {
		c.orgs = &OrgService{client: c}
	}
	return c.orgs
}

// Policies returns the PolicyService for managing policies.
func (c *Client) Policies() *PolicyService {
	if c.policies == nil {
		c.policies = &PolicyService{client: c}
	}
	return c.policies
}

// Quotas returns the QuotaService for managing quotas.
func (c *Client) Quotas() *QuotaService {
	if c.quotas == nil {
		c.quotas = &QuotaService{client: c}
	}
	return c.quotas
}

// Secrets returns the SecretService for managing secrets.
func (c *Client) Secrets() *SecretService {
	if c.secrets == nil {
		c.secrets = &SecretService{client: c}
	}
	return c.secrets
}

// ServiceAccounts returns the ServiceAccountService for managing service accounts.
func (c *Client) ServiceAccounts() *ServiceAccountService {
	if c.serviceAccounts == nil {
		c.serviceAccounts = &ServiceAccountService{client: c}
	}
	return c.serviceAccounts
}

// Tasks returns the TaskService for managing tasks.
func (c *Client) Tasks() *TaskService {
	if c.tasks == nil {
		c.tasks = &TaskService{client: c}
	}
	return c.tasks
}

// Users returns the UserService for managing users.
func (c *Client) Users() *UserService {
	if c.users == nil {
		c.users = &UserService{client: c}
	}
	return c.users
}

// VolumeSets returns the VolumeSetService for managing volume sets.
func (c *Client) VolumeSets() *VolumeSetService {
	if c.volumeSets == nil {
		c.volumeSets = &VolumeSetService{client: c}
	}
	return c.volumeSets
}

// Workloads returns the WorkloadService for managing workloads.
func (c *Client) Workloads() *WorkloadService {
	if c.workloads == nil {
		c.workloads = &WorkloadService{client: c}
	}
	return c.workloads
}
