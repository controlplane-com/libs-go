package common

type Stack[T any] struct {
	items []T
}

func NewStack[T any]() Stack[T] {
	return Stack[T]{
		items: []T{},
	}
}

func (s *Stack[T]) Pop() T {
	lenItems := len(s.items)
	i := s.items[lenItems-1]
	s.items = s.items[:lenItems-1]
	return i
}

func (s *Stack[T]) Push(item T) {
	s.items = append(s.items, item)
}

func (s *Stack[T]) Length() int {
	return len(s.items)
}
