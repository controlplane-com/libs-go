package agent

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

func TestAgentValidatorDispatch(t *testing.T) {
	// The agent kind is migrated: its type implements Validator.
	if validate.OfTypeIfValidator(reflect.TypeOf(Agent{})) == nil {
		t.Fatal("Agent does not implement Validator — kind not migrated")
	}
	s := validate.For[Agent]()

	create := &validate.Context{Org: "o1", Method: http.MethodPost}

	// The createSchema fork (agent.ts:699) belongs to data-service, not the Joi type:
	// name stays optional and description is not defaulted.
	obj := map[string]any{"name": "my-agent"}
	if err := s.ApplyCtx(obj, create, validate.Prefs{}); err != nil {
		t.Fatalf("valid agent rejected: %v", err)
	}
	if _, ok := obj["description"]; ok {
		t.Fatalf("the Joi type defaulted description: %v", obj["description"])
	}
	if err := s.ApplyCtx(map[string]any{}, create, validate.Prefs{}); err != nil {
		t.Fatalf("the Joi type required name: %v", err)
	}

	// unknown top-level field rejected (proves it's a real schema, not a no-op).
	if err := s.ApplyCtx(map[string]any{"name": "a", "bogus": 1}, create, validate.Prefs{}); err == nil {
		t.Fatal("expected unknown field to be rejected")
	}

	// the typed path agrees
	a, verr := validate.Decode[Agent](validate.Bytes(`{"name":"my-agent"}`), create, validate.Prefs{})
	if verr != nil {
		t.Fatalf("typed: valid agent rejected: %v", verr)
	}
	if a.Description != "" {
		t.Fatalf("typed: the Joi type defaulted description: %q", a.Description)
	}
	if _, verr := validate.Decode[Agent](validate.Bytes(`{"name":"a","bogus":1}`), create, validate.Prefs{}); verr == nil {
		t.Fatal("typed: expected unknown field to be rejected")
	}
}

func TestAgentHubEndpointScheme(t *testing.T) {
	// HttpHttpsUri nominal type (replaced custom=uriHttpHttps) enforces http/https.
	r := HttpHttpsUri("").ValidationRule()
	s := validate.ObjectSchema(validate.Object(map[string]*validate.Rule{"u": r}))
	if err := s.ApplyCtx(map[string]any{"u": "ftp://x"}, &validate.Context{}, validate.Prefs{}); err == nil {
		t.Fatal("expected non-http(s) scheme to be rejected")
	}
	if err := s.ApplyCtx(map[string]any{"u": "https://hub.example.com"}, &validate.Context{}, validate.Prefs{}); err != nil {
		t.Fatalf("valid https uri rejected: %v", err)
	}
}

// TestHubEndpointScheme pins agent.ts:9's `uri({scheme:['http','https']})`. Joi reports
// string.uriCustomScheme for EVERY failure once a scheme option is present — including a
// value that is not a URI at all, which the port answered with string.uri because the
// .URI() rule ran ahead of the scheme custom. The message quotes the LABEL and renders
// the schemes as an alternation, not as Go's [http https].
func TestHubEndpointScheme(t *testing.T) {
	r := HttpHttpsUri("").ValidationRule()
	path := []string{"status", "bootstrapConfig", "hubEndpoint"}
	if _, err := r.ValidateAt(path, "https://x/y", nil, validate.Prefs{}); err != nil {
		t.Fatalf("https rejected: %v", err)
	}
	for _, bad := range []string{"ftp://x/y", "not a uri"} {
		_, verr := r.ValidateAt(path, bad, nil, validate.Prefs{})
		if verr == nil {
			t.Fatalf("%q accepted", bad)
		}
		d := verr.(*validate.Error).Details()[0]
		if d.Type != "string.uriCustomScheme" {
			t.Errorf("%q: type = %q, want string.uriCustomScheme", bad, d.Type)
		}
		want := `"status.bootstrapConfig.hubEndpoint" must be a valid uri with a scheme matching the http|https pattern`
		if d.Message != want {
			t.Errorf("%q: message = %q", bad, d.Message)
		}
		if d.Context["scheme"] != "http|https" || d.Context["value"] != bad {
			t.Errorf("%q: context = %v", bad, d.Context)
		}
	}
}
