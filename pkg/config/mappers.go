package config

import (
	"github.com/controlplane-com/libs-go/pkg/scanner"
)

// SliceMapper is an alias for scanner.SliceMapper for backwards compatibility
type SliceMapper = scanner.SliceMapper

func defaultMappers() []scanner.Mapper {
	return []scanner.Mapper{
		scanner.JsonMapper{},
	}
}
