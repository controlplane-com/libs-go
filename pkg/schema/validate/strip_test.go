package validate

import (
	"encoding/json"
	"testing"
)

type stripRow struct {
	A *bool  `json:"a,omitzero"`
	B string `json:"b,omitzero"`
}

// TestStripValidatesThenDrops pins Joi's .strip(), captured from joi@17.6.0 over
// Joi.object({a: Joi.boolean().strip(), b: Joi.string()}). The key is still judged,
// then removed, so neither evaluator may store or echo it.
func TestStripValidatesThenDrops(t *testing.T) {
	rule := ObjectOrdered([]string{"a", "b"}, map[string]*Rule{
		"a": Boolean().Strip(),
		"b": String(),
	})
	for _, c := range []struct {
		body     string
		wantType string // "" = accepted
	}{
		{`{"a":true,"b":"x"}`, ""},
		{`{"b":"x"}`, ""},
		{`{"a":"notbool"}`, "boolean.base"},
		{`{"a":null}`, "boolean.base"},
	} {
		m := map[string]any{}
		if err := json.Unmarshal([]byte(c.body), &m); err != nil {
			t.Fatal(err)
		}
		err := ObjectSchema(rule).Apply(m)
		checkStrip(t, "generic", c.body, c.wantType, err)
		if err == nil {
			if _, ok := m["a"]; ok {
				t.Errorf("generic %s: stripped key survived: %v", c.body, m)
			}
		}

		var row stripRow
		d, derr := DecodeInto(&row, Bytes(c.body))
		if derr != nil {
			checkStrip(t, "typed", c.body, c.wantType, derr)
			continue
		}
		var terr error
		if verr := d.Validate(rule, nil, Prefs{}); verr != nil {
			terr = verr
		}
		checkStrip(t, "typed", c.body, c.wantType, terr)
		if terr == nil && row.A != nil {
			t.Errorf("typed %s: stripped field survived: %v", c.body, *row.A)
		}
	}
}

// A default on a stripped key is dropped too: Joi answers {} for
// Joi.object({a: Joi.boolean().strip().default(false)}).validate({}).
func TestStripDropsDefault(t *testing.T) {
	rule := ObjectOrdered([]string{"a"}, map[string]*Rule{"a": Boolean().Strip().Default(false)})
	m := map[string]any{}
	if err := ObjectSchema(rule).Apply(m); err != nil {
		t.Fatal(err)
	}
	if len(m) != 0 {
		t.Errorf("generic: default on a stripped key was kept: %v", m)
	}
	var row stripRow
	d, err := DecodeInto(&row, Bytes(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	if verr := d.Validate(rule, nil, Prefs{}); verr != nil {
		t.Fatal(verr)
	}
	if row.A != nil {
		t.Errorf("typed: default on a stripped key was kept: %v", *row.A)
	}
}

// Joi deletes a stripped key inside the key loop, so presence relations evaluated
// afterwards do not see it: xor('a','b') with a stripped `a` and no `b` is object.missing.
func TestStripIsInvisibleToPresenceRelations(t *testing.T) {
	rule := ObjectOrdered([]string{"a", "b"}, map[string]*Rule{
		"a": Boolean().Strip(),
		"b": String(),
	}).Xor("a", "b")
	m := map[string]any{"a": true}
	checkStrip(t, "generic", `{"a":true}`, "object.missing", ObjectSchema(rule).Apply(m))

	var row stripRow
	d, err := DecodeInto(&row, Bytes(`{"a":true}`))
	if err != nil {
		t.Fatal(err)
	}
	var terr error
	if verr := d.Validate(rule, nil, Prefs{}); verr != nil {
		terr = verr
	}
	checkStrip(t, "typed", `{"a":true}`, "object.missing", terr)
}

func checkStrip(t *testing.T, which, body, want string, err error) {
	t.Helper()
	if want == "" {
		if err != nil {
			t.Errorf("%s %s: rejected: %v", which, body, err)
		}
		return
	}
	ve, ok := err.(*Error)
	if !ok || ve == nil {
		t.Errorf("%s %s: accepted (or wrong error type %T), want %s", which, body, err, want)
		return
	}
	if got := ve.Details()[0].Type; got != want {
		t.Errorf("%s %s: type %q, want %q", which, body, got, want)
	}
}
