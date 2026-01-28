package common

type Iterator[T any] interface {
	Next() bool
	Item() T
	Error() error
}
