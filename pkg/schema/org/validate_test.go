package org

import (
	"net/http"
	"testing"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

func octx() *validate.Context { return &validate.Context{Org: "o1", Method: http.MethodPatch} }

// TestSyslogHostPattern pins org.ts:71's host pattern to Joi's message and context:
// the hand-rolled custom raised a bare string.pattern.base that quoted the VALUE and
// carried neither the regex nor the value in context.
func TestSyslogHostPattern(t *testing.T) {
	r := SyslogHostname("").ValidationRule()
	if _, err := r.ValidateAt([]string{"spec", "host"}, "a-b.example.com", nil, validate.Prefs{}); err != nil {
		t.Fatalf("valid host rejected: %v", err)
	}
	_, err := r.ValidateAt([]string{"spec", "host"}, "-bad-.com", nil, validate.Prefs{})
	if err == nil {
		t.Fatal("host starting with '-' accepted")
	}
	d := err.(*validate.Error).Details()[0]
	if d.Type != "string.pattern.base" {
		t.Fatalf("type = %q", d.Type)
	}
	want := `"spec.host" with value "-bad-.com" fails to match the required pattern: /^(?!-)[A-Za-z0-9-]{1,63}(?<!-)(\.(?!-)[A-Za-z0-9-]{1,63}(?<!-))*$/`
	if d.Message != want {
		t.Errorf("message =\n  %q\nwant\n  %q", d.Message, want)
	}
	if _, ok := d.Context["regex"]; !ok {
		t.Error("context.regex missing")
	}
	if d.Context["value"] != "-bad-.com" {
		t.Errorf("context.value = %v", d.Context["value"])
	}
}

// TestExtraLoggingWhen pins org.ts:80's `.when('logging', {not: exist, then: forbidden})`.
// It used to be a refine on the PARENT, so the rejection was raised at `spec` with the
// label "spec" where Joi blames `spec.extraLogging`.
func TestExtraLoggingWhen(t *testing.T) {
	// Through the kind (the typed walk), because OrgSpec's own rule carries a TYPED
	// refine that only the typed path can run.
	const s3 = `{"s3":{"bucket":"b","region":"us-east-1","credentials":"//secret/s"}}`

	_, verr := validate.Decode[Org](validate.Bytes(`{"name":"o1","spec":{"extraLogging":[`+s3+`]}}`), octx(), validate.Prefs{})
	if verr == nil {
		t.Fatal("extraLogging with no logging accepted")
	}
	d := verr.Details()[0]
	if d.Type != "any.unknown" {
		t.Fatalf("type = %q, want any.unknown", d.Type)
	}
	if len(d.Path) != 2 || d.Path[0] != "spec" || d.Path[1] != "extraLogging" {
		t.Errorf("path = %v, want [spec extraLogging]", d.Path)
	}
	if d.Context["label"] != "spec.extraLogging" {
		t.Errorf("label = %v, want spec.extraLogging", d.Context["label"])
	}

	// With `logging` present the branch does not fire.
	if _, verr := validate.Decode[Org](validate.Bytes(
		`{"name":"o1","spec":{"logging":`+s3+`,"extraLogging":[`+s3+`]}}`), octx(), validate.Prefs{}); verr != nil {
		t.Fatalf("extraLogging alongside logging rejected: %v", verr)
	}
}

// TestOrgKeyOrder pins the key declaration order to Joi's. org.ts:100 re-declares
// `name` through Base.keys({name,...}), and Joi's .keys() APPENDS — so `name` is
// validated after `links`, and abortEarly blames `version` (not `name`) for a body
// that is wrong in both.
func TestOrgKeyOrder(t *testing.T) {
	_, verr := validate.Decode[Org](validate.Bytes(`{"name":"BAD_NAME","version":-1}`),
		&validate.Context{Org: "o1", Method: http.MethodPost}, validate.Prefs{})
	if verr == nil {
		t.Fatal("expected a rejection")
	}
	d := verr.Details()[0]
	if d.Type != "number.positive" {
		t.Fatalf("abortEarly blamed %q (%s); Joi blames version first", d.Type, d.Message)
	}
}
