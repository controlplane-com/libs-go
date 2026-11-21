package wirejson

// cursor.go is the streaming face of the decoder: a caller that knows what shape it
// expects (a typed decoder filling a struct) pulls one value at a time instead of
// receiving a whole tree. Every method consumes the value under the cursor; Object
// and Array call back per entry with the cursor positioned on the entry's value,
// which the callback must consume.

// Token is the kind of the value under the cursor.
type Token uint8

const (
	TokObject Token = iota + 1
	TokArray
	TokString
	TokNumber
	TokBool
	TokNull
	// TokOther is never produced from bytes; a tree source uses it for in-process
	// values that have no JSON token (a typed slice, an int).
	TokOther
)

// Cursor is a streaming reader over one JSON document.
type Cursor struct {
	d     decoder
	depth int
}

// NewCursor positions a cursor on the document's top-level value.
func NewCursor(data []byte) *Cursor {
	c := &Cursor{d: decoder{b: data}}
	c.d.ws()
	return c
}

// Peek reports the kind of the value under the cursor without consuming it.
func (c *Cursor) Peek() (Token, error) {
	d := &c.d
	if d.i >= len(d.b) {
		return 0, errSyntax
	}
	switch ch := d.b[d.i]; ch {
	case '{':
		return TokObject, nil
	case '[':
		return TokArray, nil
	case '"':
		return TokString, nil
	case 't', 'f':
		return TokBool, nil
	case 'n':
		return TokNull, nil
	default:
		if ch == '-' || (ch >= '0' && ch <= '9') {
			return TokNumber, nil
		}
		return 0, errSyntax
	}
}

func (c *Cursor) String() (string, error) {
	if c.d.i >= len(c.d.b) || c.d.b[c.d.i] != '"' {
		return "", errSyntax
	}
	return c.d.str()
}

func (c *Cursor) Number() (float64, error) {
	v, err := c.d.number()
	if err != nil {
		return 0, err
	}
	return v.(float64), nil
}

func (c *Cursor) Bool() (bool, error) {
	v, err := c.d.value(c.depth)
	if err != nil {
		return false, err
	}
	b, ok := v.(bool)
	if !ok {
		return false, errSyntax
	}
	return b, nil
}

func (c *Cursor) Null() error {
	v, err := c.d.value(c.depth)
	if err != nil {
		return err
	}
	if v != nil {
		return errSyntax
	}
	return nil
}

// Generic consumes the value as a tree, exactly as Decode would produce it.
func (c *Cursor) Generic() (any, error) { return c.d.value(c.depth) }

// Object consumes an object, calling fn for each key with the cursor on the value.
// A duplicate key is delivered again (last wins is the callback's business).
func (c *Cursor) Object(fn func(key string) error) error {
	d := &c.d
	if c.depth > maxDepth {
		return errDepth
	}
	if d.i >= len(d.b) || d.b[d.i] != '{' {
		return errSyntax
	}
	d.i++
	d.ws()
	if d.i < len(d.b) && d.b[d.i] == '}' {
		d.i++
		return nil
	}
	c.depth++
	defer func() { c.depth-- }()
	for {
		d.ws()
		if d.i >= len(d.b) || d.b[d.i] != '"' {
			return errSyntax
		}
		k, err := d.str()
		if err != nil {
			return err
		}
		d.ws()
		if d.i >= len(d.b) || d.b[d.i] != ':' {
			return errSyntax
		}
		d.i++
		d.ws()
		if err := fn(k); err != nil {
			return err
		}
		d.ws()
		if d.i >= len(d.b) {
			return errSyntax
		}
		switch d.b[d.i] {
		case ',':
			d.i++
		case '}':
			d.i++
			return nil
		default:
			return errSyntax
		}
	}
}

// Array consumes an array, calling fn for each element with the cursor on it.
func (c *Cursor) Array(fn func(i int) error) error {
	d := &c.d
	if c.depth > maxDepth {
		return errDepth
	}
	if d.i >= len(d.b) || d.b[d.i] != '[' {
		return errSyntax
	}
	d.i++
	d.ws()
	if d.i < len(d.b) && d.b[d.i] == ']' {
		d.i++
		return nil
	}
	c.depth++
	defer func() { c.depth-- }()
	for i := 0; ; i++ {
		d.ws()
		if err := fn(i); err != nil {
			return err
		}
		d.ws()
		if d.i >= len(d.b) {
			return errSyntax
		}
		switch d.b[d.i] {
		case ',':
			d.i++
		case ']':
			d.i++
			return nil
		default:
			return errSyntax
		}
	}
}

// Done reports an error if anything but whitespace follows the consumed document.
func (c *Cursor) Done() error {
	c.d.ws()
	if c.d.i != len(c.d.b) {
		return errSyntax
	}
	return nil
}
