package cloudaccount

import (
	"net/http"
	"testing"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// TestDataErrorsCarryTheDataPath pins the provider-body errors to the path TS reports.
// cloudAccountRefine validated `data` from a FRESH root, so an unknown key inside it
// came back at ["x"] with label "x" where Joi says ["data","x"] / "data.x".
func TestDataErrorsCarryTheDataPath(t *testing.T) {
	ctx := &validate.Context{Org: "o1", Method: http.MethodPost}
	body := `{"name":"ca1","provider":"gcp","data":{"projectId":"abcdef","bogus":1}}`
	_, verr := validate.Decode[CloudAccount](validate.Bytes(body), ctx, validate.Prefs{})
	if verr == nil {
		t.Fatal("unknown key inside data accepted")
	}
	d := verr.Details()[0]
	if len(d.Path) != 2 || d.Path[0] != "data" || d.Path[1] != "bogus" {
		t.Fatalf("path = %v, want [data bogus]", d.Path)
	}
	if d.Context["label"] != "data.bogus" {
		t.Errorf("label = %v, want data.bogus", d.Context["label"])
	}
}

func TestNullDataIsNotAnObject(t *testing.T) {
	for _, body := range []string{`{"name":"ca1","provider":"aws","data":null}`, `{"name":"ca1","provider":"gcp","data":null}`} {
		_, verr := validate.Decode[CloudAccount](validate.Bytes(body), nil, validate.Prefs{})
		if verr == nil {
			t.Fatalf("%s: accepted", body)
		}
		d := verr.Details()[0]
		if d.Type != "object.base" || d.Message != `"data" must be of type object` || len(d.Path) != 1 || d.Path[0] != "data" {
			t.Fatalf("%s: got %+v", body, d)
		}
		if v, ok := d.Context["value"]; !ok || v != nil || d.Context["type"] != "object" || d.Context["key"] != "data" {
			t.Fatalf("%s: context %+v", body, d.Context)
		}
	}
	_, err := validate.Validate(CloudAccount{}.ValidationRule(), map[string]any{"name": "ca1", "provider": "aws", "data": nil}, nil, validate.Prefs{})
	if err == nil || err.Type != "object.base" {
		t.Fatalf("generic: got %v", err)
	}
	if _, verr := validate.Decode[CloudAccount](validate.Bytes(`{"name":"ca1","provider":"aws"}`), nil, validate.Prefs{}); verr != nil {
		t.Fatalf("absent data rejected: %v", verr)
	}
}
