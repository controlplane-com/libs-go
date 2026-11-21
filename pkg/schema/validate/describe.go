package validate

// describe.go ports Joi's `schema.describe()` over the compiled *Rule tree. It is
// what GET /catalog/schema/{kind} serves, and the TS service serves the Joi
// document verbatim — so the shape here is dictated by Joi, not by the Go engine:
// a node is {type, flags, rules, allow, invalid, metas, dependencies, keys,
// patterns, whens, items, matches}, in that order (derived from the oracle's own
// output; Joi emits object keys in a fixed construction order and a client diffing
// two documents sees key order).
//
// Nothing here runs during validation: Describe walks the already-built rule, and
// rules are cached per type.

import (
	"encoding/json"
	"regexp"
	"strings"
)

// descOrder is Joi's describe() key order for a schema node.
var descOrder = []string{
	"type", "flags", "rules", "allow", "invalid", "metas",
	"dependencies", "keys", "patterns", "whens", "items", "matches",
}

// ordered is a JSON object that marshals its keys in insertion order. Joi's
// describe() is order-bearing (both the node keys and, for an object, the
// DECLARATION order of `keys`), and Go's map marshalling sorts — so every object
// in the document is built through this.
type ordered struct {
	keys []string
	vals map[string]any
}

func newOrdered() *ordered { return &ordered{vals: map[string]any{}} }

func (o *ordered) set(k string, v any) {
	if _, ok := o.vals[k]; !ok {
		o.keys = append(o.keys, k)
	}
	o.vals[k] = v
}

func (o *ordered) has(k string) bool { _, ok := o.vals[k]; return ok }

func (o *ordered) len() int { return len(o.keys) }

func (o *ordered) MarshalJSON() ([]byte, error) {
	var b strings.Builder
	b.WriteByte('{')
	for i, k := range o.keys {
		if i > 0 {
			b.WriteByte(',')
		}
		kb, err := json.Marshal(k)
		if err != nil {
			return nil, err
		}
		b.Write(kb)
		b.WriteByte(':')
		vb, err := json.Marshal(o.vals[k])
		if err != nil {
			return nil, err
		}
		b.Write(vb)
	}
	b.WriteByte('}')
	return []byte(b.String()), nil
}

// Describe returns the schema's Joi `describe()` document. Nil for a legacy
// Fields-only schema (no combinator root).
func (s *Schema) Describe() any {
	r := s.root()
	if r == nil {
		return nil
	}
	return r.Describe()
}

// Describe returns the rule's Joi `describe()` document.
func (r *Rule) Describe() any { return r.describe(false) }

// describe renders one node. isTest marks an `is:` position of a .when()/
// conditional, where Joi COMPILES a literal into `any().valid(override, lit)
// .required()` — the Go builders spell that literal as a typed .Valid(lit), so the
// compiled form has to be reconstructed here or every conditional in the document
// diverges from the oracle.
func (r *Rule) describe(isTest bool) any {
	if r == nil {
		return nil
	}
	if isTest && r.isCompiledLiteral() {
		n := newOrdered()
		n.set("type", "any")
		fl := newOrdered()
		fl.set("only", true)
		fl.set("presence", "required")
		n.set("flags", fl)
		n.set("allow", append([]any{map[string]any{"override": true}}, r.valids...))
		return n
	}

	n := newOrdered()
	n.set("type", r.typeName())

	if fl, emit := r.describeFlags(); emit || fl.len() > 0 {
		n.set("flags", fl)
	}
	if rules := r.describeRules(); len(rules) > 0 {
		n.set("rules", rules)
	}
	if allow := r.describeAllow(); len(allow) > 0 {
		n.set("allow", allow)
	}
	if len(r.invalids) > 0 {
		n.set("invalid", describeValues(r.invalids))
	}
	if metas := r.describeMetas(); len(metas) > 0 {
		n.set("metas", metas)
	}
	if deps := r.describeDependencies(); len(deps) > 0 {
		n.set("dependencies", deps)
	}
	// `Joi.object({})` describes an EMPTY keys object; only `Joi.object()` (no key
	// schema at all — the keyless form) omits the member entirely.
	if len(r.keyOrder) > 0 || (r.typ == tObject && !r.keyless) {
		keys := newOrdered()
		for _, k := range r.keyOrder {
			keys.set(k, r.keys[k].describe(false))
		}
		n.set("keys", keys)
	}
	if len(r.patterns) > 0 {
		pats := make([]any, 0, len(r.patterns))
		for _, p := range r.patterns {
			pat := newOrdered()
			pat.set("schema", p.key.describe(false))
			pat.set("rule", p.val.describe(false))
			pats = append(pats, pat)
		}
		n.set("patterns", pats)
	}
	if whens := r.describeWhens(); len(whens) > 0 {
		n.set("whens", whens)
	}
	if len(r.arrItems) > 0 {
		items := make([]any, 0, len(r.arrItems))
		for _, it := range r.arrItems {
			items = append(items, it.describe(false))
		}
		n.set("items", items)
	}
	if matches := r.describeMatches(); len(matches) > 0 {
		n.set("matches", matches)
	}
	return n
}

// isCompiledLiteral reports whether the rule is exactly a one-value .Valid() with
// no other constraint — the shape a Joi literal `is:` compiles to. A multi-value
// or otherwise-constrained `is` was written as a real schema in TS too, so it is
// described as itself.
func (r *Rule) isCompiledLiteral() bool {
	if len(r.valids) != 1 || len(r.allowed) > 0 || len(r.applied) > 0 {
		return false
	}
	return r.presence == presDefault && !r.hasDefault && len(r.keyOrder) == 0 &&
		len(r.arrItems) == 0 && len(r.alts) == 0 && len(r.whens) == 0
}

func (r *Rule) typeName() string {
	switch r.typ {
	case tString:
		return "string"
	case tNumber:
		return "number"
	case tBoolean:
		return "boolean"
	case tDate:
		return "date"
	case tArray:
		return "array"
	case tObject:
		return "object"
	case tAlternatives:
		return "alternatives"
	default:
		return "any"
	}
}

// describeFlags renders the node's flags. emit forces an (even empty) flags object:
// a FUNCTION default has no JSON encoding — Joi puts the function itself on
// flags.default and JSON.stringify drops the key — so the oracle emits `flags:{}`
// where a rule's only flag is a computed default, and so must this.
func (r *Rule) describeFlags() (fl *ordered, emit bool) {
	fl = newOrdered()
	if len(r.valids) > 0 {
		fl.set("only", true)
	}
	if r.hasDefault {
		if v, ok := literalDefault(r.defaultVal); ok {
			fl.set("default", describeValue(v))
		} else {
			emit = true
		}
	}
	switch r.presence {
	case presRequired:
		fl.set("presence", "required")
	case presOptional:
		fl.set("presence", "optional")
	case presForbidden:
		fl.set("presence", "forbidden")
	}
	if r.strip {
		fl.set("result", "strip")
	}
	if r.description != "" {
		fl.set("description", r.description)
	}
	if r.label != "" {
		fl.set("label", r.label)
	}
	if r.unknown == unkAllow {
		fl.set("unknown", true)
	}
	if r.id != "" {
		fl.set("id", r.id)
	}
	return fl, emit
}

// literalDefault reports the default's JSON-encodable value. A computed default
// (any func form) has none — see describeFlags.
func literalDefault(v any) (any, bool) {
	switch v.(type) {
	case func() any, func(map[string]any) any:
		return nil, false
	}
	return v, true
}

func (r *Rule) describeRules() []any {
	if len(r.applied) == 0 {
		return nil
	}
	out := make([]any, 0, len(r.applied))
	for _, ar := range r.applied {
		e := newOrdered()
		e.set("name", ar.name)
		if ar.args != nil {
			e.set("args", describeArgs(ar.args))
		}
		out = append(out, e)
	}
	return out
}

// describeArgs renders a rule's args. Key order inside args follows Joi's own
// argument order for the rules that take more than one (pattern: regex then
// options).
func describeArgs(args map[string]any) *ordered {
	o := newOrdered()
	for _, k := range []string{"enabled", "direction", "sign", "limit", "regex", "options"} {
		v, ok := args[k]
		if !ok {
			continue
		}
		if k == "options" {
			if opts, isMap := v.(map[string]any); isMap {
				o.set(k, describeRuleOptions(opts))
				continue
			}
		}
		o.set(k, describeValue(v))
	}
	for k, v := range args {
		if !o.has(k) {
			o.set(k, describeValue(v))
		}
	}
	return o
}

// describeRuleOptions renders a rule's `options` object, keeping Joi's own key
// order (order then by, for sort) and expanding a ref to the full descriptor.
func describeRuleOptions(opts map[string]any) *ordered {
	o := newOrdered()
	for _, k := range []string{"order", "by", "urlSafe", "paddingRequired", "invert"} {
		v, ok := opts[k]
		if !ok {
			continue
		}
		if ref, isRef := v.(*Ref); isRef {
			o.set(k, describeSortByRef(ref.Key))
			continue
		}
		o.set(k, describeValue(v))
	}
	for k, v := range opts {
		if !o.has(k) {
			o.set(k, describeValue(v))
		}
	}
	return o
}

// describeAllow renders the node's accepted literal set. Joi folds .valid() into
// `allow` (with `only: true` on flags) and prefixes the `{override:true}` sentinel
// only when the valid set REPLACES an inherited one; a plain .valid() on a fresh
// type carries no sentinel, which is what the builders here produce.
func (r *Rule) describeAllow() []any {
	if len(r.allowed) == 0 && len(r.valids) == 0 {
		return nil
	}
	vals := append(append([]any(nil), r.valids...), r.allowed...)
	return describeValues(vals)
}

func describeValues(vals []any) []any {
	out := make([]any, 0, len(vals))
	for _, v := range vals {
		out = append(out, describeValue(v))
	}
	return out
}

// describeValue renders one literal/arg value in Joi's describe encoding: a Ref
// becomes {ref:{path:[…]}}, a regex its JS literal source, everything else itself.
func describeValue(v any) any {
	switch t := v.(type) {
	case *Ref:
		return describeRef(t.Key)
	case *regexp.Regexp:
		return jsRegexLiteral(t)
	case JSRegex:
		return "/" + escapeJSRegexSlashes(string(t)) + "/"
	case map[string]any:
		return t
	default:
		return v
	}
}

// describeSortByRef renders the `by` of Joi's array.sort options. Joi describes a
// ref there as the FULL ref descriptor (ref.describe() on an option), not the
// compact {ref:{path}} it uses for a min/max limit — an inconsistency in Joi that
// the oracle's output makes binding.
func describeSortByRef(key string) any {
	path := strings.Split(key, ".")
	o := newOrdered()
	o.set("adjust", nil)
	o.set("in", false)
	o.set("iterables", nil)
	o.set("map", nil)
	o.set("separator", ".")
	o.set("type", "value")
	o.set("ancestor", 0)
	o.set("path", path)
	o.set("depth", len(path))
	o.set("key", key)
	o.set("root", path[0])
	o.set("display", "ref:."+key)
	return o
}

func describeRef(key string) any {
	ref := newOrdered()
	ref.set("path", strings.Split(key, "."))
	out := newOrdered()
	out.set("ref", ref)
	return out
}

// jsRegexLiteral renders a Go regexp as the JS literal Joi reports (`/src/flags`).
// Go carries case-insensitivity as a leading `(?i)` inline flag; JS carries it as
// the trailing `i`, so it is lifted.
func jsRegexLiteral(re *regexp.Regexp) string {
	src := re.String()
	flags := ""
	if strings.HasPrefix(src, "(?i)") {
		src = src[len("(?i)"):]
		flags = "i"
	}
	return "/" + escapeJSRegexSlashes(src) + "/" + flags
}

// escapeJSRegexSlashes escapes the `/`s that would terminate a JS regex literal.
// Go regex source carries them bare; a `/` inside a character class does not
// terminate the literal and the oracle leaves those alone, so the class is tracked.
func escapeJSRegexSlashes(src string) string {
	var b strings.Builder
	inClass := false
	for i := 0; i < len(src); i++ {
		c := src[i]
		switch {
		case c == '\\' && i+1 < len(src):
			// Go's regexp source escapes a class dash; no TS pattern does, and the
			// escape is meaningless inside a class, so it is dropped to keep the
			// rendered literal identical to the Joi source.
			if inClass && src[i+1] == '-' {
				i++
				b.WriteByte('-')
				continue
			}
			b.WriteByte(c)
			i++
			b.WriteByte(src[i])
			continue
		case c == '[':
			inClass = true
		case c == ']':
			inClass = false
		case c == '/' && !inClass:
			b.WriteByte('\\')
		}
		b.WriteByte(c)
	}
	return b.String()
}

func (r *Rule) describeMetas() []any {
	out := make([]any, 0, len(r.metas))
	for _, m := range r.metas {
		out = append(out, m)
	}
	return out
}

func (r *Rule) describeDependencies() []any {
	var out []any
	add := func(rel string, groups [][]string) {
		for _, peers := range groups {
			d := newOrdered()
			d.set("rel", rel)
			d.set("peers", peers)
			out = append(out, d)
		}
	}
	add("or", r.or)
	add("xor", r.xor)
	add("oxor", r.oxor)
	add("nand", r.nand)
	return out
}

func (r *Rule) describeWhens() []any {
	if len(r.whens) == 0 {
		return nil
	}
	out := make([]any, 0, len(r.whens))
	for _, w := range r.whens {
		e := newOrdered()
		e.set("ref", describeRef(w.ref).(*ordered).vals["ref"])
		if w.is != nil {
			e.set("is", w.is.describe(true))
		}
		if w.not != nil {
			e.set("not", w.not.describe(true))
		}
		if w.then != nil {
			e.set("then", w.then.describe(false))
		}
		if w.otherwise != nil {
			e.set("otherwise", w.otherwise.describe(false))
		}
		out = append(out, e)
	}
	return out
}

func (r *Rule) describeMatches() []any {
	var out []any
	for _, alt := range r.alts {
		m := newOrdered()
		m.set("schema", alt.describe(false))
		out = append(out, m)
	}
	if c := r.altCond; c != nil {
		m := newOrdered()
		m.set("ref", describeRef(c.ref).(*ordered).vals["ref"])
		if c.is != nil {
			m.set("is", c.is.describe(true))
		}
		if c.then != nil {
			m.set("then", c.then.describe(false))
		}
		if c.otherwise != nil {
			m.set("otherwise", c.otherwise.describe(false))
		}
		out = append(out, m)
	}
	return out
}
