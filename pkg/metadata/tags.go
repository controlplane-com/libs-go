package metadata

import (
	"reflect"
	"strings"
)

const CplnTag = "cpln"

type ParsedTag map[string]string

func ParseCplnTag(rType reflect.Type, fieldName string) ParsedTag {
	if f, ok := rType.FieldByName(fieldName); ok {
		if t := f.Tag.Get(CplnTag); t != "" {
			return ParseCplnTagString(t)
		}
	}
	return nil
}

func ParseCplnTagString(t string) ParsedTag {
	parsed := ParsedTag{}
	for _, tagValue := range strings.Split(t, ";") {
		i := strings.Index(tagValue, ":")
		if i <= 0 {
			parsed[tagValue] = ""
			continue
		}
		parsed[strings.ToLower(tagValue[:i])] = tagValue[i+1:]
	}
	return parsed
}
