package database

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"reflect"
)

// JSONB is a custom type for handling JSONB fields in PostgreSQL
// It implements sql.Scanner and driver.Valuer interfaces for automatic
// JSON marshaling/unmarshaling with GORM
type JSONB map[string]any

// Value implements the driver.Valuer interface for writing to the database
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface for reading from the database
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	result := make(map[string]any)
	if err := json.Unmarshal(bytes, &result); err != nil {
		return err
	}

	*j = result
	return nil
}

// MarshalJSON implements json.Marshaler for JSONB
// This ensures JSONB serializes as a nested JSON object rather than base64-encoded bytes
func (j JSONB) MarshalJSON() ([]byte, error) {
	if j == nil {
		return []byte("null"), nil
	}
	return json.Marshal(map[string]any(j))
}

// UnmarshalJSON implements json.Unmarshaler for JSONB
// This allows JSONB to be deserialized from nested JSON objects
func (j *JSONB) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		*j = nil
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	*j = JSONB(m)
	return nil
}

// GormDataType tells GORM what SQL data type to use for this field
func (JSONB) GormDataType() string {
	return "jsonb"
}

// JSONBG is a generic JSONB wrapper that unmarshals into the concrete type parameter T.
// It implements sql.Scanner and driver.Valuer, and also json.Marshaler/json.Unmarshaler
// so it can be used seamlessly with GORM and HTTP JSON payloads.
//
// Example:
//
//	type MyCfg struct { Enabled bool `json:"enabled"` }
//	type Model struct { Cfg JSONBG[MyCfg] `gorm:"type:jsonb"` }
//
// In the database this is stored as jsonb, and when scanning it will populate MyCfg.
// When marshaling to JSON (e.g. API responses), it serializes the inner value of type T.
//
// Note: The zero value contains the zero value of T.
// If T is a pointer type, the zero value will be nil and NULL may be written to DB.
// If you want an empty object `{}` instead, use a non-pointer struct type for T.
//
// This type avoids map[string]any when a strongly-typed struct is desired.
type JSONBG[T any] struct {
	V T
}

// Value implements driver.Valuer: it marshals the inner value to JSON.
func (j *JSONBG[T]) Value() (driver.Value, error) {
	// Handle nil receiver - write SQL NULL
	// This allows optional *JSONBG fields to be nil without causing panics
	if j == nil {
		return nil, nil
	}

	// If the inner value is a nil pointer/map/slice/interface, write SQL NULL.
	// This mirrors the behavior of JSONB(map) which writes NULL for a nil map.
	v := any(j.V)
	rv := reflect.ValueOf(v)
	// When T is an interface type (e.g., any) and V is nil, reflect.ValueOf returns an invalid Value.
	if !rv.IsValid() {
		return nil, nil
	}
	switch rv.Kind() {
	case reflect.Ptr, reflect.Slice, reflect.Map, reflect.Interface:
		if rv.IsNil() {
			return nil, nil
		}
	}
	return json.Marshal(j.V)
}

// Scan implements sql.Scanner: it unmarshals the DB value into the inner value.
func (j *JSONBG[T]) Scan(value interface{}) error {
	// Handle nil receiver - this should not happen in normal GORM usage,
	// but we protect against it for safety
	if j == nil {
		return errors.New("cannot scan into nil *JSONBG receiver")
	}

	if value == nil {
		var zero T
		j.V = zero
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, &j.V)
}

// MarshalJSON delegates to the inner value.
func (j *JSONBG[T]) MarshalJSON() ([]byte, error) {
	if j == nil {
		return []byte("null"), nil
	}
	return json.Marshal(j.V)
}

// UnmarshalJSON delegates to the inner value.
func (j *JSONBG[T]) UnmarshalJSON(b []byte) error {
	if j == nil {
		return errors.New("cannot unmarshal into nil *JSONBG receiver")
	}
	return json.Unmarshal(b, &j.V)
}

// GormDataType tells GORM what SQL data type to use for this field
func (*JSONBG[T]) GormDataType() string { return "jsonb" }
