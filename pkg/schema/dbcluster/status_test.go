package dbcluster

import (
	"testing"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// TestDbClusterStatusRules pins dbcluster.ts:86-92's `lastModified: Joi.date()` and
// `version: Joi.number().positive().integer()`, which the port dropped. The type has
// no route today, so this is the only thing standing between the next handler and a
// silently unvalidated status.
func TestDbClusterStatusRules(t *testing.T) {
	r := validate.For[DbClusterStatus]()
	if err := r.ApplyCtx(map[string]any{"lastModified": "nope"}, nil, validate.Prefs{}); err == nil {
		t.Error("a non-date lastModified accepted")
	}
	if err := r.ApplyCtx(map[string]any{"version": -1}, nil, validate.Prefs{}); err == nil {
		t.Error("a negative version accepted")
	}
	if err := r.ApplyCtx(map[string]any{"version": 1.5}, nil, validate.Prefs{}); err == nil {
		t.Error("a fractional version accepted")
	}
	if err := r.ApplyCtx(map[string]any{"version": 2, "lastModified": "2020-01-02T03:04:05.000Z"}, nil, validate.Prefs{}); err != nil {
		t.Errorf("a valid status rejected: %v", err)
	}
}
