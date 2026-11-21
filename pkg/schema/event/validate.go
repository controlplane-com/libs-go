package event

// validate.go attaches validation to the canonical event types via the Validator
// interface — the go-libs home of what used to be pkg/handlers/schema_event.go +
// the event registrations in schema_reflect_orggvc.go. Both Event and EventContext
// are pure reflection ports of event.ts; EventContext carries Joi's .unknown(true)
// so arbitrary properties pass through.

import (
	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// ValidationRule reflects Event. This is the schema the "event" kind validates
// against (was RegisterReflect("event", event.Event{})).
func (Event) ValidationRule() *validate.Rule {
	return validate.Reflect(Event{})
}

// ValidationRule reflects EventContext and layers Joi's .unknown(true) (was
// RegisterObjectRules(EventContext{}, {Unknown: true})).
func (EventContext) ValidationRule() *validate.Rule {
	return validate.Reflect(EventContext{}).Unknown(true)
}
