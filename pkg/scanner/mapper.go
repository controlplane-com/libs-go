package scanner

import (
	"errors"
	"reflect"
	"strings"
)

const MapperTagName = "mapper"

var Mappers = NewMapperCollection()

type MapperCollection struct {
	registeredMappers map[string]Mapper
}

func NewMapperCollection() *MapperCollection {
	mc := &MapperCollection{registeredMappers: map[string]Mapper{}}
	mc.RegisterMapper("ConcreteMapper", ConcreteMapper{})
	mc.RegisterMapper("JsonMapper", JsonMapper{})
	mc.RegisterMapper("SliceMapper", &SliceMapper{})
	return mc
}

// SliceMapper splits a delimited string into a slice
type SliceMapper struct {
	Delimiter string
}

func (a *SliceMapper) Map(dest any, scannedValue any) (any, error) {
	if a.Delimiter == "" {
		a.Delimiter = ","
	}
	if reflect.TypeOf(dest).Kind() != reflect.Slice {
		return nil, errors.New("dest must be a slice type")
	}
	switch v := scannedValue.(type) {
	case []string:
		return v, nil
	case string:
		if v == "" {
			return nil, nil
		}
		return strings.Split(v, a.Delimiter), nil
	default:
		return nil, errors.New("scanned value must be a string or a slice of strings")
	}
}

type Mapper interface {
	Map(dest any, scannedValue any) (any, error)
}

func (mc *MapperCollection) GetMapperByName(name string) Mapper {
	return mc.registeredMappers[name]
}

func (mc *MapperCollection) RegisterMapper(name string, m Mapper) {
	mc.registeredMappers[name] = m
}
