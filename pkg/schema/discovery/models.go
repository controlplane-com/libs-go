/* hand-maintained — canonical source of truth (nodelibs/schema/src/discovery.ts).
 * `validate:` struct tags drive pkg/validate.FromType; DSL on go-libs/schema/ipSet/models.go.
 * Only five of the endpoints are `Joi.string().uri()` — the rest are plain strings. */

package discovery

type DiscoveryEndpoints struct {
	Billing           string `json:"billing,omitzero" validate:"uri"`
	Console           string `json:"console,omitzero" validate:"uri"`
	Browser           string `json:"browser,omitzero" validate:"uri"`
	Hub               string `json:"hub,omitzero" validate:"uri"`
	Registry          string `json:"registry,omitzero"`
	Audit             string `json:"audit,omitzero"`
	Logs              string `json:"logs,omitzero"`
	Metrics           string `json:"metrics,omitzero"`
	MetricsNg         string `json:"metrics-ng,omitzero"`
	Tracing           string `json:"tracing,omitzero"`
	Byok              string `json:"byok,omitzero"`
	Metering          string `json:"metering,omitzero"`
	BillingNg         string `json:"billing-ng,omitzero"`
	Grafana           string `json:"grafana,omitzero"`
	TerraformExporter string `json:"terraform-exporter,omitzero"`
	K8sCrdExporter    string `json:"k8s-crd-exporter,omitzero"`
	Mcp               string `json:"mcp,omitzero"`
	Auth              string `json:"auth,omitzero" validate:"uri"`
}

type DiscoveryFirebase struct {
	ApiKey     string `json:"apiKey,omitzero"`
	AuthDomain string `json:"authDomain,omitzero"`
	ProjectId  string `json:"projectId,omitzero"`
}

type Discovery struct {
	Endpoints *DiscoveryEndpoints `json:"endpoints,omitzero"`
	Firebase  *DiscoveryFirebase  `json:"firebase,omitzero"`
}
