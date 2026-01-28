package pipeline

import (
	"strconv"
	"strings"
)

type Mapper[In any, Out any] func(i In) (Out, error)

func Map[In any, Out any](in []In, d Mapper[In, Out]) ([]Out, error) {
	if in == nil {
		return []Out{}, nil
	}
	out := []Out{}
	for _, i := range in {
		o, err := d(i)
		if err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, nil
}

func MustMap[In any, Out any](in []In, d func(i In) Out) []Out {
	if in == nil {
		return []Out{}
	}
	out := []Out{}
	for _, i := range in {
		o := d(i)
		out = append(out, o)
	}
	return out
}

func Filter[In any](in []In, d func(i In) (bool, error)) ([]In, error) {
	if in == nil {
		return []In{}, nil
	}
	filtered := []In{}
	for _, i := range in {
		include, err := d(i)
		if err != nil {
			return nil, err
		}
		if !include {
			continue
		}
		filtered = append(filtered, i)
	}
	return filtered, nil
}

func MustFilter[In any](in []In, d func(i In) bool) []In {
	if in == nil {
		return []In{}
	}
	filtered := []In{}
	for _, i := range in {
		include := d(i)
		if !include {
			continue
		}
		filtered = append(filtered, i)
	}
	return filtered
}

func ExtractMapValues[TKey comparable, TValue any](m map[TKey]TValue) []TValue {
	values := make([]TValue, len(m))
	i := 0
	for _, v := range m {
		values[i] = v
		i++
	}
	return values
}

func Flatten[TValue any](in [][]TValue) []TValue {
	var values []TValue
	for _, v := range in {
		values = append(values, v...)
	}
	return values
}

func JoinToCamelCase(stringSlice []string) string {
	if len(stringSlice) == 0 {
		return ""
	}
	first := stringSlice[0]
	camelCase, _ := Map(stringSlice, func(s string) (string, error) {
		return strings.ToUpper(strconv.Itoa(int(s[0]))) + s[1:], nil
	})
	return strings.Join(append([]string{first}, camelCase...), "")
}

func IndexOf[T comparable](collection []T, item T) int {
	for k, i := range collection {
		if i == item {
			return k
		}
	}
	return -1
}

type Enumerable[T comparable] []T

func (e Enumerable[T]) Slice() []T {
	return e
}

func (e Enumerable[T]) Filter(d func(i T) bool) Enumerable[T] {
	return MustFilter[T](e, d)
}

func (e Enumerable[T]) IndexOf(item T) int {
	return IndexOf(e, item)
}

func Difference[T comparable](first []T, second []T) []T {
	if first == nil || second == nil {
		return []T{}
	}

	result := make([]T, 0)
	secondMap := make(map[T]struct{}, len(second))

	for _, item := range second {
		secondMap[item] = struct{}{}
	}

	for _, item := range first {
		if _, exists := secondMap[item]; !exists {
			result = append(result, item)
		}
	}

	return result
}
