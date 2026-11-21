package base

import (
	"testing"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

func TestBaseNameRule(t *testing.T) {
	r := Name("").ValidationRule()
	s := validate.ObjectSchema(validate.Object(map[string]*validate.Rule{"n": r}))
	if err := s.ApplyCtx(map[string]any{"n": "Bad_Name"}, &validate.Context{}, validate.Prefs{}); err == nil {
		t.Fatal("expected invalid name to be rejected")
	}
	if err := s.ApplyCtx(map[string]any{"n": "good-name"}, &validate.Context{}, validate.Prefs{}); err != nil {
		t.Fatalf("valid name rejected: %v", err)
	}
}

func TestBaseLocalLinkExpandAndKind(t *testing.T) {
	// base.LocalLink expands; composing ToKind("location") kind-checks — the
	// pattern gvc.LocationLink will use.
	r := LocalLinkRule().Custom(ToKind("location"))
	s := validate.ObjectSchema(validate.Object(map[string]*validate.Rule{"link": r}))
	ctx := &validate.Context{Org: "o1"}
	// //location/foo expands to /org/o1/location/foo and passes the location kind.
	obj := map[string]any{"link": "//location/foo"}
	if err := s.ApplyCtx(obj, ctx, validate.Prefs{}); err != nil {
		t.Fatalf("valid location link rejected: %v", err)
	}
	if obj["link"] != "/org/o1/location/foo" {
		t.Fatalf("link not expanded: %v", obj["link"])
	}
	// a secret link fails the location kind check.
	if err := s.ApplyCtx(map[string]any{"link": "//secret/foo"}, ctx, validate.Prefs{}); err == nil {
		t.Fatal("expected wrong-kind link to be rejected")
	}
}

func TestBaseTagsCustom(t *testing.T) {
	r := Tags(nil).ValidationRule()
	s := validate.ObjectSchema(validate.Object(map[string]*validate.Rule{"tags": r}))
	// bool coerces to string, null dropped.
	obj := map[string]any{"tags": map[string]any{"a": true, "b": nil, "c": "x"}}
	if err := s.ApplyCtx(obj, &validate.Context{}, validate.Prefs{}); err != nil {
		t.Fatalf("valid tags rejected: %v", err)
	}
	tags := obj["tags"].(map[string]any)
	if tags["a"] != "true" {
		t.Fatalf("bool tag not coerced: %v", tags["a"])
	}
	if _, ok := tags["b"]; ok {
		t.Fatal("null tag not dropped")
	}
}
