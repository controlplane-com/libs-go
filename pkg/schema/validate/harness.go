package validate

// harness.go is the table-driven parity harness the fan-out agents reuse to port
// each resource's *.test.ts vectors. It has NO dependency on `testing` so it can
// live in the engine package and be called from any package's _test.go. A thin
// loop in the test file drives t.Run over the cases.
//
// Usage (in a _test.go):
//
//	cases := []validate.Case{
//	  {Name: "good", Input: map[string]any{...}, Ctx: &validate.Context{Org: "o"},
//	   Want: map[string]any{...normalized...}},
//	  {Name: "bad", Input: ..., WantErr: "is required"},
//	}
//	for _, c := range cases {
//	  t.Run(c.Name, func(t *testing.T) {
//	    if err := c.Run(rule); err != nil { t.Fatal(err) }
//	  })
//	}

import (
	"fmt"
	"reflect"
	"strings"
)

// Case is one parity vector: an input body + context/prefs → expect accept (with
// an optional normalized-output assertion) or reject (with a message substring).
type Case struct {
	Name  string
	Input any
	Ctx   *Context
	Prefs Prefs

	// WantErr, if non-empty, asserts the case is rejected and the error message
	// contains this substring (Joi test style: expect(error.message).toContain(...)).
	WantErr string

	// Want, if non-nil, asserts the accepted+normalized output deep-equals this
	// value (Joi: expect(result.value).toEqual(...)). Numbers should be float64.
	Want any

	// Check is an optional post-normalization assertion for cases where a full
	// deep-equal is awkward; return a non-nil error to fail.
	Check func(out any) error
}

// Run validates Case.Input against rule and checks the expectation. It returns a
// descriptive error on mismatch (nil = pass).
func (c Case) Run(rule *Rule) error {
	ctx := c.Ctx
	if ctx == nil {
		ctx = &Context{}
	}
	out, verr := Validate(rule, c.Input, ctx, c.Prefs)

	if c.WantErr != "" {
		if verr == nil {
			return fmt.Errorf("expected error containing %q, got success (value=%#v)", c.WantErr, out)
		}
		if !strings.Contains(verr.Message, c.WantErr) {
			return fmt.Errorf("expected error containing %q, got %q", c.WantErr, verr.Message)
		}
		return nil
	}
	if verr != nil {
		return fmt.Errorf("unexpected error: %s (type=%s path=%v)", verr.Message, verr.Type, verr.Path)
	}
	if c.Want != nil {
		if !reflect.DeepEqual(out, c.Want) {
			return fmt.Errorf("output mismatch:\n  got:  %#v\n  want: %#v", out, c.Want)
		}
	}
	if c.Check != nil {
		if err := c.Check(out); err != nil {
			return err
		}
	}
	return nil
}

// Validate runs a rule against a value with the given context/prefs, returning
// the normalized value and a *Error (nil on success). It is the single-value
// entry point (Joi's schema.validate(value, options)); Schema.ApplyCtx is the
// in-place object entry used by the CRUD engine.
func Validate(rule *Rule, value any, ctx *Context, prefs Prefs) (any, *Error) {
	if ctx == nil {
		ctx = &Context{}
	}
	st := &evalState{ctx: ctx, prefs: prefs}
	if parent, ok := value.(map[string]any); ok {
		st.parent = mapView(parent)
	}
	// Top-level presence: a top-level required/undefined value errors like Joi.
	if value == nil && rule.presence == presRequired {
		return nil, errAt(nil, "any.required", quote(label(nil))+" is required")
	}
	return rule.validateValue(value, value != nil, nil, st)
}
