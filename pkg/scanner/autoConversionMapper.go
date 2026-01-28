package scanner

import (
	"github.com/controlplane-com/libs-go/pkg/types"
	"reflect"
)

type AutoConversionMapper struct {
	c ConcreteMapper
}

func (a *AutoConversionMapper) createArray(v reflect.Value, t reflect.Type) reflect.Value {
	array := reflect.New(reflect.ArrayOf(1, t)).Elem()
	array.Index(0).Set(v)
	return array
}

func (a *AutoConversionMapper) Map(destinationValue any, scannedValue any) (any, error) {
	v, _ := a.c.Map(destinationValue, scannedValue)
	dt, _ := types.FollowPointersToConcreteType(reflect.TypeOf(destinationValue))
	st := reflect.TypeOf(v)
	sv := reflect.ValueOf(v)
	if !st.AssignableTo(dt) {
		if dt.Kind() == reflect.Array && st.Kind() != reflect.Array {
			array := a.createArray(sv, st)
			return array.Interface(), nil
		}
		if dt.Kind() == reflect.Slice && st.Kind() != reflect.Slice {
			if st.Kind() != reflect.Array {
				sv = a.createArray(sv, st)
			}
			return sv.Slice(0, sv.Len()).Interface(), nil
		}
		rv := reflect.ValueOf(v)
		if !rv.CanConvert(dt) {
			return scannedValue, nil
		}
		return rv.Convert(dt).Interface(), nil
	}
	return v, nil
}
