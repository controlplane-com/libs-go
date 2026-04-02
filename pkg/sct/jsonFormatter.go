package sct

import "encoding/json"

type JsonFormatter[T any] struct {
}

func (j JsonFormatter[T]) Marshal(t *Token[T]) ([]byte, error) {
	return json.Marshal(t)
}

func (j JsonFormatter[T]) Unmarshal(b []byte, t *Token[T]) error {
	return json.Unmarshal(b, t)
}
