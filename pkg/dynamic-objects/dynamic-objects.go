package dynamic_objects

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/controlplane-com/libs-go/pkg/types"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type Stringable interface {
	String() string
}

type pivotedField struct {
	Path []string
	Field
}

func Pivot(obj any) [][]string {
	var paths [][]string
	for _, pivoted := range pivotInternal(GetAllFields(obj), nil) {
		paths = append(paths, pivoted.Path)
	}
	return paths
}

func PivotToMap(obj any, delimiter string) map[string]any {
	pivotedMap := map[string]any{}
	for _, pivoted := range pivotInternal(GetAllFields(obj), nil) {
		pivotedMap[strings.Join(pivoted.Path, delimiter)] = pivoted.Value
	}
	return pivotedMap
}

func ObjectToString(obj any) string {
	b := strings.Builder{}
	pivoted := pivotInternal(GetAllFields(obj), nil)
	for i, f := range pivoted {
		if i != 0 {
			b.WriteByte('|')
		}
		var value string
		switch v := f.Value.(type) {
		case string:
			value = v
			break
		case Stringable:
			value = v.String()
			break
		case bool:
			value = strconv.FormatBool(v)
			break
		case int:
			value = strconv.Itoa(v)
		case float64:
			value = strconv.FormatFloat(v, 'G', -1, 64)
		case float32:
			value = strconv.FormatFloat(float64(v), 'G', -1, 32)
		}
		b.WriteString(strings.Join(append(f.Path, value), "."))
	}
	return b.String()
}

func PivotAndJoinPaths(obj any, delimiter string) []string {
	pivotedFields := pivotInternal(GetAllFields(obj), nil)
	var paths []string
	for _, pivoted := range pivotedFields {
		paths = append(paths, strings.Join(pivoted.Path, delimiter))
	}
	return paths
}

func pivotInternal(fields []*Field, pathSoFar []string) []*pivotedField {
	var pivotedFields []*pivotedField
	for _, f := range fields {
		keyPath := make([]string, len(pathSoFar))
		copy(keyPath, pathSoFar)
		if f.Children == nil {
			pivoted := &pivotedField{
				Path:  append(keyPath, f.Name),
				Field: *f,
			}
			pivotedFields = append(pivotedFields, pivoted)
			continue
		}
		pivotedFields = append(pivotedFields, pivotInternal(f.Children, append(pathSoFar, f.Name))...)
	}
	return pivotedFields
}

type Field struct {
	Name           string
	Value          any
	ReflectedValue reflect.Value
	ReflectedType  reflect.Type
	Children       []*Field
}

func GetAllFields(obj any) []*Field {
	fields := GetFields(obj)
	for _, f := range fields {
		ct, _ := types.FollowPointersToConcreteType(f.ReflectedType)
		switch ct.Kind() {
		case reflect.Interface:
			concreteValue := types.FollowPointersToConcreteValue(f.ReflectedValue)
			if reflect.ValueOf(concreteValue).IsValid() {
				f.Children = GetAllFields(concreteValue)
			}
		case reflect.Struct, reflect.Map:
			f.Children = GetAllFields(f.Value)
		}
	}
	return fields
}

func GetFields(obj any) []*Field {
	var fields []*Field
	concreteValue := reflect.ValueOf(types.FollowPointersToConcreteValue(reflect.ValueOf(obj)))
	concreteType := concreteValue.Type()
	switch concreteType.Kind() {
	case reflect.Struct:
		numFields := concreteType.NumField()
		for i := 0; i < numFields; i++ {
			fieldReflectedValue := reflect.ValueOf(types.FollowPointersToConcreteValue(concreteValue.Field(i)))
			fieldReflectedType, _ := types.FollowPointersToConcreteType(concreteType.Field(i).Type)

			var value any = nil
			if fieldReflectedValue.IsValid() {
				value = fieldReflectedValue.Interface()
			}

			fields = append(fields, &Field{
				Name:           concreteType.Field(i).Name,
				Value:          value,
				ReflectedValue: fieldReflectedValue,
				ReflectedType:  fieldReflectedType,
			})
		}
		break
	case reflect.Map:
		keys := concreteValue.MapKeys()
		keyType := concreteType.Key()
		for index, key := range keys {
			var stringKey string
			if keyType.Kind() == reflect.String {
				stringKey = key.String()
			} else if types.TypeImplementsInterface[Stringable](key.Type()) {
				stringKey = key.Interface().(Stringable).String()
			} else {
				stringKey = strconv.Itoa(index)
			}
			val := concreteValue.MapIndex(key)
			fields = append(fields, &Field{
				Name:           stringKey,
				Value:          val.Interface(),
				ReflectedValue: val,
				ReflectedType:  concreteType.Elem(),
			})
		}
		break
	}
	return fields
}

func GetPropertyByPath(obj any, path string, delimiter string) (any, error) {
	pathPieces := strings.Split(path, delimiter)
	lenPathPieces := len(pathPieces)
	var m map[string]any
	var next = obj
	for i := 0; i < lenPathPieces; i++ {
		var ok bool
		m, ok = next.(map[string]any)
		if ok {
			next, ok = m[pathPieces[i]]
			if !ok {
				return nil, errors.New(fmt.Sprintf("invalid path: %s", path))
			}
			continue
		}
		//It's not a map. Let's try to get the next value using reflection
		value := types.FollowPointersToConcreteValue(reflect.ValueOf(next))
		reflectedValue := reflect.ValueOf(value)
		if reflectedValue.Kind() != reflect.Struct {
			return nil, errors.New(fmt.Sprintf("invalid path: %s", path))
		}
		if v := reflect.ValueOf(value).FieldByName(pathPieces[i]); !v.IsZero() {
			next = v.Interface()
			continue
		}
		return nil, errors.New(fmt.Sprintf("invalid path: %s", path))
	}
	return next, nil
}

func SetPropertyByPath(obj map[string]any, path string, delimiter string, newProperty any, createIfMissing bool) {
	pathPieces := strings.Split(path, delimiter)
	currentObj := obj
	var ok bool
	lenPath := len(pathPieces)
	for i, p := range pathPieces {
		var nextMap map[string]any
		if i == lenPath-1 {
			currentObj[p] = newProperty
			return
		}

		if nextMap, ok = currentObj[p].(map[string]any); !ok {
			if !createIfMissing {
				return
			}
			nextMap = map[string]any{}
			currentObj[p] = nextMap
			currentObj = nextMap
			continue
		}
		currentObj = nextMap
	}
}

func IsSubset(subset any, superset any, matchValues bool) bool {
	if subset == nil || superset == nil {
		return false
	}

	subsetFields := PivotToMap(subset, ".")
	for path, subsetValue := range subsetFields {
		supersetValue, err := GetPropertyByPath(superset, path, ".")
		if err != nil {
			return false
		}
		if matchValues && supersetValue != subsetValue {
			return false
		}
	}
	return true
}

func toRegex(value any) (*regexp.Regexp, error) {
	r, ok := value.(*regexp.Regexp)
	if ok {
		return r, nil
	}
	s, ok := value.(string)
	if !ok {
		return nil, errors.New("value is not a string")
	}
	l := len(s)
	if l < 2 || s[0] != '/' || s[l-1] != '/' {
		return nil, errors.New(`value does not begin and end with "/"`)
	}
	return regexp.CompilePOSIX(s[1 : l-1])
}

type Matcher[O any] interface {
	Match(candidate any, matchValues bool) bool
	Owner() O
}

type matcher[O any] struct {
	fields    map[string]any
	delimiter string
	owner     O
}

func NewMatcher[O any](fields any, delimiter string, owner O) Matcher[O] {
	m, ok := fields.(map[string]any)
	if !ok {
		m = PivotToMap(fields, delimiter)
	}
	return &matcher[O]{
		fields:    m,
		delimiter: delimiter,
		owner:     owner,
	}
}

func (m *matcher[O]) Owner() O {
	return m.owner
}

func (m *matcher[O]) Match(candidate any, matchValues bool) bool {
	if candidate == nil || m.fields == nil {
		return false
	}
	var match map[string]any
	subsetFields := PivotToMap(candidate, m.delimiter)
	for path, subsetValue := range m.fields {
		supersetValue, ok := subsetFields[path]
		if !ok {
			return false
		}
		if !matchValues {
			continue
		}
		r, err := toRegex(subsetValue)
		if err == nil {
			m.fields[path] = r
			if s, ok := supersetValue.(string); !ok || !r.MatchString(s) {
				return false
			}
			if match == nil {
				match = map[string]any{}
			}
			continue
		}
		if subsetValue != supersetValue {
			return false
		}
		if match == nil {
			match = map[string]any{}
		}
		SetPropertyByPath(match, path, m.delimiter, supersetValue, true)
	}
	return true
}

func Extract(obj any, paths []string, delimiter string) (any, error) {
	subset := map[string]any{}
	for _, p := range paths {
		prop, err := GetPropertyByPath(obj, p, delimiter)
		if err != nil {
			return nil, err
		}
		SetPropertyByPath(subset, p, delimiter, prop, true)
	}
	return subset, nil
}

func Fingerprint(obj any, paths []string, delimiter string) (any, string, error) {
	subset, err := Extract(obj, paths, delimiter)
	if err != nil {
		return nil, "", err
	}
	b, err := json.Marshal(subset)
	if err != nil {
		return nil, "", err
	}
	return subset, string(b), nil
}
