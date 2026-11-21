package orgLogging

import (
	"testing"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// ctx supplies the org the credentials localLink expands against.
func ctx() *validate.Context { return &validate.Context{Org: "o1"} }

// TestRegionEnums pins the two region enums that orgLogging.ts declares and the port
// dropped: both leaf types were unconstrained strings (a models.go comment claimed a
// registration in a pkg/handlers file that no longer exists), so an org could be
// configured to ship logs to a region that does not exist.
func TestRegionEnums(t *testing.T) {
	r := validate.Reflect(CloudWatchLogging{})
	obj := map[string]any{"region": "moon-base-1", "credentials": "//secret/s",
		"groupName": "g", "streamName": "s"}
	err := r.ApplyAt(nil, obj, ctx(), validate.Prefs{})
	if err == nil {
		t.Fatal("cloudWatch.region accepted an unknown AWS region")
	}
	d := err.(*validate.Error).Details()[0]
	if d.Type != "any.only" {
		t.Errorf("type = %q, want any.only", d.Type)
	}
	// The valids order is the TS declaration order — Joi prints it in the message.
	if vs, _ := d.Context["valids"].([]any); len(vs) != 18 || vs[0] != "us-east-1" || vs[17] != "af-south-1" {
		t.Errorf("valids = %v", d.Context["valids"])
	}
	ok := map[string]any{"region": "eu-west-3", "credentials": "//secret/s", "groupName": "g", "streamName": "s"}
	if err := r.ApplyAt(nil, ok, ctx(), validate.Prefs{}); err != nil {
		t.Fatalf("valid region rejected: %v", err)
	}

	sd := validate.Reflect(StackdriverLogging{})
	if err := sd.ApplyAt(nil, map[string]any{"credentials": "//secret/s", "location": "moon-1"}, ctx(), validate.Prefs{}); err == nil {
		t.Fatal("stackdriver.location accepted an unknown GCP region")
	}
	if err := sd.ApplyAt(nil, map[string]any{"credentials": "//secret/s", "location": "us-west4"}, ctx(), validate.Prefs{}); err != nil {
		t.Fatalf("valid location rejected: %v", err)
	}
}

// TestElasticXor pins ElasticLogging's .xor(elasticCloud, aws, generic): exactly one
// provider. Neither half was enforced, so an elastic block with none or with several
// was accepted and persisted — and its consumers read "the first key present".
func TestElasticXor(t *testing.T) {
	rule := ElasticLogging{}.ValidationRule()
	_, err := rule.ValidateAt(nil, map[string]any{}, nil, validate.Prefs{})
	if err == nil {
		t.Fatal("empty elastic accepted")
	}
	if d := err.(*validate.Error).Details()[0]; d.Type != "object.missing" {
		t.Errorf("type = %q, want object.missing", d.Type)
	}
	two := map[string]any{
		"aws":          map[string]any{"host": "a.es.amazonaws.com", "port": 443, "index": "i", "type": "t", "credentials": "//secret/s", "region": "us-east-1"},
		"elasticCloud": map[string]any{"index": "i", "type": "t", "credentials": "//secret/s", "cloudId": "c"},
	}
	_, err = rule.ValidateAt(nil, two, ctx(), validate.Prefs{})
	if err == nil {
		t.Fatal("two elastic providers accepted")
	}
	d := err.(*validate.Error).Details()[0]
	if d.Type != "object.xor" {
		t.Errorf("type = %q, want object.xor", d.Type)
	}
	// Joi renders the peers in the .xor() argument order, not the struct field order.
	if want := `"value" contains a conflict between exclusive peers [elasticCloud, aws, generic]`; d.Message != want {
		t.Errorf("message = %q, want %q", d.Message, want)
	}
}

// TestEsHostname pins the elastic host leaves to Joi's string.hostname (they used to
// raise an any.custom quoting the value).
func TestEsHostname(t *testing.T) {
	_, err := EsHost("").ValidationRule().ValidateAt([]string{"host"}, "-bad-", nil, validate.Prefs{})
	if d := err.(*validate.Error).Details()[0]; d.Type != "string.hostname" {
		t.Errorf("EsHost type = %q, want string.hostname", d.Type)
	}
	// EsAmazonHost keeps its own suffix check AFTER the hostname check.
	_, err = EsAmazonHost("").ValidationRule().ValidateAt([]string{"host"}, "a.example.com", nil, validate.Prefs{})
	if err == nil {
		t.Fatal("EsAmazonHost accepted a non-amazon host")
	}
	if _, err := EsAmazonHost("").ValidationRule().ValidateAt([]string{"host"}, "a.es.amazonaws.com", nil, validate.Prefs{}); err != nil {
		t.Fatalf("EsAmazonHost rejected a valid host: %v", err)
	}
}
