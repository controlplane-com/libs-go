package validate

// rule.go is the Joi-equivalent combinator engine — a faithful Go port of the
// runtime behaviour of the `joi` library as it is used across
// nodelibs/schema/src/*.ts. It coexists with the legacy declarative FieldRule
// path in validate.go: a *Schema may carry a *Rule Root (combinator) OR the old
// Fields/Custom slices. base.go:runValidate drives both through Schema.Apply /
// Schema.ApplyCtx.
//
// Design:
//   - The engine lives in pkg/validate with ZERO dependency on pkg/handlers.
//   - Schema DEFINITIONS are authored in pkg/handlers (so they can call links.go)
//     using the fluent builders here; a CustomFn closure carries the irreducible
//     `.custom()` logic and receives a *Helper carrying the validation Context.
//   - Validation is faithful to Joi defaults as the TS service invokes them:
//     convert:true (coercion), abortEarly:true (first error wins), unknown keys
//     REJECTED by default, presence/.default honoured. Per-call-site overrides
//     (allowUnknown/stripUnknown/presence:'optional') are threaded via Prefs.

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// ruleType is the primitive Joi type of a Rule.
type ruleType int

const (
	tAny ruleType = iota
	tString
	tNumber
	tBoolean
	tDate
	tArray
	tObject
	tAlternatives
)

// presence mirrors Joi presence: default (honour required/optional), required,
// optional, forbidden.
type presence int

const (
	presDefault presence = iota
	presRequired
	presOptional
	presForbidden
)

// unknownMode is an object's policy for keys not in its schema.
type unknownMode int

const (
	unkInherit unknownMode = iota // no explicit policy on the schema
	unkReject                     // Joi default when validated with default prefs
	unkAllow                      // .unknown(true) / allowUnknown pref
	unkStrip                      // .unknown('strip') / stripUnknown pref
)

// Context is the Joi validation context (makeJoiContext in base.ts): org, gvc and
// HTTP method. It is threaded to every CustomFn via *Helper.
type Context struct {
	Org    string
	Gvc    string
	Method string // POST | PUT | PATCH
}

// Prefs are per-call-site preference overrides (the options object passed to
// schema.validate). The zero value = Joi defaults (reject unknown, honour
// presence). Specific TS call sites set these (workload.ts:282, deployment
// processor, secret.ts, orgMember invite, inviteToOrg).
type Prefs struct {
	// There is deliberately no `presence:'optional'` preference: in Joi it only sets
	// the DEFAULT presence, which is already optional, and never relaxes a key's
	// explicit .required() — so it is a no-op, and a port that relaxed requireds
	// under it accepted bodies TS rejects.
	// AllowUnknown maps to allowUnknown:true — keep unknown keys.
	AllowUnknown bool
	// StripUnknown maps to stripUnknown:true — drop unknown keys (wins over
	// AllowUnknown, matching Joi).
	StripUnknown bool
}

// CustomFn is a `.custom((value, helpers) => ...)` port. It receives the current
// (already type-coerced) value and returns the possibly-transformed value, or an
// error to fail validation. Returning a *Error preserves type/path; any other
// error is wrapped with the field's path.
type CustomFn func(value any, h *Helper) (any, error)

// Helper is the second argument to a CustomFn — Joi's `helpers`. Original is the
// pre-coercion value (helpers.original); Ctx is prefs.context.
type Helper struct {
	Ctx      *Context
	Original any
	Path     []string
	// present, for a container custom, reports whether a key of the object was set
	// (sent, or defaulted) — the typed twin of `_, ok := obj[key]`.
	present func(key string) bool
	// st is the validation state the custom runs in, for Ancestor().
	st *evalState
}

// Ancestor returns the object n levels above the value being validated, as a generic
// map: 0 is the immediate parent object, 1 its parent, and so on — Joi's
// helpers.state.ancestors. nil when there is no such level.
//
// Arrays do not count as a level (an element shares its array's state), so a custom on
// `spec.containers[0]` reads the container with Ancestor(0), the spec with Ancestor(1)
// and the document with Ancestor(2). It exists because a cross-field check often has to
// read a grandparent — a container rule needing the workload's own tags — which the
// immediate-parent-only view could not reach at all.
func (h *Helper) Ancestor(n int) map[string]any {
	if h == nil || n < 0 {
		return nil
	}
	st := h.st
	for i := 0; i < n && st != nil; i++ {
		st = st.up
	}
	if st == nil || st.parent == nil {
		return nil
	}
	return st.parent.asMap()
}

// Present reports whether key was set on the object a container custom runs on:
// sent by the client, or given a default. Always false for a leaf custom.
func (h *Helper) Present(key string) bool { return h != nil && h.present != nil && h.present(key) }

// Errorf raises a typed validation error anchored at the field being validated
// (helpers.error / thrown Error in Joi customs).
func (h *Helper) Errorf(typ, format string, args ...any) *Error {
	return &Error{Type: typ, Message: fmt.Sprintf(format, args...), Path: append([]string(nil), h.Path...)}
}

// Label is the field's Joi label ("data.payload", or "value" at the root) — for a
// custom composing its own message, so it names the field the caller wrote instead of
// a hardcoded one.
func (h *Helper) Label() string { return label(h.Path) }

// WithCtx adds context keys to an error a custom is about to return, for the cases
// where a custom has to raise one of Joi's OWN error types by hand and must carry the
// context that type always has (a string.pattern.base carries `regex` and `value`).
func (h *Helper) WithCtx(e *Error, kv ...any) *Error {
	// Seed label/key exactly as errAt would: a non-nil Ctx tells the envelope the
	// error is already anchored, so whatever is missing here is missing for good.
	if e.Ctx == nil {
		e = errCtx(e, "label", label(e.Path))
		if len(e.Path) > 0 {
			e = errCtx(e, "key", pathSeg(e.Path[len(e.Path)-1]))
		}
	}
	for i := 0; i+1 < len(kv); i += 2 {
		k, _ := kv[i].(string)
		e = errCtx(e, k, kv[i+1])
	}
	return e
}

// ErrorAt is Errorf anchored at a path RELATIVE to the node being validated, for a
// cross-field check that must report against a NESTED field rather than the object it
// runs on. rel segments are object keys (string) or array indices (int); value is the
// value at that path (Joi's context.value).
//
// It exists because a check needing the parent's data still has to blame the child:
// the container cross-checks read the workload's options, so they run on the workload,
// but Joi reports them at `spec.containers[0]` — which is the field the caller has to
// fix, and the only thing an editor can highlight.
func (h *Helper) ErrorAt(rel []any, value any, typ, format string, args ...any) *Error {
	path := append([]string(nil), h.Path...)
	for _, seg := range rel {
		switch v := seg.(type) {
		case int:
			path = append(path, idxSeg(v))
		case string:
			path = append(path, v)
		default:
			path = append(path, fmt.Sprint(v))
		}
	}
	e := &Error{Type: typ, Message: fmt.Sprintf(format, args...), Path: path}
	e.Ctx = map[string]any{"label": label(path), "value": value}
	if len(path) > 0 {
		e.Ctx["key"] = pathSeg(path[len(path)-1])
	}
	return e
}

// metaFlags carries the .meta() flags the engine acts on.
type metaFlags struct {
	readOnly     bool
	unmodifiable bool
}

// whenClause is one `.when(ref, {is|not, then, otherwise})`.
type whenClause struct {
	ref       string
	is        *Rule
	not       *Rule
	then      *Rule
	otherwise *Rule
}

// altBranch is one conditional alternative (`Joi.alternatives().conditional`).
type altBranch struct {
	ref       string
	is        *Rule
	then      *Rule
	otherwise *Rule
}

// Rule is a compiled Joi schema node. Build via the constructors (String,
// Number, …) and fluent methods; each fluent method returns a copy (Joi schemas
// are immutable), so shared sub-schema symbols can be safely re-decorated.
type Rule struct {
	typ ruleType

	presence   presence
	strip      bool // .strip(): validate, then drop the key
	hasDefault bool
	defaultVal any    // literal, *Ref, or func() any
	label      string // .label() override (rare)
	id         string // .id() — describe()-only, Joi's flags.id
	meta       metaFlags
	// description is .description() text: describe()-only, no validation effect.
	description string
	// metas is the ORDERED .meta({...}) list, as Joi reports it. meta above holds
	// the two flags the engine acts on; everything else is carried for describe().
	metas []map[string]any
	// applied records the fluent rule methods in APPLICATION order, because Joi's
	// describe() reports `rules` as an ordered array and the flattened fields below
	// lose that order. Build-time only (rules are cached), so validation is untouched.
	applied []appliedRule

	allowed  []any // .allow(...) — bypass values accepted verbatim
	valids   []any // .valid(...) — only these accepted (any.only)
	invalids []any // .invalid(...)
	customs  []CustomFn

	// string
	strTrim, strLower, strUpper          bool
	strFromNumber                        bool // coerce a JSON number to its string form (JoiNumberToString)
	strEmail, strURI, strUUID, strDomain bool
	strDomainTLDs                        bool
	strBase64                            bool
	// strURIScheme is the `|`-joined alternation of `.uri({scheme:[…]})`. It is the
	// string Joi reports verbatim in the string.uriCustomScheme context.
	strURIScheme           string
	strMin, strMax, strLen *int
	strRegex               *regexp.Regexp
	strRegexInvert         bool
	// strRegexSrc is the JAVASCRIPT source to report when it differs from the Go
	// pattern. RE2 has no lookahead, so a Joi regex using one has to be split into
	// a Go regex plus a guard — but the message and the describe document must still
	// quote the regex the TS schema was written with.
	strRegexSrc string

	// number
	numMin, numMax         *float64
	numMinRef, numMaxRef   string
	numInt, numPos, numNeg bool
	numPort                bool
	numPrecision           *int

	// date
	dateIso bool // .iso() — the value must be an ISO-8601 string

	// array
	arrItems    []*Rule
	arrMin      *int
	arrMax      *int
	arrLen      *int
	arrSort     bool
	arrSortBy   string
	arrUnique   bool
	arrUniqueBy string
	// arrUniquePath marks the STRING-comparator spelling of .unique(). Joi stamps the
	// comparator into the error context as `path` only for that form; the function
	// form (`.unique((a,b) => a.k === b.k)`) carries no path at all.
	arrUniquePath bool
	arrDedupe     bool
	arrDedupeBy   string
	arrSingle     bool

	// object
	keys     map[string]*Rule
	keyOrder []string
	unknown  unknownMode
	// patterns are the object's .pattern(keySchema, valSchema) clauses, IN ORDER: Joi
	// tries each against an undeclared key and the FIRST match wins (keys.js breaks
	// unless the pattern is fallthrough). A single pair silently dropped the second
	// pattern — envoyHttp declares two on one object.
	patterns []keyPattern
	objMin   *int // .min(n) — minimum number of keys
	objMax   *int // .max(n) — maximum number of keys
	or       [][]string
	xor      [][]string
	oxor     [][]string
	nand     [][]string

	// alternatives
	alts    []*Rule
	altCond *altBranch

	// conditional (applies to any typed rule)
	whens []whenClause

	// keyless marks `Joi.object()` — an object with no key schema, which accepts
	// unknown keys without carrying an `unknown` flag (see Object). It does NOT apply
	// once a .pattern() is attached: the pattern is then the key schema.
	keyless bool

	// untypedCustom marks a container rule whose customs were written for the
	// generic shapes (a map or []any); the typed walk refuses such a rule rather
	// than hand its customs a pointer they would silently ignore.
	untypedCustom bool

	// msgOverrides is Joi's `.message(text)`: the message to substitute for the error
	// of a given Joi code, keyed by that code. Only the message changes — the error
	// type and its context are the rule's own, which is what clients match on.
	msgOverrides map[string]string

	// resolved marks a rule whose .when() branches have already been merged in, so the
	// second resolveEffective on the same value is a no-op. Set only on the freshly
	// cloned result of resolveEffective; a schema-built rule never carries it.
	resolved bool
}

// appliedRule is one entry of Joi's describe() `rules` array: the rule's Joi name
// and its args (nil when the rule takes none).
type appliedRule struct {
	name string
	args map[string]any
}

// addRule records a fluent rule application on the (already cloned) rule.
func (r *Rule) addRule(name string, args map[string]any) {
	r.applied = append(r.applied, appliedRule{name: name, args: args})
}

// isBareType reports whether the rule constrains nothing beyond its base type —
// i.e. it is a plain Joi.string()/number()/… . Used to decide which of Joi's two
// alternatives error shapes to raise (alternatives.types vs alternatives.match).
func (r *Rule) isBareType() bool {
	return len(r.allowed) == 0 && len(r.valids) == 0 && len(r.invalids) == 0 &&
		len(r.customs) == 0 && len(r.whens) == 0 && len(r.alts) == 0 &&
		len(r.keys) == 0 && len(r.arrItems) == 0 &&
		!r.strEmail && !r.strURI && !r.strUUID && !r.strDomain && !r.strDomainTLDs &&
		!r.strBase64 &&
		r.strMin == nil && r.strMax == nil && r.strLen == nil && r.strRegex == nil &&
		r.numMin == nil && r.numMax == nil && r.numPrecision == nil &&
		!r.numInt && !r.numPos && !r.numNeg && !r.numPort &&
		r.arrMin == nil && r.arrMax == nil && r.arrLen == nil &&
		r.objMin == nil && r.objMax == nil &&
		len(r.or) == 0 && len(r.xor) == 0 && len(r.oxor) == 0 && len(r.nand) == 0
}

// Ref is a Joi.ref('sibling') — resolved against the immediate parent object.
type Ref struct{ Key string }

// RefTo builds a Joi.ref.
func RefTo(key string) *Ref { return &Ref{Key: key} }

func (r *Rule) clone() *Rule {
	c := *r
	// copy the slices we append to so a fluent call never mutates the source
	c.allowed = append([]any(nil), r.allowed...)
	c.valids = append([]any(nil), r.valids...)
	c.invalids = append([]any(nil), r.invalids...)
	c.customs = append([]CustomFn(nil), r.customs...)
	if r.msgOverrides != nil {
		c.msgOverrides = make(map[string]string, len(r.msgOverrides))
		for k, v := range r.msgOverrides {
			c.msgOverrides[k] = v
		}
	}
	c.arrItems = append([]*Rule(nil), r.arrItems...)
	c.or = append([][]string(nil), r.or...)
	c.xor = append([][]string(nil), r.xor...)
	c.oxor = append([][]string(nil), r.oxor...)
	c.nand = append([][]string(nil), r.nand...)
	c.alts = append([]*Rule(nil), r.alts...)
	c.whens = append([]whenClause(nil), r.whens...)
	c.patterns = append([]keyPattern(nil), r.patterns...)
	c.applied = append([]appliedRule(nil), r.applied...)
	c.metas = append([]map[string]any(nil), r.metas...)
	// keys map is shared until Keys()/Append() rebuilds it
	return &c
}

// --- constructors ---

// String builds Joi.string(). By default rejects ” (Joi string.empty) unless
// .Allow("").
func String() *Rule { return &Rule{typ: tString} }

// Number builds Joi.number().
func Number() *Rule { return &Rule{typ: tNumber} }

// Boolean builds Joi.boolean().
func Boolean() *Rule { return &Rule{typ: tBoolean} }

// DateRule builds Joi.date(); normalized output is an ISO-8601 millis UTC string
// (matching JSON.stringify of a JS Date).
func DateRule() *Rule { return &Rule{typ: tDate} }

// Any builds Joi.any().
func Any() *Rule { return &Rule{typ: tAny} }

// Array builds Joi.array().items(items...).
func Array(items ...*Rule) *Rule { return &Rule{typ: tArray, arrItems: items} }

// Object builds Joi.object(keys). A nil map is Joi.object() (any keys subject to
// unknown policy). Go maps have no order, so the keys are sorted for
// determinism; use ObjectOrdered where the DECLARATION order matters.
func Object(keys map[string]*Rule) *Rule {
	r := &Rule{typ: tObject, keys: map[string]*Rule{}}
	if keys == nil {
		// `Joi.object()` with NO key schema accepts unknown keys; only `Joi.object({})`
		// rejects them. A nil map is how the port spells the first form, and defaulting
		// it to reject refused bodies TS takes (mk8sTriton's `none` / `gateway`).
		// Recorded as its own flag rather than .unknown(true) because Joi reports no
		// `unknown` flag here — the permissiveness comes from having no keys at all.
		r.keyless = true
	}
	for k, v := range keys {
		r.keys[k] = v
		r.keyOrder = append(r.keyOrder, k)
	}
	sort.Strings(r.keyOrder) // deterministic error ordering
	return r
}

// ObjectOrdered builds Joi.object(keys) preserving the given key order.
//
// abortEarly means only the FIRST failing key is reported, so the order decides
// WHICH error the caller sees. Joi walks keys in schema-declaration order; the
// canonical structs are declared in that same order, so a rule reflected from a
// struct must keep its field order rather than sort (a body missing both `name`
// and `data` reported "data" where TS reports "name").
func ObjectOrdered(order []string, keys map[string]*Rule) *Rule {
	r := &Rule{typ: tObject, keys: map[string]*Rule{}}
	for _, k := range order {
		if v, ok := keys[k]; ok {
			r.keys[k] = v
			r.keyOrder = append(r.keyOrder, k)
		}
	}
	return r
}

// Alternatives builds Joi.alternatives().try(alts...).
func Alternatives(alts ...*Rule) *Rule { return &Rule{typ: tAlternatives, alts: alts} }

// --- shared fluent methods ---

// Required marks the field .required().
func (r *Rule) Required() *Rule { c := r.clone(); c.presence = presRequired; return c }

// Optional marks the field .optional().
func (r *Rule) Optional() *Rule { c := r.clone(); c.presence = presOptional; return c }

// Strip is Joi's `.strip()`: the key is validated as usual and then removed from the
// result, defaults included, so it is also invisible to xor/nand/oxor and object.min.
// It lets a retired field keep validating on write while never being stored or echoed.
func (r *Rule) Strip() *Rule {
	c := r.clone()
	c.strip = true
	return c
}

// Forbidden marks the field .forbidden().
func (r *Rule) Forbidden() *Rule { c := r.clone(); c.presence = presForbidden; return c }

// Default sets .default(v). v may be a literal, a *Ref (Joi.ref), or a func() any.
func (r *Rule) Default(v any) *Rule {
	c := r.clone()
	c.hasDefault = true
	c.defaultVal = v
	return c
}

// Valid sets .valid(vals...) — only these values are accepted (any.only).
func (r *Rule) Valid(vals ...any) *Rule {
	c := r.clone()
	c.valids = append(c.valids, vals...)
	return c
}

// Invalid sets .invalid(vals...).
func (r *Rule) Invalid(vals ...any) *Rule {
	c := r.clone()
	c.invalids = append(c.invalids, vals...)
	return c
}

// Allow sets .allow(vals...) — these bypass all other checks and pass verbatim.
func (r *Rule) Allow(vals ...any) *Rule {
	c := r.clone()
	c.allowed = append(c.allowed, vals...)
	return c
}

// Custom appends a .custom(fn).
func (r *Rule) Custom(fn CustomFn) *Rule {
	c := r.clone()
	c.customs = append(c.customs, fn)
	if c.typ == tObject || c.typ == tArray {
		c.untypedCustom = true
	}
	c.addRule("custom", map[string]any{})
	return c
}

// CustomTyped attaches a container custom that handles the typed form: on the typed
// walk it receives a POINTER to the object/array (see CustomT for the checked
// generic form), on the generic evaluator the map/[]any. For a custom that inspects
// its argument by reflection and so serves both.
func (r *Rule) CustomTyped(fn CustomFn) *Rule {
	c := r.clone()
	c.customs = append(c.customs, fn)
	c.addRule("custom", map[string]any{})
	return c
}

// Refine attaches an object-level cross-field custom (Joi `.custom()` on the
// object). obj is the validated object (post type-coercion, siblings readable /
// mutatable); h carries the validation Context. Return an error (a *Error or
// h.Errorf) to fail. It is the builder form of the old Refiner interface, used
// inside a resource's ValidationRule to layer cross-field checks onto Reflect().
func (r *Rule) Refine(fn func(map[string]any, *Helper) error) *Rule { return withRefine(r, fn) }

// ReadOnly marks .meta({readOnly:true}).
func (r *Rule) ReadOnly() *Rule {
	c := r.clone()
	c.meta.readOnly = true
	c.metas = append(c.metas, map[string]any{"readOnly": true})
	return c
}

// Unmodifiable marks .meta({unmodifiable:true}).
func (r *Rule) Unmodifiable() *Rule {
	c := r.clone()
	c.meta.unmodifiable = true
	c.metas = append(c.metas, map[string]any{"unmodifiable": true})
	return c
}

// Meta appends a .meta({...}) the engine does not act on (type/pointer/hidden —
// codegen and UI hints). It is carried verbatim into Describe().
func (r *Rule) Meta(m map[string]any) *Rule {
	c := r.clone()
	c.metas = append(c.metas, m)
	return c
}

// Id sets .id(name) — Joi's schema identifier, reported as flags.id by describe().
// It carries no validation behaviour.
func (r *Rule) Id(name string) *Rule { c := r.clone(); c.id = name; return c }

// Label overrides the error label.
func (r *Rule) Label(l string) *Rule { c := r.clone(); c.label = l; return c }

// Description sets .description(text). It carries no validation behaviour; it is
// reported by Describe() (GET /catalog/schema/{kind}).
func (r *Rule) Description(text string) *Rule { c := r.clone(); c.description = text; return c }

// When appends a `.when(ref, …)` conditional.
func (r *Rule) When(ref string, w When) *Rule {
	c := r.clone()
	c.whens = append(c.whens, whenClause{ref: ref, is: w.condition(), not: w.Not, then: w.Then, otherwise: w.Otherwise})
	return c
}

// When holds the branches of a `.when` clause.
//
// The two condition forms are NOT interchangeable, because Joi treats them differently
// (compile.js exports.when: "Only apply required if this wasn't already a schema or a
// ref"):
//
//	IsEq — the TS `is: 'cpu'` LITERAL form. Compiled to `valid(lit).required()`, so an
//	       ABSENT sibling does not match.
//	Is   — the TS `is: Joi.string().regex(...)` SCHEMA form. Joi adds no .required(), so
//	       an absent sibling DOES match unless the schema itself is .Required().
//
// Both shapes are live in the schema set — workloadOptions.ts:145 is a literal,
// envoyCommon.ts:8 is a schema — so the engine cannot pick one. Port a TS literal with
// IsEq; reach for Is only when the TS wrote a schema.
type When struct {
	// IsEq is the literal condition (see the type doc). Mutually exclusive with Is.
	IsEq      any
	Is        *Rule
	Not       *Rule
	Then      *Rule
	Otherwise *Rule
}

func (w When) condition() *Rule {
	if w.IsEq != nil {
		return literalCondition(w.IsEq)
	}
	return w.Is
}

// literalCondition is Joi's compile of a literal `is:` — `Joi.valid(lit).required()`.
func literalCondition(lit any) *Rule {
	if vals, ok := lit.([]any); ok {
		return Any().Valid(vals...).Required()
	}
	return Any().Valid(lit).Required()
}

// --- string fluent ---

// FromNumber coerces a JSON number input to its string form before string rules run,
// porting the TS JoiNumberToString extend (workload.ts WholeNumberOrPercent /
// IntegerOrPercent, master f94fb6d78).
func (r *Rule) FromNumber() *Rule {
	c := r.clone()
	c.strFromNumber = true
	return c
}

func (r *Rule) Trim() *Rule {
	c := r.clone()
	c.strTrim = true
	c.addRule("trim", map[string]any{"enabled": true})
	return c
}

func (r *Rule) Lowercase() *Rule {
	c := r.clone()
	c.strLower = true
	c.addRule("case", map[string]any{"direction": "lower"})
	return c
}

func (r *Rule) Uppercase() *Rule {
	c := r.clone()
	c.strUpper = true
	c.addRule("case", map[string]any{"direction": "upper"})
	return c
}

func (r *Rule) Email() *Rule {
	c := r.clone()
	c.strEmail = true
	c.addRule("email", nil)
	return c
}

func (r *Rule) URI() *Rule {
	c := r.clone()
	c.strURI = true
	c.addRule("uri", nil)
	return c
}

// URIScheme is `.uri({scheme: [...]})`. Joi regenerates the whole URI grammar with the
// scheme production replaced, so a failure is reported as string.uriCustomScheme (with
// the alternation in context) rather than the plain string.uri.
func (r *Rule) URIScheme(schemes ...string) *Rule {
	c := r.clone()
	c.strURI = true
	c.strURIScheme = strings.Join(schemes, "|")
	c.addRule("uri", map[string]any{"options": map[string]any{"scheme": toAnySlice(schemes)}})
	return c
}

// Base64 is `.base64()` with Joi's defaults (padding required, not url-safe).
func (r *Rule) Base64() *Rule {
	c := r.clone()
	c.strBase64 = true
	c.addRule("base64", map[string]any{"options": map[string]any{"paddingRequired": true, "urlSafe": false}})
	return c
}

// Message is Joi's `.message(text)`: it replaces the message of the rule applied
// immediately before it, leaving the error type and context alone. Hand-rolling the
// override as a .Custom() instead (the shape this replaced) loses the type and the
// limit/regex context that clients rendering their own message read.
func (r *Rule) Message(text string) *Rule {
	c := r.clone()
	code := c.lastRuleCode()
	if c.msgOverrides == nil {
		c.msgOverrides = map[string]string{}
	}
	c.msgOverrides[code] = text
	return c
}

// Messages is Joi's `.messages({type: text})`: it replaces the message of every error
// of that TYPE the rule raises — including the presence errors (any.required,
// any.unknown) its parent object raises for it, which no `.message()` can reach.
func (r *Rule) Messages(byType map[string]string) *Rule {
	c := r.clone()
	if c.msgOverrides == nil {
		c.msgOverrides = map[string]string{}
	}
	for typ, text := range byType {
		c.msgOverrides[typ] = text
	}
	return c
}

// lastRuleCode is the Joi error code the most recently applied rule raises — what
// `.message()` binds to. An unmapped rule name panics at BUILD time (schemas are built
// at init) rather than silently dropping the override at request time.
func (r *Rule) lastRuleCode() string {
	if len(r.applied) == 0 {
		panic("validate: Message() with no preceding rule to override")
	}
	name := r.applied[len(r.applied)-1].name
	switch name {
	case "pattern":
		if r.strRegexInvert {
			return "string.pattern.invert.base"
		}
		return "string.pattern.base"
	case "min", "max", "length":
		return typeName(r.typ) + "." + name
	case "uri":
		if r.strURIScheme != "" {
			return "string.uriCustomScheme"
		}
		return "string.uri"
	case "base64", "email", "guid", "domain", "trim", "lowercase", "uppercase":
		return "string." + name
	case "integer", "port", "precision", "positive", "negative":
		return "number." + name
	default:
		panic("validate: Message() does not know the error code of rule " + name)
	}
}

func (r *Rule) UUID() *Rule {
	c := r.clone()
	c.strUUID = true
	c.addRule("guid", nil)
	return c
}

func (r *Rule) Domain() *Rule {
	c := r.clone()
	c.strDomain = true
	c.addRule("domain", nil)
	return c
}

// DomainTLDs is .domain(opts) — Joi checks the last segment against the bundled IANA
// list as soon as ANY options object is passed, `{minDomainSegments:2}` included, and
// only `tlds:false` turns it back off. Bare .domain() skips the list, so the two are
// genuinely different rules and the call site has to pick.
func (r *Rule) DomainTLDs() *Rule {
	c := r.clone()
	c.strDomain = true
	c.strDomainTLDs = true
	c.addRule("domain", map[string]any{"options": map[string]any{"minDomainSegments": 2}})
	return c
}

// Regex sets .regex(re) / .pattern(re). name is the optional named-pattern label.
func (r *Rule) Regex(re *regexp.Regexp) *Rule {
	c := r.clone()
	c.strRegex = re
	c.addRule("pattern", map[string]any{"regex": re})
	return c
}

// RegexAs is Regex with an explicit JavaScript source for the message and the
// describe document — for a Joi pattern RE2 cannot express verbatim (lookahead).
// The Go regexp still does the matching, so it must be equivalent for the inputs
// that reach it; anything it cannot express belongs in a guard beside it.
func (r *Rule) RegexAs(re *regexp.Regexp, jsSource string) *Rule {
	c := r.clone()
	c.strRegex = re
	c.strRegexSrc = jsSource
	c.addRule("pattern", map[string]any{"regex": JSRegex(jsSource)})
	return c
}

// JSRegex is a regex carried as its JavaScript SOURCE, for describe() and messages
// where the Go pattern is not the pattern the schema was written with.
type JSRegex string

// regexSource is the pattern as the TS schema spells it.
func (r *Rule) regexSource() string {
	if r.strRegexSrc != "" {
		return r.strRegexSrc
	}
	if r.strRegex == nil {
		return ""
	}
	return r.strRegex.String()
}

// RegexInvert sets .pattern(re, {invert:true}) / .regex(re, {invert:true}).
func (r *Rule) RegexInvert(re *regexp.Regexp) *Rule {
	c := r.clone()
	c.strRegex = re
	c.strRegexInvert = true
	c.addRule("pattern", map[string]any{"regex": re, "options": map[string]any{"invert": true}})
	return c
}

// Min sets .min(n): string length, number bound, or array length by type.
func (r *Rule) Min(n int) *Rule {
	c := r.clone()
	switch c.typ {
	case tNumber:
		f := float64(n)
		c.numMin = &f
	case tArray:
		c.arrMin = &n
	case tObject:
		c.objMin = &n
	default:
		c.strMin = &n
	}
	c.addRule("min", map[string]any{"limit": n})
	return c
}

// Max sets .max(n) by type.
func (r *Rule) Max(n int) *Rule {
	c := r.clone()
	switch c.typ {
	case tNumber:
		f := float64(n)
		c.numMax = &f
	case tArray:
		c.arrMax = &n
	case tObject:
		c.objMax = &n
	default:
		c.strMax = &n
	}
	c.addRule("max", map[string]any{"limit": n})
	return c
}

// Length sets .length(n) by type.
func (r *Rule) Length(n int) *Rule {
	c := r.clone()
	switch c.typ {
	case tArray:
		c.arrLen = &n
	default:
		c.strLen = &n
	}
	c.addRule("length", map[string]any{"limit": n})
	return c
}

// --- number fluent ---

// NumMin sets a numeric minimum (float, for non-integer bounds like precision).
func (r *Rule) NumMin(f float64) *Rule {
	c := r.clone()
	c.numMin = &f
	c.addRule("min", map[string]any{"limit": f})
	return c
}

// NumMax sets a numeric maximum (float).
func (r *Rule) NumMax(f float64) *Rule {
	c := r.clone()
	c.numMax = &f
	c.addRule("max", map[string]any{"limit": f})
	return c
}

// MinRef sets .min(Joi.ref(key)) for numbers.
func (r *Rule) MinRef(key string) *Rule {
	c := r.clone()
	c.numMinRef = key
	c.addRule("min", map[string]any{"limit": RefTo(key)})
	return c
}

// MaxRef sets .max(Joi.ref(key)) for numbers.
func (r *Rule) MaxRef(key string) *Rule {
	c := r.clone()
	c.numMaxRef = key
	c.addRule("max", map[string]any{"limit": RefTo(key)})
	return c
}

func (r *Rule) Integer() *Rule {
	c := r.clone()
	c.numInt = true
	c.addRule("integer", nil)
	return c
}

func (r *Rule) Positive() *Rule {
	c := r.clone()
	c.numPos = true
	c.addRule("sign", map[string]any{"sign": "positive"})
	return c
}

func (r *Rule) Negative() *Rule {
	c := r.clone()
	c.numNeg = true
	c.addRule("sign", map[string]any{"sign": "negative"})
	return c
}

func (r *Rule) Port() *Rule {
	c := r.clone()
	c.numPort = true
	c.addRule("port", nil)
	return c
}

// Precision sets .precision(n) — rounds to n decimal places (Joi convert).
func (r *Rule) Precision(n int) *Rule {
	c := r.clone()
	c.numPrecision = &n
	c.addRule("precision", map[string]any{"limit": n})
	return c
}

// Iso sets Joi.date().iso(): the value must be a STRING in ISO-8601 (Joi's own isoDate
// grammar, see isodate.go), so a slash date, a month name, a number and an epoch-millis
// string are all rejected. It used to return the rule unchanged, which accepted all of
// them.
func (r *Rule) Iso() *Rule {
	c := r.clone()
	c.dateIso = true
	c.addRule("format", map[string]any{"format": "iso"})
	return c
}

// --- array fluent ---

// Items sets .items(items...).
func (r *Rule) Items(items ...*Rule) *Rule { c := r.clone(); c.arrItems = items; return c }

// Sort sets .sort() (ascending, matching Joi's default order).
func (r *Rule) Sort() *Rule {
	c := r.clone()
	c.arrSort = true
	c.addRule("sort", map[string]any{"options": map[string]any{"order": "ascending"}})
	return c
}

// SortBy adds Joi's `.sort({ by: field })` — order the elements by one of their
// object keys, ascending. Distinct from Sort(), which orders whole values.
func (r *Rule) SortBy(field string) *Rule {
	c := r.clone()
	c.arrSort = true
	c.arrSortBy = field
	c.addRule("sort", map[string]any{"options": map[string]any{"order": "ascending", "by": RefTo(field)}})
	return c
}

// Single sets .single() — a non-array scalar is coerced to a one-element array
// before item validation (Joi array.single with convert:true).
func (r *Rule) Single() *Rule { c := r.clone(); c.arrSingle = true; return c }

// Unique sets .unique(); by is an optional key for uniqueness by property. This is
// Joi's `.array().unique()` — a REJECT on a duplicate (array.unique error).
func (r *Rule) Unique(by ...string) *Rule {
	c := r.clone()
	c.arrUnique = true
	var args map[string]any
	if len(by) > 0 {
		c.arrUniqueBy = by[0]
		c.arrUniquePath = true
		// Joi's comparator arg is a function, which JSON.stringify drops — the oracle
		// reports the empty args object, and a bare .unique() reports no args at all.
		args = map[string]any{}
	}
	c.addRule("unique", args)
	return c
}

// UniqueComparing is `.unique((a, b) => a[field] === b[field])` — Joi's FUNCTION
// comparator. It rejects like Unique but, unlike the string-path spelling, its error
// carries no `path` context, because Joi has no path to report for a function.
func (r *Rule) UniqueComparing(field string) *Rule {
	c := r.clone()
	c.arrUnique = true
	c.arrUniqueBy = field
	// Joi's comparator arg is a function, which JSON.stringify drops — the oracle
	// reports an empty args object.
	c.addRule("unique", map[string]any{})
	return c
}

// uniqueCtxPath is the comparator Joi puts in the array.unique context, or "" when
// the comparator was a function.
func (r *Rule) uniqueCtxPath() string {
	if r.arrUniquePath {
		return r.arrUniqueBy
	}
	return ""
}

// Dedupe sets TS's `.custom(makeUnique(by?))` semantics (base.ts:138-149): a duplicate
// is SILENTLY DROPPED (keep first, _.uniq/_.uniqBy), never rejected. Distinct from
// Unique, which rejects — a field must map to whichever its Joi schema uses.
func (r *Rule) Dedupe(by ...string) *Rule {
	c := r.clone()
	c.arrDedupe = true
	if len(by) > 0 {
		c.arrDedupeBy = by[0]
	}
	// TS spells dedupe as `.custom(makeUnique())` (base.ts:138), so Joi reports it
	// as a custom rule, not array.unique.
	c.addRule("custom", map[string]any{})
	return c
}

// --- object fluent ---

// Keys replaces/extends the object's keyed sub-schemas.
func (r *Rule) Keys(keys map[string]*Rule) *Rule {
	c := r.clone()
	nk := map[string]*Rule{}
	for k, v := range c.keys {
		nk[k] = v
	}
	// Joi's .keys()/.append() adds keys AFTER the existing ones, so only the appended
	// tail is sorted (a Go map arg has no order) — re-sorting the whole list would
	// destroy a reflected rule's declaration order. A key that ALREADY exists is
	// filtered out of its old position and re-appended at the tail (keys.js: the
	// overridden child is dropped and the replacement pushed), which decides which of
	// several missing required keys abortEarly reports.
	var added []string
	for k, v := range keys {
		added = append(added, k)
		nk[k] = v
	}
	sort.Strings(added)
	order := make([]string, 0, len(c.keyOrder)+len(added))
	for _, k := range c.keyOrder {
		if _, overridden := keys[k]; !overridden {
			order = append(order, k)
		}
	}
	c.keys = nk
	c.keyOrder = append(order, added...)
	return c
}

// Append is Joi's schema.append(keys) — same as Keys for our purposes.
func (r *Rule) Append(keys map[string]*Rule) *Rule { return r.Keys(keys) }

// Fork replaces one key's schema (a narrowed Joi.fork over a single path).
func (r *Rule) Fork(key string, fn func(*Rule) *Rule) *Rule {
	c := r.clone()
	nk := map[string]*Rule{}
	for k, v := range c.keys {
		nk[k] = v
	}
	if cur, ok := nk[key]; ok {
		nk[key] = fn(cur)
		c.keys = nk
		c.orderAfterRef(key)
		return c
	}
	c.keys = nk
	return c
}

// orderAfterRef reproduces Joi's reference-driven key reordering: a key whose rule
// references a SIBLING is validated after that sibling, so Joi moves it to just after
// the referent. `Joi.object({a,b,c,d,e}).fork(['b'], s => s.default(Joi.ref('d')))`
// describes as a,c,d,b,e — b followed d.
//
// Which constructs trigger it is not uniform, and guessing gets it backwards. Measured
// against joi@17.6.0 on 2026-09-16, over `Joi.object({a,b,c,d,e})`:
//
//	a.default(Joi.ref('e'))                   -> b,c,d,e,a   REORDERS
//	a.when('e', {...})                        -> b,c,d,e,a   REORDERS
//	a: alternatives().conditional('e', {...})  -> a,b,c,d,e   does NOT
//
// Only the default-ref case is implemented here, because it is the one the shared create
// fork depends on and the one a divergence was measured for. `.when()` is a KNOWN gap:
// no live schema is affected today (the only `.when()` referent carrying a default is
// query's `op`, and `value` already sorts after it), but a new one would be. Landing it
// would move keys across many schemas and so change which key abortEarly blames, which
// is not something to do without the route sweep to check it against.
//
// This is what decides which key abortEarly blames. The shared create fork gives
// `description` a `$name` default, so in TS `name` is checked first and a create
// missing both is "name is required"; here description kept its declared position and
// the answer was `"description" must be a string`. Every kind runs that fork, so this
// belongs in Fork rather than in the handler that calls it.
func (r *Rule) orderAfterRef(key string) {
	kr := r.keys[key]
	if kr == nil || !kr.hasDefault {
		return
	}
	ref, ok := kr.defaultVal.(*Ref)
	if !ok || ref.Key == key {
		return
	}
	if _, sibling := r.keys[ref.Key]; !sibling {
		return
	}
	order := make([]string, 0, len(r.keyOrder))
	for _, k := range r.keyOrder {
		if k == key {
			continue
		}
		order = append(order, k)
		if k == ref.Key {
			order = append(order, key)
		}
	}
	r.keyOrder = order
}

// Unknown sets the object's unknown-key policy: true → keep unknown keys.
//
// `.unknown(false)` clears the flag rather than setting a reject POLICY, because that
// is what Joi does — `unknown(false)` writes the flag's own default, and Joi drops a
// flag equal to its default, so the object falls back to `Common.default(flags.unknown,
// prefs.allowUnknown)` and an allowUnknown preference still applies. Recording it as an
// explicit reject made `.unknown(false)` block the preference.
func (r *Rule) Unknown(allow bool) *Rule {
	c := r.clone()
	if allow {
		c.unknown = unkAllow
	} else {
		c.unknown = unkInherit
	}
	return c
}

// StripUnknown sets .unknown('strip') behaviour on the object.
func (r *Rule) StripUnknown() *Rule { c := r.clone(); c.unknown = unkStrip; return c }

// Pattern adds a .pattern(keySchema, valSchema). Several may be attached; they are
// tried in the order they were added.
func (r *Rule) Pattern(key, val *Rule) *Rule {
	c := r.clone()
	c.patterns = append(c.patterns, keyPattern{key: key, val: val})
	return c
}

// keyPattern is one .pattern(keySchema, valSchema).
type keyPattern struct{ key, val *Rule }

// hasPatterns reports whether the object declares any key pattern.
func (r *Rule) hasPatterns() bool { return len(r.patterns) > 0 }

// matchPattern returns the rule for the first pattern whose key schema accepts key.
func (r *Rule) matchPattern(key string, kp []string, st *evalState) *Rule {
	for _, p := range r.patterns {
		if _, err := p.key.validateValue(key, true, kp, st); err == nil {
			return p.val
		}
	}
	return nil
}

// Or adds .or(peers...) — at least one of peers must be present.
func (r *Rule) Or(peers ...string) *Rule { c := r.clone(); c.or = append(c.or, peers); return c }

// Xor adds .xor(peers...) — exactly one of peers must be present.
func (r *Rule) Xor(peers ...string) *Rule { c := r.clone(); c.xor = append(c.xor, peers); return c }

// Oxor adds .oxor(peers...) — at most one of peers may be present.
func (r *Rule) Oxor(peers ...string) *Rule { c := r.clone(); c.oxor = append(c.oxor, peers); return c }

// Nand adds .nand(peers...) — peers must not all be present together.
func (r *Rule) Nand(peers ...string) *Rule { c := r.clone(); c.nand = append(c.nand, peers); return c }

// --- alternatives fluent ---

// Try sets .try(alts...).
func (r *Rule) Try(alts ...*Rule) *Rule { c := r.clone(); c.alts = append(c.alts, alts...); return c }

// Conditional sets alternatives().conditional(ref, {is, then, otherwise}).
func (r *Rule) Conditional(ref string, b AltCase) *Rule {
	c := r.clone()
	c.altCond = &altBranch{ref: ref, is: b.condition(), then: b.Then, otherwise: b.Otherwise}
	return c
}

// AltCase holds one alternatives().conditional branch. IsEq / Is carry the same
// distinction they do on When — see that type's doc.
type AltCase struct {
	IsEq      any
	Is        *Rule
	Then      *Rule
	Otherwise *Rule
}

func (b AltCase) condition() *Rule {
	if b.IsEq != nil {
		return literalCondition(b.IsEq)
	}
	return b.Is
}

// ObjectSchema wraps a *Rule object as a top-level *Schema for Register/For.
func ObjectSchema(root *Rule) *Schema { return &Schema{Root: root} }

// MinNum / MaxNum are Min / Max with a float64 bound, for limits outside int range
// (envoy's UInt64 caps at 2^64-1, which Joi computes as a float64).
func (r *Rule) MinNum(f float64) *Rule {
	c := r.clone()
	c.numMin = &f
	c.addRule("min", map[string]any{"limit": f})
	return c
}

func (r *Rule) MaxNum(f float64) *Rule {
	c := r.clone()
	c.numMax = &f
	c.addRule("max", map[string]any{"limit": f})
	return c
}

// ToFloat coerces a decoded JSON number (or a numeric string, as convert:true does)
// to float64. For customs that have to do arithmetic on a value the schema has
// already accepted.
func ToFloat(v any) (float64, bool) {
	switch t := v.(type) {
	case float64:
		return t, true
	case float32:
		return float64(t), true
	case int:
		return float64(t), true
	case int64:
		return float64(t), true
	case json.Number:
		f, err := t.Float64()
		return f, err == nil
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(t), 64)
		return f, err == nil
	}
	return 0, false
}
