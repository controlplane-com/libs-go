package wirejson

/* Reading one property out of a stored column does not require the column. Lookup
 * scans the bytes to the path and decodes only what it lands on, so an authz term on
 * `spec.containers` or a sort on `tags.env` costs a scan instead of a full parse plus a
 * map for every object on the way down. */

// Lookup returns the value at a dot-path inside a JSON document, decoded exactly as
// json.Unmarshal into `any` would produce it. found is false when a segment is missing
// or its parent is not an object — the same answer walking a map[string]any gives.
//
// An empty path returns the whole document. Duplicate keys resolve last-wins, as
// encoding/json does.
func Lookup(data []byte, path []string) (any, bool) {
	d := decoder{b: data}
	d.ws()
	return d.lookup(path, 0)
}

func (d *decoder) lookup(path []string, depth int) (any, bool) {
	if len(path) == 0 {
		v, err := d.value(depth)
		if err != nil {
			return nil, false
		}
		return v, true
	}
	if depth > maxDepth || d.i >= len(d.b) || d.b[d.i] != '{' {
		return nil, false
	}
	d.i++
	d.ws()
	if d.i < len(d.b) && d.b[d.i] == '}' {
		return nil, false
	}
	hit := -1
	for {
		d.ws()
		if d.i >= len(d.b) || d.b[d.i] != '"' {
			return nil, false
		}
		match, err := d.keyEquals(path[0])
		if err != nil {
			return nil, false
		}
		d.ws()
		if d.i >= len(d.b) || d.b[d.i] != ':' {
			return nil, false
		}
		d.i++
		d.ws()
		// Record rather than return: a later duplicate of the same key wins.
		if match {
			hit = d.i
		}
		if err := d.skip(depth + 1); err != nil {
			return nil, false
		}
		d.ws()
		if d.i >= len(d.b) {
			return nil, false
		}
		switch d.b[d.i] {
		case ',':
			d.i++
		case '}':
			if hit < 0 {
				return nil, false
			}
			d.i = hit
			return d.lookup(path[1:], depth+1)
		default:
			return nil, false
		}
	}
}

// skip advances past one value without building it. Composites are scanned by nesting
// depth (strings consumed properly, so a brace inside one cannot unbalance the count);
// scalars go through the real parser, so a malformed number is still an error.
func (d *decoder) skip(depth int) error {
	if depth > maxDepth {
		return errDepth
	}
	if d.i >= len(d.b) {
		return errSyntax
	}
	switch d.b[d.i] {
	case '"':
		return d.skipString()
	case '{', '[':
		nest := 0
		for d.i < len(d.b) {
			switch d.b[d.i] {
			case '"':
				if err := d.skipString(); err != nil {
					return err
				}
			case '{', '[':
				nest++
				d.i++
			case '}', ']':
				nest--
				d.i++
				if nest == 0 {
					return nil
				}
			default:
				d.i++
			}
		}
		return errSyntax
	default:
		_, err := d.value(depth)
		return err
	}
}

// keyEquals consumes the object key at d.i and reports whether it equals want. A plain
// ASCII key is compared against the raw bytes, so scanning past a large object does not
// allocate a string per key; anything escaped or non-ASCII goes through the real string
// parser, whose unescaping and U+FFFD substitution decide the comparison.
func (d *decoder) keyEquals(want string) (bool, error) {
	start := d.i
	for i := d.i + 1; i < len(d.b); i++ {
		c := d.b[i]
		if c == '"' {
			seg := d.b[d.i+1 : i]
			d.i = i + 1
			return string(seg) == want, nil
		}
		if c == '\\' || c >= 0x80 {
			break
		}
		if c < 0x20 {
			return false, errSyntax
		}
	}
	d.i = start
	k, err := d.str()
	if err != nil {
		return false, err
	}
	return k == want, nil
}

// skipString advances past a JSON string without building it.
func (d *decoder) skipString() error {
	for i := d.i + 1; i < len(d.b); i++ {
		switch d.b[i] {
		case '\\':
			i++
		case '"':
			d.i = i + 1
			return nil
		}
	}
	return errSyntax
}
