package validate

// order.go carries the BODY'S key order from the decoder to the rules that need it.
//
// JavaScript iterates an object's keys in insertion order, so a TS custom asking for
// "the first key matching X" is answering from the wire. A Go map has no order, so
// every such port sorted instead — and that is not cosmetic: envoy's
// uiFilterNamingValidator picks the `cpln_` provider that decides which cluster/jwt
// pair the guard demands, so this port stored a config TS refuses and refused one TS
// stores, in both directions.
//
// Order is keyed by PATH, not by map identity: the value a custom sees has usually
// been rebuilt (validateObject returns a fresh map, and a custom may build its own),
// so an identity table would silently go empty. Both evaluators thread the same path
// through the same evalState, so neither can see an order the other cannot.

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

// wireOrder is the decoder's record of what each object's keys were, in wire order.
// base is how many path segments the validation prefixes that the decode did not see
// (Decoded.ValidateAt reporting a sub-document under "spec").
type wireOrder struct {
	tab  map[string][]string
	base int
}

// pathKey flattens a path to a table key. Segments are length-prefixed rather than
// joined by a separator: any byte, the engine's own array-index marker included, is
// legal inside a JSON key, so a separator would let one body's key collide with
// another's path.
func pathKey(path []string) string {
	var b strings.Builder
	for _, seg := range path {
		b.WriteString(strconv.Itoa(len(seg)))
		b.WriteByte(':')
		b.WriteString(seg)
	}
	return b.String()
}

func (w *wireOrder) at(path []string) []string {
	if w == nil || w.tab == nil || len(path) < w.base {
		return nil
	}
	return w.tab[pathKey(path[w.base:])]
}

// relPath appends rel — object keys (string) and array indices (int) — to base.
func relPath(base []string, rel []any) []string {
	path := append([]string(nil), base...)
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
	return path
}

// OrderedKeys returns v's keys the way `Object.keys(v)` orders them in the TS service:
// the order the request body listed them. rel addresses an object NESTED under the one
// the custom runs on (object keys as strings, array indices as ints); omit it for the
// value the custom was handed.
//
// v must be map-shaped — an object the engine keeps as a map, which is what a custom
// reading keys by name is looking at. An object decoded into a STRUCT has no key order
// to ask about: its keys are its schema fields, and both evaluators already walk those
// in schema order.
//
// Keys the wire order does not know — added by a default or a custom, or a value that
// never came from a body at all — fall in behind the ones it does, in the order a
// jsonb column hands them back. That is the right fallback rather than a plain sort:
// the order-less objects this engine sees are stored rows (the PATCH merge tree), and
// TS reads those out of a jsonb column, so jsonb's order IS its insertion order. A
// lexicographic fallback split the verdict on PATCH for `cpln_z` vs `cpln_aa`, where
// the two orders disagree.
func OrderedKeys(v any, h *Helper, rel ...any) []string {
	return orderKeys(mapKeys(v), h.wireKeys(rel))
}

// mapKeys is every string key v actually holds, whatever map type it is.
func mapKeys(v any) []string {
	rv := reflect.ValueOf(v)
	for rv.IsValid() && rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
	}
	if !rv.IsValid() || rv.Kind() != reflect.Map || rv.Type().Key().Kind() != reflect.String {
		return nil
	}
	out := make([]string, 0, rv.Len())
	for _, k := range rv.MapKeys() {
		out = append(out, k.String())
	}
	return out
}

// wireKeys is the recorded order for the object at h.Path + rel, or nil.
func (h *Helper) wireKeys(rel []any) []string {
	if h == nil || h.st == nil {
		return nil
	}
	if len(rel) == 0 {
		return h.st.order.at(h.Path)
	}
	return h.st.order.at(relPath(h.Path, rel))
}

// orderKeys is ES2015 OrdinaryOwnPropertyKeys over the keys present: array-index keys
// first in ascending numeric order however they were written, then the rest in wire
// order, then whatever the wire never carried, sorted.
func orderKeys(present, wire []string) []string {
	if len(present) < 2 {
		return present
	}
	left := make(map[string]bool, len(present))
	var idx []string
	for _, k := range present {
		if isArrayIndex(k) {
			idx = append(idx, k)
			continue
		}
		left[k] = true
	}
	sort.Slice(idx, func(i, j int) bool {
		return len(idx[i]) < len(idx[j]) || (len(idx[i]) == len(idx[j]) && idx[i] < idx[j])
	})
	out := make([]string, 0, len(present))
	out = append(out, idx...)
	for _, k := range wire {
		if left[k] {
			out = append(out, k)
			delete(left, k)
		}
	}
	rest := make([]string, 0, len(left))
	for k := range left {
		rest = append(rest, k)
	}
	sortJsonb(rest)
	return append(out, rest...)
}

// sortJsonb is postgres jsonb's key order: length first, then bytewise. For the
// canonical integer keys above it is also their numeric order.
func sortJsonb(keys []string) {
	sort.Slice(keys, func(i, j int) bool {
		if len(keys[i]) != len(keys[j]) {
			return len(keys[i]) < len(keys[j])
		}
		return keys[i] < keys[j]
	})
}

// isArrayIndex is JS's own test: the canonical decimal form of an integer in
// [0, 2^32-1). Those keys jump the queue regardless of when they were written.
func isArrayIndex(s string) bool {
	if s == "" || len(s) > 10 || (len(s) > 1 && s[0] == '0') {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	n, err := strconv.ParseUint(s, 10, 64)
	return err == nil && n < 1<<32-1
}
