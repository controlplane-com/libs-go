/* auto-generated */

package discovery

type DiscoveryEndpoints struct {
	Billing           string `json:"billing,omitempty"`
	Console           string `json:"console,omitempty"`
	Browser           string `json:"browser,omitempty"`
	Hub               string `json:"hub,omitempty"`
	Registry          string `json:"registry,omitempty"`
	Audit             string `json:"audit,omitempty"`
	Logs              string `json:"logs,omitempty"`
	Metrics           string `json:"metrics,omitempty"`
	MetricsNg         string `json:"metrics-ng,omitempty"`
	Tracing           string `json:"tracing,omitempty"`
	Byok              string `json:"byok,omitempty"`
	Metering          string `json:"metering,omitempty"`
	BillingNg         string `json:"billing-ng,omitempty"`
	Grafana           string `json:"grafana,omitempty"`
	TerraformExporter string `json:"terraform-exporter,omitempty"`
	K8sCrdExporter    string `json:"k8s-crd-exporter,omitempty"`
	Mcp               string `json:"mcp,omitempty"`
	Auth              string `json:"auth,omitempty"`
}

type DiscoveryFirebase struct {
	ApiKey     string `json:"apiKey,omitempty"`
	AuthDomain string `json:"authDomain,omitempty"`
	ProjectId  string `json:"projectId,omitempty"`
}

type Discovery struct {
	Endpoints DiscoveryEndpoints `json:"endpoints,omitempty"`
	Firebase  DiscoveryFirebase  `json:"firebase,omitempty"`
}
