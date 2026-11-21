package cronjob

// validate.go attaches the two cronjob.ts object rules that a `validate:` struct tag
// cannot carry. Both were recorded only as a models.go comment claiming they were
// "RegisterRefine'd in pkg/handlers" — no such registration ever existed, so a job
// execution condition came back without its two timestamps.

import (
	"time"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// conditionTimeDefault mirrors cronjob.ts:11,14 `.default(new Date(Date.now()).toUTCString())`.
// The Joi default is evaluated ONCE, when the module is loaded, so what TS stamps is the
// service's start time rendered as an HTTP-date — not the time of the request. Capturing
// it at package init is the same thing on this side.
var conditionTimeDefault = time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")

// ValidationRule ports JobExecutionCondition's two `.default(...)` timestamps.
func (JobExecutionCondition) ValidationRule() *validate.Rule {
	return validate.Reflect(JobExecutionCondition{}).
		Fork("lastDetectionTime", func(r *validate.Rule) *validate.Rule { return r.Default(conditionTimeDefault) }).
		Fork("lastTransitionTime", func(r *validate.Rule) *validate.Rule { return r.Default(conditionTimeDefault) })
}

// ValidationRule ports JobExecutionStatus's object `.custom()` (cronjob.ts:38-43):
// `if (value.completionTime == null) delete value.completionTime`. `.allow(null)` lets
// the null THROUGH, and without this the key was persisted and returned as an explicit
// `"completionTime": null` that TS never writes.
func (JobExecutionStatus) ValidationRule() *validate.Rule {
	return validate.Reflect(JobExecutionStatus{}).CustomTyped(dropNullCompletionTime)
}

// dropNullCompletionTime serves both evaluators: the generic one hands over the
// object as a map (delete the key), the typed walk as a *JobExecutionStatus (the
// zero string is already omitted by the encoder).
func dropNullCompletionTime(value any, h *validate.Helper) (any, error) {
	if m, ok := value.(map[string]any); ok {
		if v, present := m["completionTime"]; present && v == nil {
			delete(m, "completionTime")
		}
		return m, nil
	}
	if s, ok := value.(*JobExecutionStatus); ok && s.CompletionTime == "" {
		s.CompletionTime = ""
	}
	return value, nil
}
