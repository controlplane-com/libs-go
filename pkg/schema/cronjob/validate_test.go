package cronjob

import (
	"testing"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// TestConditionTimeDefaults pins cronjob.ts:11,14 — both timestamps carry a default,
// so a condition that omits them comes back carrying them. The port had neither (a
// models.go comment claimed they were RegisterRefine'd in pkg/handlers; nothing was),
// so every stored job-execution condition was missing two fields TS writes.
func TestConditionTimeDefaults(t *testing.T) {
	obj := map[string]any{"type": "Complete"}
	if err := validate.For[JobExecutionCondition]().ApplyCtx(obj, nil, validate.Prefs{}); err != nil {
		t.Fatalf("condition rejected: %v", err)
	}
	for _, k := range []string{"lastDetectionTime", "lastTransitionTime"} {
		s, _ := obj[k].(string)
		if s == "" {
			t.Fatalf("%s was not defaulted: %#v", k, obj[k])
		}
		// Joi's default is `new Date(Date.now()).toUTCString()` — an HTTP-date, not ISO.
		if s[len(s)-4:] != " GMT" {
			t.Errorf("%s = %q, want an HTTP-date ending in GMT", k, s)
		}
	}
	// A supplied value wins over the default.
	given := map[string]any{"type": "Complete", "lastDetectionTime": "2020-01-02T03:04:05.000Z"}
	if err := validate.For[JobExecutionCondition]().ApplyCtx(given, nil, validate.Prefs{}); err != nil {
		t.Fatalf("condition rejected: %v", err)
	}
	if given["lastDetectionTime"] == conditionTimeDefault {
		t.Error("an explicit lastDetectionTime was overwritten by the default")
	}
}

// TestNullCompletionTimeDropped pins cronjob.ts:38-43's object custom: `.allow(null)`
// lets the null through and the custom then DELETES the key, so TS never stores or
// returns `"completionTime": null`. The port kept it.
func TestNullCompletionTimeDropped(t *testing.T) {
	obj := map[string]any{"workloadVersion": 1, "name": "j1", "completionTime": nil}
	if err := validate.For[JobExecutionStatus]().ApplyCtx(obj, nil, validate.Prefs{}); err != nil {
		t.Fatalf("job execution rejected: %v", err)
	}
	if _, present := obj["completionTime"]; present {
		t.Errorf("completionTime survived as %#v", obj["completionTime"])
	}
	// A real value is untouched.
	obj2 := map[string]any{"workloadVersion": 1, "name": "j1", "completionTime": "2020-01-02T03:04:05.000Z"}
	if err := validate.For[JobExecutionStatus]().ApplyCtx(obj2, nil, validate.Prefs{}); err != nil {
		t.Fatalf("job execution rejected: %v", err)
	}
	if obj2["completionTime"] == nil {
		t.Error("a non-null completionTime was dropped")
	}
}
