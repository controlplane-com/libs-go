package scanner

import (
	"encoding/json"
	"github.com/controlplane-com/libs-go/pkg/types"
	"reflect"
)

type JsonMapper struct {
	m ConcreteMapper
}

func (j JsonMapper) Map(destinationValue any, scannedValue any) (any, error) {
	mappedValue, _ := j.m.Map(destinationValue, scannedValue)

	ct, _ := types.FollowPointersToConcreteType(reflect.TypeOf(destinationValue))
	v := reflect.New(ct)
	pV := v.Interface()

	if s, ok := mappedValue.(string); ok {
		err := json.Unmarshal([]byte(s), pV)
		if err != nil {
			return nil, err
		}
	}
	return reflect.ValueOf(pV).Elem().Interface(), nil
}
