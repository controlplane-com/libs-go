package types

import (
	"errors"
	"fmt"
	"github.com/controlplane-com/libs-go/pkg/common"
	"reflect"
	"strings"
)

func FollowPointersToConcreteType(t reflect.Type) (reflect.Type, int) {
	l := 0
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
		l++
	}
	return t, l
}

func FollowPointersToConcreteValue(v reflect.Value) any {
	if !v.IsValid() || (v.Kind() == reflect.Pointer && v.IsNil()) {
		return nil
	}
	l := 0
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
		l++
	}
	return v.Interface()
}

// GetNilValue returns a nil value for the given type, if the type can be set to nil. Otherwise, it returns an error
func GetNilValue(t reflect.Type) (any, error) {
	z := reflect.New(t).Elem()
	err := common.Try(func() error {
		//IsNil will panic for non-nillable types
		z.IsNil()
		return nil
	})
	if err != nil {
		return nil, errors.New(fmt.Sprintf("cannot assign nil to a field of type %s", GetTypeName(t)))
	}
	return z.Interface(), nil
}

func GetTypeName(t reflect.Type) string {
	ct, l := FollowPointersToConcreteType(t)
	return fmt.Sprintf("%s%s", strings.Repeat("*", l), ct.Name())
}

func CopyAndIndirect(n int, s any) any {
	v := reflect.New(reflect.TypeOf(s)).Elem()
	v.Set(reflect.ValueOf(s))
	return indirectInternal(n, v).Interface()
}

func Indirect(n int, s any) (any, error) {
	r := reflect.ValueOf(s)
	if r.Kind() != reflect.Pointer {
		return nil, errors.New("unaddressable value given")
	}
	return indirectInternal(n, r.Elem()).Interface(), nil
}

func indirectInternal(n int, v reflect.Value) reflect.Value {
	for i := 0; i < n; i++ {
		p := reflect.New(reflect.PointerTo(v.Type())).Elem()
		p.Set(v.Addr())
		v = p
	}
	return v
}

// CountIndirectionsBetweenTypes returns the number of indirections between t1 and t2, where t1 points through some level of indirection to t2
func CountIndirectionsBetweenTypes(t1 reflect.Type, t2 reflect.Type) (int, error) {
	ct1, it1 := FollowPointersToConcreteType(t1)
	ct2, it2 := FollowPointersToConcreteType(t2)
	if ct1 != ct2 || it2 >= it1 {
		return 0, errors.New("t1 and t2 does not eventually point to t2")
	}
	return it1 - it2, nil
}

func CountIndirectionsBetweenTypesUnsafely(t1 reflect.Type, t2 reflect.Type) int {
	t := t1
	i := 0
	for t != t2 {
		t = t.Elem()
		i++
	}
	return i
}

func TypeImplementsInterface[T any](t reflect.Type) bool {
	interfaceType := reflect.TypeOf((*T)(nil)).Elem()
	return t.Implements(interfaceType) || reflect.PointerTo(t).Implements(interfaceType)
}

func AsInterface[T any](v reflect.Value) (T, error) {
	t := v.Type()
	interfaceType := reflect.TypeOf((*T)(nil)).Elem()
	if t.Implements(interfaceType) {
		if v.Kind() == reflect.Pointer && v.IsNil() {
			if !v.CanSet() {
				var z T
				return z, errors.New("un-settable value given")
			}
			v.Set(reflect.New(v.Type().Elem()))
		}
		return v.Interface().(T), nil
	}
	if reflect.PointerTo(t).Implements(interfaceType) {
		if !v.CanAddr() {
			var z T
			return z, errors.New("v must be addressable")
		}
		return v.Addr().Interface().(T), nil
	}
	var z T
	return z, errors.New("The given value does not implement the given interface")
}

func EnsureConcreteValue(v reflect.Value) (reflect.Value, error) {
	if !v.CanSet() {
		return reflect.Value{}, errors.New("un-settable value given")
	}
	cv := FollowPointersToConcreteValue(v)
	reflected := reflect.ValueOf(cv)
	if cv == nil {
		ct, indirections := FollowPointersToConcreteType(v.Type())
		reflected = reflect.New(ct).Elem()
		indirect := indirectInternal(indirections, reflected)
		v.Set(indirect)
	}
	return reflected, nil
}

type Field struct {
	Name  string
	Value any
}

func GetFields[T any](v T) []Field {
	var fields []Field
	rt, _ := FollowPointersToConcreteType(reflect.TypeOf(v))
	rv := reflect.ValueOf(FollowPointersToConcreteValue(reflect.ValueOf(v)))
	numFields := rt.NumField()
	for i := 0; i < numFields; i++ {
		field := rv.Field(i)
		if !field.CanInterface() {
			continue
		}
		name := rt.Field(i).Name
		fields = append(fields, Field{Name: name, Value: rv.FieldByName(name).Interface()})
	}
	return fields
}
