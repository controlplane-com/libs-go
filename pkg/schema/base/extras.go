package base

import (
	"encoding/json"
	"reflect"
	"strings"
)

/* Some schema objects are declared `.unknown(true)`: they name a few properties and
 * accept any others, which round-trip untouched. A plain struct cannot hold those, so a
 * type that models one embeds Extras and delegates its JSON to the two helpers below —
 * otherwise decoding a stored value and re-encoding it silently deletes every property
 * the struct does not name. */

// Extras holds the unnamed properties of an `unknown(true)` object.
type Extras map[string]any

// ExtrasMarker tells the validate engine's decoder that allowed unknown keys of the
// enclosing struct belong here.
func (Extras) ExtrasMarker() {}

// namedKeys returns the json names the struct type itself declares.
func namedKeys(t reflect.Type) map[string]bool {
	out := map[string]bool{}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.PkgPath != "" {
			continue
		}
		name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
		if name == "" {
			name = f.Name
		}
		if name != "-" {
			out[name] = true
		}
	}
	return out
}

// MarshalExtras encodes v (which MUST be a defined-type alias of the real struct, so
// this does not recurse through its MarshalJSON) and splices the extras back in. A named
// property always wins over an extra of the same name.
func MarshalExtras(v any, extras Extras) ([]byte, error) {
	named, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	if len(extras) == 0 {
		return named, nil
	}
	var merged map[string]json.RawMessage
	if err := json.Unmarshal(named, &merged); err != nil {
		return nil, err
	}
	if merged == nil {
		merged = map[string]json.RawMessage{}
	}
	for k, x := range extras {
		if _, taken := merged[k]; taken {
			continue
		}
		b, err := json.Marshal(x)
		if err != nil {
			return nil, err
		}
		merged[k] = b
	}
	return json.Marshal(merged)
}

// UnmarshalExtras decodes into v (a pointer to the alias) and collects every property the
// struct does not name into *extras.
func UnmarshalExtras(b []byte, v any, extras *Extras) error {
	if err := json.Unmarshal(b, v); err != nil {
		return err
	}
	var all map[string]json.RawMessage
	if err := json.Unmarshal(b, &all); err != nil {
		return err
	}
	known := namedKeys(reflect.TypeOf(v).Elem())
	*extras = nil
	for k, raw := range all {
		if known[k] {
			continue
		}
		var x any
		if err := json.Unmarshal(raw, &x); err != nil {
			return err
		}
		if *extras == nil {
			*extras = Extras{}
		}
		(*extras)[k] = x
	}
	return nil
}
