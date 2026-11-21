/* hand-maintained — canonical source of truth for the shared Tracing sub-schema.
 *
 * `validate:` struct tags drive pkg/validate.FromType. See
 * nodelibs/schema/src/tracing.ts and the tag DSL on go-libs/schema/ipSet/models.go.
 * The provider .xor(lightstep, otel, controlplane) and the top-level .custom that
 * mirrors the deprecated top-level `lightstep` under `provider.lightstep` (and
 * vice-versa) are wired in pkg/handlers/schema_reflect_orggvc.go.
 */

package tracing

// SecretLink ports LocalLink().custom(toKind("secret")): a localLink pointing to
// a resource of kind secret.
type SecretLink string

type LightstepTracing struct {
	Endpoint    string     `json:"endpoint" validate:"required"`
	Credentials SecretLink `json:"credentials,omitzero" validate:"optional"`
}

type OpenTelemetry struct {
	Endpoint string `json:"endpoint" validate:"required"`
}

type TracingCustomTags map[string]TracingCustomTag

type TracingProviderControlplane struct {
}

// TracingProvider ports tracing.ts provider. .xor(lightstep, otel, controlplane).
type TracingProvider struct {
	Otel         *OpenTelemetry               `json:"otel,omitzero"`
	Lightstep    *LightstepTracing            `json:"lightstep,omitzero"`
	Controlplane *TracingProviderControlplane `json:"controlplane,omitzero"`
}

// Tracing ports tracing.ts:20. tracingRefine handles the lightstep<->provider mirror.
type Tracing struct {
	Sampling   float64            `json:"sampling" validate:"min=0,max=100,precision=2,required"`
	Lightstep  *LightstepTracing  `json:"lightstep,omitzero"`
	CustomTags *TracingCustomTags `json:"customTags,omitzero"`
	Provider   *TracingProvider   `json:"provider,omitzero"`
}

type TracingCustomTagLiteral struct {
	Value string `json:"value,omitzero" validate:"max=50"`
}

type TracingCustomTag struct {
	Literal *TracingCustomTagLiteral `json:"literal,omitzero"`
}
