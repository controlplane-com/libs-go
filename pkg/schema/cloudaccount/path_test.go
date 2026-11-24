package cloudaccount

import (
	"net/http"
	"reflect"
	"strings"
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

func TestAwsRoleNameAndBoundary(t *testing.T) {
	long := strings.Repeat("a", 65)
	for body, ok := range map[string]bool{
		`{"name":"ca1","provider":"aws","data":{"roleArn":"arn:aws:iam::1:role/` + long[:64] + `"}}`:                                               true,
		`{"name":"ca1","provider":"aws","data":{"roleArn":"arn:aws:iam::1:role/cpln/` + long[:64] + `"}}`:                                          true,
		`{"name":"ca1","provider":"aws","data":{"roleArn":"arn:aws:iam::1:role/` + long + `"}}`:                                                    false,
		`{"name":"ca1","provider":"aws","data":{"roleArn":"arn:aws:iam::1:role/x","boundaryPolicyArn":"arn:aws:iam::1:policy/cpln-x-boundary"}}`:   true,
		`{"name":"ca1","provider":"aws","data":{"roleArn":"arn:aws:iam::1:role/x","boundaryPolicyArn":"arn:aws:iam::1:policy/cpln/x/boundary"}}`:   true,
		`{"name":"ca1","provider":"aws","data":{"roleArn":"arn:aws:iam::1:role/x","boundary":"arn:aws:iam::1:policy/x"}}`:                          false,
		`{"name":"ca1","provider":"aws","data":{"roleArn":"arn:aws:iam::1:role/x","boundaryPolicyArn":"arn:aws:iam::1:role/x"}}`:                   false,
		`{"name":"ca1","provider":"aws","data":{"roleArn":"arn:aws:iam::1:role/cpln/` + long + `"}}`:                                               false,
		`{"name":"ca1","provider":"aws","data":{"roleArn":"arn:aws:iam::1:role/x","boundaryPolicyArn":"arn:aws:iam::aws:policy/PowerUserAccess"}}`: true,
		`{"name":"ca1","provider":"gcp","data":{"projectId":"abcdef","boundaryPolicyArn":"arn:aws:iam::1:policy/x"}}`:                              false,
	} {
		if _, verr := validate.Decode[CloudAccount](validate.Bytes(body), nil, validate.Prefs{}); (verr == nil) != ok {
			t.Errorf("%s: accepted=%v, want %v (%v)", body, verr == nil, ok, verr)
		}
	}
}

func TestSuffixRule(t *testing.T) {
	rule := Suffix("").ValidationRule()
	for v, ok := range map[string]bool{"a": true, "eu2": true, "abcdefg": true, "abcdefgh": false, "-abc": false, "a-b": false, "ABC": false, "": false} {
		if _, err := rule.ValidateAt(nil, v, nil, validate.Prefs{}); (err == nil) != ok {
			t.Errorf("%q: accepted=%v, want %v", v, err == nil, ok)
		}
	}
	for _, body := range []string{
		`{"name":"ca1","provider":"aws","data":{"roleArn":"arn:aws:iam::1:role/cpln-x-abc","suffix":"abc"}}`,
		`{"name":"ca1","provider":"gcp","data":{"projectId":"abcdef","suffix":"abc"}}`,
	} {
		if _, verr := validate.Decode[CloudAccount](validate.Bytes(body), nil, validate.Prefs{}); verr == nil {
			t.Errorf("suffix stored: %s", body)
		}
	}
}

func TestAwsBoundarySameAccountOrAwsManaged(t *testing.T) {
	for boundary, ok := range map[string]bool{
		"arn:aws:iam::111122223333:policy/cpln/acme-boundary": true,
		"arn:aws:iam::aws:policy/PowerUserAccess":             true,
		"arn:aws:iam::999999999999:policy/other":              false,
		"arn:aws-cn:iam::111122223333:policy/x":               false,
		"arn:aws:iam::111122223333:policy/path/":              false,
	} {
		body := `{"name":"ca1","provider":"aws","data":{"roleArn":"arn:aws:iam::111122223333:role/cpln-acme","boundaryPolicyArn":"` + boundary + `"}}`
		if _, verr := validate.Decode[CloudAccount](validate.Bytes(body), nil, validate.Prefs{}); (verr == nil) != ok {
			t.Errorf("%s: accepted=%v, want %v (%v)", boundary, verr == nil, ok, verr)
		}
	}
}

func TestDataIsImmutable(t *testing.T) {
	for _, f := range validate.ImmutableFields(reflect.TypeFor[CloudAccount]()) {
		if f == "data" {
			return
		}
	}
	t.Fatal("data is not among the immutable fields")
}
