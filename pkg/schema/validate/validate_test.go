package validate

import "testing"

func TestApply(t *testing.T) {
	schema := &Schema{
		Fields: []FieldRule{
			{Path: "name", Required: true, Type: "string"},
			{Path: "alias", Default: "$name", Type: "string"},
			{Path: "spec.initialCapacity", Required: true, Type: "number", Integer: true, Min: Num(10)},
			{Path: "spec.performanceClass", Enum: []any{"general-purpose-ssd", "high-throughput-ssd"}},
			{Path: "spec.fileSystemType", Enum: []any{"ext4", "xfs"}, Default: "ext4"},
		},
	}

	cases := []struct {
		name    string
		obj     map[string]any
		wantErr bool
		check   func(map[string]any) bool
	}{
		{
			name: "valid + defaults applied",
			obj:  map[string]any{"name": "g1", "spec": map[string]any{"initialCapacity": float64(20), "performanceClass": "general-purpose-ssd"}},
			check: func(o map[string]any) bool {
				return o["alias"] == "g1" && o["spec"].(map[string]any)["fileSystemType"] == "ext4"
			},
		},
		{name: "missing required name", obj: map[string]any{"spec": map[string]any{"initialCapacity": float64(20)}}, wantErr: true},
		{name: "missing required nested", obj: map[string]any{"name": "g1", "spec": map[string]any{}}, wantErr: true},
		{name: "below min", obj: map[string]any{"name": "g1", "spec": map[string]any{"initialCapacity": float64(5)}}, wantErr: true},
		{name: "non-integer", obj: map[string]any{"name": "g1", "spec": map[string]any{"initialCapacity": float64(12.5)}}, wantErr: true},
		{name: "invalid enum", obj: map[string]any{"name": "g1", "spec": map[string]any{"initialCapacity": float64(20), "performanceClass": "nope"}}, wantErr: true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := schema.Apply(c.obj)
			if c.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !c.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if c.check != nil && !c.check(c.obj) {
				t.Fatalf("post-condition failed: %+v", c.obj)
			}
		})
	}
}

func TestCustom(t *testing.T) {
	schema := &Schema{
		Custom: []func(map[string]any) error{
			func(o map[string]any) error {
				if o["x"] == "bad" {
					return errTest
				}
				return nil
			},
		},
	}
	if err := schema.Apply(map[string]any{"x": "ok"}); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if err := schema.Apply(map[string]any{"x": "bad"}); err == nil {
		t.Fatal("expected custom error")
	}
}

var errTest = &testErr{}

type testErr struct{}

func (*testErr) Error() string { return "bad" }

// express's json parser accepts an ARRAY body under strict mode and hands it to the
// route as-is, so a route that passes req.body straight to Joi must be able to report
// Joi's own object.base for one. ApplyCtx takes a map and cannot express that.
func TestApplyValueJudgesANonObjectTopLevel(t *testing.T) {
	s := ObjectSchema(Object(map[string]*Rule{"name": String()}))

	out, err := s.ApplyValue(map[string]any{"name": "n"}, nil, Prefs{})
	if err != nil {
		t.Fatalf("an object must still pass: %v", err)
	}
	if m, _ := out.(map[string]any); m["name"] != "n" {
		t.Fatalf("normalized value lost: %#v", out)
	}

	_, err = s.ApplyValue([]any{}, nil, Prefs{})
	ve, ok := err.(*Error)
	if !ok {
		t.Fatalf("an array must be rejected by the SCHEMA, got %v", err)
	}
	d := ve.Details()
	if len(d) != 1 || d[0].Type != "object.base" {
		t.Fatalf("details: %#v", d)
	}
}
