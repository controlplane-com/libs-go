package validate

// cursor.go abstracts where a body comes from. The decoder reads through a cursor so
// it is written once for both sources: request bytes (parsed by the wirejson scanner,
// no intermediate tree) and an in-memory tree (the merge path, whose result is a map
// until the merge itself is typed).

import (
	"errors"

	"github.com/controlplane-com/libs-go/pkg/schema/wirejson"
)

// Source is a body to decode: Bytes or Tree. ordered reports whether the source can
// still say what order the keys came in — bytes can, an already-parsed tree cannot.
type Source interface {
	cursor() cursor
	ordered() bool
}

// Bytes is a JSON document.
type Bytes []byte

// Tree is an already-parsed document in encoding/json's generic shapes.
type Tree struct{ V any }

func (b Bytes) cursor() cursor { return wirejson.NewCursor([]byte(b)) }
func (t Tree) cursor() cursor  { return &treeCursor{stack: []any{t.V}} }

func (b Bytes) ordered() bool { return true }
func (t Tree) ordered() bool  { return false }

type tok = wirejson.Token

// cursor is what the decoder drives. Every method consumes the value at the cursor;
// Object and Array hand each entry's value to the callback, which must consume it.
type cursor interface {
	Peek() (tok, error)
	String() (string, error)
	Number() (float64, error)
	Bool() (bool, error)
	Null() error
	Object(func(key string) error) error
	Array(func(i int) error) error
	Generic() (any, error)
	// Done reports trailing garbage after the top-level value.
	Done() error
}

// treeCursor walks a generic tree with a stack whose top is the current value.
type treeCursor struct{ stack []any }

var errTreeShape = errors.New("validate: value does not match the expected shape")

func (c *treeCursor) top() any { return c.stack[len(c.stack)-1] }
func (c *treeCursor) pop()     { c.stack = c.stack[:len(c.stack)-1] }

func (c *treeCursor) Peek() (tok, error) {
	switch c.top().(type) {
	case map[string]any:
		return wirejson.TokObject, nil
	case []any:
		return wirejson.TokArray, nil
	case string:
		return wirejson.TokString, nil
	case float64:
		return wirejson.TokNumber, nil
	case bool:
		return wirejson.TokBool, nil
	case nil:
		return wirejson.TokNull, nil
	}
	// An in-process tree may carry values the wire never does (a typed slice, an int,
	// a time). Those are handled as generic values by the decoder.
	return wirejson.TokOther, nil
}

func (c *treeCursor) String() (string, error) {
	s, ok := c.top().(string)
	if !ok {
		return "", errTreeShape
	}
	c.pop()
	return s, nil
}

func (c *treeCursor) Number() (float64, error) {
	f, ok := c.top().(float64)
	if !ok {
		return 0, errTreeShape
	}
	c.pop()
	return f, nil
}

func (c *treeCursor) Bool() (bool, error) {
	b, ok := c.top().(bool)
	if !ok {
		return false, errTreeShape
	}
	c.pop()
	return b, nil
}

func (c *treeCursor) Null() error {
	if c.top() != nil {
		return errTreeShape
	}
	c.pop()
	return nil
}

func (c *treeCursor) Object(fn func(key string) error) error {
	m, ok := c.top().(map[string]any)
	if !ok {
		return errTreeShape
	}
	c.pop()
	for _, k := range sortedKeys(m) {
		c.stack = append(c.stack, m[k])
		if err := fn(k); err != nil {
			return err
		}
	}
	return nil
}

func (c *treeCursor) Array(fn func(i int) error) error {
	arr, ok := c.top().([]any)
	if !ok {
		return errTreeShape
	}
	c.pop()
	for i, el := range arr {
		c.stack = append(c.stack, el)
		if err := fn(i); err != nil {
			return err
		}
	}
	return nil
}

func (c *treeCursor) Generic() (any, error) {
	v := c.top()
	c.pop()
	return v, nil
}

func (c *treeCursor) Done() error { return nil }
