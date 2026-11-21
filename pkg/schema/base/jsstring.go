package base

// jsstring.go ports JavaScript's String(value) coercion for the places a ported
// message interpolates a value into a template literal. Go's %v renders a map as
// `map[b:1]` and a slice as `[1 2]`, where JS renders `[object Object]` and `1,2` —
// so a faithful message has to do the coercion JS does, not Go's.

import (
	"reflect"
	"strconv"
	"strings"
)

// JSString is `${v}` — JavaScript's ToString:
//   - nil            -> "null" (an undefined value renders as "undefined"; callers
//     that distinguish the two pass JSUndefined explicitly)
//   - bool           -> "true" / "false"
//   - number         -> the shortest round-tripping form ("1", "1.5", "1e+21")
//   - string         -> itself
//   - array/slice    -> Array.prototype.toString: elements joined by "," with
//     null/undefined rendering as the empty string (recursively)
//   - anything else  -> "[object Object]"
func JSString(v any) string {
	if v == nil {
		return "null"
	}
	switch t := v.(type) {
	case string:
		return t
	case bool:
		return strconv.FormatBool(t)
	case float64:
		return strconv.FormatFloat(t, 'g', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(t), 'g', -1, 32)
	case int:
		return strconv.Itoa(t)
	case int32:
		return strconv.FormatInt(int64(t), 10)
	case int64:
		return strconv.FormatInt(t, 10)
	case uint, uint32, uint64:
		return strconv.FormatUint(reflect.ValueOf(v).Uint(), 10)
	}
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			return "null"
		}
		rv = rv.Elem()
	}
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		parts := make([]string, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			el := rv.Index(i).Interface()
			if el == nil {
				parts[i] = "" // JS renders null/undefined elements as empty
				continue
			}
			parts[i] = JSString(el)
		}
		return strings.Join(parts, ",")
	case reflect.String:
		return rv.String()
	case reflect.Bool:
		return strconv.FormatBool(rv.Bool())
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(rv.Float(), 'g', -1, 64)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(rv.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(rv.Uint(), 10)
	}
	return "[object Object]"
}

// jsWhitespace is ECMA-262 WhiteSpace + LineTerminator — what JavaScript's `\s`
// matches. It is wider than Go's ASCII whitespace set (NBSP, the U+2000 block, the
// BOM…), so `strings.ContainsAny(v, " \t\n\r\f\v")` let an image link carrying a NBSP
// or a BOM through where base.ts:198's /\s+/ rejects it. U+0085 (NEL) is deliberately
// NOT in the set: Go's unicode.IsSpace counts it, JavaScript's \s does not.
//
// The same set lives in validate.isJSSpace (unexported); if the engine ever exports
// it, delete this copy.
func jsWhitespace(r rune) bool {
	switch r {
	case '\t', '\v', '\f', '\r', '\n', ' ', 0x00A0, 0xFEFF, 0x1680,
		0x2028, 0x2029, 0x202F, 0x205F, 0x3000:
		return true
	}
	return r >= 0x2000 && r <= 0x200A
}

// ContainsJSWhitespace reports whether s contains a character JavaScript's `\s`
// matches.
func ContainsJSWhitespace(s string) bool { return strings.ContainsFunc(s, jsWhitespace) }
