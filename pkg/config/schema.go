package config

import (
	"errors"
	"fmt"
	dynamic_objects "github.com/controlplane-com/libs-go/pkg/dynamic-objects"
	"github.com/controlplane-com/libs-go/pkg/metadata"
	"github.com/controlplane-com/libs-go/pkg/scanner"
	"github.com/controlplane-com/libs-go/pkg/types"
	"github.com/iancoleman/strcase"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Entry struct {
	Name    string `json:"name"`
	Default any
}

type schemaMapper struct {
}

func (s schemaMapper) Map(dest any, scannedValue any) (any, error) {
	scannedString := scannedValue.(string)
	switch dest.(type) {
	case time.Duration:
		return time.ParseDuration(scannedString)
	case time.Time:
		return time.Parse(time.RFC3339, scannedString)
	case int:
		return strconv.Atoi(scannedString)
	case int8:
		return strconv.ParseInt(scannedString, 10, 8)
	case int16:
		v, err := strconv.ParseInt(scannedString, 10, 16)
		return int16(v), err
	case int32:
		v, err := strconv.ParseInt(scannedString, 10, 32)
		return int32(v), err
	case int64:
		return strconv.ParseInt(scannedString, 10, 64)
	case uint:
		return strconv.Atoi(scannedString)
	case uint8:
		v, err := strconv.ParseUint(scannedString, 10, 8)
		return uint8(v), err
	case uint16:
		v, err := strconv.ParseUint(scannedString, 10, 16)
		return uint16(v), err
	case uint32:
		v, err := strconv.ParseUint(scannedString, 10, 32)
		return uint16(v), err
	case uint64:
		return strconv.ParseUint(scannedString, 10, 64)
	case float32:
		v, err := strconv.ParseFloat(scannedString, 32)
		return float32(v), err
	case float64:
		return strconv.ParseFloat(scannedString, 64)
	case bool:
		return scannedString == "1" || strings.ToLower(scannedString) == "true", nil
	case string:
		return scannedString, nil
	case regexp.Regexp:
		return regexp.Compile(scannedString)
	}
	return nil, errors.New(fmt.Sprintf("Unsupported type %s", reflect.TypeOf(dest).Name()))
}

func backToString(val any) string {
	switch v := val.(type) {
	case time.Time:
		return v.Format(time.RFC3339)
	case int:
		return strconv.Itoa(v)
	case int8:
		return strconv.Itoa(int(v))
	case int16:
		return strconv.Itoa(int(v))
	case int32:
		return strconv.Itoa(int(v))
	case int64:
		return strconv.Itoa(int(v))
	case uint:
		return strconv.Itoa(int(v))
	case uint8:
		return strconv.Itoa(int(v))
	case uint16:
		return strconv.Itoa(int(v))
	case uint32:
		return strconv.Itoa(int(v))
	case uint64:
		return strconv.Itoa(int(v))
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		if v {
			return "true"
		}
		return "false"
	case string:
		return v
	case []string:
		return strings.Join(v, ", ")
	case fmt.Stringer:
		return v.String()
	default:
		return ""
	}
}

func ParseSchema(schema any) error {
	st := scanner.NewScanTarget(schema)
	st.AddAllFields(schemaMapper{})
	v := st.GetReflectedModel()
	t := v.Type()
	for i := 0; i < v.Type().NumField(); i++ {
		f := t.Field(i)
		defaultValue := ""
		envName := strcase.ToScreamingSnake(f.Name)
		if tag, ok := f.Tag.Lookup("cpln"); ok {
			parsedTag := metadata.ParseCplnTagString(tag)
			if d, ok := parsedTag["default"]; ok {
				defaultValue = d
			}
			if e, ok := parsedTag["env"]; ok {
				envName = e
			}
		}
		envValue, ok := os.LookupEnv(envName)
		if !ok {
			envValue = defaultValue
		}
		if err := st.SetFieldValue(f.Name, envValue); err != nil {
			return err
		}
	}
	return nil
}

var config = map[string]any{}

func Initialize[T any](schema *T, mappers ...scanner.Mapper) error {
	if schema == nil {
		return errors.New("unexpected nil argument for parameter schema")
	}
	typeName := reflect.TypeOf(schema).Elem().Name()
	c, ok := config[typeName]
	if ok && c != nil {
		return nil
	}
	mappers = append(mappers, defaultMappers()...)
	for _, m := range mappers {
		t, _ := types.FollowPointersToConcreteType(reflect.TypeOf(m))
		scanner.Mappers.RegisterMapper(t.Name(), m)
	}
	err := ParseSchema(schema)
	if err == nil {
		config[typeName] = schema
	}
	return err
}

func GetConfig[T any]() *T {
	typeName := reflect.TypeOf((*T)(nil)).Elem().Name()
	c, ok := config[typeName]
	if !ok {
		return nil
	}
	if t, ok := c.(*T); ok {
		return t
	}
	return nil
}

func Summarize(schema any) string {
	st := scanner.NewScanTarget(schema)
	st.AddAllFields(schemaMapper{})
	v := st.GetReflectedModel()
	t := v.Type()
	b := strings.Builder{}
	b.Grow(t.NumField() * 75)
	longestField := 0
	stringValues := map[string]string{}
	longestStringValue := 0
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		name := f.Name
		lenName := len(name)
		if longestField < lenName {
			longestField = lenName
		}
		var s string
		var sensitive bool
		if tag, ok := f.Tag.Lookup("cpln"); ok {
			parsedTag := metadata.ParseCplnTagString(tag)
			if _, ok := parsedTag["sensitive"]; ok {
				sensitive = true
			}
		}
		if sensitive {
			s = "************"
		} else if f.Type.AssignableTo(reflect.TypeOf((*dynamic_objects.Stringable)(nil)).Elem()) {
			fieldValue := v.Field(i).Interface()
			if reflect.ValueOf(fieldValue).Kind() == reflect.Pointer && reflect.ValueOf(fieldValue).IsNil() {
				s = ""
			} else {
				s = fieldValue.(fmt.Stringer).String()
			}
		} else {
			s = backToString(st.GetFieldValue(name))
		}
		lenS := len(s)
		if lenS > longestStringValue {
			longestStringValue = lenS
		}
		stringValues[name] = s
	}

	for i := 0; i < t.NumField(); i++ {
		name := t.Field(i).Name
		b.WriteString(fmt.Sprintf("%-*s", longestField+1, name))
		b.WriteRune('|')
		b.WriteString(fmt.Sprintf("%*s", longestStringValue+1, stringValues[name]))
		b.WriteRune('\n')
	}
	return b.String()
}
