package validate

// eval.go executes a compiled *Rule against a value — the runtime half of the
// Joi port. Semantics mirror Joi with convert:true, abortEarly:true.

import (
	"encoding/json"
	"math"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Error is a validation failure. It carries the Joi error type, human message,
// and the dot/bracket path so the HTTP layer can render error.details[] at parity
// with the TS service (isDataInvalid → {status, message, details}).
type Error struct {
	Type    string
	Message string
	Path    []string
	Ctx     map[string]any

	// customWrapped guards against wrapping an any.custom message twice as it
	// propagates through nested customs.
	customWrapped bool
}

func (e *Error) Error() string { return e.Message }

// Detail is one element of the Joi error.details[] array.
//
// Path is []any, not []string, because Joi's path mixes types: an object key is a
// string and an ARRAY INDEX is a number (`["spec","ports",0,"protocol"]`). A client
// walking the path to point at the offending field — which is the only reason the
// field exists — indexes an array with the number, so emitting "0" there is not a
// cosmetic difference.
type Detail struct {
	Message string         `json:"message"`
	Path    []any          `json:"path"`
	Type    string         `json:"type"`
	Context map[string]any `json:"context,omitzero"`
}

// Details returns the details[] array (abortEarly → a single element).
func (e *Error) Details() []Detail {
	ctx := e.Ctx
	if ctx == nil {
		ctx = map[string]any{"label": label(e.Path)}
	}
	// `key` is the last path segment, typed the same way the path is (Joi:
	// context.key is 0, not "0", inside an array).
	if _, ok := ctx["key"]; !ok && len(e.Path) > 0 {
		ctx["key"] = pathSeg(e.Path[len(e.Path)-1])
	}
	// A top-level failure has an EMPTY path, not a null one: Joi always emits an array.
	out := make([]any, 0, len(e.Path))
	for _, seg := range e.Path {
		out = append(out, pathSeg(seg))
	}
	return []Detail{{Message: e.Message, Path: out, Type: e.Type, Context: ctx}}
}

// pathSeg renders one internal path segment as Joi would type it: a marked array
// index becomes an int, anything else stays the string key.
func pathSeg(seg string) any {
	if n, ok := indexOf(seg); ok {
		return n
	}
	return seg
}

// evalState is threaded through nested validation so refs (Joi.ref / .when) can
// resolve against the immediate parent object and customs get the context.
type evalState struct {
	ctx    *Context
	prefs  Prefs
	parent parentView // the object currently being validated (for sibling refs)
	// up is the state the ENCLOSING object was validated in, giving customs Joi's
	// ancestor chain (helpers.state.ancestors). It is a link rather than a slice so
	// descending costs nothing on the hot path. Arrays do not add a level: an element's
	// items share their array's state, exactly as the sibling view does.
	up *evalState
	// order is the body's key order, by path (order.go). It rides the state rather
	// than the value so BOTH evaluators read the same table at the same path — a
	// custom cannot be handed an order on one path and not the other.
	order *wireOrder
}

// descend builds the state for the keys of an object whose sibling view is p.
func (st *evalState) descend(p parentView) *evalState {
	// The map entry points seed the state with a view of the very object about to be
	// validated, so a ROOT-level .when() can resolve its own keys. Descending into that
	// same object would stack the document twice in the ancestor chain — and the typed
	// entry point seeds nothing, so the two evaluators would disagree on depth.
	if st.up == nil && sameObject(st.parent, p) {
		return &evalState{ctx: st.ctx, prefs: st.prefs, parent: p, order: st.order}
	}
	return &evalState{ctx: st.ctx, prefs: st.prefs, parent: p, up: st, order: st.order}
}

// sameObject reports a seeded root view and the descent view over the same map.
func sameObject(seed, child parentView) bool {
	sm, ok := seed.(mapView)
	if !ok {
		return false
	}
	pv, ok := child.(*progressView)
	if !ok {
		return false
	}
	return reflect.ValueOf(map[string]any(sm)).Pointer() == reflect.ValueOf(pv.in).Pointer()
}

func label(path []string) string {
	if len(path) == 0 {
		return "value"
	}
	var b strings.Builder
	for i, seg := range path {
		if n, ok := indexOf(seg); ok {
			b.WriteString("[")
			b.WriteString(strconv.Itoa(n))
			b.WriteString("]")
			continue
		}
		if i > 0 {
			b.WriteByte('.')
		}
		b.WriteString(seg)
	}
	return b.String()
}

// idxMark tags a path segment as an ARRAY INDEX rather than an object key. The
// distinction cannot be recovered from the text — `labels.123` is a legal object key
// and would otherwise be rendered `labels[123]` and typed as a number — so it is
// recorded where it is known (the array walk) and read back at the two places that
// care: label() and pathSeg(). The mark never leaves this package.
const idxMark = "\x00#"

// idxSeg builds the marked segment for array element i.
func idxSeg(i int) string { return idxMark + strconv.Itoa(i) }

// IndexPath builds a path segment for array element i, for a caller outside this
// package composing a base path (see Rule.ValidateAt / Rule.ApplyAt). Using it rather
// than strconv.Itoa is what keeps the segment typed as a NUMBER in the reported path.
func IndexPath(i int) string { return idxSeg(i) }

// indexOf returns the numeric index of a marked segment.
func indexOf(s string) (int, bool) {
	if !strings.HasPrefix(s, idxMark) {
		return 0, false
	}
	n, err := strconv.Atoi(s[len(idxMark):])
	if err != nil {
		return 0, false
	}
	return n, true
}

func isIndex(s string) bool { _, ok := indexOf(s); return ok }

// errCtx adds one Joi context field to an error (the per-type extras Joi puts in
// details[].context: valids, regex, child, types).
func errCtx(e *Error, key string, val any) *Error {
	if e.Ctx == nil {
		e.Ctx = map[string]any{}
	}
	e.Ctx[key] = val
	return e
}

// JSTrim is JavaScript's String.prototype.trim, whose whitespace set is NOT the
// same as Go's unicode.IsSpace. Two characters differ, in opposite directions:
//
//	U+FEFF (ZWNBSP/BOM) — JS strips it, Go does not
//	U+0085 (NEL)        — Go strips it, JS does not
//
// Because .trim() runs on every trimmed string BEFORE the length and pattern
// rules and the result is what gets PERSISTED, the disagreement changed both
// which values were accepted and what was stored: a leading BOM survived here
// and was stripped there, and vice versa for NEL.
//
// The set is ECMA-262 WhiteSpace + LineTerminator.
func JSTrim(s string) string {
	return strings.TrimFunc(s, isJSSpace)
}

func isJSSpace(r rune) bool {
	switch r {
	case '\t', '\v', '\f', '\r', '\n', ' ', 0x00A0, 0xFEFF, 0x1680,
		0x2028, 0x2029, 0x202F, 0x205F, 0x3000:
		return true
	}
	// U+2000..U+200A are the remaining Space_Separator characters.
	return r >= 0x2000 && r <= 0x200A
}

// JSLen is JavaScript's String.prototype.length: the number of UTF-16 code
// units, NOT runes and NOT bytes.
//
// Joi's .min/.max/.length count what JS counts, so a character outside the BMP
// (emoji, most CJK extensions, many scripts) counts as TWO. Counting runes let
// this port accept a 200-emoji value against a 250-"character" cap that TS
// rejects at 400 — i.e. up to 2x the documented limit on EVERY string rule in the
// schema set, persisting values TS's own validator considers illegal.
func JSLen(s string) int {
	n := 0
	for _, r := range s {
		n++
		if r > 0xFFFF {
			n++ // encoded as a surrogate pair in UTF-16
		}
	}
	return n
}

// typeBaseError is the wrong-type error for a typed schema. It is ONE builder because
// the null branch and the wrong-type branch must agree: they used to be separate, and
// object disagreed with itself — `{"spec":null}` said "must be a object" (Joi's generic
// template) while `{"spec":5}` said "must be of type object" (Joi's object-specific
// one). Joi's object.base message interpolates a `type` context var, which is also why
// object is the only type here carrying one.
func typeBaseError(path []string, typ ruleType) *Error {
	if typ == tObject {
		return errCtx(errAt(path, "object.base", quote(label(path))+" must be of type object"), "type", "object")
	}
	article := " must be a "
	if typ == tArray {
		article = " must be an "
	}
	return errAt(path, typeName(typ)+".base", quote(label(path))+article+humanType(typ))
}

// errLimit is errAt plus Joi's `limit` context field, which every min/max/length
// message interpolates — a client rendering its own message needs the bound, and
// re-parsing it out of the English is not a reasonable ask.
func errLimit(path []string, typ, msg string, limit any) *Error {
	return errCtx(errAt(path, typ, msg), "limit", limit)
}

func errAt(path []string, typ, msg string) *Error {
	l := label(path)
	ctx := map[string]any{"label": l}
	if len(path) > 0 {
		ctx["key"] = pathSeg(path[len(path)-1])
	}
	return &Error{Type: typ, Message: msg, Path: append([]string(nil), path...), Ctx: ctx}
}

// slashRegex renders a regex pattern the way Joi's messages do — the JS
// RegExp.toString() form, wrapped in slashes (e.g. /^[a-z].../). Go's
// Regexp.String() returns the bare source, so the TS-parity error message
// ("... required pattern: /^.../") must add the delimiters
// (outstanding_wireformat item 2).
func slashRegex(pattern string) string { return "/" + escapeRegexSlashes(pattern) + "/" }

// regexLiteral is slashRegex plus the JS FLAGS. Go carries case-insensitivity as a
// leading `(?i)` inline flag and JS as a trailing `i`, so a `(?i)` pattern rendered
// bare produced `/(?i)^sha…$/` where Joi prints `/^sha…$/i` — a message mismatch on
// every case-insensitive pattern in the schema set.
func (r *Rule) regexLiteral() string {
	src := r.regexSource()
	flags := ""
	if strings.HasPrefix(src, "(?i)") {
		src, flags = src[len("(?i)"):], "i"
	} else if r.strRegex != nil && strings.HasPrefix(r.strRegex.String(), "(?i)") {
		// RegexAs: the JS source is spelled without the inline flag, but the pattern
		// it stands for is still case-insensitive.
		flags = "i"
	}
	return slashRegex(src) + flags
}

// escapeRegexSlashes reproduces V8's EscapeRegExpPattern: a forward slash that would
// otherwise TERMINATE the literal is escaped, and one inside a character class is not
// (a `/` there cannot end the literal). So `(/org/[^/]+/.*)` renders as
// `(\/org\/[^/]+\/.*)` — escaped twice outside the class, left alone inside it.
//
// Go's Regexp.String() returns the bare source, so without this the pattern in the
// message differed from TS's for every link and path regex in the schema set.
func escapeRegexSlashes(pattern string) string {
	var b strings.Builder
	inClass := false
	for i := 0; i < len(pattern); i++ {
		c := pattern[i]
		switch {
		case c == '\\' && i+1 < len(pattern):
			b.WriteByte(c)
			i++
			b.WriteByte(pattern[i])
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

// resolveDefaultVal resolves a Default value (literal, *Ref, or func).
//
// The `func(map[string]any) any` form is Joi's PARENT-AWARE default —
// `.default(p => ...)` — where the value depends on sibling keys (e.g.
// workloadOptions ResourceTarget, which is 95 unless `multi` is set or the metric
// is keda). Returning nil from it means "no default", matching a JS undefined.
func resolveDefaultVal(v any, parent parentView) any {
	switch d := v.(type) {
	case *Ref:
		sib, _ := siblingPath(parent, d.Key)
		return sib
	case func() any:
		return d()
	case func(map[string]any) any:
		if parent == nil {
			return d(nil)
		}
		return d(parent.asMap())
	default:
		return v
	}
}

// whenMatches decides a `.when()` condition the way Joi does.
//
// An ABSENT sibling is not automatically a miss: Joi validates `undefined` against the
// condition SCHEMA, and only a required condition rejects it. Joi adds that .required()
// itself when `is` was written as a LITERAL (Compile.when: "Only apply required if this
// wasn't already a schema or a ref"), which is why `{is: 'x'}` misses on an absent
// sibling while `{is: Joi.string()}` matches. A port that demanded presence for every
// condition silently took the wrong branch on every schema-shaped `is` —
// envoyCommon's `is: Joi.string().regex(/^cpln_.*/)` is exactly that shape.
//
// With no condition at all Joi uses `invalid(null, false, 0, ”).required()`: present
// and not one of JS's falsey values.
func whenMatches(cond *Rule, sib any, present bool, ref string, st *evalState) bool {
	if cond == nil {
		if !present {
			return false
		}
		switch v := sib.(type) {
		case nil:
			return false
		case bool:
			return v
		case float64:
			return v != 0
		case int:
			return v != 0
		case string:
			return v != ""
		}
		return true
	}
	if !present {
		return cond.presence != presRequired
	}
	_, err := cond.validateValue(sib, present, []string{ref}, st)
	return err == nil
}

// resolveEffective applies any .when clauses given the parent object, returning a
// concrete rule with then/otherwise merged in.
//
// The result is MARKED resolved, because the seam is reached twice for the same value:
// the object loop resolves a key's rule before it decides presence, and
// noValueContext lists the errors Joi raises WITHOUT a context.value. an any.ref blames
// the LIMIT REFERENCE, not the value being validated, and never carries one; any.required
// has no value by definition.
var noValueContext = map[string]bool{"any.ref": true, "any.required": true}

// validateValueInner resolves again on the way into the leaf. mergeRules APPENDS
// customs/valids/allows, so resolving twice ran a `then` branch's .custom() twice — a
// transforming custom applied its transform twice — and printed its valids twice
// ("must be one of [only, only]").
func (r *Rule) resolveEffective(st *evalState) *Rule {
	if r.resolved {
		return r
	}
	// An alternatives().conditional() carries its PRESENCE on the branches, not on the
	// alternatives node — `payload: alternatives().conditional('encoding', {then:
	// string().base64().required(), otherwise: string().required()})`. The object loop
	// reads presence off the key's rule, so without resolving the branch here an ABSENT
	// key looked optional and a secret with no payload at all was accepted.
	if r.altCond != nil {
		if p := r.altCondPresence(st); p != presDefault {
			c := r.clone()
			c.presence = p
			r = c
		}
	}
	if len(r.whens) == 0 {
		return r
	}
	eff := r
	for _, w := range r.whens {
		var sib any
		present := false
		sib, present = siblingPath(st.parent, w.ref)
		// `not` is not a negated match in Joi: Compile.when rewrites {not, then,
		// otherwise} as {is: not, then: otherwise, otherwise: then}, so the branches
		// swap and the condition itself is evaluated normally.
		cond, thenR, elseR := w.is, w.then, w.otherwise
		if w.not != nil {
			cond, thenR, elseR = w.not, w.otherwise, w.then
		}
		var pick *Rule
		if whenMatches(cond, sib, present, w.ref, st) {
			pick = thenR
		} else {
			pick = elseR
		}
		if pick != nil {
			// Joi generates the branch before concatenating it (base.js _generate calls
			// then._generate), so a .when() inside a branch resolves here and not later.
			eff = mergeRules(eff, pick.resolveEffective(st))
		}
	}
	if eff == r {
		return r
	}
	eff = eff.clone()
	eff.whens = nil
	eff.resolved = true
	return eff
}

// mergeRules concatenates add onto base (Joi schema.concat): presence/default and
// scalar constraints from add override, slices append, flags OR. Every field of Rule is
// handled — the hand-written list used to stop at the common ones, so a `.when()` branch
// silently dropped its .length()/.unique()/.sort()/.pattern()/.xor()/.try() and its
// label/description.
func mergeRules(base, add *Rule) *Rule {
	c := base.clone()
	if add.typ != tAny {
		c.typ = add.typ
	}
	if add.presence != presDefault {
		c.presence = add.presence
	}
	if add.hasDefault {
		c.hasDefault = true
		c.defaultVal = add.defaultVal
	}
	if add.label != "" {
		c.label = add.label
	}
	if add.description != "" {
		c.description = add.description
	}
	c.meta.readOnly = c.meta.readOnly || add.meta.readOnly
	c.meta.unmodifiable = c.meta.unmodifiable || add.meta.unmodifiable
	c.metas = append(c.metas, add.metas...)
	c.applied = append(c.applied, add.applied...)
	c.valids = append(c.valids, add.valids...)
	c.invalids = append(c.invalids, add.invalids...)
	c.allowed = append(c.allowed, add.allowed...)
	c.customs = append(c.customs, add.customs...)
	c.whens = append(c.whens, add.whens...)
	c.untypedCustom = c.untypedCustom || add.untypedCustom
	// string
	c.strTrim = c.strTrim || add.strTrim
	c.strFromNumber = c.strFromNumber || add.strFromNumber
	c.strLower = c.strLower || add.strLower
	c.strUpper = c.strUpper || add.strUpper
	c.strEmail = c.strEmail || add.strEmail
	c.strURI = c.strURI || add.strURI
	c.strUUID = c.strUUID || add.strUUID
	c.strip = c.strip || add.strip
	c.strDomain = c.strDomain || add.strDomain
	c.strDomainTLDs = c.strDomainTLDs || add.strDomainTLDs
	c.strBase64 = c.strBase64 || add.strBase64
	if add.strURIScheme != "" {
		c.strURIScheme = add.strURIScheme
	}
	for k, v := range add.msgOverrides {
		if c.msgOverrides == nil {
			c.msgOverrides = map[string]string{}
		}
		c.msgOverrides[k] = v
	}
	if add.strMin != nil {
		c.strMin = add.strMin
	}
	if add.strMax != nil {
		c.strMax = add.strMax
	}
	if add.strLen != nil {
		c.strLen = add.strLen
	}
	if add.strRegex != nil {
		c.strRegex = add.strRegex
		c.strRegexInvert = add.strRegexInvert
		c.strRegexSrc = add.strRegexSrc
	}
	// number
	if add.numMin != nil {
		c.numMin = add.numMin
	}
	if add.numMax != nil {
		c.numMax = add.numMax
	}
	if add.numMinRef != "" {
		c.numMinRef = add.numMinRef
	}
	if add.numMaxRef != "" {
		c.numMaxRef = add.numMaxRef
	}
	c.numInt = c.numInt || add.numInt
	c.numPos = c.numPos || add.numPos
	c.numNeg = c.numNeg || add.numNeg
	c.numPort = c.numPort || add.numPort
	if add.numPrecision != nil {
		c.numPrecision = add.numPrecision
	}
	// date
	c.dateIso = c.dateIso || add.dateIso
	// array
	if len(add.arrItems) > 0 {
		c.arrItems = add.arrItems
	}
	if add.arrMin != nil {
		c.arrMin = add.arrMin
	}
	if add.arrMax != nil {
		c.arrMax = add.arrMax
	}
	if add.arrLen != nil {
		c.arrLen = add.arrLen
	}
	if add.arrSort {
		c.arrSort = true
		c.arrSortBy = add.arrSortBy
	}
	if add.arrUnique {
		c.arrUnique = true
		c.arrUniqueBy = add.arrUniqueBy
		c.arrUniquePath = add.arrUniquePath
	}
	if add.arrDedupe {
		c.arrDedupe = true
		c.arrDedupeBy = add.arrDedupeBy
	}
	c.arrSingle = c.arrSingle || add.arrSingle
	// object
	c.or = append(c.or, add.or...)
	c.xor = append(c.xor, add.xor...)
	c.oxor = append(c.oxor, add.oxor...)
	c.nand = append(c.nand, add.nand...)
	if len(add.keys) > 0 {
		// clone() SHARES the keys map (it is rebuilt only by Keys/Fork), so writing the
		// branch's keys straight into it poisoned the cached schema rule for every later
		// request — a .when() whose branch carries keys permanently rewrote the base.
		nk := make(map[string]*Rule, len(c.keys)+len(add.keys))
		for k, v := range c.keys {
			nk[k] = v
		}
		var added []string
		for k, v := range add.keys {
			if _, ok := nk[k]; !ok {
				added = append(added, k)
			}
			nk[k] = v
		}
		c.keys = nk
		sort.Strings(added) // appended tail only — see Rule.Keys
		c.keyOrder = append(append([]string(nil), c.keyOrder...), added...)
		c.keyless = false
	}
	if len(add.patterns) > 0 {
		c.patterns = append(c.patterns, add.patterns...)
		c.keyless = false
	}
	if add.unknown != unkInherit {
		c.unknown = add.unknown
	}
	if add.objMin != nil {
		c.objMin = add.objMin
	}
	if add.objMax != nil {
		c.objMax = add.objMax
	}
	// alternatives
	c.alts = append(c.alts, add.alts...)
	if add.altCond != nil {
		c.altCond = add.altCond
	}
	return c
}

// validateValue validates a present-or-absent value at path. Returns the
// normalized value (coerced/defaulted) or an *Error. Absent handling (required /
// default / forbidden) is done by the OBJECT loop; validateValue is called only
// for values considered present (or explicitly for condition probing).
func (r *Rule) validateValue(value any, present bool, path []string, st *evalState) (any, *Error) {
	out, err := r.validateValueInner(value, present, path, st)
	// Joi's error.details[].context carries the failing `value` on every leaf
	// error. Stamping it at this ONE seam (rather than at ~30 errAt call sites)
	// covers every rule; only errors raised AT this node are stamped, so a nested
	// failure keeps the value of the node that actually failed.
	if err != nil && err.Ctx != nil && len(err.Path) == len(path) {
		if _, has := err.Ctx["value"]; !has && !noValueContext[err.Type] {
			err.Ctx["value"] = value
		}
		err = r.relabel(err, path)
	}
	// .message(text) overrides the wording of the rule it was attached to, and only
	// for an error raised AT this node — a nested key's failure keeps its own message.
	if err != nil && len(r.msgOverrides) > 0 && len(err.Path) == len(path) {
		if m, ok := r.msgOverrides[err.Type]; ok {
			err.Message = m
		}
	}
	return out, err
}

func (r *Rule) validateValueInner(value any, present bool, path []string, st *evalState) (any, *Error) {
	r = r.resolveEffective(st)

	// allow(...) bypass: verbatim-accept an allowed value (e.g. '', null)
	for _, a := range r.allowed {
		if looseEqual(a, value) {
			return value, nil
		}
	}

	// null handling: a null value that is not allowed fails, unless the type is
	// Any (which accepts null) — Joi rejects null for typed schemas by default.
	if value == nil {
		if r.typ == tAny && len(r.valids) == 0 {
			return nil, nil
		}
		if len(r.valids) > 0 && containsLoose(r.valids, nil) {
			return nil, nil
		}
		if len(r.valids) > 0 {
			return nil, r.checkOnly(value, path) // any.only, not <type>.base — see below
		}
		if r.typ == tAlternatives {
			// null is just a value the branches reject; letting it fall through to
			// typeBaseError produced `"value" must be a any`, a message Joi never emits
			return r.validateAlternatives(value, path, st)
		}
		return value, typeBaseError(path, r.typ)
	}

	// A `.valid()` set makes the rule an ENUM: Joi checks membership before it checks
	// the base type, so a wrong-TYPED value (5, true, {}, []) reports any.only listing
	// the valid values — not string.base. This is every enum field on every kind, which
	// is why it is decided here rather than inside each typed validator.
	if len(r.valids) > 0 && !isCoercibleTo(r.typ, value) {
		return nil, r.checkOnly(value, path)
	}

	switch r.typ {
	case tString:
		return r.validateString(value, path, st)
	case tNumber:
		return r.validateNumber(value, path, st)
	case tBoolean:
		return r.validateBoolean(value, path, st)
	case tDate:
		return r.validateDate(value, path, st)
	case tArray:
		return r.validateArray(value, path, st)
	case tObject:
		return r.validateObject(value, path, st)
	case tAlternatives:
		return r.validateAlternatives(value, path, st)
	default: // tAny
		return r.runValids(value, path, func(v any) (any, *Error) {
			return r.runCustoms(v, value, path, st)
		})
	}
}

// runValids enforces .valid()/.invalid() then calls next with the value.
func (r *Rule) runValids(value any, path []string, next func(any) (any, *Error)) (any, *Error) {
	if err := r.checkOnly(value, path); err != nil {
		return nil, err
	}
	return next(value)
}

// checkOnly is the .valid()/.invalid() pair, shared by every typed validator.
// It lives apart from runValids because the typed leaves (string/number/boolean/
// date) check the COERCED value, not the raw one, and each reaches this point at
// its own moment. .invalid() used to be enforced only on the untyped branch, so a
// `.Invalid()` on any typed rule was silently dropped — env.ts's
// .disallow('K_SERVICE',…) had to be hand-written as a custom, which then reported
// a bare field name where Joi reports the full path.
func (r *Rule) checkOnly(value any, path []string) *Error {
	if len(r.valids) > 0 && !containsLoose(r.valids, value) {
		return errCtx(errAt(path, "any.only", quote(label(path))+" must be "+onlyList(r.valids)), "valids", r.valids)
	}
	if len(r.invalids) > 0 && containsLoose(r.invalids, value) {
		// `value` is filled in by the caller's ctx pass, as it is for any.only.
		return errCtx(errAt(path, "any.invalid", quote(label(path))+" contains an invalid value"), "invalids", r.invalids)
	}
	return nil
}

// runCustoms runs the .custom() chain.
func (r *Rule) runCustoms(value, original any, path []string, st *evalState) (any, *Error) {
	cur := value
	// Joi's `helpers.error` closes over the value as it was on ENTRY to the rule chain
	// (validator.js:306 — `internals.rules` never reassigns that binding), so a custom
	// that fails AFTER an earlier custom transformed the value still reports the
	// pre-transform value in context.value. Reporting `cur` made every
	// LocalLink.custom(toKind) rejection echo the EXPANDED link where TS echoes what
	// the client sent.
	for _, fn := range r.customs {
		h := &Helper{Ctx: st.ctx, Original: original, Path: append([]string(nil), path...), st: st}
		out, err := fn(cur, h)
		if err != nil {
			if ne, ok := err.(*NestedValidationError); ok {
				e := errAt(path, "any.custom", ne.Err.Message)
				e.Ctx["error"] = map[string]any{"_original": ne.Original, "details": ne.Err.Details()}
				return nil, wrapCustom(e, value)
			}
			if ve, ok := err.(*Error); ok {
				if len(ve.Path) == 0 && !isRootAnchored(ve) {
					ve.Path = append([]string(nil), path...)
				}
				return nil, wrapCustom(ve, value)
			}
			return nil, wrapCustom(errAt(path, "any.custom", err.Error()), value)
		}
		cur = out
	}
	return cur, nil
}

// isRootAnchored distinguishes "no path set" from "deliberately anchored at the root".
// An ErrorAt on a root-level object stamps Ctx, so a nil Ctx means unset.
func isRootAnchored(e *Error) bool { return e.Ctx != nil }

// wrapCustom applies Joi's `.custom()` failure envelope: a THROWN error becomes
// `any.custom` whose message is `"<label>" failed custom validation because <thrown>`,
// carrying the thrown error in context. Errors that name a specific Joi type
// (helpers.error('string.pattern.base')) are Joi's OWN errors and pass through
// untouched — only the any.custom ones are wrapped.
//
// Wrapping matters beyond cosmetics: without the prefix the message names no field at
// all, so a caller with a 30-container workload was told "CPU must be at least 25
// millicore." and left to find which one.
func wrapCustom(e *Error, value any) *Error {
	if e == nil || e.Type != "any.custom" || e.customWrapped {
		return e
	}
	e.Message = quote(label(e.Path)) + " failed custom validation because " + e.Message
	e.customWrapped = true
	if e.Ctx == nil {
		e.Ctx = map[string]any{"label": label(e.Path)}
		if len(e.Path) > 0 {
			e.Ctx["key"] = pathSeg(e.Path[len(e.Path)-1])
		}
	}
	if _, has := e.Ctx["value"]; !has {
		e.Ctx["value"] = value
	}
	// Joi puts the THROWN object here. A plain `new Error(msg)` has no enumerable own
	// properties and serializes as {}; a thrown Joi ValidationError serializes as its
	// {_original, details}. NestedValidationError is how a custom says which it threw.
	if _, has := e.Ctx["error"]; !has {
		e.Ctx["error"] = map[string]any{}
	}
	return e
}

func (r *Rule) validateString(value any, path []string, st *evalState) (any, *Error) {
	if r.strFromNumber {
		if f, isNum := numericValue(value); isNum {
			if cs, exp := jsNumString(f); !exp {
				value = cs
			}
		}
	}
	s, ok := value.(string)
	if !ok {
		return nil, errAt(path, "string.base", quote(label(path))+" must be a string")
	}
	if r.strTrim {
		s = JSTrim(s)
	}
	if r.strLower {
		s = strings.ToLower(s)
	}
	if r.strUpper {
		s = strings.ToUpper(s)
	}
	// allow()/valid() are matched on the COERCED string, and a match short-circuits
	// everything after it — including the string.empty guard below.
	if r.acceptsCoerced(s) {
		return s, nil
	}
	if err := r.checkOnly(s, path); err != nil {
		return nil, err
	}
	if s == "" {
		// Joi.string() rejects '' unless .allow('')
		return nil, errAt(path, "string.empty", quote(label(path))+" is not allowed to be empty")
	}
	if r.strLen != nil && JSLen(s) != *r.strLen {
		return nil, errLimit(path, "string.length", quote(label(path))+" length must be "+itoa(*r.strLen)+" characters long", *r.strLen)
	}
	if r.strMin != nil && JSLen(s) < *r.strMin {
		return nil, errLimit(path, "string.min", quote(label(path))+" length must be at least "+itoa(*r.strMin)+" characters long", *r.strMin)
	}
	if r.strMax != nil && JSLen(s) > *r.strMax {
		return nil, errLimit(path, "string.max", quote(label(path))+" length must be less than or equal to "+itoa(*r.strMax)+" characters long", *r.strMax)
	}
	if r.strRegex != nil {
		m := r.strRegex.MatchString(s)
		if m == r.strRegexInvert {
			if r.strRegexInvert {
				return nil, errCtx(errAt(path, "string.pattern.invert.base", quote(label(path))+" with value "+quote(s)+" matches the inverted pattern: "+r.regexLiteral()), "regex", map[string]any{})
			}
			return nil, errCtx(errAt(path, "string.pattern.base", quote(label(path))+" with value "+quote(s)+" fails to match the required pattern: "+r.regexLiteral()), "regex", map[string]any{})
		}
	}
	if r.strEmail && !isEmail(s) {
		// Joi's string.email context lists the addresses that failed (it validates a
		// comma-separated list when `multiple` is set); for a single value that is the
		// value itself.
		return nil, errCtx(errAt(path, "string.email", quote(label(path))+" must be a valid email"), "invalids", []any{s})
	}
	if r.strBase64 && !isBase64(s) {
		return nil, errAt(path, "string.base64", quote(label(path))+" must be a valid base64 string")
	}
	if r.strURI {
		if r.strURIScheme != "" {
			// Joi special-cases these two exact values back to the plain string.uri
			// even when a scheme is configured (string.js's uri rule).
			if s == "http:/" || s == "https:/" {
				return nil, errAt(path, "string.uri", quote(label(path))+" must be a valid uri")
			}
			if !isURIScheme(s, r.strURIScheme) {
				return nil, errCtx(errAt(path, "string.uriCustomScheme", quote(label(path))+" must be a valid uri with a scheme matching the "+r.strURIScheme+" pattern"), "scheme", r.strURIScheme)
			}
		} else if !isURI(s) {
			return nil, errAt(path, "string.uri", quote(label(path))+" must be a valid uri")
		}
	}
	if r.strUUID && !isGUID(s) {
		return nil, errAt(path, "string.guid", quote(label(path))+" must be a valid GUID")
	}
	if r.strDomain && !isDomain(s, r.strDomainTLDs) {
		return nil, errAt(path, "string.domain", quote(label(path))+" must contain a valid domain name")
	}
	return r.runCustoms(s, value, path, st)
}

func (r *Rule) validateNumber(value any, path []string, st *evalState) (any, *Error) {
	f, lossy, ok := toNumberDetail(value)
	if !ok {
		return nil, errAt(path, "number.base", quote(label(path))+" must be a number")
	}
	// Joi's coerce-from-string rejects a literal whose value the double cannot hold —
	// "9007199254740993" parses to ...992 and no longer round-trips.
	if lossy {
		return nil, unsafeNumber(path)
	}
	// valids/invalids are checked on the COERCED value but BEFORE the base validate,
	// which is where Joi rounds .precision() — so `.valid(1).precision(0)` rejects 0.6
	// rather than rounding it into range. Rounding first made the enum unreachable.
	// Matching post-coercion is also what makes `.allow(0)` on the string "0" store the
	// NUMBER 0 rather than the string the client sent.
	if r.acceptsCoerced(f) {
		return f, nil
	}
	if err := r.checkOnly(f, path); err != nil {
		return nil, err
	}
	if math.IsInf(f, 0) {
		return nil, errAt(path, "number.infinity", quote(label(path))+" cannot be infinity")
	}
	if r.numPrecision != nil {
		p := math.Pow(10, float64(*r.numPrecision))
		f = math.Round(f*p) / p
	}
	if f == 0 {
		f = 0 // Joi normalises -0 to 0 twice over; JSON.stringify(-0) is "0"
	}
	// Joi rejects anything outside the safe-integer range unless .unsafe() — a value
	// past 2^53 does not survive the round trip through a double, so accepting it
	// stored a number the client never sent.
	if f > maxSafeInteger || f < -maxSafeInteger {
		return nil, unsafeNumber(path)
	}
	if r.numInt && f != math.Trunc(f) {
		return nil, errAt(path, "number.integer", quote(label(path))+" must be an integer")
	}
	if r.numPos && f <= 0 {
		return nil, errAt(path, "number.positive", quote(label(path))+" must be a positive number")
	}
	if r.numNeg && f >= 0 {
		return nil, errAt(path, "number.negative", quote(label(path))+" must be a negative number")
	}
	if r.numPort && (f < 0 || f > 65535 || f != math.Trunc(f)) {
		return nil, errAt(path, "number.port", quote(label(path))+" must be a valid port")
	}
	min, minLimit, err := r.refBound(r.numMin, r.numMinRef, path, st)
	if err != nil {
		return nil, err
	}
	if min != nil && f < *min {
		return nil, errLimit(path, "number.min", quote(label(path))+" must be greater than or equal to "+limitStr(minLimit), minLimit)
	}
	max, maxLimit, err := r.refBound(r.numMax, r.numMaxRef, path, st)
	if err != nil {
		return nil, err
	}
	if max != nil && f > *max {
		return nil, errLimit(path, "number.max", quote(label(path))+" must be less than or equal to "+limitStr(maxLimit), maxLimit)
	}
	return r.runCustoms(f, value, path, st)
}

// refBound resolves a min/max bound that may be a Joi.ref. A ref that does not resolve
// to a NUMBER is Joi's any.ref — the port used to drop the bound silently, so a body
// with `maxSize` but no `minSize` (the shape on all ten mk8s providers) was rejected by
// TS and accepted here. The second result is what the message and `limit` context carry:
// the ref descriptor for a ref, the number itself otherwise.
func (r *Rule) refBound(lit *float64, ref string, path []string, st *evalState) (*float64, any, *Error) {
	if ref == "" {
		if lit == nil {
			return nil, nil, nil
		}
		return lit, *lit, nil
	}
	var sib any
	if st != nil {
		sib, _ = siblingPath(st.parent, ref)
	}
	v, ok := refNumber(sib)
	if !ok {
		e := errAt(path, "any.ref", quote(label(path))+" limit references "+quote("ref:"+ref)+" which must be a number")
		e = errCtx(errCtx(e, "arg", "limit"), "reason", "must be a number")
		return nil, nil, errCtx(e, "ref", refDescriptor(ref))
	}
	return &v, refDescriptor(ref), nil
}

// refNumber accepts only an actual number: Joi asserts Common.isNumber on a resolved
// limit ref, so a sibling holding the STRING "5" fails the assert rather than coercing.
func refNumber(v any) (float64, bool) {
	switch t := v.(type) {
	case float64:
		return t, true
	case float32:
		return float64(t), true
	case int:
		return float64(t), true
	case int64:
		return float64(t), true
	}
	return 0, false
}

// refDescriptor is the object Joi puts in a message's `limit` / an any.ref's `ref`.
func refDescriptor(ref string) map[string]any {
	segs := strings.Split(ref, ".")
	path := make([]any, len(segs))
	for i, seg := range segs {
		path[i] = seg
	}
	return map[string]any{
		"adjust": nil, "in": false, "iterables": nil, "map": nil,
		"separator": ".", "type": "value", "ancestor": 1,
		"path": path, "depth": len(segs),
		"key": segs[len(segs)-1], "root": segs[0], "display": "ref:" + ref,
	}
}

// limitStr renders a bound the way Joi interpolates it: a ref as `ref:<path>`.
func limitStr(limit any) string {
	if m, ok := limit.(map[string]any); ok {
		if d, ok := m["display"].(string); ok {
			return d
		}
	}
	if f, ok := limit.(float64); ok {
		return ftoa(f)
	}
	return ""
}

// siblingPath resolves a ref key that may be DOTTED ("o.f"), walking into the nested
// object. A flat lookup read every dotted ref as absent, and silently.
func siblingPath(p parentView, key string) (any, bool) {
	if p == nil {
		return nil, false
	}
	if !strings.Contains(key, ".") {
		return p.sibling(key)
	}
	segs := strings.Split(key, ".")
	cur, ok := p.sibling(segs[0])
	if !ok {
		return nil, false
	}
	for _, seg := range segs[1:] {
		m, isMap := cur.(map[string]any)
		if !isMap {
			return nil, false
		}
		cur, ok = m[seg]
		if !ok {
			return nil, false
		}
	}
	return cur, true
}

func (r *Rule) validateBoolean(value any, path []string, st *evalState) (any, *Error) {
	var b bool
	switch v := value.(type) {
	case bool:
		b = v
	case string:
		switch strings.ToLower(v) {
		case "true":
			b = true
		case "false":
			b = false
		default:
			return nil, errAt(path, "boolean.base", quote(label(path))+" must be a boolean")
		}
	default:
		return nil, errAt(path, "boolean.base", quote(label(path))+" must be a boolean")
	}
	if r.acceptsCoerced(b) {
		return b, nil
	}
	if err := r.checkOnly(b, path); err != nil {
		return nil, err
	}
	return r.runCustoms(b, value, path, st)
}

func (r *Rule) validateDate(value any, path []string, st *evalState) (any, *Error) {
	var t time.Time
	switch v := value.(type) {
	case string:
		if r.dateIso {
			// Joi's .iso() gate: a non-string never reaches here (date.base below), and a
			// string that is not ISO-8601 is date.format, not date.base.
			if !isISODate(v) {
				return nil, errCtx(errAt(path, "date.format", quote(label(path))+" must be in ISO 8601 date format"), "format", "iso")
			}
			parsed, ok := parseDateString(v, true)
			if !ok {
				return nil, errCtx(errAt(path, "date.format", quote(label(path))+" must be in ISO 8601 date format"), "format", "iso")
			}
			t = parsed
			break
		}
		parsed, ok := parseDate(v)
		if !ok {
			return nil, errAt(path, "date.base", quote(label(path))+" must be a valid date")
		}
		t = parsed
	case float64, int, int64:
		// .iso() is a STRING format: Joi's isoDate test never matches a number, and the
		// value then reaches validate() as a non-string, which is date.base not date.format.
		if r.dateIso {
			return nil, errAt(path, "date.base", quote(label(path))+" must be a valid date")
		}
		switch n := v.(type) {
		case float64:
			t = time.UnixMilli(int64(n)).UTC()
		case int:
			t = time.UnixMilli(int64(n)).UTC()
		case int64:
			t = time.UnixMilli(n).UTC()
		}
	case time.Time:
		t = v
	default:
		return nil, errAt(path, "date.base", quote(label(path))+" must be a valid date")
	}
	// normalized output = ISO-8601 millis UTC (JS Date.toISOString())
	norm := t.UTC().Format("2006-01-02T15:04:05.000Z07:00")
	return r.runCustoms(norm, value, path, st)
}

func (r *Rule) validateArray(value any, path []string, st *evalState) (any, *Error) {
	arr, ok := value.([]any)
	if !ok {
		// A body decoded from JSON is always []any, but an object shaped IN PROCESS
		// can hold a concretely-typed slice — the engine's own link builder produces
		// []map[string]any. Those are arrays too; a bare `.([]any)` called them
		// `must be an array`.
		if s, converted := anySlice(value); converted {
			arr, ok = s, true
		}
	}
	if !ok {
		// .single(): coerce a lone scalar to a one-element array (Joi array.single).
		if r.arrSingle {
			arr = []any{value}
			ok = true
		}
	}
	if !ok {
		return nil, errAt(path, "array.base", quote(label(path))+" must be an array")
	}
	out := make([]any, 0, len(arr))
	for i, el := range arr {
		ep := append(append([]string(nil), path...), idxSeg(i))
		nv, err := r.validateArrayItem(el, ep, st)
		if err != nil {
			return nil, err
		}
		out = append(out, nv)
	}
	// Joi's `.items(X.required())`: each REQUIRED item schema must be satisfied by at
	// least one element, or the array errors as a whole. Absent entirely is still fine
	// (the key's own presence rule decides that) — it is an empty/short array that
	// fails. Without this an `.items(required)` array silently accepted `[]`.
	if misses := r.unmatchedRequiredItems(out, st); misses > 0 {
		e := errAt(path, "array.includesRequiredUnknowns",
			quote(label(path))+" does not contain "+itoa(misses)+" required value(s)")
		return nil, errCtx(e, "unknownMisses", misses)
	}
	if r.arrLen != nil && len(out) != *r.arrLen {
		return nil, errLimit(path, "array.length", quote(label(path))+" must contain "+itoa(*r.arrLen)+" items", *r.arrLen)
	}
	if r.arrMin != nil && len(out) < *r.arrMin {
		return nil, errLimit(path, "array.min", quote(label(path))+" must contain at least "+itoa(*r.arrMin)+" items", *r.arrMin)
	}
	if r.arrMax != nil && len(out) > *r.arrMax {
		return nil, errLimit(path, "array.max", quote(label(path))+" must contain less than or equal to "+itoa(*r.arrMax)+" items", *r.arrMax)
	}
	if r.arrUnique {
		if pos, dupePos, ok := findDuplicate(out, r.arrUniqueBy); ok {
			return nil, uniqueError(path, pos, dupePos, out[pos], out[dupePos], r.uniqueCtxPath())
		}
	}
	if r.arrSort {
		keys := make([]any, len(out))
		for i, el := range out {
			keys[i] = el
			if r.arrSortBy != "" {
				keys[i] = elemKey(el, r.arrSortBy)
			}
		}
		if !sortableTogether(keys) {
			return nil, errAt(path, "array.sort.mismatching", quote(label(path))+" cannot be sorted due to mismatching types")
		}
		if r.arrSortBy != "" {
			sortValuesBy(out, r.arrSortBy)
		} else {
			sortValues(out)
		}
	}
	// TS `.custom(makeUnique())` runs AFTER .sort() and SILENTLY drops repeats (keep
	// first), so the kept-element order matches Joi's makeUnique.
	if r.arrDedupe {
		out = dedupeValues(out, r.arrDedupeBy)
	}
	return r.runCustoms(out, value, path, st)
}

// dedupeValues collapses duplicates keeping the first occurrence (TS makeUnique →
// _.uniq / _.uniqBy). by="" keys on the whole element; else on element[by].
func dedupeValues(arr []any, by string) []any {
	seen := map[string]bool{}
	out := make([]any, 0, len(arr))
	for _, el := range arr {
		if by == "" && isContainer(el) {
			out = append(out, el)
			continue
		}
		key := toKey(el)
		if by != "" {
			if m, ok := el.(map[string]any); ok {
				key = toKey(m[by])
			}
		}
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, el)
	}
	return out
}

// validateArrayItem validates one element against the (possibly multiple) item
// schemas; Joi accepts an element if ANY item schema matches.
// anySlice widens any Go slice (other than []byte) to []any by reflection.
func anySlice(v any) ([]any, bool) {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() || rv.Kind() != reflect.Slice {
		return nil, false
	}
	if rv.Type().Elem().Kind() == reflect.Uint8 {
		return nil, false // []byte is a scalar, not an array
	}
	out := make([]any, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		out[i] = rv.Index(i).Interface()
	}
	return out, true
}

func (r *Rule) validateArrayItem(el any, path []string, st *evalState) (any, *Error) {
	if len(r.arrItems) == 0 {
		return el, nil
	}
	var firstErr *Error
	for _, it := range r.arrItems {
		nv, err := it.validateValue(el, true, path, st)
		if err == nil {
			return nv, nil
		}
		if firstErr == nil {
			firstErr = err
		}
	}
	if len(r.arrItems) == 1 {
		return nil, firstErr
	}
	return nil, includesError(path, el, firstErr)
}

// unknownMode is the policy an object applies to a key its schema does not declare,
// as Joi computes it (keys.js internals.unknown): a schema with NO keys and NO pattern
// allows everything and the stripUnknown PREFERENCE never reaches it, an explicit
// .unknown(true) likewise outranks the preference (`prefs.stripUnknown &&
// !schema._flags.unknown`), and only an unset flag inherits from the preferences.
//
// It is one function because the two evaluators must not drift here: the typed walk
// used to let stripUnknown win over .unknown(true) and to reject a pattern miss
// outright, so the same body was 200 on the tree path and 400 on the write path.
func (r *Rule) unknownMode(st *evalState) unknownMode {
	mode := r.unknown
	if mode == unkInherit && r.keyless && !r.hasPatterns() {
		mode = unkAllow // Joi.object() with no key schema — see Rule.keyless
	}
	if mode != unkAllow {
		if st.prefs.StripUnknown {
			mode = unkStrip
		} else if mode == unkInherit && st.prefs.AllowUnknown {
			mode = unkAllow
		}
	}
	if mode == unkInherit {
		mode = unkReject // Joi default
	}
	return mode
}

func (r *Rule) validateObject(value any, path []string, st *evalState) (any, *Error) {
	m, ok := value.(map[string]any)
	if !ok {
		return nil, typeBaseError(path, tObject)
	}
	out := map[string]any{}
	// Joi validates keys against a CLONE of the object it is mutating in place, so a
	// later key's ref (.when / Joi.ref / min(ref)) sees the earlier keys already
	// coerced and DEFAULTED. Reading the raw input instead made `value.when('op')`
	// miss an `op` that only exists because of its own .default('=') — and the typed
	// walk, which mutates the struct in place, already behaved Joi's way, so the two
	// evaluators disagreed on every defaulted ref.
	child := st.descend(&progressView{out: out, in: m})

	// declared keys (deterministic order)
	handled := map[string]bool{}
	for _, k := range r.keyOrder {
		sub := r.keys[k].resolveEffective(child)
		handled[k] = true
		raw, present := m[k]
		kp := append(append([]string(nil), path...), k)

		if !present {
			// required beats default, as in Joi: an absent required key errors even when
			// it carries a .default() (walk.go does the same for the typed path).
			if sub.presence == presRequired {
				return nil, requiredError(sub, kp)
			}
			// absent → default / presence
			if sub.hasDefault {
				dv := resolveDefaultVal(sub.defaultVal, mapView(m))
				// Run the default through the key's own rule so a defaulted OBJECT
				// picks up its nested defaults. `default={}` on memcache spec.options
				// otherwise stored a literal `{}` where TS stores the four inner
				// defaults — and TS gets those because its default is the populated
				// object. Validating the default is also just correct: a default that
				// does not satisfy its own schema is a bug either way.
				// A default that resolves to nothing (a Joi.ref to an absent sibling) leaves
				// the key UNSET, as Joi does — writing the nil stored a JSON null / a null
				// column where TS omits the key. walk.go already skipped it.
				if dv == nil {
					continue
				}
				if nv, verr := sub.validateValue(dv, true, append(append([]string(nil), path...), k), child); verr == nil {
					dv = nv
				}
				if !sub.strip {
					out[k] = dv
				}
				continue
			}
			eff := sub.presence
			if eff == presRequired {
				return nil, requiredError(sub, kp)
			}
			continue
		}
		if sub.presence == presForbidden {
			return nil, errCtx(forbiddenError(sub, kp), "value", raw)
		}
		// readOnly fields ARE validated. The earlier assumption that TS strips them
		// before validating was wrong: Joi's `.meta({readOnly:true})` is metadata
		// only, so TS answers 400 for `version: 0`, `version: 1.5`, a non-uuid `id`
		// and a malformed `links` entry. Skipping them here meant those returned
		// 201 — and worse, a bad `created`/`lastModified` sailed past the schema and
		// blew up later in the row unmarshal as a 500 leaking a Go time-parse error,
		// because the store decodes those columns into time.Time.
		//
		// The server still OWNS the values: the create path overwrites them and the
		// merge path pins them to the stored row (retainProps / the readOnly union),
		// so validating here only decides whether a client's echoed value is
		// well-formed, exactly as it does in TS.
		nv, err := sub.validateValue(raw, true, kp, child)
		if err != nil {
			return nil, err
		}
		if !sub.strip {
			out[k] = nv
		}
	}

	// unknown keys
	mode := r.unknownMode(st)
	unknownKeys := make([]string, 0)
	for k := range m {
		if !handled[k] {
			unknownKeys = append(unknownKeys, k)
		}
	}
	sort.Strings(unknownKeys)
	for _, k := range unknownKeys {
		kp := append(append([]string(nil), path...), k)
		// pattern map: the key must match one pattern's key schema, and that pattern's
		// value schema then judges the value
		if pv := r.matchPattern(k, kp, child); pv != nil {
			nv, err := pv.validateValue(m[k], true, kp, child)
			if err != nil {
				return nil, err
			}
			out[k] = nv
			continue
		}
		switch mode {
		case unkAllow:
			out[k] = m[k]
		case unkStrip:
			// drop
		default: // reject
			return nil, errCtx(errCtx(errAt(kp, "object.unknown", quote(label(kp))+" is not allowed"), "child", k), "value", m[k])
		}
	}

	// key-count bounds (Joi object.min / object.max)
	if r.objMin != nil && len(out) < *r.objMin {
		return nil, errLimit(path, "object.min", quote(label(path))+" must have at least "+itoa(*r.objMin)+" "+keyWord(*r.objMin), *r.objMin)
	}
	if r.objMax != nil && len(out) > *r.objMax {
		return nil, errLimit(path, "object.max", quote(label(path))+" must have less than or equal to "+itoa(*r.objMax)+" "+keyWord(*r.objMax), *r.objMax)
	}

	// presence relations
	if err := r.checkPresenceRelations(out, path); err != nil {
		return nil, err
	}
	return r.runCustoms(out, value, path, st)
}

func (r *Rule) checkPresenceRelations(out map[string]any, path []string) *Error {
	has := func(k string) bool { _, ok := out[k]; return ok }
	// Joi carries the peer set (and, where it applies, which peers were present) in the
	// context of every dependency error; a caller renders its own message from those.
	// Emitting the message alone left `peers`/`present`/`main` missing on 46 rules.
	peersCtx := func(e *Error, peers []string) *Error {
		ps := append([]string(nil), peers...)
		// `value` is the OBJECT the dependency is declared on, which Joi carries on
		// every one of these errors.
		e = errCtx(e, "value", out)
		return errCtx(errCtx(e, "peers", ps), "peersWithLabels", ps)
	}
	presentCtx := func(e *Error, peers []string) *Error {
		present := make([]string, 0, len(peers))
		for _, p := range peers {
			if has(p) {
				present = append(present, p)
			}
		}
		return errCtx(errCtx(peersCtx(e, peers), "present", present), "presentWithLabels", present)
	}
	missing := func(peers []string) *Error {
		return peersCtx(errAt(path, "object.missing",
			quote(label(path))+" must contain at least one of ["+strings.Join(peers, ", ")+"]"), peers)
	}
	count := func(peers []string) int {
		n := 0
		for _, p := range peers {
			if has(p) {
				n++
			}
		}
		return n
	}
	for _, peers := range r.or {
		if count(peers) == 0 {
			return missing(peers)
		}
	}
	for _, peers := range r.xor {
		switch n := count(peers); {
		case n == 0:
			return missing(peers)
		case n > 1:
			return presentCtx(errAt(path, "object.xor",
				quote(label(path))+" contains a conflict between exclusive peers ["+strings.Join(peers, ", ")+"]"), peers)
		}
	}
	for _, peers := range r.oxor {
		if count(peers) > 1 {
			return presentCtx(errAt(path, "object.oxor",
				quote(label(path))+" contains a conflict between optional exclusive peers ["+strings.Join(peers, ", ")+"]"), peers)
		}
	}
	for _, peers := range r.nand {
		if len(peers) > 0 && count(peers) == len(peers) {
			// nand names the FIRST peer and lists the rest — it is not the oxor message,
			// which is what this used to emit, with the wrong label as well.
			main, rest := peers[0], peers[1:]
			e := errAt(path, "object.nand",
				quote(main)+" must not exist simultaneously with ["+strings.Join(rest, ", ")+"]")
			e = errCtx(errCtx(e, "main", main), "mainWithLabel", main)
			return peersCtx(e, rest)
		}
	}
	return nil
}

func (r *Rule) validateAlternatives(value any, path []string, st *evalState) (any, *Error) {
	// conditional: pick branch by sibling ref
	if r.altCond != nil {
		var sib any
		present := false
		sib, present = siblingPath(st.parent, r.altCond.ref)
		var pick *Rule
		if whenMatches(r.altCond.is, sib, present, r.altCond.ref, st) {
			pick = r.altCond.then
		} else {
			pick = r.altCond.otherwise
		}
		if pick == nil {
			// no branch applies: Joi has nothing to match against and raises
			// alternatives.any rather than accepting the value
			return nil, errAt(path, "alternatives.any", quote(label(path))+" does not match any of the allowed types")
		}
		nv, err := pick.validateValue(value, true, path, st)
		if err != nil {
			return nil, err
		}
		// the alternatives node's OWN rules run after the branch, as they do for every
		// other type — returning here dropped a .custom() on the alternatives silently
		return r.runCustoms(nv, value, path, st)
	}
	var failures []*Error
	for _, alt := range r.alts {
		nv, err := alt.validateValue(value, true, path, st)
		if err == nil {
			return r.runCustoms(nv, value, path, st)
		}
		failures = append(failures, err)
	}
	return nil, alternativesError(failures, path)
}

// alternativesError ports Joi's internals.errors (alternatives.js): the reported error
// is chosen from the branch ERRORS, not from the shape of the branch schemas.
//
//   - no branch at all               → alternatives.any
//   - exactly one branch             → that branch's own error, verbatim
//   - every failure is a `*.base` or an `any.only` → alternatives.types listing the
//     type names and the valid values together
//   - exactly one failure is anything else (a nested error, a length/pattern rule) →
//     that one, because it is the informative one
//   - otherwise                      → alternatives.match carrying every branch's detail
//
// Deciding on the schema shape instead made every alternative in the schema set answer
// "does not match any of the allowed types", naming neither the field nor the reason.
func alternativesError(failures []*Error, path []string) *Error {
	if len(failures) == 0 {
		return errAt(path, "alternatives.any", quote(label(path))+" does not match any of the allowed types")
	}
	if len(failures) == 1 {
		return failures[0]
	}
	var types []string
	seen := map[string]bool{}
	addType := func(t string) {
		if !seen[t] {
			seen[t] = true
			types = append(types, t)
		}
	}
	var complex []*Error
	for _, f := range failures {
		if len(f.Path) != len(path) {
			complex = append(complex, f) // an error from INSIDE an object/array branch
			continue
		}
		if f.Type == "any.only" {
			if vs, ok := f.Ctx["valids"].([]any); ok {
				for _, v := range vs {
					addType(valueStr(v))
				}
			}
			continue
		}
		typ, code, ok := strings.Cut(f.Type, ".")
		if !ok || code != "base" {
			complex = append(complex, f)
			continue
		}
		addType(typ)
	}
	if len(complex) == 0 {
		e := errAt(path, "alternatives.types", quote(label(path))+" must be one of ["+strings.Join(types, ", ")+"]")
		return errCtx(e, "types", types)
	}
	if len(complex) == 1 {
		return complex[0]
	}
	details := make([]Detail, 0, len(failures))
	msgs := make([]string, 0, len(failures))
	for _, f := range failures {
		details = append(details, f.Details()...)
		msgs = append(msgs, f.Message)
	}
	e := errAt(path, "alternatives.match", quote(label(path))+" does not match any of the allowed types")
	return errCtx(errCtx(e, "message", joinMessages(msgs)), "details", details)
}

// joinMessages is Joi's Errors.details message join (joi/lib/errors.js): identical
// messages collapse to one before the ". " join, so N branches failing the same way
// report it once. details[] is NOT deduped, matching Joi.
func joinMessages(msgs []string) string {
	if len(msgs) > 1 {
		seen := make(map[string]bool, len(msgs))
		uniq := msgs[:0:0]
		for _, m := range msgs {
			if seen[m] {
				continue
			}
			seen[m] = true
			uniq = append(uniq, m)
		}
		msgs = uniq
	}
	return strings.Join(msgs, ". ")
}

// --- helpers ---

// maxSafeInteger is JS's Number.MAX_SAFE_INTEGER, the bound Joi's number.unsafe uses.
const maxSafeInteger = 9007199254740991

func unsafeNumber(path []string) *Error {
	return errAt(path, "number.unsafe", quote(label(path))+" must be a safe number")
}

func toNumber(v any) (float64, bool) {
	f, _, ok := toNumberDetail(v)
	return f, ok
}

// toNumberDetail is toNumber plus Joi's coerce-from-string precision check: lossy
// reports a numeric literal the double silently rounded (number.unsafe), as opposed to
// one that is not a number at all (number.base).
func toNumberDetail(v any) (f float64, lossy, ok bool) {
	switch t := v.(type) {
	case float64:
		return t, false, true
	case float32:
		return float64(t), false, true
	case int:
		return float64(t), false, true
	case int64:
		return float64(t), false, true
	case string:
		// Joi coerces a string with its own numberRx, not the platform parser. Go's
		// ParseFloat additionally accepts NaN/Inf/Infinity and digit separators
		// (`1_0`), so those validated here and then failed to MARSHAL — a 500, or a row
		// Postgres would not take — on any of the ~250 number fields.
		s := strings.TrimSpace(t)
		if !joiNumberRe.MatchString(s) {
			return 0, false, false
		}
		n, err := strconv.ParseFloat(s, 64)
		if err != nil {
			// An exponent the double cannot hold ("1e400", "1e-400") is Joi's
			// number.unsafe, not number.base: the literal IS a number, it just does not
			// survive the conversion.
			if ne, isNum := err.(*strconv.NumError); isNum && ne.Err == strconv.ErrRange {
				return 0, true, true
			}
			return 0, false, false
		}
		if math.IsNaN(n) {
			return 0, false, false
		}
		if n == 0 {
			n = 0 // -0
		}
		if math.IsInf(n, 0) {
			// "1e400": Joi reports number.unsafe, not number.base
			return 0, true, true
		}
		if n == 0 && reNumNonZero.MatchString(mantissaOf(s)) {
			// "1e-400" underflows to zero: the literal is a number the double cannot
			// hold, which is Joi's number.unsafe (its exponent round-trip yields NaN).
			return 0, true, true
		}
		if !strings.ContainsAny(s, "eE") {
			// Joi compares the parsed value's own toString with the normalised literal;
			// a disagreement means the double lost digits.
			if js, exp := jsNumString(n); !exp && js != normalizeDecimal(s) {
				return 0, true, true
			}
		}
		return n, false, true
	}
	return 0, false, false
}

// jsNumString is Number.prototype.toString for a double in the range JS renders
// WITHOUT an exponent; exp reports that JS would have used one (in which case Joi's
// coerce check passes the value through untouched).
func jsNumString(f float64) (s string, exp bool) {
	a := math.Abs(f)
	if f != 0 && (a >= 1e21 || a < 1e-6) {
		return "", true
	}
	return ftoa(f), false
}

// mantissaOf is the part of a numeric literal before its exponent.
func mantissaOf(s string) string {
	if i := strings.IndexAny(s, "eE"); i >= 0 {
		return s[:i]
	}
	return s
}

var (
	reNumNonZero       = regexp.MustCompile(`[1-9]`)
	reNumTrailDotZeros = regexp.MustCompile(`\.0*$`)
	reNumLeadingDot    = regexp.MustCompile(`^(-?)\.([^.]*)$`)
	reNumLeadingZeros  = regexp.MustCompile(`^(-?)0+([0-9])`)
	reNumTrailZeros    = regexp.MustCompile(`0+$`)
)

// normalizeDecimal ports Joi's internals.normalizeDecimal (number.js): the canonical
// spelling of a decimal literal, so "5.0", "+5" and "005" all compare equal to "5".
func normalizeDecimal(s string) string {
	s = strings.TrimPrefix(s, "+")
	s = reNumTrailDotZeros.ReplaceAllString(s, "")
	s = reNumLeadingDot.ReplaceAllString(s, "${1}0.${2}")
	s = reNumLeadingZeros.ReplaceAllString(s, "${1}${2}")
	if strings.Contains(s, ".") && strings.HasSuffix(s, "0") {
		s = reNumTrailZeros.ReplaceAllString(s, "")
	}
	if s == "-0" {
		return "0"
	}
	return s
}

// joiNumberRe is Joi's own numberRx (lib/types/number.js): an optional sign, digits
// with an optional fraction, and an optional exponent. Deliberately NOT Go's parser.
var joiNumberRe = regexp.MustCompile(`^[+-]?(?:\d+(?:\.\d*)?|\.\d+)(?:[eE][+-]?\d+)?$`)

// dateLayoutsUTC are the forms JS's Date parser treats as UTC (date-only) or that
// carry their own offset.
var dateLayoutsUTC = []string{
	time.RFC3339Nano, time.RFC3339, "2006-01-02T15:04:05.000Z07:00",
	"2006-01-02", "2006-01", "2006",
}

// dateLayoutsLocal are the forms with a time but NO offset — and the slash date, which
// ECMA-262 also reads as LOCAL (a slash date carries no offset, so `new Date("2020/01/02")`
// is midnight where the server stands). Listing it as UTC shifted every one of them.
var dateLayoutsLocal = []string{
	"2006-01-02T15:04:05.999999999", "2006-01-02 15:04:05.999999999",
	"2006-01-02T15:04:05", "2006-01-02 15:04:05", "2006-01-02T15:04", "2006-01-02 15:04",
	"2006/01/02", "2006/01/02 15:04:05",
	"Jan 2 2006", "Jan 02 2006", "January 2 2006",
}

// dateNumericRe is Joi's own "normalize number string" test (date.js internals.parse):
// a string of digits is an epoch-millis TIMESTAMP, checked BEFORE any calendar format.
// Running the layouts first read "2020" as the year 2020 where Joi reads it as 2020ms
// after the epoch.
var dateNumericRe = regexp.MustCompile(`^[+-]?\d+(\.\d+)?$`)

var reDigitComma = regexp.MustCompile(`\d,\d`)

func parseDate(s string) (time.Time, bool) { return parseDateString(s, false) }

// parseDateString parses a date the way `new Date(s)` does. iso skips the
// numeric-timestamp branch, which Joi's iso format never takes.
func parseDateString(s string, iso bool) (time.Time, bool) {
	if !iso && dateNumericRe.MatchString(strings.TrimSpace(s)) {
		f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
		if err != nil {
			return time.Time{}, false
		}
		return time.UnixMilli(int64(f)).UTC(), true
	}
	// Go's time parser accepts a COMMA as the fractional-second separator; JS's Date
	// does not, so "2020-01-02T03:04:05,5" is a date here and Invalid Date there.
	if reDigitComma.MatchString(s) {
		return time.Time{}, false
	}
	for _, layout := range dateLayoutsUTC {
		if t, err := time.Parse(layout, s); err == nil {
			return t, true
		}
	}
	// A datetime without an offset is local time in JS — `new Date("2030-01-01T00:00:00")`
	// is midnight in the server's zone, not UTC. Rejecting these outright made Go
	// refuse the most common hand-written form, which TS accepts.
	for _, layout := range dateLayoutsLocal {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

// looseEqual is Joi's Values.get — deep equality, plus the one bridge Go needs that JS
// does not: a literal written as int and a JSON number decoded as float64 are the same
// value. It deliberately does NOT bridge strings to numbers: Joi compares the COERCED
// value against the set, so `.valid(1)` never matches the string "1" (a number schema
// has already turned "1" into 1 by then, and any other schema has not).
func looseEqual(a, b any) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	if reflect.DeepEqual(a, b) {
		return true
	}
	if af, aok := numericValue(a); aok {
		if bf, bok := numericValue(b); bok {
			return af == bf
		}
	}
	return false
}

// numericValue is the float value of a Go NUMBER (never a numeric string).
func numericValue(v any) (float64, bool) {
	switch t := v.(type) {
	case float64:
		return t, true
	case float32:
		return float64(t), true
	case int:
		return float64(t), true
	case int8:
		return float64(t), true
	case int16:
		return float64(t), true
	case int32:
		return float64(t), true
	case int64:
		return float64(t), true
	case uint:
		return float64(t), true
	case uint8:
		return float64(t), true
	case uint16:
		return float64(t), true
	case uint32:
		return float64(t), true
	case uint64:
		return float64(t), true
	}
	return 0, false
}

// acceptsCoerced reports a value that .allow()/.valid() lets through AFTER coercion.
// Joi checks the valids set between coercion and the base validate, so a value that only
// becomes allowed once trimmed or parsed still passes, and it short-circuits the base
// validate and every rule — `.trim().allow(”)` on "   " is accepted as "", where
// comparing the RAW value hit string.empty instead.
func (r *Rule) acceptsCoerced(v any) bool {
	return containsLoose(r.allowed, v) || (len(r.valids) > 0 && containsLoose(r.valids, v))
}

func containsLoose(list []any, v any) bool {
	for _, e := range list {
		if looseEqual(e, v) {
			return true
		}
	}
	return false
}

// findDuplicate returns the index of the first repeat and the index it repeats, which
// is what Joi's array.unique error points at.
func findDuplicate(arr []any, by string) (pos, dupePos int, found bool) {
	seen := map[string]int{}
	for i, el := range arr {
		var key string
		if by != "" {
			if m, ok := el.(map[string]any); ok {
				key = toKey(m[by])
			} else {
				key = toKey(el)
			}
		} else {
			key = toKey(el)
		}
		if j, ok := seen[key]; ok {
			return i, j, true
		}
		seen[key] = i
	}
	return 0, 0, false
}

// uniqueError is Joi's array.unique: reported at the DUPLICATE ELEMENT, not at the
// array, and carrying the two positions and the repeated value. Blaming the array left
// a caller with a 40-entry list and no way to tell which two entries collided.
// dupe is the element that REPEATS (context.value, at pos) and prior is the element it
// repeats (context.dupeValue, at dupePos) — Joi reports both, and reporting the repeat
// twice told a caller the two colliding entries were identical when they were not.
func uniqueError(path []string, pos, dupePos int, dupe, prior any, byPath string) *Error {
	ep := append(append([]string(nil), path...), idxSeg(pos))
	e := errAt(ep, "array.unique", quote(label(ep))+" contains a duplicate value")
	e = errCtx(errCtx(errCtx(e, "pos", pos), "dupePos", dupePos), "dupeValue", prior)
	if byPath != "" {
		e = errCtx(e, "path", byPath)
	}
	return errCtx(e, "value", dupe)
}

// sortValuesBy is Joi's `.sort({ by: field })`: order elements by one of their
// object keys. Only the plain `.sort()` form was implemented, so the by-field
// arrays (env collections, probe headers, quota overrides) persisted in whatever
// order the client submitted — a stored-document difference, not just a response
// one, since these arrays are stored sorted.
func sortValuesBy(arr []any, field string) {
	sort.SliceStable(arr, func(i, j int) bool {
		return compareValues(elemKey(arr[i], field), elemKey(arr[j], field)) < 0
	})
}

// elemKey reads the sort key off an element; a non-object (or a missing key)
// sorts as nil, which compareValues orders consistently.
func elemKey(v any, field string) any {
	if m, ok := v.(map[string]any); ok {
		return m[field]
	}
	return nil
}

func sortValues(arr []any) {
	sort.SliceStable(arr, func(i, j int) bool {
		return compareValues(arr[i], arr[j]) < 0
	})
}

func compareValues(a, b any) int {
	af, aok := toNumber(a)
	bf, bok := toNumber(b)
	if aok && bok {
		switch {
		case af < bf:
			return -1
		case af > bf:
			return 1
		default:
			return 0
		}
	}
	return strings.Compare(toKey(a), toKey(b))
}

// toKey is the identity of a value for unique/dedupe/sort: scalars by value, a
// container by its canonical JSON (Joi's array.unique compares objects deeply).
func toKey(v any) string {
	switch t := v.(type) {
	case nil:
		return "z:"
	case string:
		return "s:" + t
	case float64:
		return "n:" + strconv.FormatFloat(t, 'g', -1, 64)
	case bool:
		if t {
			return "b:1"
		}
		return "b:0"
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return "x:" + strconv.Quote(reflect.TypeOf(v).String())
		}
		return "j:" + string(b)
	}
}

// isContainer reports a value lodash's _.uniq would compare by reference — which
// never matches, so makeUnique() (Dedupe without a key) leaves such elements alone.
func isContainer(v any) bool {
	switch v.(type) {
	case map[string]any, []any:
		return true
	}
	return false
}

func typeName(t ruleType) string {
	switch t {
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
	default:
		return "any"
	}
}

func humanType(t ruleType) string {
	switch t {
	case tString:
		return "string"
	case tNumber:
		return "number"
	case tBoolean:
		return "boolean"
	case tDate:
		return "valid date"
	case tArray:
		return "array"
	case tObject:
		return "object"
	default:
		return "any"
	}
}

func onlyList(valids []any) string {
	if len(valids) == 1 {
		return "[" + valueStr(valids[0]) + "]"
	}
	parts := make([]string, len(valids))
	for i, v := range valids {
		parts[i] = valueStr(v)
	}
	return "one of [" + strings.Join(parts, ", ") + "]"
}

func valueStr(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case nil:
		return "null"
	case bool:
		if t {
			return "true"
		}
		return "false"
	default:
		if f, ok := numericValue(v); ok {
			return ftoa(f)
		}
		return ""
	}
}

func quote(s string) string { return "\"" + s + "\"" }
func itoa(n int) string     { return strconv.Itoa(n) }

// ftoa renders a number the way JavaScript's String(n) does, which is what Joi
// interpolates into every limit and valid-value message. Go's 'g' verb switches to
// exponent form far earlier and zero-pads the exponent, so a `.max(2**32-1)` read
// "must be less than or equal to 4.294967295e+09" here and "… 4294967295" in TS — on
// every number rule in the schema set, not just the large ones (1e6 also differs).
func ftoa(f float64) string {
	switch {
	case math.IsNaN(f):
		return "NaN"
	case math.IsInf(f, 1):
		return "Infinity"
	case math.IsInf(f, -1):
		return "-Infinity"
	case f == 0:
		return "0" // String(-0) is "0"
	}
	if a := math.Abs(f); a >= 1e-6 && a < 1e21 {
		return strconv.FormatFloat(f, 'f', -1, 64)
	}
	// exponent form: JS writes e+21 / e-7, Go writes e+21 / e-07
	e := strconv.FormatFloat(f, 'e', -1, 64)
	i := strings.LastIndexByte(e, 'e')
	mant, exp := e[:i], e[i+1:]
	sign := ""
	if exp[0] == '+' || exp[0] == '-' {
		sign, exp = string(exp[0]), exp[1:]
	}
	exp = strings.TrimLeft(exp, "0")
	if exp == "" {
		exp = "0"
	}
	if sign == "-" {
		return mant + "e-" + exp
	}
	return mant + "e+" + exp
}

// altCondPresence resolves the presence of the branch an alternatives().conditional()
// would pick for the current parent. presDefault when no branch applies.
func (r *Rule) altCondPresence(st *evalState) presence {
	var sib any
	present := false
	if st != nil {
		sib, present = siblingPath(st.parent, r.altCond.ref)
	}
	pick := r.altCond.otherwise
	if whenMatches(r.altCond.is, sib, present, r.altCond.ref, st) {
		pick = r.altCond.then
	}
	if pick == nil {
		return presDefault
	}
	return pick.presence
}

// keyWord is Joi's pluralization for the object key-count messages
// ({if(#limit == 1, "key", "keys")}); the port said "children".
func keyWord(n int) string {
	if n == 1 {
		return "key"
	}
	return "keys"
}

// NestedValidationError is the failure of a schema run INSIDE a custom — base.ts's
// checkImageName does `throw ImageName.validate(v).error`, and Joi serializes that
// thrown ValidationError into `context.error` as {_original, details}, where a plain
// Error serializes as {}. A custom returns this to say it threw the former.
type NestedValidationError struct {
	Original any
	Err      *Error
}

func (e *NestedValidationError) Error() string { return e.Err.Message }

// Nested wraps a sub-schema failure for return from a custom (see the type doc).
// A nil inner error returns nil, so a caller can pass a Validate result straight through.
func Nested(original any, verr *Error) error {
	if verr == nil {
		return nil
	}
	return &NestedValidationError{Original: original, Err: verr}
}

// CustomErrorAt builds the error Joi raises when a `.custom()` at `path` throws a plain
// Error: an `any.custom` carrying the wrapped message and the usual label/key/value/error
// context. Exported for the passes that run OUTSIDE the schema walk but must answer in
// the schema's envelope — the data-service expands links before validating, where TS
// expands inside a Joi custom, and a bare error there escapes the envelope entirely.
func CustomErrorAt(path []string, value any, msg string) *Error {
	return wrapCustom(errAt(path, "any.custom", msg), value)
}

// isCoercibleTo reports whether Joi's convert:true could turn `value` into the rule's
// type at all. Used only to decide enum-vs-base error precedence.
func isCoercibleTo(t ruleType, value any) bool {
	switch t {
	case tString:
		_, ok := value.(string)
		return ok
	case tNumber:
		_, ok := toNumber(value)
		return ok
	case tBoolean:
		switch v := value.(type) {
		case bool:
			return true
		case string:
			return v == "true" || v == "false"
		}
		return false
	}
	return true
}

// progressView is the object a key's refs are resolved against while the object is
// still being validated: a key already processed reads back its NORMALIZED value, one
// not yet reached reads the raw input. It mirrors Joi mutating its `value` clone in
// place as the key loop advances.
type progressView struct {
	out map[string]any
	in  map[string]any
}

func (p *progressView) sibling(key string) (any, bool) {
	if v, ok := p.out[key]; ok {
		return v, true
	}
	v, ok := p.in[key]
	return v, ok
}

func (p *progressView) asMap() map[string]any {
	m := make(map[string]any, len(p.in)+len(p.out))
	for k, v := range p.in {
		m[k] = v
	}
	for k, v := range p.out {
		m[k] = v
	}
	return m
}

// sortableTogether reports whether every sort key is of the same comparable kind. Joi
// raises array.sort.mismatching rather than falling back on a stringified compare, which
// silently reordered a mixed array instead of rejecting it.
func sortableTogether(keys []any) bool {
	kind := ""
	for _, k := range keys {
		var this string
		switch k.(type) {
		case string:
			this = "string"
		case float64, float32, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			this = "number"
		default:
			return false
		}
		if kind == "" {
			kind = this
		} else if kind != this {
			return false
		}
	}
	return true
}

// labelOf is the label an error raised AT this rule's node carries: the .label()
// override when the schema set one, otherwise the path. The override used to be stored
// and never read, so every `.Label(...)` in the schema set was ineffective.
func (r *Rule) labelOf(path []string) string {
	if r.label != "" {
		return r.label
	}
	return label(path)
}

// relabel rewrites an error raised at this rule's own node to use its .label() override.
func (r *Rule) relabel(e *Error, path []string) *Error {
	if e == nil || r.label == "" || len(e.Path) != len(path) {
		return e
	}
	old := label(path)
	if old == r.label {
		return e
	}
	e.Message = strings.Replace(e.Message, quote(old), quote(r.label), 1)
	if e.Ctx != nil {
		e.Ctx["label"] = r.label
	}
	return e
}

// unmatchedRequiredItems counts the rule's REQUIRED item schemas that no element of the
// array satisfies — Joi's `unknownMisses`.
func (r *Rule) unmatchedRequiredItems(out []any, st *evalState) int {
	misses := 0
	for _, item := range r.arrItems {
		if item == nil || item.presence != presRequired {
			continue
		}
		matched := false
		for _, el := range out {
			if _, err := item.validateValue(el, true, nil, st); err == nil {
				matched = true
				break
			}
		}
		if !matched {
			misses++
		}
	}
	return misses
}
