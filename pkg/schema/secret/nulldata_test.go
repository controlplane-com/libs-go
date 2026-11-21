package secret

import (
	"testing"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

func TestNullDataFailsTheTypeBranch(t *testing.T) {
	for _, c := range []struct{ body, typ, msg string }{
		{`{"name":"s1","type":"opaque","data":null}`, "object.base", `"data" must be of type object`},
		{`{"name":"s1","type":"dictionary","data":null}`, "object.base", `"data" must be of type object`},
		{`{"name":"s1","type":"docker","data":null}`, "string.base", `"data" must be a string`},
	} {
		_, verr := validate.Decode[Secret](validate.Bytes(c.body), nil, validate.Prefs{})
		if verr == nil {
			t.Fatalf("%s: accepted", c.body)
		}
		d := verr.Details()[0]
		if d.Type != c.typ || d.Message != c.msg || len(d.Path) != 1 || d.Path[0] != "data" {
			t.Fatalf("%s: got %+v", c.body, d)
		}
	}
	_, err := validate.Validate(Secret{}.ValidationRule(), map[string]any{"name": "s1", "type": "opaque", "data": nil}, nil, validate.Prefs{})
	if err == nil || err.Type != "object.base" {
		t.Fatalf("generic: got %v", err)
	}
	if _, verr := validate.Decode[Secret](validate.Bytes(`{"name":"s1","type":"opaque"}`), nil, validate.Prefs{}); verr != nil {
		t.Fatalf("absent data rejected: %v", verr)
	}
}
