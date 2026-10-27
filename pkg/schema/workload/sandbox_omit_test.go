package workload

import (
	"encoding/json"
	"testing"
)

// A dev sandbox written from Go must not carry empty scaling or volume objects: the server rejects scaling on a dev box and treats volume as immutable.
func TestSandboxOptionalObjectsAreOmitted(t *testing.T) {
	b, err := json.Marshal(WorkloadSpec{Sandbox: &WorkloadSpecSandbox{Type: "dev"}})
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"scaling", "volume"} {
		if _, ok := m["sandbox"][k]; ok {
			t.Errorf("sandbox.%s is sent as an empty object: %s", k, b)
		}
	}
}
