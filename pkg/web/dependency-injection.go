package web

import (
	"errors"
)

var (
	ContainerItemNotFoundErr error = errors.New("no matching item found")
)

type Container[T any] interface {
	Append(key string, item T)
	List() []T
	Get(key string) T
	Delete(key string)
}

type container[T any] struct {
	items map[string]T
}

func NewContainer[T any]() Container[T] {
	return &container[T]{
		items: map[string]T{},
	}
}

func (s *container[T]) Append(key string, item T) {
	s.items[key] = item
}

func (s *container[T]) List() []T {
	lenItems := len(s.items)
	values := make([]T, lenItems)
	i := 0
	for _, v := range s.items {
		values[i] = v
		i++
	}
	return values
}

func (s *container[T]) Delete(key string) {
	delete(s.items, key)
}

func (s *container[T]) Get(key string) T {
	return s.items[key]
}

func GetItem[DesiredType any, T any](c Container[T]) (DesiredType, error) {
	for _, s := range c.List() {
		var a any = s
		if itemOfDesiredType, ok := a.(DesiredType); ok {
			return itemOfDesiredType, nil
		}
	}

	var zeroValue DesiredType
	return zeroValue, ContainerItemNotFoundErr
}
