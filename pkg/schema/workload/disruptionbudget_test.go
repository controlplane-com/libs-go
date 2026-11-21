package workload

import (
	"testing"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// Mirrors nodelibs/schema/src/workload.test.ts spec.disruptionBudget (master f94fb6d78).

func TestMinAvailableBlocksDrains(t *testing.T) {
	cases := []struct {
		minAvailable string
		minScale     int
		blocks       bool
	}{
		// blocks (rejected)
		{"1", 1, true}, {"3", 3, true}, {"5", 3, true}, {"50%", 1, true}, {"80%", 4, true},
		{"1", 0, true}, // minScale 0 judged at one replica
		// does not block (accepted)
		{"2", 3, false}, {"50%", 2, false}, {"75%", 10, false}, {"0", 1, false},
	}
	for _, c := range cases {
		if got := minAvailableBlocksDrains(c.minAvailable, c.minScale); got != c.blocks {
			t.Errorf("minAvailableBlocksDrains(%q, %d) = %v, want %v", c.minAvailable, c.minScale, got, c.blocks)
		}
	}
}

func TestDisruptionBudgetFieldValidation(t *testing.T) {
	rule := DisruptionBudget{}.ValidationRule()
	valid := func(v map[string]any) error {
		_, err := rule.ValidateAt([]string{"spec", "disruptionBudget"}, v, nil, validate.Prefs{})
		return err
	}
	ok := []map[string]any{
		{"minAvailable": float64(3)}, // JSON number coerced to "3"
		{"minAvailable": "3"},
		{"minAvailable": "90%"},
		{"minAvailable": float64(0)},
		{"maxUnavailable": float64(2)},
		{"maxUnavailable": "20%"},
		{"maxUnavailable": "100%"},
	}
	for _, v := range ok {
		if err := valid(v); err != nil {
			t.Errorf("%v rejected: %v", v, err)
		}
	}
	bad := []map[string]any{
		{"minAvailable": "3", "maxUnavailable": "2"}, // oxor
		{"maxUnavailable": float64(0)},
		{"maxUnavailable": "0"},
		{"maxUnavailable": "0%"},
		{"maxUnavailable": "101%"},
		{"minAvailable": "100%"},
		{"minAvailable": float64(2.5)},
		{"minAvailable": "2.5"},
		{"minAvailable": float64(-3)},
		{"minAvailable": "-3"},
		{"minAvailable": "02"},
	}
	for _, v := range bad {
		if err := valid(v); err == nil {
			t.Errorf("%v accepted, want rejected", v)
		}
	}
}
