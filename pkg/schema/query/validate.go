package query

// validate.go attaches the parts of nodelibs/schema/src/query.ts that a struct
// tag cannot express: the Term `.xor('tag','property','rel')`, the
// `value.when('op', …)` presence, the value alternatives, and the boolean-tag
// coercion. Everything else (enums, trim, defaults) lives on the tags.
//
// Only Term needs a ValidationRule; Query/Spec/SpecSort are reached by
// reflection over their tags.

import (
	"strconv"

	"github.com/controlplane-com/libs-go/pkg/schema/validate"
)

// noValueOps are the two operators that take no `value` — the complement of the
// `.when('op', {is: valid('=','>','>=','<','<=','!=','~','=~','contains'), then:
// required()})` list in query.ts. Expressed as the complement so an ABSENT op
// (which Joi defaults to '=') still requires a value: the engine resolves a
// `when` ref against the raw sibling, so an `Is` on the value-ops would not match
// an omitted op and the field would wrongly become optional.
var noValueOps = []any{string(TermOpExists), string(TermOpNotExists)}

// ValidationRule ports the Term object rule (query.ts:6-37): the tagged fields
// plus the value presence/type rules, the xor, and the tag-value coercion.
func (Term) ValidationRule() *validate.Rule {
	return validate.Reflect(Term{}).
		Fork("value", func(*validate.Rule) *validate.Rule {
			return validate.Alternatives(
				validate.String(), validate.Number(), validate.Boolean(), validate.DateRule(),
			).When("op", validate.When{
				Is:        validate.Any().Valid(noValueOps...),
				Then:      validate.Any().Optional(),
				Otherwise: validate.Any().Required(),
			})
		}).
		Xor("tag", "property", "rel").
		CustomTyped(coerceTagValue)
}

// coerceTagValue ports the Term `.custom()` at query.ts:33-36: a tag term's
// boolean value is compared as its STRING form, because the write path stores
// boolean tag values stringified (base.ts Tags custom). A query arrives as a tree
// (the -query body) or as a typed Term (a group's memberQuery, a policy's
// targetQuery), so it serves both forms.
func coerceTagValue(value any, _ *validate.Helper) (any, error) {
	switch term := value.(type) {
	case map[string]any:
		if tag, _ := term["tag"].(string); tag != "" {
			if b, isBool := term["value"].(bool); isBool {
				term["value"] = strconv.FormatBool(b)
			}
		}
	case *Term:
		if term.Tag != "" {
			if b, isBool := term.Value.(bool); isBool {
				term.Value = strconv.FormatBool(b)
			}
		}
	}
	return value, nil
}

// Specialize builds the per-kind Query rule — the port of catalog.ts
// specializeQuerySchema (L20-43). A registered kind does not validate against the
// generic Query schema: its `property`, `rel` and `sort.by` are narrowed to what that
// kind actually exposes, `kind` becomes required, and `spec.sort` gains a default of
// the kind's natural order.
//
// Doing it in the SCHEMA rather than as a post-validation check is what produces Joi's
// `any.only` error naming the alternatives ("must be one of [name, id, …]") instead of
// a bespoke sentence. The caller gets the list of properties they could have used,
// which is the whole point of the message.
//
// props/rels are ORDERED: the order reaches the user as the order of the valids in
// that message, and TS's comes from Object.keys() over the registration literal.
func Specialize(ordProp string, props, rels []string) *validate.Rule {
	propValids := anyOf(append(append([]string{}, props...), "*"))

	term := Term{}.ValidationRule().
		Fork("property", func(*validate.Rule) *validate.Rule {
			return validate.String().Valid(propValids...)
		}).
		Fork("rel", func(*validate.Rule) *validate.Rule {
			// A kind with no relations forbids `rel` outright rather than accepting
			// a term that could never match anything.
			if len(rels) == 0 {
				return validate.String().Forbidden()
			}
			return validate.String().Valid(anyOf(rels)...)
		})

	return validate.Reflect(Query{}).
		Fork("kind", func(r *validate.Rule) *validate.Rule { return r.Required() }).
		Fork("spec", func(spec *validate.Rule) *validate.Rule {
			return spec.
				Fork("terms", func(*validate.Rule) *validate.Rule { return validate.Array(term) }).
				Fork("sort", func(sort *validate.Rule) *validate.Rule {
					return sort.
						Fork("by", func(by *validate.Rule) *validate.Rule {
							return by.Valid(anyOf(props)...)
						}).
						// A fresh map per use: Joi's default here is a FUNCTION, and a
						// shared map would be handed to every query and mutated by one.
						Default(func() any {
							return map[string]any{"by": ordProp, "order": "asc"}
						})
				})
		})
}

func anyOf(ss []string) []any {
	out := make([]any, 0, len(ss))
	for _, s := range ss {
		out = append(out, s)
	}
	return out
}
