package command

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// TestShrinkVolumeStatusRevertFields pins the fields the revert path records
// (command.ts ShrinkVolumeStatus, master 2d6d8442b and 17b92872b): the old attributes
// map takes empty values, the old disk type is held to the shared DiskTypes list, and
// both timestamps are dates. They also round-trip in declaration order.
func TestShrinkVolumeStatusRevertFields(t *testing.T) {
	decode := func(body string) (*ShrinkVolumeStatus, *validate.Error) {
		return validate.Decode[ShrinkVolumeStatus](validate.Bytes(body), nil, validate.Prefs{})
	}

	body := `{"stage":"await-data-ready","storageDeviceIdToRemove":"dev-old",` +
		`"oldVolumeAttributes":{"a":"","b":"v"},"oldVolumeSize":100,"oldDiskType":"gp3",` +
		`"newStorageDeviceId":"dev-new","awaitDataReadySince":"2026-09-30T00:00:00.000Z"}`
	st, verr := decode(body)
	if verr != nil {
		t.Fatal(verr)
	}
	if st.OldVolumeAttributes["a"] != "" || st.OldVolumeAttributes["b"] != "v" {
		t.Errorf("oldVolumeAttributes = %v", st.OldVolumeAttributes)
	}
	if st.OldVolumeSize == nil || *st.OldVolumeSize != 100 || st.OldDiskType != "gp3" {
		t.Errorf("oldVolumeSize = %v, oldDiskType = %q", st.OldVolumeSize, st.OldDiskType)
	}

	if _, verr := decode(`{"stage":"fail","oldDiskType":"not-a-disk-type"}`); verr == nil {
		t.Error("an unknown oldDiskType was accepted")
	}
	if _, verr := decode(`{"stage":"fail","awaitDataReadySince":"whenever"}`); verr == nil {
		t.Error("a non-date awaitDataReadySince was accepted")
	}

	out, err := json.Marshal(st)
	if err != nil {
		t.Fatal(err)
	}
	for _, pair := range [][2]string{
		{"storageDeviceIdToRemove", "oldVolumeAttributes"},
		{"oldDiskType", "newStorageDeviceId"},
		{"newStorageDeviceId", "awaitDataReadySince"},
	} {
		if strings.Index(string(out), pair[0]) > strings.Index(string(out), pair[1]) {
			t.Errorf("%s follows %s on the wire: %s", pair[0], pair[1], out)
		}
	}
}
