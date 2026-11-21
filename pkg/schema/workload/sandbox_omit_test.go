package workload

import (
	"encoding/json"
	"testing"

	"github.com/controlplane-com/libs-go/pkg/schema/sandbox"
)

// A dev sandbox written from Go must not carry empty scaling or volume objects: the
// server rejects scaling on a dev box and treats volume as immutable. The port models
// workload.sandbox as sandbox.SandboxSpec (which has no scaling key and a nil-omitted
// volume pointer), so neither surfaces as an empty object.
func TestSandboxOptionalObjectsAreOmitted(t *testing.T) {
	b, err := json.Marshal(WorkloadSpec{Sandbox: &sandbox.SandboxSpec{Type: "dev"}})
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
