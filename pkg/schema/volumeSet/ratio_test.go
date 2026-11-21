package volumeSet

import (
	"testing"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// TestMountResourcesZeroMin pins the JS division semantics of volumeSet.ts:293-307.
// A `minCpu` of "0" makes maxCpu/minCpu Infinity in JS, which trips the ratio check;
// the port guarded the division with `minCPU > 0` and so SKIPPED the check, accepting
// a mount-resources block TS rejects. 0/0 is NaN, which compares false — so min==max=="0"
// stays accepted on both sides.
func TestMountResourcesZeroMin(t *testing.T) {
	cases := []struct {
		body string
		ok   bool
	}{
		{`{"minCpu":"0","maxCpu":"2000m"}`, false},
		{`{"minMemory":"0","maxMemory":"2Gi"}`, false},
		{`{"minCpu":"0","maxCpu":"0"}`, true}, // 0/0 is NaN: no ratio failure
		{`{"minCpu":"500m","maxCpu":"2000m"}`, true},
		{`{"minCpu":"100m","maxCpu":"2000m"}`, false}, // 20:1 > 4:1
	}
	for _, c := range cases {
		// The typed path: the ratio check is a RefineT on the resources struct.
		_, verr := validate.Decode[VolumeSetSpecMountOptionsResources](validate.Bytes(c.body), nil, validate.Prefs{})
		if (verr == nil) != c.ok {
			t.Errorf("%s: ok=%v, err=%v", c.body, c.ok, verr)
		}
	}
}
