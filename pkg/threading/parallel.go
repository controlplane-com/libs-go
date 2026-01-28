package threading

import (
	"github.com/controlplane-com/libs-go/pkg/errors"
	"sync"
)

func WaitGroup(copies int, worker func() error) error {
	wg := sync.WaitGroup{}
	wg.Add(copies)
	errs := &errors.CompoundError{}
	for i := 0; i < copies; i++ {
		go func() {
			err := worker()
			if err != nil {
				errs.Errs = append(errs.Errs, err)
			}
			wg.Done()
		}()
	}
	wg.Wait()

	if len(errs.Errs) == 0 {
		return nil
	}
	return errs
}

func Parallel[In any, Out any](items []In, worker func(arg In, index int) ([]Out, error)) ([]Out, error) {
	wg := sync.WaitGroup{}
	numItems := len(items)
	wg.Add(numItems)
	m := sync.Mutex{}
	var results = make([][]Out, numItems)
	var errs []error
	for index, item := range items {
		closureIndex := index
		closureItem := item
		go func() {
			defer wg.Done()
			out, err := worker(closureItem, closureIndex)
			m.Lock()
			defer m.Unlock()
			if err != nil {
				errs = append(errs, err)
				return
			}
			results[closureIndex] = out
		}()
	}
	wg.Wait()
	if len(errs) == 0 {
		var flatResults []Out
		for _, r := range results {
			flatResults = append(flatResults, r...)
		}
		return flatResults, nil
	}
	return nil, errors.NewCompoundError("one or more errors occurred", errs...)
}
