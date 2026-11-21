package mk8sAddons

import (
	"testing"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// TestAwsEFSRoleArnOptional pins master 405bb677c: roleArn is optional, because the EFS
// CSI IAM role is created when it is absent. It is still pattern-checked when present.
func TestAwsEFSRoleArnOptional(t *testing.T) {
	for _, c := range []struct {
		body, want string
	}{
		{`{}`, ""},
		{`{"roleArn":"arn:aws:iam::123:role/x"}`, ""},
		{`{"roleArn":"bad"}`, "string.pattern.base"},
	} {
		_, verr := validate.Decode[AwsEFSAddonConfig](validate.Bytes(c.body), nil, validate.Prefs{})
		got := ""
		if verr != nil {
			got = verr.Details()[0].Type
		}
		if got != c.want {
			t.Errorf("%s: got %q, want %q", c.body, got, c.want)
		}
	}
}
