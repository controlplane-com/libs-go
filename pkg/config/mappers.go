package config

import (
	"errors"
	"github.com/controlplane-com/libs-go/pkg/scanner"
	"reflect"
	"strings"
)

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

func defaultMappers() []scanner.Mapper {
	return []scanner.Mapper{
		scanner.JsonMapper{},
		&SliceMapper{},
	}
}
