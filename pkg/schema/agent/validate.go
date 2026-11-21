package agent

// validate.go attaches validation to the canonical agent types via the Validator
// interface — the go-libs home of what used to be pkg/handlers/schema_agent.go +
// schema_reflect_agent.go. agent.ts reflects from Agent. The old `custom=uriHttpHttps`
// tag is replaced by the strongly-typed HttpHttpsUri field type.

import (
	"strings"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// ValidationRule reflects Agent.
func (Agent) ValidationRule() *validate.Rule {
	return validate.Reflect(Agent{})
}

// ValidationRule reflects the -info beat payload (validated by the -info route).
func (AgentInfo) ValidationRule() *validate.Rule { return validate.Reflect(AgentInfo{}) }

// ValidationRule ports agent.ts:9 `Joi.string().uri({scheme:['http','https']})`. Joi
// reports string.uriCustomScheme for EVERY failure once a scheme option is present —
// including a value that is not a URI at all — so the .URI() rule cannot run first
// (it answered string.uri for "not a uri" where TS answers string.uriCustomScheme).
// The scheme list renders as Joi's alternation ("http|https"), not as Go's [http https],
// and the message quotes the LABEL, not the value.
func (HttpHttpsUri) ValidationRule() *validate.Rule {
	return validate.String().Custom(func(value any, h *validate.Helper) (any, error) {
		s, _ := value.(string)
		if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
			return s, nil
		}
		return nil, h.WithCtx(
			h.Errorf("string.uriCustomScheme", "%q must be a valid uri with a scheme matching the http|https pattern", h.Label()),
			"scheme", "http|https")
	})
}
