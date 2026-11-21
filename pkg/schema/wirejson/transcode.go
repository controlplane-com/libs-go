package wirejson

import (
	"strconv"
	"unicode/utf8"
)

// Transcode rewrites a jsonb column's bytes as the bytes this service puts on the wire,
// in ONE pass and without building a single map.
//
// A jsonb column arrives from Postgres as JSON text. The read path used to parse it into
// map[string]any, copy that into the API object, and encode it again — measured at 354µs
// per 50-row page for the decode alone, and the largest single item in the service. None
// of that is needed to answer a list or query: nothing reads `spec`.
//
// It is a transcode rather than a straight copy because Postgres's text is close to, but
// not the same as, what encoding/json would have produced:
//
//   - Numbers. jsonb keeps `numeric`, so 1e21 comes back as 1000000000000000000000 and
//     1e-7 as 0.0000001. Go renders the float64 as 1e+21 and 1e-07. Numbers are therefore
//     re-rendered through float64 — the same lossy step the old decode took, so the bytes
//     do not move and neither does the precision.
//   - Whitespace. Postgres emits `{"a": 1, "b": 2}`; the encoder emits it tight.
//   - U+2028/U+2029. Postgres passes them through raw; the encoder escapes them, and so
//     must this, or a spec string could break a JavaScript consumer.
//
// KEY ORDER is deliberately left as Postgres gives it (length, then bytewise). That is a
// change: the map path sorted keys alphabetically. It is a change TOWARD the TS service,
// which hands the driver's parsed object straight to JSON.stringify and therefore emits
// Postgres's order too — see INTENTIONAL-DIVERGENCES C12.
//
// Invalid input is reported rather than guessed at; callers fall back to the decode.
func Transcode(dst, src []byte) ([]byte, error) {
	t := transcoder{src: src, dst: dst}
	t.ws()
	if err := t.value(0); err != nil {
		return nil, err
	}
	t.ws()
	if t.i != len(t.src) {
		return nil, errSyntax
	}
	return t.dst, nil
}

type transcoder struct {
	src []byte
	dst []byte
	i   int
}

func (t *transcoder) ws() {
	for t.i < len(t.src) {
		switch t.src[t.i] {
		case ' ', '\t', '\r', '\n':
			t.i++
		default:
			return
		}
	}
}

func (t *transcoder) value(depth int) error {
	if depth > maxDepth {
		return errDepth
	}
	if t.i >= len(t.src) {
		return errSyntax
	}
	switch c := t.src[t.i]; {
	case c == '{':
		return t.object(depth)
	case c == '[':
		return t.array(depth)
	case c == '"':
		return t.str()
	case c == 't':
		return t.lit("true")
	case c == 'f':
		return t.lit("false")
	case c == 'n':
		return t.lit("null")
	case c == '-' || (c >= '0' && c <= '9'):
		return t.number()
	}
	return errSyntax
}

func (t *transcoder) lit(want string) error {
	if len(t.src)-t.i < len(want) || string(t.src[t.i:t.i+len(want)]) != want {
		return errSyntax
	}
	t.dst = append(t.dst, want...)
	t.i += len(want)
	return nil
}

func (t *transcoder) object(depth int) error {
	t.i++
	t.dst = append(t.dst, '{')
	t.ws()
	if t.i < len(t.src) && t.src[t.i] == '}' {
		t.i++
		t.dst = append(t.dst, '}')
		return nil
	}
	for n := 0; ; n++ {
		t.ws()
		if t.i >= len(t.src) || t.src[t.i] != '"' {
			return errSyntax
		}
		if n > 0 {
			t.dst = append(t.dst, ',')
		}
		if err := t.str(); err != nil {
			return err
		}
		t.ws()
		if t.i >= len(t.src) || t.src[t.i] != ':' {
			return errSyntax
		}
		t.i++
		t.dst = append(t.dst, ':')
		t.ws()
		if err := t.value(depth + 1); err != nil {
			return err
		}
		t.ws()
		if t.i >= len(t.src) {
			return errSyntax
		}
		switch t.src[t.i] {
		case ',':
			t.i++
		case '}':
			t.i++
			t.dst = append(t.dst, '}')
			return nil
		default:
			return errSyntax
		}
	}
}

func (t *transcoder) array(depth int) error {
	t.i++
	t.dst = append(t.dst, '[')
	t.ws()
	if t.i < len(t.src) && t.src[t.i] == ']' {
		t.i++
		t.dst = append(t.dst, ']')
		return nil
	}
	for n := 0; ; n++ {
		t.ws()
		if n > 0 {
			t.dst = append(t.dst, ',')
		}
		if err := t.value(depth + 1); err != nil {
			return err
		}
		t.ws()
		if t.i >= len(t.src) {
			return errSyntax
		}
		switch t.src[t.i] {
		case ',':
			t.i++
		case ']':
			t.i++
			t.dst = append(t.dst, ']')
			return nil
		default:
			return errSyntax
		}
	}
}

// number re-renders through float64 so the bytes match what the map path produced.
func (t *transcoder) number() error {
	start := t.i
	if t.src[t.i] == '-' {
		t.i++
	}
	if t.i >= len(t.src) {
		return errSyntax
	}
	if t.src[t.i] == '0' {
		t.i++
	} else if t.src[t.i] >= '1' && t.src[t.i] <= '9' {
		t.digits()
	} else {
		return errSyntax
	}
	if t.i < len(t.src) && t.src[t.i] == '.' {
		t.i++
		if !t.digits() {
			return errSyntax
		}
	}
	if t.i < len(t.src) && (t.src[t.i] == 'e' || t.src[t.i] == 'E') {
		t.i++
		if t.i < len(t.src) && (t.src[t.i] == '+' || t.src[t.i] == '-') {
			t.i++
		}
		if !t.digits() {
			return errSyntax
		}
	}
	f, err := strconv.ParseFloat(string(t.src[start:t.i]), 64)
	if err != nil {
		return errSyntax
	}
	t.dst, err = appendFloat(t.dst, f)
	return err
}

func (t *transcoder) digits() bool {
	n := t.i
	for t.i < len(t.src) && t.src[t.i] >= '0' && t.src[t.i] <= '9' {
		t.i++
	}
	return t.i > n
}

// str copies a string through, re-escaping only what the encoder escapes differently
// from Postgres. Postgres already emits the short escapes (\n, \t, \") and \uXXXX for
// other control characters, so the only rewrite is U+2028/U+2029, which it passes raw.
func (t *transcoder) str() error {
	t.i++ // opening quote
	t.dst = append(t.dst, '"')
	start := t.i
	for t.i < len(t.src) {
		c := t.src[t.i]
		switch {
		case c == '"':
			t.dst = append(t.dst, t.src[start:t.i]...)
			t.dst = append(t.dst, '"')
			t.i++
			return nil
		case c == '\\':
			// Validate the escape rather than skipping two bytes: a transcoder that
			// accepts more than the decoder would let malformed jsonb reach the wire
			// unexamined. Postgres only ever emits these.
			if t.i+1 >= len(t.src) {
				return errSyntax
			}
			switch t.src[t.i+1] {
			case '"', '\\', '/', 'b', 'f', 'n', 'r', 't':
				t.i += 2
			case 'u':
				if t.i+6 > len(t.src) {
					return errSyntax
				}
				for _, h := range t.src[t.i+2 : t.i+6] {
					if !isHex(h) {
						return errSyntax
					}
				}
				t.i += 6
			default:
				return errSyntax
			}
		case c < 0x20:
			return errSyntax
		case c < utf8.RuneSelf:
			t.i++
		default:
			r, size := utf8.DecodeRune(t.src[t.i:])
			if r == ' ' || r == ' ' {
				t.dst = append(t.dst, t.src[start:t.i]...)
				t.dst = append(t.dst, '\\', 'u', '2', '0', '2', hex[r&0xF])
				t.i += size
				start = t.i
				continue
			}
			// An invalid byte is left as-is: encoding/json would substitute U+FFFD,
			// but jsonb cannot hold invalid UTF-8 — Postgres rejects it on write.
			t.i += size
		}
	}
	return errSyntax
}

func isHex(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}
