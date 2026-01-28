package scanner

import (
	"github.com/controlplane-com/libs-go/pkg/types"
	"reflect"
)

// ConcreteMapper converts nil scanned values into the zero value for the corresponding type
type ConcreteMapper struct {
}

func (s ConcreteMapper) Map(dest any, scannedValue any) (any, error) {
	r := reflect.ValueOf(scannedValue)
	cv := types.FollowPointersToConcreteValue(r)
	if cv == nil {
		ct, _ := types.FollowPointersToConcreteType(reflect.TypeOf(dest))
		return reflect.New(ct).Elem().Interface(), nil
	}
	return cv, nil
}
