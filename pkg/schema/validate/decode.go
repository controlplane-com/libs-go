package validate

// decode.go is phase one of typed validation: fill a struct straight from the body,
// remembering what the wire actually said. The struct is the value the rest of the
// request works on; the record (a node per container) is what lets the walk in
// walk.go tell absent from null from a zero value, see undeclared keys, and report
// a value that did not fit its field with the rule's own message instead of a
// decode error. Nothing is rejected here except malformed JSON: every semantic
// decision belongs to the rule, and the rule runs in schema order so the first
// error reported is the one Joi would report.

import (
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"

	"github.com/controlplane-com/libs-go/pkg/schema/wirejson"
)

// Field states recorded per declared key / element.
const (
	stAbsent uint8 = iota
	stFits         // decoded into the field
	stNull         // explicit null
	stRaw          // did not fit the field's type; the generic value is kept for the rule
)

type kv struct {
	key string
	val any
}

// node records what the wire held at one container value.
type node struct {
	// struct destinations, indexed by destPlan field
	state   []uint8
	raw     map[int]any
	kids    map[int]*node
	unknown []kv
	// slice destinations, by element
	elemState []uint8
	elemRaw   map[int]any
	elems     map[int]*node
	// map destinations, by key
	mstate map[string]uint8
	mraw   map[string]any
	mkids  map[string]*node
}

// Decoded is a struct filled from a body plus the record of what the body held. It
// is handed to Validate once the caller has done whatever it does between decoding
// and validating (the pre-schema seams rewrite the row there).
type Decoded struct {
	dst   reflect.Value // the struct (addressable)
	root  *node
	order map[string][]string // wire key order, by path (see order.go)
}

// ErrSyntax is returned for a body that is not valid JSON.
var ErrSyntax = errors.New("invalid JSON")

// DecodeInto fills dst (a pointer to a struct) from src. dst's existing contents are
// kept where the body is silent, so a caller can pre-populate server-owned fields.
func DecodeInto(dst any, src Source) (*Decoded, error) {
	rv := reflect.ValueOf(dst)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return nil, fmt.Errorf("validate: DecodeInto needs a non-nil pointer, got %T", dst)
	}
	c := src.cursor()
	d := &dec{c: c}
	if src.ordered() {
		d.order = map[string][]string{}
	}
	inner, kind, _ := target(rv.Elem())
	if kind != dkStruct {
		return nil, fmt.Errorf("validate: DecodeInto needs a struct destination, got %s", rv.Elem().Type())
	}
	t, err := c.Peek()
	if err != nil {
		return nil, ErrSyntax
	}
	if t != wirejson.TokObject {
		// A non-object body: keep the generic value so the root rule reports the
		// object.base error itself.
		g, err := d.generic()
		if err != nil {
			return nil, ErrSyntax
		}
		if err := c.Done(); err != nil {
			return nil, ErrSyntax
		}
		return &Decoded{dst: inner, root: &node{raw: map[int]any{-1: g}}, order: d.order}, nil
	}
	n, err := d.structValue(inner, planFor(inner.Type()))
	if err != nil {
		return nil, wrapSyntax(err)
	}
	if err := c.Done(); err != nil {
		return nil, ErrSyntax
	}
	return &Decoded{dst: inner, root: n, order: d.order}, nil
}

func wrapSyntax(err error) error {
	if errors.Is(err, errTreeShape) {
		return err
	}
	return ErrSyntax
}

type dec struct {
	c    cursor
	path []string
	// order is the wire key order of every object the body carried, by path. Only a
	// byte source has one: a tree source's maps have no order left to record.
	order map[string][]string
}

func (d *dec) push(seg string) { d.path = append(d.path, seg) }
func (d *dec) pop()            { d.path = d.path[:len(d.path)-1] }

// record files the wire key order of the object the cursor just finished. One key
// cannot be out of order, so only a real choice is worth remembering.
func (d *dec) record(keys []string) {
	if d.order != nil && len(keys) > 1 {
		d.order[pathKey(d.path)] = keys
	}
}

// generic reads the value under the cursor as a tree, recording the key order of
// every object inside it. The cursor's own Generic() builds the whole subtree in one
// go with no path to hang the order on — and a generic subtree (`http[]`, whose Go
// type is the union's `any`) is exactly where a custom asks for it.
func (d *dec) generic() (any, error) {
	if d.order == nil {
		return d.c.Generic()
	}
	t, err := d.c.Peek()
	if err != nil {
		return nil, err
	}
	switch t {
	case wirejson.TokObject:
		m := map[string]any{}
		var keys []string
		err := d.c.Object(func(key string) error {
			d.push(key)
			v, err := d.generic()
			d.pop()
			if err != nil {
				return err
			}
			if _, dup := m[key]; !dup { // duplicate: last value wins, first POSITION wins
				keys = append(keys, key)
			}
			m[key] = v
			return nil
		})
		if err != nil {
			return nil, err
		}
		d.record(keys)
		return m, nil
	case wirejson.TokArray:
		out := []any{}
		err := d.c.Array(func(i int) error {
			d.push(idxSeg(i))
			v, err := d.generic()
			d.pop()
			if err != nil {
				return err
			}
			out = append(out, v)
			return nil
		})
		if err != nil {
			return nil, err
		}
		return out, nil
	default:
		return d.c.Generic()
	}
}

// value decodes the value under the cursor into v (a settable field or element).
func (d *dec) value(v reflect.Value) (state uint8, raw any, kid *node, err error) {
	t, err := d.c.Peek()
	if err != nil {
		return 0, nil, nil, err
	}
	if t == wirejson.TokNull {
		if err := d.c.Null(); err != nil {
			return 0, nil, nil, err
		}
		setNull(v)
		return stNull, nil, nil, nil
	}
	if t == wirejson.TokOther {
		// an in-process value from a tree source: take it if it is already the field's
		// type, else keep it for the rule
		g, err := d.c.Generic()
		if err != nil {
			return 0, nil, nil, err
		}
		if g != nil && reflect.TypeOf(g).AssignableTo(v.Type()) {
			v.Set(DeepCopy(reflect.ValueOf(g)))
			return stFits, nil, nil, nil
		}
		inner, _, _ := target(v)
		if g != nil && reflect.TypeOf(g).AssignableTo(inner.Type()) {
			inner.Set(DeepCopy(reflect.ValueOf(g)))
			return stFits, nil, nil, nil
		}
		clear(v)
		return stRaw, g, nil, nil
	}
	inner, kind, _ := target(v)
	switch kind {
	case dkString:
		if t == wirejson.TokString {
			s, err := d.c.String()
			if err != nil {
				return 0, nil, nil, err
			}
			inner.SetString(s)
			return stFits, nil, nil, nil
		}
	case dkNumber:
		if t == wirejson.TokNumber {
			f, err := d.c.Number()
			if err != nil {
				return 0, nil, nil, err
			}
			if setNumber(inner, f) {
				return stFits, nil, nil, nil
			}
			clear(v)
			return stRaw, f, nil, nil
		}
	case dkBool:
		if t == wirejson.TokBool {
			b, err := d.c.Bool()
			if err != nil {
				return 0, nil, nil, err
			}
			inner.SetBool(b)
			return stFits, nil, nil, nil
		}
	case dkOpaque:
		if t == wirejson.TokString {
			s, err := d.c.String()
			if err != nil {
				return 0, nil, nil, err
			}
			if setOpaque(inner, s) {
				return stFits, nil, nil, nil
			}
			clear(v)
			return stRaw, s, nil, nil
		}
	case dkStruct:
		if t == wirejson.TokObject {
			n, err := d.structValue(inner, planFor(inner.Type()))
			return stFits, nil, n, err
		}
	case dkSlice:
		if t == wirejson.TokArray {
			n, err := d.sliceValue(inner)
			return stFits, nil, n, err
		}
	case dkMap:
		if t == wirejson.TokObject {
			n, err := d.mapValue(inner)
			return stFits, nil, n, err
		}
	case dkAny:
		g, err := d.generic()
		if err != nil {
			return 0, nil, nil, err
		}
		inner.Set(reflect.ValueOf(&g).Elem())
		return stFits, nil, nil, nil
	}
	// the value does not fit the field: keep it for the rule to judge
	g, err := d.generic()
	if err != nil {
		return 0, nil, nil, err
	}
	clear(v)
	return stRaw, g, nil, nil
}

func (d *dec) structValue(v reflect.Value, p *destPlan) (*node, error) {
	n := &node{state: make([]uint8, len(p.fields))}
	err := d.c.Object(func(key string) error {
		// No order is recorded for a struct destination: its keys are the schema's
		// fields, which both evaluators walk in schema order, and there is no map for
		// a custom to read keys off. The path is still pushed — a map or a generic
		// subtree BELOW here is addressed through it.
		d.push(key)
		err := d.structKey(v, p, n, key)
		d.pop()
		return err
	})
	return n, err
}

// structKey decodes the value of one key, with d.path already on it.
func (d *dec) structKey(v reflect.Value, p *destPlan, n *node, key string) error {
	i, ok := p.byName[key]
	if !ok {
		g, err := d.generic()
		if err != nil {
			return err
		}
		for j := range n.unknown {
			if n.unknown[j].key == key { // duplicate: last wins
				n.unknown[j].val = g
				return nil
			}
		}
		n.unknown = append(n.unknown, kv{key, g})
		return nil
	}
	f := p.field(v, i)
	if n.state[i] != stAbsent { // duplicate key: last wins
		clear(f)
		delete(n.raw, i)
		delete(n.kids, i)
	}
	st, raw, kid, err := d.value(f)
	if err != nil {
		return err
	}
	n.state[i] = st
	if st == stRaw {
		if n.raw == nil {
			n.raw = map[int]any{}
		}
		n.raw[i] = raw
	}
	if kid != nil {
		if n.kids == nil {
			n.kids = map[int]*node{}
		}
		n.kids[i] = kid
	}
	return nil
}

func (d *dec) sliceValue(v reflect.Value) (*node, error) {
	n := &node{}
	et := v.Type().Elem()
	out := reflect.MakeSlice(v.Type(), 0, 4)
	err := d.c.Array(func(i int) error {
		el := reflect.New(et).Elem()
		d.push(idxSeg(i))
		st, raw, kid, err := d.value(el)
		d.pop()
		if err != nil {
			return err
		}
		out = reflect.Append(out, el)
		n.elemState = append(n.elemState, st)
		if st == stRaw {
			if n.elemRaw == nil {
				n.elemRaw = map[int]any{}
			}
			n.elemRaw[i] = raw
		}
		if kid != nil {
			if n.elems == nil {
				n.elems = map[int]*node{}
			}
			n.elems[i] = kid
		}
		return nil
	})
	v.Set(out)
	return n, err
}

func (d *dec) mapValue(v reflect.Value) (*node, error) {
	n := &node{mstate: map[string]uint8{}}
	et := v.Type().Elem()
	if v.IsNil() {
		v.Set(reflect.MakeMap(v.Type()))
	}
	var keys []string
	err := d.c.Object(func(key string) error {
		el := reflect.New(et).Elem()
		d.push(key)
		st, raw, kid, err := d.value(el)
		d.pop()
		if err != nil {
			return err
		}
		if _, dup := n.mstate[key]; !dup { // duplicate: last value wins, first POSITION wins
			keys = append(keys, key)
		}
		v.SetMapIndex(reflect.ValueOf(key).Convert(v.Type().Key()), el)
		n.mstate[key] = st
		if st == stRaw {
			if n.mraw == nil {
				n.mraw = map[string]any{}
			}
			n.mraw[key] = raw
		}
		if kid != nil {
			if n.mkids == nil {
				n.mkids = map[string]*node{}
			}
			n.mkids[key] = kid
		}
		return nil
	})
	d.record(keys)
	return n, err
}

// setNumber stores f into a numeric field, reporting false when the field's kind
// cannot hold it exactly (a fraction or overflow into an integer) — the rule then
// judges the raw number and says "must be an integer" rather than the decoder
// truncating it.
func setNumber(v reflect.Value, f float64) bool {
	switch v.Kind() {
	case reflect.Float32, reflect.Float64:
		v.SetFloat(f)
		return true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if f != math.Trunc(f) || f < math.MinInt64 || f > math.MaxInt64 {
			return false
		}
		i := int64(f)
		if v.OverflowInt(i) {
			return false
		}
		v.SetInt(i)
		return true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if f != math.Trunc(f) || f < 0 || f > math.MaxUint64 {
			return false
		}
		u := uint64(f)
		if v.OverflowUint(u) {
			return false
		}
		v.SetUint(u)
		return true
	}
	return false
}

// setOpaque stores a JSON string into a self-unmarshalling scalar type.
func setOpaque(v reflect.Value, s string) bool {
	p := v.Addr().Interface()
	if tu, ok := p.(encoding.TextUnmarshaler); ok {
		return tu.UnmarshalText([]byte(s)) == nil
	}
	if ju, ok := p.(json.Unmarshaler); ok {
		return ju.UnmarshalJSON(wirejson.AppendString(nil, s)) == nil
	}
	return false
}
