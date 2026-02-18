/* auto-generated */

package tracing

type LightstepTracing struct {
	Endpoint    string `json:"endpoint"`
	Credentials string `json:"credentials,omitempty"`
}

type OpenTelemetry struct {
	Endpoint string `json:"endpoint"`
}

type TracingCustomTags map[string]TracingCustomTag

type TracingProviderControlplane struct {
}

type TracingProvider struct {
	Otel         *OpenTelemetry               `json:"otel,omitempty"`
	Lightstep    *LightstepTracing            `json:"lightstep,omitempty"`
	Controlplane *TracingProviderControlplane `json:"controlplane,omitempty"`
}

type Tracing struct {
	Sampling   float32            `json:"sampling"`
	Lightstep  *LightstepTracing  `json:"lightstep,omitempty"`
	CustomTags *TracingCustomTags `json:"customTags,omitempty"`
	Provider   *TracingProvider   `json:"provider,omitempty"`
}

type TracingCustomTagLiteral struct {
	Value string `json:"value,omitempty"`
}

type TracingCustomTag struct {
	Literal *TracingCustomTagLiteral `json:"literal,omitempty"`
}
