// Package wirejson serializes an API object — the map[string]any shape every handler
// builds — to JSON bytes.
//
// It exists because that one operation is the single largest CPU line item in the
// service. Profiling an org-wide workload `-query` under load put
// encoding/json.mapEncoder at 34% of all CPU: for every response the reflect-based
// encoder walks an interface tree, and for every nested object it allocates a slice of
// reflect.Values to sort. The values in these maps come from a JSON decode, so the type
// set is closed and tiny — a type switch does the same work without reflection.
//
// It is NOT a general JSON encoder and must never become one. The contract is that
// Append produces the bytes httpx.WriteJSON's encoder already produced — an
// encoding/json.Encoder with SetEscapeHTML(FALSE), which is what matches
// JSON.stringify and therefore the TS service: `<`, `>` and `&` go out as themselves.
// Anything this package does not recognize is handed back to encoding/json rather than
// guessed at. The contract is enforced differentially (encode_test.go): the two
// encoders are compared byte-for-byte, including over random trees of adversarial
// strings.
package wirejson

import (
	"bytes"
	"encoding/json"
	"math"
	"slices"
	"strconv"
	"sync"
	"unicode/utf8"
)

// Append appends the JSON encoding of v to dst.
func Append(dst []byte, v any) ([]byte, error) {
	return appendValue(dst, v)
}

// Marshal returns the JSON encoding of v.
func Marshal(v any) ([]byte, error) {
	b, err := appendValue(make([]byte, 0, 512), v)
	if err != nil {
		return nil, err
	}
	return b, nil
}

var keyPool = sync.Pool{New: func() any { s := make([]string, 0, 32); return &s }}

func appendValue(dst []byte, v any) ([]byte, error) {
	switch t := v.(type) {
	case nil:
		return append(dst, "null"...), nil
	case string:
		return appendString(dst, t), nil
	case bool:
		if t {
			return append(dst, "true"...), nil
		}
		return append(dst, "false"...), nil
	case float64:
		return appendFloat(dst, t)
	case int:
		return strconv.AppendInt(dst, int64(t), 10), nil
	case int64:
		return strconv.AppendInt(dst, t, 10), nil
	case json.RawMessage:
		// Already JSON. encoding/json compacts it; so do we, via the fallback, because
		// re-deriving compaction here would be a second parser to keep honest.
		return appendViaStdlib(dst, t)
	case map[string]any:
		return appendObject(dst, t)
	case []any:
		if t == nil {
			// a nil slice is `null`, not `[]` — encoding/json distinguishes them
			return append(dst, "null"...), nil
		}
		dst = append(dst, '[')
		var err error
		for i, e := range t {
			if i > 0 {
				dst = append(dst, ',')
			}
			if dst, err = appendValue(dst, e); err != nil {
				return nil, err
			}
		}
		return append(dst, ']'), nil
	case []map[string]any:
		if t == nil {
			// a nil slice is `null`, not `[]` — encoding/json distinguishes them
			return append(dst, "null"...), nil
		}
		dst = append(dst, '[')
		var err error
		for i, e := range t {
			if i > 0 {
				dst = append(dst, ',')
			}
			if dst, err = appendObject(dst, e); err != nil {
				return nil, err
			}
		}
		return append(dst, ']'), nil
	case []string:
		if t == nil {
			// a nil slice is `null`, not `[]` — encoding/json distinguishes them
			return append(dst, "null"...), nil
		}
		dst = append(dst, '[')
		for i, e := range t {
			if i > 0 {
				dst = append(dst, ',')
			}
			dst = appendString(dst, e)
		}
		return append(dst, ']'), nil
	default:
		// A row struct (or a slice of them) takes the plan-based appender the store
		// registers, so a typed response never pays the reflection encoder — or the
		// re-scan encoding/json runs over every MarshalJSON result.
		if structAppender != nil {
			if out, ok, err := structAppender(dst, v); ok {
				return out, err
			}
		}
		return appendViaStdlib(dst, v)
	}
}

// structAppender is consulted for any value appendValue has no case for. It reports
// ok=false to decline, in which case the standard library encodes the value.
var structAppender func(dst []byte, v any) ([]byte, bool, error)

// SetStructAppender installs the appender for row structs. The store registers its
// plan-based encoder here at init; wirejson cannot import the store.
func SetStructAppender(fn func(dst []byte, v any) ([]byte, bool, error)) { structAppender = fn }

// The scalar appenders, exported for the row encoder so a field is appended without
// boxing it into an interface first.

func AppendString(dst []byte, s string) []byte          { return appendString(dst, s) }
func AppendFloat(dst []byte, f float64) ([]byte, error) { return appendFloat(dst, f) }
func AppendBool(dst []byte, b bool) []byte {
	if b {
		return append(dst, "true"...)
	}
	return append(dst, "false"...)
}
func AppendInt(dst []byte, i int64) []byte   { return strconv.AppendInt(dst, i, 10) }
func AppendUint(dst []byte, u uint64) []byte { return strconv.AppendUint(dst, u, 10) }

// AppendStdlib encodes v with encoding/json (SetEscapeHTML(false)) — the fallback for a
// field type the row encoder does not model.
func AppendStdlib(dst []byte, v any) ([]byte, error) { return appendViaStdlib(dst, v) }

func appendObject(dst []byte, m map[string]any) ([]byte, error) {
	if m == nil {
		return append(dst, "null"...), nil
	}
	kp := keyPool.Get().(*[]string)
	keys := (*kp)[:0]
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys) // encoding/json sorts map keys; the wire order must not drift
	dst = append(dst, '{')
	var err error
	for i, k := range keys {
		if i > 0 {
			dst = append(dst, ',')
		}
		dst = appendString(dst, k)
		dst = append(dst, ':')
		if dst, err = appendValue(dst, m[k]); err != nil {
			*kp = keys
			keyPool.Put(kp)
			return nil, err
		}
	}
	*kp = keys
	keyPool.Put(kp)
	return append(dst, '}'), nil
}

// appendViaStdlib is the escape hatch: a type this package does not model is encoded by
// encoding/json, so an unrecognized value is slow, never wrong. SetEscapeHTML(false) to
// match the rest of this encoder and the writer it replaces.
func appendViaStdlib(dst []byte, v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return append(dst, bytes.TrimSuffix(buf.Bytes(), []byte("\n"))...), nil
}

// appendFloat reproduces encoding/json's floatEncoder, including its 'e'-format
// exponent cleanup. A JSON number decoded into `any` is always a float64, so this runs
// for every numeric field in every response.
func appendFloat(dst []byte, f float64) ([]byte, error) {
	if math.IsInf(f, 0) || math.IsNaN(f) {
		return nil, &json.UnsupportedValueError{Str: strconv.FormatFloat(f, 'g', -1, 64)}
	}
	abs := math.Abs(f)
	fmtByte := byte('f')
	if abs != 0 && (abs < 1e-6 || abs >= 1e21) {
		fmtByte = 'e'
	}
	dst = strconv.AppendFloat(dst, f, fmtByte, -1, 64)
	if fmtByte == 'e' {
		// clean up e-09 to e-9
		n := len(dst)
		if n >= 4 && dst[n-4] == 'e' && dst[n-3] == '-' && dst[n-2] == '0' {
			dst[n-2] = dst[n-1]
			dst = dst[:n-1]
		}
	}
	return dst, nil
}

const hex = "0123456789abcdef"

// appendString reproduces encoding/json's string encoder with HTML escaping ON, which
// is encoding/json.Marshal's default: <, > and & become <, > and &, and
// U+2028/U+2029 are escaped so the output is safe to embed in a script tag. Getting
// this wrong is invisible until a tag or an ampersand appears in a description.
func appendString(dst []byte, s string) []byte {
	dst = append(dst, '"')
	start := 0
	for i := 0; i < len(s); {
		if b := s[i]; b < utf8.RuneSelf {
			if safeSet[b] {
				i++
				continue
			}
			dst = append(dst, s[start:i]...)
			switch b {
			case '\\', '"':
				dst = append(dst, '\\', b)
			case '\b':
				dst = append(dst, '\\', 'b')
			case '\f':
				dst = append(dst, '\\', 'f')
			case '\n':
				dst = append(dst, '\\', 'n')
			case '\r':
				dst = append(dst, '\\', 'r')
			case '\t':
				dst = append(dst, '\\', 't')
			default:
				// \u00xx for the remaining control characters, and for <, >, & the
				// same escape encoding/json uses.
				dst = append(dst, '\\', 'u', '0', '0', hex[b>>4], hex[b&0xF])
			}
			i++
			start = i
			continue
		}
		c, size := utf8.DecodeRuneInString(s[i:])
		// U+2028/U+2029 are escaped even with HTML escaping off — they are valid JSON
		// but terminate a JavaScript string literal. encoding/json does this
		// unconditionally, so matching it is not optional.
		if c == '\u2028' || c == '\u2029' {
			dst = append(dst, s[start:i]...)
			dst = append(dst, '\\', 'u', '2', '0', '2', hex[c&0xF])
			i += size
			start = i
			continue
		}
		if c == utf8.RuneError && size == 1 {
			// An invalid byte becomes U+FFFD as a LITERAL character, not the `\ufffd`
			// escape: that is what encoding/json emits (it changed in go1.27; the escape
			// was the pre-1.27 form) and what JSON.stringify emits, so escaping it here
			// put a byte difference on the wire for any string carrying invalid UTF-8.
			dst = append(dst, s[start:i]...)
			dst = utf8.AppendRune(dst, utf8.RuneError)
			i += size
			start = i
			continue
		}
		i += size
	}
	dst = append(dst, s[start:]...)
	return append(dst, '"')
}

// safeSet reports whether a byte can appear in a JSON string unescaped, matching
// encoding/json's safeSet — the escapeHTML=false table, so `<`, `>` and `&` are safe.
var safeSet = func() [utf8.RuneSelf]bool {
	var s [utf8.RuneSelf]bool
	for i := 0x20; i < utf8.RuneSelf; i++ {
		s[i] = true
	}
	s['"'] = false
	s['\\'] = false
	return s
}()
