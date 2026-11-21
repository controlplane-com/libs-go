package validate

import "testing"

// TestForkReordersAfterARef pins Joi's reference-driven key ordering, which decides
// which key abortEarly blames.
//
// `Joi.object({a,b,c,d,e}).fork(['b'], s => s.default(Joi.ref('d')))` describes as
// a,c,d,b,e: the referencing key moves to just after its referent, so the referent is
// validated first. Captured from joi@17.6.0.
func TestForkReordersAfterARef(t *testing.T) {
	mk := func() *Rule {
		return ObjectOrdered([]string{"a", "b", "c", "d", "e"}, map[string]*Rule{
			"a": String(), "b": String(), "c": String(), "d": String(), "e": String(),
		})
	}
	join := func(r *Rule) string {
		out := ""
		for i, k := range r.keyOrder {
			if i > 0 {
				out += ","
			}
			out += k
		}
		return out
	}
	for _, c := range []struct {
		name string
		got  string
		want string
	}{
		{"base", join(mk()), "a,b,c,d,e"},
		// A fork that adds no ref leaves the order alone.
		{"required", join(mk().Fork("a", func(r *Rule) *Rule { return r.Required() })), "a,b,c,d,e"},
		{"b refs d", join(mk().Fork("b", func(r *Rule) *Rule { return r.Default(RefTo("d")) })), "a,c,d,b,e"},
		{"a refs e", join(mk().Fork("a", func(r *Rule) *Rule { return r.Default(RefTo("e")) })), "b,c,d,e,a"},
		// A ref to a key that is not a sibling cannot be ordered against.
		{"a refs absent", join(mk().Fork("a", func(r *Rule) *Rule { return r.Default(RefTo("zz")) })), "a,b,c,d,e"},
	} {
		if c.got != c.want {
			t.Errorf("%s: order %q, want %q", c.name, c.got, c.want)
		}
	}
}

// TestForkedRefKeyIsBlamedSecond is the behavior the ordering exists for: the referent
// is reported, not the referencing key.
func TestForkedRefKeyIsBlamedSecond(t *testing.T) {
	r := ObjectOrdered([]string{"a", "b"}, map[string]*Rule{
		"a": String(),
		"b": String(),
	}).Fork("a", func(r *Rule) *Rule { return r.Default(RefTo("b")) }).
		Fork("b", func(r *Rule) *Rule { return r.Required() })

	err := ObjectSchema(r).Apply(map[string]any{"a": 1})
	if err == nil {
		t.Fatal("no error")
	}
	verr, ok := err.(*Error)
	if !ok {
		t.Fatalf("error type %T, want *Error", err)
	}
	if got := verr.Details()[0].Type; got != "any.required" {
		t.Errorf("type %q, want any.required (the REFERENT, checked first)", got)
	}
}
