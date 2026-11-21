package wirejson

import (
	"errors"
	"strconv"
	"unicode/utf16"
	"unicode/utf8"
)

// Decode parses JSON into the generic Go shapes encoding/json produces for an `any`
// target: map[string]any, []any, string, float64, bool, nil.
//
// It is the read-side twin of Append, and it exists for the same reason. A jsonb column
// arrives as bytes and is parsed into a map so handlers can work with it; profiling a
// paged query put encoding/json.Unmarshal inside store.JSONB.Scan at 24% of all CPU.
// The reflect-driven decoder is doing generic work that a parser targeting exactly
// these six shapes does not have to do.
//
// Same contract as the encoder, and the same enforcement: the result must be
// DeepEqual to what json.Unmarshal would have produced — including float64 for every
// number, last-wins duplicate keys, and U+FFFD for invalid UTF-8 — with anything
// malformed reported as an error rather than guessed at. decode_test.go compares the
// two over random documents.
func Decode(data []byte, out *any) error {
	d := decoder{b: data}
	d.ws()
	v, err := d.value(0)
	if err != nil {
		return err
	}
	d.ws()
	if d.i != len(d.b) {
		return errSyntax
	}
	*out = v
	return nil
}

var (
	errSyntax = errors.New("wirejson: invalid JSON")
	errDepth  = errors.New("wirejson: exceeded max depth")
)

// maxDepth matches encoding/json's nesting limit, so a hostile document is refused
// here rather than by a stack overflow.
const maxDepth = 10000

type decoder struct {
	b []byte
	i int
}

func (d *decoder) ws() {
	for d.i < len(d.b) {
		switch d.b[d.i] {
		case ' ', '\t', '\r', '\n':
			d.i++
		default:
			return
		}
	}
}

func (d *decoder) value(depth int) (any, error) {
	if depth > maxDepth {
		return nil, errDepth
	}
	if d.i >= len(d.b) {
		return nil, errSyntax
	}
	switch c := d.b[d.i]; c {
	case '{':
		return d.object(depth)
	case '[':
		return d.array(depth)
	case '"':
		return d.str()
	case 't':
		if len(d.b)-d.i >= 4 && string(d.b[d.i:d.i+4]) == "true" {
			d.i += 4
			return true, nil
		}
		return nil, errSyntax
	case 'f':
		if len(d.b)-d.i >= 5 && string(d.b[d.i:d.i+5]) == "false" {
			d.i += 5
			return false, nil
		}
		return nil, errSyntax
	case 'n':
		if len(d.b)-d.i >= 4 && string(d.b[d.i:d.i+4]) == "null" {
			d.i += 4
			return nil, nil
		}
		return nil, errSyntax
	default:
		if c == '-' || (c >= '0' && c <= '9') {
			return d.number()
		}
		return nil, errSyntax
	}
}

func (d *decoder) object(depth int) (any, error) {
	d.i++ // '{'
	m := make(map[string]any, 8)
	d.ws()
	if d.i < len(d.b) && d.b[d.i] == '}' {
		d.i++
		return m, nil
	}
	for {
		d.ws()
		if d.i >= len(d.b) || d.b[d.i] != '"' {
			return nil, errSyntax
		}
		k, err := d.str()
		if err != nil {
			return nil, err
		}
		d.ws()
		if d.i >= len(d.b) || d.b[d.i] != ':' {
			return nil, errSyntax
		}
		d.i++
		d.ws()
		v, err := d.value(depth + 1)
		if err != nil {
			return nil, err
		}
		m[k] = v // last wins, as encoding/json does
		d.ws()
		if d.i >= len(d.b) {
			return nil, errSyntax
		}
		switch d.b[d.i] {
		case ',':
			d.i++
		case '}':
			d.i++
			return m, nil
		default:
			return nil, errSyntax
		}
	}
}

func (d *decoder) array(depth int) (any, error) {
	d.i++ // '['
	d.ws()
	if d.i < len(d.b) && d.b[d.i] == ']' {
		d.i++
		return []any{}, nil
	}
	out := make([]any, 0, 8)
	for {
		d.ws()
		v, err := d.value(depth + 1)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
		d.ws()
		if d.i >= len(d.b) {
			return nil, errSyntax
		}
		switch d.b[d.i] {
		case ',':
			d.i++
		case ']':
			d.i++
			return out, nil
		default:
			return nil, errSyntax
		}
	}
}

// number scans a JSON number by the grammar, NOT by "digits and punctuation".
// strconv.ParseFloat is more permissive than JSON — it accepts a leading zero (01) and
// a trailing point (1.) — so scanning loosely and letting ParseFloat decide would
// silently accept documents encoding/json rejects.
func (d *decoder) number() (any, error) {
	start := d.i
	if d.i < len(d.b) && d.b[d.i] == '-' {
		d.i++
	}
	// int: 0 | [1-9][0-9]*
	if d.i >= len(d.b) {
		return nil, errSyntax
	}
	if d.b[d.i] == '0' {
		d.i++
	} else if d.b[d.i] >= '1' && d.b[d.i] <= '9' {
		d.digits()
	} else {
		return nil, errSyntax
	}
	// frac
	if d.i < len(d.b) && d.b[d.i] == '.' {
		d.i++
		if !d.digits() {
			return nil, errSyntax
		}
	}
	// exp
	if d.i < len(d.b) && (d.b[d.i] == 'e' || d.b[d.i] == 'E') {
		d.i++
		if d.i < len(d.b) && (d.b[d.i] == '+' || d.b[d.i] == '-') {
			d.i++
		}
		if !d.digits() {
			return nil, errSyntax
		}
	}
	// strconv.ParseFloat is what encoding/json uses, so the rounding is identical.
	f, err := strconv.ParseFloat(string(d.b[start:d.i]), 64)
	if err != nil {
		return nil, errSyntax
	}
	return f, nil
}

// digits consumes one or more digits and reports whether it consumed any.
func (d *decoder) digits() bool {
	n := d.i
	for d.i < len(d.b) && d.b[d.i] >= '0' && d.b[d.i] <= '9' {
		d.i++
	}
	return d.i > n
}

// str reads a JSON string. The common case — no backslash — is a single allocation
// over the raw bytes; only an escaped string takes the slow path.
func (d *decoder) str() (string, error) {
	d.i++ // opening quote
	start := d.i
	for d.i < len(d.b) {
		switch c := d.b[d.i]; {
		case c == '"':
			s := d.b[start:d.i]
			d.i++
			if !utf8.Valid(s) {
				// encoding/json replaces invalid UTF-8 with U+FFFD rather than
				// failing; a jsonb column can carry it.
				return string([]rune(string(s))), nil
			}
			return string(s), nil
		case c == '\\':
			return d.strEscaped(start)
		case c < 0x20:
			return "", errSyntax
		default:
			d.i++
		}
	}
	return "", errSyntax
}

func (d *decoder) strEscaped(start int) (string, error) {
	buf := make([]byte, 0, (len(d.b)-start)/2+16)
	buf = append(buf, d.b[start:d.i]...)
	for d.i < len(d.b) {
		c := d.b[d.i]
		switch {
		case c == '"':
			d.i++
			if !utf8.Valid(buf) {
				return string([]rune(string(buf))), nil
			}
			return string(buf), nil
		case c == '\\':
			d.i++
			if d.i >= len(d.b) {
				return "", errSyntax
			}
			switch e := d.b[d.i]; e {
			case '"', '\\', '/':
				buf = append(buf, e)
				d.i++
			case 'b':
				buf = append(buf, '\b')
				d.i++
			case 'f':
				buf = append(buf, '\f')
				d.i++
			case 'n':
				buf = append(buf, '\n')
				d.i++
			case 'r':
				buf = append(buf, '\r')
				d.i++
			case 't':
				buf = append(buf, '\t')
				d.i++
			case 'u':
				r, err := d.hex4()
				if err != nil {
					return "", err
				}
				if utf16.IsSurrogate(r) {
					// A lone surrogate becomes U+FFFD, as encoding/json does; a valid
					// pair is combined.
					if d.i+1 < len(d.b) && d.b[d.i] == '\\' && d.b[d.i+1] == 'u' {
						save := d.i
						// hex4 expects d.i ON the 'u' and steps past it itself; skipping
						// both bytes here ate the first hex digit, so every VALID
						// surrogate pair — `\ud83d\ude00`, what json.dumps and every
						// ASCII-safe encoder emit for an emoji — failed to parse.
						d.i++
						r2, err := d.hex4()
						if err != nil {
							return "", err
						}
						if dec := utf16.DecodeRune(r, r2); dec != utf8.RuneError {
							buf = utf8.AppendRune(buf, dec)
							continue
						}
						d.i = save
					}
					buf = utf8.AppendRune(buf, utf8.RuneError)
					continue
				}
				buf = utf8.AppendRune(buf, r)
			default:
				return "", errSyntax
			}
		case c < 0x20:
			return "", errSyntax
		default:
			buf = append(buf, c)
			d.i++
		}
	}
	return "", errSyntax
}

// hex4 reads the four hex digits of a \u escape; d.i points just past the 'u'.
func (d *decoder) hex4() (rune, error) {
	d.i++ // past 'u'
	if d.i+4 > len(d.b) {
		return 0, errSyntax
	}
	var r rune
	for _, c := range d.b[d.i : d.i+4] {
		var v rune
		switch {
		case c >= '0' && c <= '9':
			v = rune(c - '0')
		case c >= 'a' && c <= 'f':
			v = rune(c-'a') + 10
		case c >= 'A' && c <= 'F':
			v = rune(c-'A') + 10
		default:
			return 0, errSyntax
		}
		r = r*16 + v
	}
	d.i += 4
	return r, nil
}
