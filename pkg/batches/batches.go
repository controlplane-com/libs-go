package batches

import (
	"github.com/controlplane-com/libs-go/pkg/common"
)

// BatchSlice returns a batch cut out of the given slice, and the index of the next batch
func NewBatchIterator[S ~[]T, T any](slice S, batchSize int) common.Iterator[S] {
	return &batchIterator[S, T]{
		slice:     slice,
		index:     0,
		batchSize: batchSize,
	}
}

type batchIterator[S ~[]T, T any] struct {
	slice     S
	nextSlice S
	err       error
	index     int
	batchSize int
}

func (b *batchIterator[S, T]) Next() bool {
	if b.err != nil {
		return false
	}
	newSlice, newIndex := BatchSlice[S, T](b.slice, b.batchSize, b.index)
	b.index = newIndex
	b.nextSlice = newSlice
	return len(newSlice) > 0
}

func (b *batchIterator[S, T]) Item() S {
	return b.nextSlice
}

func (b *batchIterator[S, T]) Error() error {
	return b.err
}

func BatchSlice[S ~[]T, T any](slice S, batchSize, index int) (S, int) {
	lenSlice := len(slice)
	if index >= lenSlice || index < 0 {
		return nil, -1
	}
	endIndex := index + batchSize
	if endIndex >= lenSlice {
		return slice[index:], -1
	}
	return slice[index:endIndex], endIndex
}
