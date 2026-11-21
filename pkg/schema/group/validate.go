package group

// validate.go attaches validation to the canonical group types via the Validator
// interface — the go-libs home of what used to be pkg/handlers/schema_group.go plus
// the group entries in pkg/handlers/schema_reflect_iam.go (RegisterType(MemberLink),
// RegisterRefine(GroupIdentityMatcher), RegisterReflect("group", Group)).
//
// group.ts reflects from Group; there is NO root object rule on the reflected
// schema (data-service's createSchema fork, .or() selector requirement, memberQuery
// fork and superuser self-removal guard live in the data-service GroupHandler).
//
// Ported here:
//   - MemberLink: LocalLink.custom(toKind('user','serviceaccount')) (group.ts:11).
//   - MemberLinks: the array-level .max(200).sort().custom(makeUnique()) (group.ts:
//     10-13) — kept in the nominal slice type's rule (not field tags) so the exact
//     Joi rule ORDER is preserved.
//   - GroupIdentityMatcher: the object .custom() that parses `expression` as JMESPath
//     or JavaScript (group.ts:20-36).

import (
	"github.com/dop251/goja"
	"github.com/jmespath/go-jmespath"
	"github.com/controlplane-com/libs-go/pkg/schema/base"
	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// ValidationRule reflects Group (group.ts: Base.keys({...})).
func (Group) ValidationRule() *validate.Rule { return validate.Reflect(Group{}) }

// ValidationRule ports `LocalLink.custom(toKind('user','serviceaccount'))`
// (group.ts:11): a local link that must resolve to a user or serviceaccount in the
// same org.
func (MemberLink) ValidationRule() *validate.Rule {
	return base.LocalLinkRule().Custom(base.ToKind("user", "serviceaccount"))
}

// ValidationRule ports the memberLinks array rule
// `Joi.array().items(LocalLink.custom(toKind(...))).max(200).sort().custom(makeUnique())`
// (group.ts:10-13) with the same operation order (max -> sort -> makeUnique), so the
// element rule comes from MemberLink and the array-level normalization lives here.
func (MemberLinks) ValidationRule() *validate.Rule {
	return validate.Array(MemberLink("").ValidationRule()).
		Max(200).
		Sort().
		Dedupe()
}

// ValidationRule reflects GroupIdentityMatcher and layers the expression-parse
// object .custom (group.ts:20): the `language` selects whether `expression` is
// parsed as JMESPath or JavaScript, throwing "Cannot parse ..." on a syntax error.
func (GroupIdentityMatcher) ValidationRule() *validate.Rule {
	return validate.RefineT(validate.Reflect(GroupIdentityMatcher{}), identityMatcherRefine)
}

// identityMatcherRefine ports the group identityMatcher object .custom (group.ts:20):
// parse the expression as JMESPath / JavaScript, mirroring the TS
// `jmespath.search`/`new Script()` throws.
func identityMatcherRefine(m *GroupIdentityMatcher, h *validate.Helper) error {
	lang, expr := string(m.Language), m.Expression

	switch lang {
	case "jmespath":
		if err := parseJMESPath(expr); err != nil {
			return h.Errorf("any.custom", "Cannot parse JMESPath expression: %s", err.Error())
		}
	case "javascript":
		if err := parseJSExpression(expr); err != nil {
			return h.Errorf("any.custom", "Cannot parse JavaScript expression: %s", err.Error())
		}
	}
	return nil
}

// parseJSExpression validates a JavaScript identityMatcher by actually PARSING it,
// standing in for TS's `new Script(expression)` (nodelibs/schema/src/group.ts:29-35,
// which throws "Cannot parse JavaScript expression: ..." on a SyntaxError).
//
// This used to be a byte-filter + bracket-balance scan approximating "V8's two
// reported SyntaxError categories", on the premise that "no embeddable JS engine is
// available in go-libs". That approximation accepted syntactically invalid programs TS
// rejects — `var var var` and `return 1 +` both stored happily (TS 400s) — and an
// unparseable matcher then silently matches NOTHING at authz time, because the
// evaluator swallows the compile error. A group rule that never fires, with no error at
// write time and none at evaluation, is the worst shape available.
//
// goja is a real ECMAScript parser, so this rejects what V8 rejects for ordinary
// programs. It is a PARSE, not an evaluation: nothing from the expression runs here
// (evaluation stays in the isolated-vm sidecar, which is what enforces the memory/time
// limits). Exotic disagreements between goja's and V8's grammar acceptance are possible
// at the margins; the alternative — a hand-rolled scanner — disagrees with V8 on
// ordinary programs, which is strictly worse.
func parseJSExpression(s string) error {
	if _, err := goja.Compile("", s, false); err != nil {
		return err
	}
	return nil
}

// isIllegalJSChar reports whether c cannot begin/continue any JavaScript token
// outside a string literal (so V8 raises "Invalid or unexpected token").
func isIllegalJSChar(c byte) bool {
	switch {
	case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9':
		return false
	case c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\f' || c == '\v':
		return false
	}
	switch c {
	case '_', '$', '.', ',', ';', ':', '+', '-', '*', '/', '%',
		'=', '!', '<', '>', '&', '|', '^', '~', '?':
		return false
	}
	return true
}

// parseJMESPath validates a JMESPath expression by actually COMPILING it, which is
// what TS does (nodelibs/schema/src/group.ts:17-36 runs jmespath.search({}, expr) and
// throws "Cannot parse JMESPath expression: ..." on failure).
//
// This used to be a bracket/quote BALANCE SCAN standing in for a parser, on the stated
// premise that go-jmespath "cannot be imported here without adding it to
// go-libs/go.mod". That premise was stale — the dependency is already a direct one of
// the data-service module and is imported by its authz matcher; only go-libs/go.mod
// lacked it, so the rule regressed when it moved here from pkg/handlers.
//
// The gap was not cosmetic. A balance scan accepts anything bracket-balanced, so
// `foo..bar` and `|||` were stored happily (TS 400s) — and then silently matched
// NOTHING at authz time, because the matcher swallows the compile error and returns
// false. A group whose membership rule never fires, with no error at write time and
// none at evaluation, is close to the worst failure shape available: the operator has
// every reason to believe the group works.
func parseJMESPath(s string) error {
	if _, err := jmespath.Compile(s); err != nil {
		return err
	}
	return nil
}
