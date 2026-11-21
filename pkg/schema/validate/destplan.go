package validate

// destplan.go describes the Go value a body decodes INTO. The decoder (decode.go)
// fills a struct straight from the JSON, and the typed walk (walk.go) validates the
// struct in place; both need the same view of a destination type: which exported,
// json-named fields it has (embedded structs flattened the way encoding/json does),
// and what kind of value each accepts. Plans are built once per type and cached.

import (
	"encoding"
	"encoding/json"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"
)

// Wrapper is a container field type holding ONE inner value — the shape of a jsonb
// column wrapper. The decoder writes through it and the walk reads through it, so a
// row's `spec` column decodes straight into the typed spec the column carries.
type Wrapper interface {
	// InnerForWrite returns a pointer to the inner value, allocating it if needed,
	// and marks the wrapper as holding a value (not null).
	InnerForWrite() any
	// SetNull records an explicit JSON null.
	SetNull()
	// IsNull reports an explicit null.
	IsNull() bool
}

// ExtrasMarker is implemented by the map type an `unknown(true)` object keeps its
// undeclared keys in (base.Extras). A struct with such a field receives allowed
// unknown keys there; a struct without one has them stripped.
type ExtrasMarker interface{ ExtrasMarker() }

// destKind classifies what a destination Go type accepts from the wire.
type destKind uint8

const (
	dkInvalid destKind = iota
	dkString           // string kinds (nominal included)
	dkNumber           // ints, uints, floats
	dkBool
	dkStruct
	dkSlice
	dkMap    // map[string]X
	dkAny    // interface{}: takes the generic tree
	dkOpaque // a scalar type that unmarshals itself (time.Time, WireTime, …)
	dkBytes  // []byte
)

type destPlan struct {
	t      reflect.Type
	fields []destField
	byName map[string]int
	extras []int // index path of the ExtrasMarker field, nil if none
}

type destField struct {
	name  string
	index []int
	typ   reflect.Type // the field's declared type (pointer/wrapper NOT stripped)
	// skip marks a field tagged `validate:"-"`: decoded, never validated and never
	// reported as unknown — a row column that is not part of the wire schema.
	skip bool
	// via is the index of a Wrapper field the plan flattens (`wire:"inline"`); index
	// is then relative to the wrapper's inner struct.
	via []int
}

var (
	destPlans   sync.Map // reflect.Type → *destPlan
	timeType    = reflect.TypeOf(time.Time{})
	extrasType  = reflect.TypeOf((*ExtrasMarker)(nil)).Elem()
	wrapperType = reflect.TypeOf((*Wrapper)(nil)).Elem()
	textUnmType = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
	jsonUnmType = reflect.TypeOf((*json.Unmarshaler)(nil)).Elem()
)

// planFor returns the cached destination plan for a struct type.
func planFor(t reflect.Type) *destPlan {
	if p, ok := destPlans.Load(t); ok {
		return p.(*destPlan)
	}
	p := &destPlan{t: t, byName: map[string]int{}}
	p.collect(t, nil)
	actual, _ := destPlans.LoadOrStore(t, p)
	return actual.(*destPlan)
}

// collect walks a struct's fields, flattening untagged embedded structs the way
// encoding/json does (a tagged embedded struct is a named object; a `json:"-"` field
// is skipped). The first field declared for a name wins, which is also json's rule
// for the depth-0 case that matters here.
func (p *destPlan) collect(t reflect.Type, prefix []int) {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		index := append(append([]int(nil), prefix...), i)
		if f.Type.Implements(extrasType) {
			p.extras = index
			continue
		}
		if f.Anonymous {
			ft := f.Type
			if ft.Kind() == reflect.Ptr {
				ft = ft.Elem()
			}
			if _, tagged := f.Tag.Lookup("json"); !tagged && ft.Kind() == reflect.Struct {
				p.collect(ft, index)
				continue
			}
		}
		if f.PkgPath != "" {
			continue
		}
		if f.Tag.Get("wire") == "inline" {
			// a column wrapper whose inner struct's fields sit at this level on the wire
			inner := reflect.New(innerType(f.Type)).Elem()
			sub := planFor(inner.Type())
			for _, sf := range sub.fields {
				if _, dup := p.byName[sf.name]; dup {
					continue
				}
				p.byName[sf.name] = len(p.fields)
				p.fields = append(p.fields, destField{name: sf.name, index: sf.index, typ: sf.typ, skip: sf.skip, via: index})
			}
			continue
		}
		name := jsonName(f)
		if name == "-" {
			continue
		}
		if _, dup := p.byName[name]; dup {
			continue
		}
		p.byName[name] = len(p.fields)
		p.fields = append(p.fields, destField{name: name, index: index, typ: f.Type, skip: f.Tag.Get("validate") == "-"})
	}
}

// innerType is the struct type a Wrapper field holds.
func innerType(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	w, ok := reflect.New(t).Interface().(Wrapper)
	if !ok {
		panic("validate: wire:\"inline\" needs a Wrapper field, got " + t.String())
	}
	return reflect.TypeOf(w.InnerForWrite()).Elem()
}

// field returns the addressable field value inside v (a struct value), allocating
// nil pointers along an embedded-pointer path and opening an inline wrapper.
func (p *destPlan) field(v reflect.Value, i int) reflect.Value {
	f := p.fields[i]
	if f.via != nil {
		w := fieldByIndexAlloc(v, f.via)
		inner, _, _ := target(w)
		return fieldByIndexAlloc(inner, f.index)
	}
	return fieldByIndexAlloc(v, f.index)
}

func fieldByIndexAlloc(v reflect.Value, index []int) reflect.Value {
	for n, i := range index {
		if n > 0 && v.Kind() == reflect.Ptr {
			if v.IsNil() {
				v.Set(reflect.New(v.Type().Elem()))
			}
			v = v.Elem()
		}
		v = v.Field(i)
	}
	return v
}

// kindOf classifies a destination type. Pointers and wrappers are stripped first
// (see settable), so this sees the type a value is actually written into.
func kindOf(t reflect.Type) destKind {
	if t == timeType || (t.Kind() != reflect.Struct && t.Kind() != reflect.Map && t.Kind() != reflect.Slice &&
		(reflect.PtrTo(t).Implements(jsonUnmType) || reflect.PtrTo(t).Implements(textUnmType))) {
		return dkOpaque
	}
	switch t.Kind() {
	case reflect.String:
		return dkString
	case reflect.Bool:
		return dkBool
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return dkNumber
	case reflect.Struct:
		if reflect.PtrTo(t).Implements(jsonUnmType) && !hasExportedFields(t) {
			return dkOpaque
		}
		return dkStruct
	case reflect.Slice:
		if t.Elem().Kind() == reflect.Uint8 {
			return dkBytes
		}
		return dkSlice
	case reflect.Map:
		if t.Key().Kind() != reflect.String {
			return dkInvalid
		}
		return dkMap
	case reflect.Interface:
		return dkAny
	}
	return dkInvalid
}

func hasExportedFields(t reflect.Type) bool {
	for i := 0; i < t.NumField(); i++ {
		if t.Field(i).PkgPath == "" {
			return true
		}
	}
	return false
}

// target resolves a field value to the value that is actually written: pointers are
// allocated and dereferenced, wrappers are opened. Returns the inner value and its
// kind. The wrapper, if any, is returned so a null can be recorded on it.
func target(v reflect.Value) (inner reflect.Value, kind destKind, w Wrapper) {
	for {
		if v.CanAddr() && v.Addr().Type().Implements(wrapperType) {
			w = v.Addr().Interface().(Wrapper)
			v = reflect.ValueOf(w.InnerForWrite()).Elem()
			continue
		}
		if v.Kind() == reflect.Ptr {
			if v.Type().Implements(wrapperType) {
				if v.IsNil() {
					v.Set(reflect.New(v.Type().Elem()))
				}
				w = v.Interface().(Wrapper)
				v = reflect.ValueOf(w.InnerForWrite()).Elem()
				continue
			}
			if v.IsNil() {
				v.Set(reflect.New(v.Type().Elem()))
			}
			v = v.Elem()
			continue
		}
		return v, kindOf(v.Type()), w
	}
}

// peel is target's read-only twin: it follows pointers and wrappers WITHOUT
// allocating. ok is false when a nil pointer / null wrapper is hit.
func peel(v reflect.Value) (reflect.Value, destKind, bool) {
	for {
		if v.CanAddr() && v.Addr().Type().Implements(wrapperType) {
			w := v.Addr().Interface().(Wrapper)
			if w.IsNull() {
				return v, dkInvalid, false
			}
			v = reflect.ValueOf(w.InnerForWrite()).Elem()
			continue
		}
		if v.Kind() == reflect.Ptr {
			if v.IsNil() {
				return v, dkInvalid, false
			}
			if v.Type().Implements(wrapperType) {
				w := v.Interface().(Wrapper)
				if w.IsNull() {
					return v, dkInvalid, false
				}
				v = reflect.ValueOf(w.InnerForWrite()).Elem()
				continue
			}
			v = v.Elem()
			continue
		}
		return v, kindOf(v.Type()), true
	}
}

// clear resets a field to absent: nil for pointers/maps/slices/interfaces, the zero
// value otherwise, and a null-free empty wrapper for wrappers.
func clear(v reflect.Value) {
	v.Set(reflect.Zero(v.Type()))
}

// setNull records an explicit null on a field: wrappers remember it, everything
// else is cleared.
func setNull(v reflect.Value) {
	if v.CanAddr() && v.Addr().Type().Implements(wrapperType) {
		v.Addr().Interface().(Wrapper).SetNull()
		return
	}
	if v.Kind() == reflect.Ptr && v.Type().Implements(wrapperType) {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v.Interface().(Wrapper).SetNull()
		return
	}
	clear(v)
}

// isZero reports whether a field holds nothing: the zero value, a nil pointer, or
// an empty (unset, non-null) wrapper.
func isZero(v reflect.Value) bool {
	if v.CanAddr() && v.Addr().Type().Implements(wrapperType) {
		w := v.Addr().Interface().(Wrapper)
		return !w.IsNull() && v.IsZero()
	}
	return v.IsZero()
}

// trimName strips a json tag's options.
func trimName(tag string) string { return strings.Split(tag, ",")[0] }

// DeepCopy copies a value so the result shares nothing mutable with the source. A
// Wrapper is rebuilt around a copy of what it holds. The decoder uses it when a tree
// source hands it an in-process value of the field's own type: the row being built
// must not write through to wherever that value came from.
func DeepCopy(v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Pointer && v.Type().Implements(wrapperType) {
		if v.IsNil() {
			return v
		}
		return cloneWrapper(v.Interface().(Wrapper), v.Type().Elem())
	}
	if v.Kind() == reflect.Struct && v.CanAddr() && v.Addr().Type().Implements(wrapperType) {
		return cloneWrapper(v.Addr().Interface().(Wrapper), v.Type()).Elem()
	}
	switch v.Kind() {
	case reflect.Pointer:
		if v.IsNil() {
			return v
		}
		out := reflect.New(v.Type().Elem())
		out.Elem().Set(DeepCopy(v.Elem()))
		return out
	case reflect.Interface:
		if v.IsNil() {
			return v
		}
		out := reflect.New(v.Type()).Elem()
		out.Set(DeepCopy(v.Elem()))
		return out
	case reflect.Struct:
		out := reflect.New(v.Type()).Elem()
		out.Set(v)
		for i := 0; i < v.NumField(); i++ {
			if f := out.Field(i); f.CanSet() {
				f.Set(DeepCopy(v.Field(i)))
			}
		}
		return out
	case reflect.Slice:
		if v.IsNil() || v.Type().Elem().Kind() == reflect.Uint8 {
			return v
		}
		out := reflect.MakeSlice(v.Type(), v.Len(), v.Len())
		for i := 0; i < v.Len(); i++ {
			out.Index(i).Set(DeepCopy(v.Index(i)))
		}
		return out
	case reflect.Map:
		if v.IsNil() {
			return v
		}
		out := reflect.MakeMapWithSize(v.Type(), v.Len())
		for _, k := range v.MapKeys() {
			out.SetMapIndex(k, DeepCopy(v.MapIndex(k)))
		}
		return out
	}
	return v
}

// cloneWrapper builds a fresh wrapper of type t holding a copy of w's value.
func cloneWrapper(w Wrapper, t reflect.Type) reflect.Value {
	out := reflect.New(t)
	ow := out.Interface().(Wrapper)
	if w.IsNull() {
		ow.SetNull()
		return out
	}
	src := reflect.ValueOf(w.InnerForWrite()).Elem()
	dst := reflect.ValueOf(ow.InnerForWrite()).Elem()
	dst.Set(DeepCopy(src))
	return out
}

// ServerColumns returns the json names of a struct type's `validate:"-"` fields —
// columns the engine round-trips (scan + serialize + carried through the merge) but
// that are NOT part of the write schema. A request body may not set them: the handler
// rejects any that a client sends, the way TS's closed schemas reject them as unknown
// keys once the handler has stripped them from the merge base (e.g. serviceaccount
// `tokens`, org `account_id`). Cached per type. Top-level fields only, which is what
// every such column is.
func ServerColumns(t reflect.Type) []string {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil
	}
	if v, ok := serverColumns.Load(t); ok {
		return v.([]string)
	}
	var out []string
	for _, f := range planFor(t).fields {
		if f.skip {
			out = append(out, f.name)
		}
	}
	sort.Strings(out)
	actual, _ := serverColumns.LoadOrStore(t, out)
	return actual.([]string)
}

var serverColumns sync.Map // reflect.Type → []string
