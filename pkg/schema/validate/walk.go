package validate

// walk.go is phase two of typed validation: run a rule graph over a struct the
// decoder filled, in schema order, writing defaults and coercions back into the
// fields. Leaf semantics (string/number/boolean/date/any/alternatives, allow/valid,
// customs) are the SAME code the generic evaluator runs — a leaf is read out of its
// field as the generic value, judged by validateValue, and the normalized result is
// written back — so the two modes cannot drift on a leaf. Only the container logic
// (objects, arrays, maps) is typed here, and it mirrors validateObject/validateArray
// line for line.

import (
	"encoding"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"time"

	"github.com/controlplane-com/libs-go/pkg/schema/wirejson"
)

// Validate runs rule over the decoded struct. Fields the body did not carry but a
// caller set between DecodeInto and Validate count as present (a pre-schema hook
// writes the row the same way TS rewrites req.body before Joi runs).
func (d *Decoded) Validate(rule *Rule, ctx *Context, prefs Prefs) *Error {
	return d.ValidateAt(rule, nil, ctx, prefs)
}

// ValidateAt is Validate with every error reported UNDER base — for a struct that is
// one field of a larger document (a command's `spec`).
func (d *Decoded) ValidateAt(rule *Rule, base []string, ctx *Context, prefs Prefs) *Error {
	if ctx == nil {
		ctx = &Context{}
	}
	st := &evalState{ctx: ctx, prefs: prefs, order: &wireOrder{tab: d.order, base: len(base)}}
	if g, ok := d.root.raw[-1]; ok {
		// the body was not an object; the root rule says so in its own words
		_, err := rule.validateValue(g, true, base, st)
		if err == nil {
			err = rule.stampGeneric(typeBaseError(base, tObject), base, g)
		}
		return err
	}
	return rule.walkTyped(d.dst, d.root, base, st)
}

// Value returns the decoded struct (the pointer DecodeInto was given).
func (d *Decoded) Value() any { return d.dst.Addr().Interface() }

// Present reports whether a top-level key was in the body, or has since been set on
// the struct (the same notion of presence Validate uses).
func (d *Decoded) Present(jsonName string) bool {
	if d.root == nil || d.root.state == nil {
		return false
	}
	p := planFor(d.dst.Type())
	i, ok := p.byName[jsonName]
	if !ok {
		return false
	}
	return d.root.state[i] != stAbsent || !isZero(p.field(d.dst, i))
}

// MarkPresent records a top-level key as present even though the body did not carry
// it, so the walk descends into it (applying its nested defaults) and an apply-merge
// treats it as written rather than copying the stored value over. TS expresses this by
// rewriting req.body before Joi runs (volumeset.ts:914 `if (!req.body.status)
// req.body.status = {}`); a Go struct field is always physically there, so assignment
// alone cannot say it.
func (d *Decoded) MarkPresent(jsonName string) {
	if d.root == nil || d.root.state == nil {
		return
	}
	p := planFor(d.dst.Type())
	i, ok := p.byName[jsonName]
	if !ok || i >= len(d.root.state) {
		return
	}
	if d.root.state[i] == stAbsent {
		d.root.state[i] = stFits
	}
}

// Decode is DecodeInto + Validate against the type's own schema, for callers with
// nothing to do in between.
func Decode[T any](src Source, ctx *Context, prefs Prefs) (*T, *Error) {
	out := new(T)
	d, err := DecodeInto(out, src)
	if err != nil {
		return nil, &Error{Type: "syntax", Message: err.Error()}
	}
	if verr := d.Validate(For[T]().Rule(), ctx, prefs); verr != nil {
		return nil, verr
	}
	return out, nil
}

// walkField validates one field given what the decoder recorded for it.
func (r *Rule) walkField(v reflect.Value, state uint8, raw any, kid *node, path []string, st *evalState) *Error {
	switch state {
	case stNull:
		out, err := r.validateValue(nil, true, path, st)
		if err != nil {
			return err
		}
		if out != nil && !tryAssign(v, out) {
			// a custom turned the null into a value the field cannot hold: report it as
			// the field's own type error. assign() panics, and this seam is request-data
			// driven (which custom branch runs depends on the body), so it was a 500.
			return r.stampGeneric(typeBaseError(path, destRuleType(v)), path, out)
		}
		return nil
	case stRaw:
		out, err := r.validateValue(raw, true, path, st)
		if err != nil {
			return err
		}
		if !tryAssign(v, out) {
			// the rule accepted a shape the field cannot hold (an `any` rule over a typed
			// column): say so in the field's terms rather than storing nothing
			return r.stampGeneric(typeBaseError(path, destRuleType(v)), path, raw)
		}
		return nil
	default:
		return r.walkTyped(v, kid, path, st)
	}
}

// walkTyped validates a field that holds a value of its own type.
func (r *Rule) walkTyped(v reflect.Value, kid *node, path []string, st *evalState) *Error {
	inner, kind, ok := peel(v)
	if !ok {
		return r.walkField(v, stNull, nil, nil, path, st)
	}
	switch {
	case r.typ == tObject && kind == dkStruct:
		return r.stamp(r.walkStruct(inner, kid, path, st), path, inner)
	case r.typ == tObject && kind == dkMap:
		return r.stamp(r.walkMap(inner, kid, path, st), path, inner)
	case r.typ == tArray && kind == dkSlice:
		return r.stamp(r.walkSlice(inner, kid, path, st), path, inner)
	case kind == dkAny:
		g := inner.Interface()
		out, err := r.validateValue(g, true, path, st)
		if err != nil {
			return err
		}
		inner.Set(reflect.ValueOf(&out).Elem())
		return nil
	case kind == dkStruct || kind == dkSlice || kind == dkMap:
		if r.typ == tAny {
			if len(r.customs) == 0 && len(r.valids) == 0 && len(r.invalids) == 0 {
				return nil
			}
			// an `any` rule that still has something to say judges the container as the
			// generic tree, which is what the generic evaluator does with it
			out, err := r.validateValue(treeOf(inner), true, path, st)
			if err != nil {
				return err
			}
			if out != nil && !tryAssign(inner, out) {
				return r.stampGeneric(typeBaseError(path, destRuleType(inner)), path, out)
			}
			return nil
		}
		// A rule whose type the field cannot hold is a schema/struct skew — but it is
		// reached per REQUEST (the decoder only gets here once the body fits the field),
		// and the generic evaluator answers it with a clean <type>.base. Panicking here
		// turned that 400 into a 500 on the main write path.
		return r.stampGeneric(typeBaseError(path, r.typ), path, treeOf(inner))
	default:
		return r.walkScalar(inner, kind, path, st)
	}
}

// walkScalar judges a leaf through the generic evaluator and writes the result back.
func (r *Rule) walkScalar(inner reflect.Value, kind destKind, path []string, st *evalState) *Error {
	g := scalarOf(inner, kind)
	out, err := r.validateValue(g, true, path, st)
	if err != nil {
		return err
	}
	if out == nil {
		clear(inner)
		return nil
	}
	if !tryAssign(inner, out) {
		return r.stampGeneric(typeBaseError(path, destRuleType(inner)), path, out)
	}
	return nil
}

func (r *Rule) walkStruct(v reflect.Value, kid *node, path []string, st *evalState) *Error {
	plan := planFor(v.Type())
	if kid == nil {
		kid = &node{state: make([]uint8, len(plan.fields))}
	}
	if r.untypedCustom {
		panic(fmt.Sprintf("validate: %s at %q carries a custom written for a map; use CustomT/RefineT", v.Type(), label(path)))
	}
	child := st.descend(&structView{plan: plan, v: v, n: kid})
	set := make([]bool, len(plan.fields))
	handled := map[string]bool{}

	for _, k := range r.keyOrder {
		sub := r.keys[k].resolveEffective(child)
		handled[k] = true
		i, ok := plan.byName[k]
		if !ok {
			// A key the schema declares but the row has no field for (a gvc-scoped
			// kind's `gvc`, which the wire accepts and the row derives from its scope):
			// the rule still judges what was sent, and the value then has nowhere to go.
			kp := append(append([]string(nil), path...), k)
			if u, sent := kid.unknownValue(k); sent {
				if _, err := sub.validateValue(u, true, kp, child); err != nil {
					return err
				}
				kid.dropUnknown(k)
				continue
			}
			eff := sub.presence
			if eff == presRequired {
				return requiredError(sub, kp)
			}
			continue
		}
		if plan.fields[i].skip {
			continue
		}
		f := plan.field(v, i)
		state, raw, kd := kid.state[i], kid.raw[i], kid.kids[i]
		kp := append(append([]string(nil), path...), k)
		if state == stAbsent && !isZero(f) {
			state = stFits // set in process between decode and validate
		}
		if state == stFits && isNilish(f) {
			state = stAbsent // decoded, then cleared in process: a hook removed it
		}
		if state == stAbsent {
			// required beats default: Joi errors on an absent required key even when it
			// carries a .default(), so taking the default branch first silently supplied
			// a value where TS answers any.required.
			if sub.presence == presRequired {
				return requiredError(sub, kp)
			}
			if sub.hasDefault {
				dv := resolveDefaultVal(sub.defaultVal, child.parent)
				// A default the field cannot hold (a ref to a sibling that is itself
				// malformed — description defaulting to a numeric name) is left unset:
				// Joi does not validate defaults, and the sibling's own rule reports the
				// error when the walk reaches it.
				if sub.strip {
					continue
				}
				if dv != nil && tryAssign(f, dv) {
					// the default runs through the key's own rule (see validateObject); a
					// default that fails it is kept as written
					if err := sub.walkTyped(f, nil, kp, child); err != nil {
						assign(f, dv)
					}
					// A default only counts as PRESENT once it actually produced a value:
					// Joi never sets the key when the default resolves to undefined, so a
					// conditional default must not make an xor/nand/oxor peer look present.
					set[i] = true
				}
				continue
			}
			eff := sub.presence
			if eff == presRequired {
				return requiredError(sub, kp)
			}
			continue
		}
		if sub.presence == presForbidden {
			return errCtx(forbiddenError(sub, kp), "value", fieldGeneric(f, kid, i))
		}
		if err := sub.walkField(f, state, raw, kd, kp, child); err != nil {
			return err
		}
		if sub.strip {
			clear(f)
			kid.state[i] = stAbsent
			continue
		}
		set[i] = true
	}

	// unknown keys
	mode := r.unknownMode(st)
	unknown := append([]kv(nil), kid.unknown...)
	// a field the plan has but the schema does not declare is unknown to the schema
	for i, f := range plan.fields {
		if !handled[f.name] && !f.skip && kid.state[i] != stAbsent {
			unknown = append(unknown, kv{f.name, fieldGeneric(plan.field(v, i), kid, i)})
			clear(plan.field(v, i))
		}
	}
	sort.Slice(unknown, func(i, j int) bool { return unknown[i].key < unknown[j].key })
	kept := 0
	for _, u := range unknown {
		kp := append(append([]string(nil), path...), u.key)
		if pv := r.matchPattern(u.key, kp, child); pv != nil {
			nv, err := pv.validateValue(u.val, true, kp, child)
			if err != nil {
				return err
			}
			keepExtra(v, plan, u.key, nv)
			kept++
			continue
		}
		switch mode {
		case unkAllow:
			if keepExtra(v, plan, u.key, u.val) {
				kept++
			}
		case unkStrip:
		default:
			return errCtx(errCtx(errAt(kp, "object.unknown", quote(label(kp))+" is not allowed"), "child", u.key), "value", u.val)
		}
	}

	n := kept
	for _, s := range set {
		if s {
			n++
		}
	}
	if r.objMin != nil && n < *r.objMin {
		return errLimit(path, "object.min", quote(label(path))+" must have at least "+itoa(*r.objMin)+" "+keyWord(*r.objMin), *r.objMin)
	}
	if r.objMax != nil && n > *r.objMax {
		return errLimit(path, "object.max", quote(label(path))+" must have less than or equal to "+itoa(*r.objMax)+" "+keyWord(*r.objMax), *r.objMax)
	}

	has := func(k string) bool {
		i, ok := plan.byName[k]
		return ok && set[i]
	}
	if err := r.checkPresenceRelationsWith(has, path, v); err != nil {
		return err
	}
	return r.runContainerCustoms(v, has, path, st)
}

// isNilish reports a field that holds nothing at all — a nil pointer, map, slice or
// interface, or an unset wrapper — as opposed to a zero scalar, which is a value.
func isNilish(f reflect.Value) bool {
	if f.CanAddr() && f.Addr().Type().Implements(wrapperType) {
		w := f.Addr().Interface().(Wrapper)
		return !w.IsNull() && f.IsZero()
	}
	switch f.Kind() {
	case reflect.Pointer, reflect.Map, reflect.Slice, reflect.Interface:
		return f.IsNil()
	}
	return false
}

// unknownValue is the generic value recorded for an undeclared key, if the body sent it.
func (n *node) unknownValue(key string) (any, bool) {
	for _, u := range n.unknown {
		if u.key == key {
			return u.val, true
		}
	}
	return nil, false
}

// dropUnknown forgets an undeclared key the schema consumed.
func (n *node) dropUnknown(key string) {
	for i, u := range n.unknown {
		if u.key == key {
			n.unknown = append(n.unknown[:i], n.unknown[i+1:]...)
			return
		}
	}
}

// runContainerCustoms runs the typed customs of an object/array rule with a pointer
// to the container (CustomT / RefineT assert the concrete pointer type).
func (r *Rule) runContainerCustoms(v reflect.Value, present func(string) bool, path []string, st *evalState) *Error {
	if len(r.customs) == 0 {
		return nil
	}
	ptr := v.Addr().Interface()
	for _, fn := range r.customs {
		h := &Helper{Ctx: st.ctx, Original: ptr, Path: append([]string(nil), path...), present: present, st: st}
		if _, err := fn(ptr, h); err != nil {
			if ve, ok := err.(*Error); ok {
				if len(ve.Path) == 0 && !isRootAnchored(ve) {
					ve.Path = append([]string(nil), path...)
				}
				return wrapCustom(ve, treeOf(v))
			}
			return wrapCustom(errAt(path, "any.custom", err.Error()), treeOf(v))
		}
	}
	return nil
}

func (r *Rule) walkMap(v reflect.Value, kid *node, path []string, st *evalState) *Error {
	if kid == nil {
		kid = &node{mstate: map[string]uint8{}}
	}
	child := st.descend(&mapValueView{v: v})
	keyType := v.Type().Key()
	mkey := func(k string) reflect.Value { return reflect.ValueOf(k).Convert(keyType) }

	// Joi validates every DECLARED key first and only then looks at what is left over
	// (keys.js: `unprocessed` is drained by the key loop before internals.unknown runs),
	// so under abortEarly a missing required key is reported before an unknown one.
	// This used to be a single pass over the value's own keys, which reversed that.
	handled := map[string]bool{}
	for _, k := range r.keyOrder {
		handled[k] = true
		sub := r.keys[k].resolveEffective(child)
		kp := append(append([]string(nil), path...), k)
		kv := mkey(k)
		mv := v.MapIndex(kv)
		if !mv.IsValid() {
			// required beats default, as in validateObject / walkStruct
			if sub.presence == presRequired {
				return requiredError(sub, kp)
			}
			if sub.hasDefault {
				if dv := resolveDefaultVal(sub.defaultVal, child.parent); dv != nil {
					el := reflect.New(v.Type().Elem()).Elem()
					if tryAssign(el, dv) {
						v.SetMapIndex(kv, el)
					}
				}
			}
			continue
		}
		// A present key the schema FORBIDS is an error, as it is in the struct walk
		// and in eval.go. walkMap used to skip the check, so the typed write path —
		// the one every create and merge runs — accepted and STORED a forbidden key.
		if sub.presence == presForbidden {
			return errCtx(forbiddenError(sub, kp), "value", mapEntryGeneric(v, kv, kid, k))
		}
		el := reflect.New(v.Type().Elem()).Elem()
		el.Set(mv)
		state := kid.mstate[k]
		if state == stAbsent {
			state = stFits
		}
		if err := sub.walkField(el, state, kid.mraw[k], kid.mkids[k], kp, child); err != nil {
			return err
		}
		v.SetMapIndex(kv, el)
	}

	mode := r.unknownMode(st)
	keys := make([]string, 0, v.Len())
	for _, k := range v.MapKeys() {
		if !handled[k.String()] {
			keys = append(keys, k.String())
		}
	}
	sort.Strings(keys)
	for _, k := range keys {
		kp := append(append([]string(nil), path...), k)
		kv := mkey(k)
		if pv := r.matchPattern(k, kp, child); pv != nil {
			el := reflect.New(v.Type().Elem()).Elem()
			el.Set(v.MapIndex(kv))
			state := kid.mstate[k]
			if state == stAbsent {
				state = stFits
			}
			if err := pv.walkField(el, state, kid.mraw[k], kid.mkids[k], kp, child); err != nil {
				return err
			}
			v.SetMapIndex(kv, el)
			continue
		}
		// A key no pattern matches is merely UNKNOWN: the object's unknown policy
		// decides, exactly as it does in validateObject. Rejecting it here made
		// .Unknown(true) / stripUnknown 400 on the write path alone.
		switch mode {
		case unkAllow:
		case unkStrip:
			v.SetMapIndex(kv, reflect.Value{})
		default:
			return errCtx(errCtx(errAt(kp, "object.unknown", quote(label(kp))+" is not allowed"), "child", k), "value", mapEntryGeneric(v, kv, kid, k))
		}
	}

	n := v.Len()
	if r.objMin != nil && n < *r.objMin {
		return errLimit(path, "object.min", quote(label(path))+" must have at least "+itoa(*r.objMin)+" "+keyWord(*r.objMin), *r.objMin)
	}
	if r.objMax != nil && n > *r.objMax {
		return errLimit(path, "object.max", quote(label(path))+" must have less than or equal to "+itoa(*r.objMax)+" "+keyWord(*r.objMax), *r.objMax)
	}
	has := func(k string) bool { return v.MapIndex(mkey(k)).IsValid() }
	if err := r.checkPresenceRelationsWith(has, path, v); err != nil {
		return err
	}
	if r.untypedCustom {
		panic(fmt.Sprintf("validate: %s at %q carries a custom written for a map; use CustomT", v.Type(), label(path)))
	}
	return r.runContainerCustoms(v, has, path, st)
}

func (r *Rule) walkSlice(v reflect.Value, kid *node, path []string, st *evalState) *Error {
	if kid == nil {
		kid = &node{}
	}
	if r.untypedCustom {
		panic(fmt.Sprintf("validate: %s at %q carries a custom written for []any; use CustomT", v.Type(), label(path)))
	}
	for i := 0; i < v.Len(); i++ {
		ep := append(append([]string(nil), path...), idxSeg(i))
		el := v.Index(i)
		state := stFits
		if i < len(kid.elemState) {
			state = kid.elemState[i]
		}
		if err := r.walkItem(el, state, kid.elemRaw[i], kid.elems[i], ep, st); err != nil {
			return err
		}
	}
	n := v.Len()
	// Joi's `.items(X.required())`: each REQUIRED item schema must be satisfied by at
	// least one element, or the array errors as a whole. Kept identical to the map
	// evaluator's check — the two paths must blame the same thing.
	if misses := r.unmatchedRequiredItemsTyped(v, kid, st); misses > 0 {
		e := errAt(path, "array.includesRequiredUnknowns",
			quote(label(path))+" does not contain "+itoa(misses)+" required value(s)")
		return errCtx(e, "unknownMisses", misses)
	}
	if r.arrLen != nil && n != *r.arrLen {
		return errLimit(path, "array.length", quote(label(path))+" must contain "+itoa(*r.arrLen)+" items", *r.arrLen)
	}
	if r.arrMin != nil && n < *r.arrMin {
		return errLimit(path, "array.min", quote(label(path))+" must contain at least "+itoa(*r.arrMin)+" items", *r.arrMin)
	}
	if r.arrMax != nil && n > *r.arrMax {
		return errLimit(path, "array.max", quote(label(path))+" must contain less than or equal to "+itoa(*r.arrMax)+" items", *r.arrMax)
	}
	if r.arrUnique {
		seen := map[string]int{}
		for i := 0; i < n; i++ {
			k := elemKeyOf(v.Index(i), r.arrUniqueBy)
			if j, ok := seen[k]; ok {
				return uniqueError(path, i, j, treeOf(v.Index(i)), treeOf(v.Index(j)), r.uniqueCtxPath())
			}
			seen[k] = i
		}
	}
	if r.arrSort {
		keys := make([]any, n)
		for i := 0; i < n; i++ {
			keys[i] = elemSortKey(v.Index(i), r.arrSortBy)
		}
		if !sortableTogether(keys) {
			return errAt(path, "array.sort.mismatching", quote(label(path))+" cannot be sorted due to mismatching types")
		}
		sortSlice(v, reflect.Swapper(v.Interface()), r.arrSortBy)
	}
	if r.arrDedupe {
		seen := map[string]bool{}
		out := reflect.MakeSlice(v.Type(), 0, n)
		for i := 0; i < n; i++ {
			if r.arrDedupeBy == "" && isContainerValue(v.Index(i)) {
				out = reflect.Append(out, v.Index(i)) // _.uniq compares objects by reference: never equal
				continue
			}
			k := elemKeyOf(v.Index(i), r.arrDedupeBy)
			if seen[k] {
				continue
			}
			seen[k] = true
			out = reflect.Append(out, v.Index(i))
		}
		v.Set(out)
	}
	return r.runContainerCustoms(v, nil, path, st)
}

// walkItem validates one element against the item rules; Joi accepts an element if
// ANY item schema matches, which a typed element can only do when there is a single
// item rule (the schemas here) or the element is a scalar.
func (r *Rule) walkItem(el reflect.Value, state uint8, raw any, kid *node, path []string, st *evalState) *Error {
	switch len(r.arrItems) {
	case 0:
		return nil
	case 1:
		return r.arrItems[0].walkField(el, state, raw, kid, path, st)
	}
	inner, kind, ok := peel(el)
	if ok && (kind == dkStruct || kind == dkSlice || kind == dkMap) {
		// A CONTAINER element under several item rules used to panic — a 500 on a shape
		// Joi accepts. Try each rule against a copy and commit the first that passes;
		// the copy keeps a rejected candidate's partial coercions out of the element.
		var firstErr *Error
		for _, it := range r.arrItems {
			try := reflect.New(el.Type()).Elem()
			try.Set(el)
			err := it.walkField(try, state, raw, kid, path, st)
			if err == nil {
				el.Set(try)
				return nil
			}
			if firstErr == nil {
				firstErr = err
			}
		}
		return includesError(path, treeOf(el), firstErr)
	}
	var g any
	switch {
	case state == stRaw:
		g = raw
	case !ok || state == stNull:
		g = nil
	default:
		g = scalarOf(inner, kind)
	}
	var firstErr *Error
	for _, it := range r.arrItems {
		nv, err := it.validateValue(g, true, path, st)
		if err == nil {
			if !tryAssign(el, nv) {
				return it.stampGeneric(typeBaseError(path, destRuleType(el)), path, nv)
			}
			return nil
		}
		if firstErr == nil {
			firstErr = err
		}
	}
	return includesError(path, g, firstErr)
}

// includesError is the error Joi raises for an element that matched NONE of several
// item schemas: with a single inclusion the inner error passes through, with more than
// one Joi replaces it with array.includes (array.js: `Errors.report('array.includes')`
// once `il > 1`), because "must be a string" is misleading when a number would have
// done as well.
func includesError(path []string, value any, inner *Error) *Error {
	if inner == nil {
		return nil
	}
	e := errAt(path, "array.includes", quote(label(path))+" does not match any of the allowed types")
	return errCtx(errCtx(e, "pos", lastIndexOf(path)), "value", value)
}

// lastIndexOf is the array position a path ends at, for array.includes' `pos` context.
func lastIndexOf(path []string) int {
	if len(path) == 0 {
		return 0
	}
	n, _ := indexOf(path[len(path)-1])
	return n
}

// sortSlice is Joi's .sort() / .sort({by}) over a typed slice.
func sortSlice(v reflect.Value, swap func(i, j int), by string) {
	n := v.Len()
	keys := make([]any, n)
	for i := 0; i < n; i++ {
		keys[i] = elemSortKey(v.Index(i), by)
	}
	// insertion sort keeps the swapper and the key slice in step (stable, small n)
	for i := 1; i < n; i++ {
		for j := i; j > 0 && compareValues(keys[j-1], keys[j]) > 0; j-- {
			swap(j-1, j)
			keys[j-1], keys[j] = keys[j], keys[j-1]
		}
	}
}

// elemSortKey reads the sort key off a typed element (the element itself, or one of
// its fields), as elemKey does off a generic one.
func elemSortKey(el reflect.Value, by string) any {
	inner, kind, ok := peel(el)
	if !ok {
		return nil
	}
	if by == "" {
		return scalarOf(inner, kind)
	}
	if kind == dkStruct {
		p := planFor(inner.Type())
		if i, ok := p.byName[by]; ok {
			fi, fk, fok := peel(p.field(inner, i))
			if fok {
				return scalarOf(fi, fk)
			}
		}
		return nil
	}
	if kind == dkMap {
		mv := inner.MapIndex(reflect.ValueOf(by).Convert(inner.Type().Key()))
		if mv.IsValid() {
			if fi, fk, fok := peel(mv); fok {
				return scalarOf(fi, fk)
			}
		}
	}
	return nil
}

// elemKeyOf is toKey for a typed element: scalars key exactly as their generic
// value does, and containers key on their content — the generic evaluator keyed
// every object identically, which made any two objects "duplicates".
func elemKeyOf(el reflect.Value, by string) string {
	if by != "" {
		return toKey(elemSortKey(el, by))
	}
	inner, kind, ok := peel(el)
	if !ok {
		return toKey(nil)
	}
	switch kind {
	case dkStruct, dkSlice, dkMap, dkAny:
		b, _ := json.Marshal(inner.Interface())
		return "j:" + string(b)
	}
	return toKey(scalarOf(inner, kind))
}

func isContainerValue(el reflect.Value) bool {
	inner, kind, ok := peel(el)
	if !ok {
		return false
	}
	switch kind {
	case dkStruct, dkSlice, dkMap:
		return true
	case dkAny:
		return isContainer(inner.Interface())
	}
	return false
}

// checkPresenceRelationsWith is checkPresenceRelations over a presence predicate, with
// `obj` the object the dependency is declared on — Joi puts it in `context.value`, and
// a presence map there reported `{"otel":true}` where TS reports the real block.
func (r *Rule) checkPresenceRelationsWith(has func(string) bool, path []string, obj any) *Error {
	if len(r.or) == 0 && len(r.xor) == 0 && len(r.oxor) == 0 && len(r.nand) == 0 {
		return nil
	}
	out := map[string]any{}
	for _, peers := range r.or {
		for _, p := range peers {
			if has(p) {
				out[p] = true
			}
		}
	}
	for _, group := range [][][]string{r.xor, r.oxor, r.nand} {
		for _, peers := range group {
			for _, p := range peers {
				if has(p) {
					out[p] = true
				}
			}
		}
	}
	err := r.checkPresenceRelations(out, path)
	if err != nil && err.Ctx != nil {
		if rv, ok := obj.(reflect.Value); ok && rv.IsValid() {
			err.Ctx["value"] = treeOf(rv)
		}
	}
	return err
}

// --- reading typed values as the generic evaluator's shapes ---

// scalarOf reads a leaf field as the generic value the evaluator judges.
func scalarOf(inner reflect.Value, kind destKind) any {
	switch kind {
	case dkString:
		return inner.String()
	case dkNumber:
		switch inner.Kind() {
		case reflect.Float32, reflect.Float64:
			return inner.Float()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return float64(inner.Int())
		default:
			return float64(inner.Uint())
		}
	case dkBool:
		return inner.Bool()
	case dkOpaque:
		if inner.Type().ConvertibleTo(timeType) {
			return inner.Convert(timeType).Interface()
		}
		if tm, ok := inner.Interface().(encoding.TextMarshaler); ok {
			b, _ := tm.MarshalText()
			return string(b)
		}
		b, _ := json.Marshal(inner.Interface())
		var g any
		_ = wirejson.Decode(b, &g)
		return g
	case dkAny:
		return inner.Interface()
	case dkBytes:
		return string(inner.Bytes())
	}
	return treeOf(inner)
}

// treeOf renders a typed value as the generic tree — used for error contexts and the
// rare untyped custom, never on the success path.
func treeOf(v reflect.Value) any {
	b, err := json.Marshal(v.Interface())
	if err != nil {
		return nil
	}
	var g any
	_ = wirejson.Decode(b, &g)
	return g
}

// fieldGeneric is the generic value of a field for an error context.
func fieldGeneric(f reflect.Value, kid *node, i int) any {
	if kid != nil {
		if raw, ok := kid.raw[i]; ok {
			return raw
		}
		if kid.state[i] == stNull {
			return nil
		}
	}
	inner, kind, ok := peel(f)
	if !ok {
		return nil
	}
	return scalarOf(inner, kind)
}

func mapEntryGeneric(m, k reflect.Value, kid *node, key string) any {
	if kid != nil {
		if raw, ok := kid.mraw[key]; ok {
			return raw
		}
		if kid.mstate[key] == stNull {
			return nil
		}
	}
	inner, kind, ok := peel(m.MapIndex(k))
	if !ok {
		return nil
	}
	return scalarOf(inner, kind)
}

// keepExtra stores an allowed unknown key in the struct's ExtrasMarker map, if it has
// one; without one the key is dropped, as it has nowhere to live.
func keepExtra(v reflect.Value, p *destPlan, key string, val any) bool {
	if p.extras == nil {
		return false
	}
	f := fieldByIndexAlloc(v, p.extras)
	if f.IsNil() {
		f.Set(reflect.MakeMap(f.Type()))
	}
	f.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(&val).Elem())
	return true
}

// destRuleType is the Joi type a field's Go type corresponds to, for an error message.
func destRuleType(v reflect.Value) ruleType {
	inner, kind, _ := target(v)
	_ = inner
	switch kind {
	case dkStruct, dkMap:
		return tObject
	case dkSlice:
		return tArray
	case dkString, dkOpaque:
		return tString
	case dkNumber:
		return tNumber
	case dkBool:
		return tBoolean
	}
	return tAny
}

// tryAssign is assign that reports a value the field cannot hold instead of panicking.
func tryAssign(v reflect.Value, g any) (ok bool) {
	defer func() {
		if recover() != nil {
			ok = false
		}
	}()
	assign(v, g)
	return true
}

// assign writes a generic value (the evaluator's normalized output, or a default)
// into a typed field. A value the field cannot hold is a schema/type mismatch, which
// is a programming error rather than a request error.
func assign(v reflect.Value, g any) {
	if g == nil {
		// nil means absent: a caller clearing a field, or a custom returning nothing.
		// An explicit null is the decoder's record alone, never written back here.
		clear(v)
		return
	}
	if t, ok := g.(time.Time); ok {
		inner, kind, _ := target(v)
		if kind == dkOpaque {
			if inner.Type().ConvertibleTo(timeType) {
				inner.Set(reflect.ValueOf(t).Convert(inner.Type()))
				return
			}
			g = t.UTC().Format(time.RFC3339Nano)
		} else if kind == dkString {
			g = t.UTC().Format(time.RFC3339Nano)
		}
	}
	// already the field's own type (a typed custom returned it)
	if gv := reflect.ValueOf(g); gv.Type().AssignableTo(v.Type()) {
		v.Set(gv)
		return
	} else if inner, _, _ := target(v); gv.Type().AssignableTo(inner.Type()) {
		inner.Set(gv)
		return
	} else if gv.Kind() == reflect.Ptr && gv.Elem().Type().AssignableTo(inner.Type()) {
		inner.Set(gv.Elem())
		return
	} else if gv.Type().ConvertibleTo(inner.Type()) && gv.Kind() == inner.Kind() {
		inner.Set(gv.Convert(inner.Type()))
		return
	}
	d := &dec{c: Tree{g}.cursor()}
	state, raw, _, err := d.value(v)
	if err != nil || state == stRaw {
		panic(fmt.Sprintf("validate: cannot store %T (%v) into %s: %v", raw, raw, v.Type(), err))
	}
}

// --- sibling views for refs and whens ---

// parentView is the object a rule's siblings are read from: the map being validated
// in generic mode, the struct (with its decode record) in typed mode.
type parentView interface {
	sibling(key string) (any, bool)
	asMap() map[string]any
}

type mapView map[string]any

func (m mapView) sibling(key string) (any, bool) { v, ok := m[key]; return v, ok }
func (m mapView) asMap() map[string]any          { return m }

type structView struct {
	plan *destPlan
	v    reflect.Value
	n    *node
}

func (s *structView) sibling(key string) (any, bool) {
	i, ok := s.plan.byName[key]
	if !ok {
		for _, u := range s.n.unknown {
			if u.key == key {
				return u.val, true
			}
		}
		return nil, false
	}
	f := s.plan.field(s.v, i)
	present := s.n.state[i] != stAbsent || !isZero(f)
	if !present {
		return nil, false
	}
	return fieldGeneric(f, s.n, i), true
}

func (s *structView) asMap() map[string]any {
	m := map[string]any{}
	for _, f := range s.plan.fields {
		if v, ok := s.sibling(f.name); ok {
			m[f.name] = v
		}
	}
	return m
}

type mapValueView struct{ v reflect.Value }

func (m *mapValueView) sibling(key string) (any, bool) {
	mv := m.v.MapIndex(reflect.ValueOf(key).Convert(m.v.Type().Key()))
	if !mv.IsValid() {
		return nil, false
	}
	inner, kind, ok := peel(mv)
	if !ok {
		return nil, true
	}
	return scalarOf(inner, kind), true
}

func (m *mapValueView) asMap() map[string]any {
	out := map[string]any{}
	for _, k := range m.v.MapKeys() {
		out[k.String()], _ = m.sibling(k.String())
	}
	return out
}

// --- typed hooks ---

// CustomT is a custom that receives a pointer to the typed container it is declared
// on (an object or array rule) and may mutate it in place. It is the typed form of
// Custom for containers; the generic evaluator refuses it, and the typed walk refuses
// a Custom written for a map or []any, so neither can silently skip the other's hook.
func CustomT[T any](fn func(*T, *Helper) error) CustomFn {
	return func(v any, h *Helper) (any, error) {
		p, ok := v.(*T)
		if !ok {
			panic(fmt.Sprintf("validate: typed custom for *%s ran on %T at %q", reflect.TypeOf((*T)(nil)).Elem(), v, h.Label()))
		}
		if err := fn(p, h); err != nil {
			return nil, err
		}
		return p, nil
	}
}

// RefineT attaches a typed object-level cross-field check to an object rule.
func RefineT[T any](r *Rule, fn func(*T, *Helper) error) *Rule {
	c := r.clone()
	c.customs = append(c.customs, CustomT(fn))
	return c
}

// Str reads a string field of the object a container custom runs on — a map on the
// generic evaluator, a struct pointer on the typed walk — by json name.
func Str(obj any, jsonName string) string {
	if m, ok := obj.(map[string]any); ok {
		s, _ := m[jsonName].(string)
		return s
	}
	f, ok := Field(obj, jsonName)
	if !ok {
		return ""
	}
	inner, kind, ok := peel(f)
	if !ok || kind != dkString {
		return ""
	}
	return inner.String()
}

// Get reads a field of the object a container custom runs on, in the form the
// custom can work with: a scalar as itself, a struct as a pointer to it, a map with
// interface values as map[string]any (sharing the storage), nil when absent or null.
func Get(obj any, jsonName string) any {
	if m, ok := obj.(map[string]any); ok {
		return m[jsonName]
	}
	f, ok := Field(obj, jsonName)
	if !ok {
		return nil
	}
	inner, kind, ok := peel(f)
	if !ok {
		return nil
	}
	switch kind {
	case dkStruct:
		return inner.Addr().Interface()
	case dkMap:
		if inner.Type().Elem().Kind() == reflect.Interface {
			return inner.Convert(reflect.TypeOf(map[string]any{})).Interface()
		}
		return inner.Interface()
	case dkSlice:
		return inner.Addr().Interface()
	}
	return scalarOf(inner, kind)
}

// Sub reads a struct or slice field as a typed pointer, nil when absent, null, or
// not of that type.
func Sub[T any](obj any, jsonName string) *T {
	if m, ok := obj.(map[string]any); ok {
		p, _ := m[jsonName].(*T)
		return p
	}
	p, _ := Get(obj, jsonName).(*T)
	return p
}

// Set writes a value into a field (or map key) of the object a container custom
// runs on. A struct pointer, a generic tree or a scalar are all accepted.
func Set(obj any, jsonName string, val any) {
	if m, ok := obj.(map[string]any); ok {
		m[jsonName] = val
		return
	}
	f, ok := Field(obj, jsonName)
	if !ok {
		panic(fmt.Sprintf("validate: Set: %T has no field %q", obj, jsonName))
	}
	assign(f, val)
}

// Field returns the addressable field of a struct pointer by its json name, for a
// container custom that is written once for several types (reflectively).
func Field(ptr any, jsonName string) (reflect.Value, bool) {
	rv := reflect.ValueOf(ptr)
	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Struct {
		return reflect.Value{}, false
	}
	return FieldOf(rv.Elem(), jsonName)
}

// FieldOf is Field on an addressable struct value: the settable field for a json
// name, through embedded structs and inline columns, allocating what it must.
func FieldOf(v reflect.Value, jsonName string) (reflect.Value, bool) {
	if v.Kind() != reflect.Struct || !v.CanAddr() {
		return reflect.Value{}, false
	}
	p := planFor(v.Type())
	i, ok := p.byName[jsonName]
	if !ok {
		return reflect.Value{}, false
	}
	return p.field(v, i), true
}

// Peel follows pointers and wrappers to the value they hold, without allocating; ok is
// false at a nil pointer or a null wrapper.
func Peel(v reflect.Value) (reflect.Value, bool) {
	inner, _, ok := peel(v)
	return inner, ok
}

// PeelForWrite is Peel that ALLOCATES: a nil pointer is filled in and a null wrapper
// opened, so the caller gets a settable inner value. It is how a caller materializes a
// subtree the body omitted — the Go equivalent of TS assigning `req.body.status = {}`
// before Joi runs, which is what makes the schema's nested defaults land there.
func PeelForWrite(v reflect.Value) (reflect.Value, bool) {
	if !v.CanAddr() && v.Kind() != reflect.Ptr {
		return reflect.Value{}, false
	}
	inner, _, _ := target(v)
	return inner, inner.IsValid()
}

// Assign writes a value into a field the way the walk writes a normalized result: a
// value of the field's type, a pointer to one, a convertible scalar, a time, or a
// generic tree.
func Assign(field reflect.Value, val any) { assign(field, val) }

// sortedKeys returns a map's keys in order.
func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// UntypedContainerCustoms lists the paths in a rule graph whose container customs were
// written for the generic shapes (Custom / Refine on an object or array rule). Such a
// rule can only run on the generic evaluator; a resource schema, which the write path
// runs on the row, must have none. A gate test asks this of every registered kind.
func (s *Schema) UntypedContainerCustoms() []string {
	var out []string
	seen := map[*Rule]bool{}
	var walk func(r *Rule, path string)
	walk = func(r *Rule, path string) {
		if r == nil || seen[r] {
			return
		}
		seen[r] = true
		if r.untypedCustom {
			out = append(out, path)
		}
		for _, k := range r.keyOrder {
			walk(r.keys[k], path+"."+k)
		}
		for _, p := range r.patterns {
			walk(p.val, path+".*")
		}
		for _, it := range r.arrItems {
			walk(it, path+"[]")
		}
		for _, a := range r.alts {
			walk(a, path+"|")
		}
		for _, w := range r.whens {
			walk(w.then, path+"?then")
			walk(w.otherwise, path+"?otherwise")
		}
		if r.altCond != nil {
			walk(r.altCond.then, path+"?then")
			walk(r.altCond.otherwise, path+"?otherwise")
		}
	}
	walk(s.Rule(), "")
	return out
}

// requiredError / forbiddenError are the object loop's own presence errors, honouring
// the key rule's .label() override the same way the leaf errors do.
func requiredError(sub *Rule, kp []string) *Error {
	return sub.presenceMessage(sub.relabel(errAt(kp, "any.required", quote(sub.labelOf(kp))+" is required"), kp))
}

func forbiddenError(sub *Rule, kp []string) *Error {
	return sub.presenceMessage(sub.relabel(errAt(kp, "any.unknown", quote(sub.labelOf(kp))+" is not allowed"), kp))
}

// presenceMessage applies the key rule's `.messages()` to a presence error its parent
// raised for it, as Joi does.
func (r *Rule) presenceMessage(e *Error) *Error {
	if e != nil {
		if m, ok := r.msgOverrides[e.Type]; ok {
			e.Message = m
		}
	}
	return e
}

// stamp completes an error the TYPED container walks raise at their own node: Joi puts
// the failing value in context.value and honours the rule's .label(), and the generic
// evaluator gets both from the validateValue seam — which walkStruct/walkMap/walkSlice
// bypass. So array.min, object.min, array.sort.mismatching and friends came back without
// a value on the write path and with one on the tree path.
func (r *Rule) stamp(e *Error, path []string, v reflect.Value) *Error {
	if e == nil || e.Ctx == nil || len(e.Path) != len(path) {
		return e
	}
	if _, has := e.Ctx["value"]; !has {
		e.Ctx["value"] = treeOf(v)
	}
	return r.relabel(e, path)
}

// stampGeneric is stamp for the typed walk's typeBaseError seams, where the failing
// value is already generic (the raw body fragment the field could not hold) — they
// answered without Joi's context.value, which every other error path carries.
func (r *Rule) stampGeneric(e *Error, path []string, v any) *Error {
	if e == nil || e.Ctx == nil || len(e.Path) != len(path) {
		return e
	}
	if _, has := e.Ctx["value"]; !has && !noValueContext[e.Type] {
		e.Ctx["value"] = v
	}
	return r.relabel(e, path)
}

// unmatchedRequiredItemsTyped counts the REQUIRED item schemas that no element of the
// typed slice satisfies — Joi's `unknownMisses`. Each trial runs against a copy so a
// rejected candidate leaves no coercion behind on the element.
func (r *Rule) unmatchedRequiredItemsTyped(v reflect.Value, kid *node, st *evalState) int {
	misses := 0
	for _, item := range r.arrItems {
		if item == nil || item.presence != presRequired {
			continue
		}
		matched := false
		for i := 0; i < v.Len() && !matched; i++ {
			state := stFits
			if i < len(kid.elemState) {
				state = kid.elemState[i]
			}
			try := reflect.New(v.Type().Elem()).Elem()
			try.Set(v.Index(i))
			if err := item.walkField(try, state, kid.elemRaw[i], kid.elems[i], nil, st); err == nil {
				matched = true
			}
		}
		if !matched {
			misses++
		}
	}
	return misses
}
