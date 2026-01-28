package maps

func Merge[K comparable, V any](a map[K]V, b map[K]V, mergeFunc func(V, V) (V, error)) (map[K]V, error) {
	if a == nil && b == nil {
		return nil, nil
	}
	if a == nil {
		a = map[K]V{}
	}
	if b == nil {
		b = map[K]V{}
	}
	merged := map[K]V{}
	var err error
	for k, valueA := range a {
		if valueB, ok := b[k]; ok {
			if mergeFunc == nil {
				merged[k] = valueB
				continue
			}
			if merged[k], err = mergeFunc(valueA, valueB); err != nil {
				return nil, err
			}
			continue
		}
		merged[k] = valueA
	}
	for k, valueB := range b {
		if _, ok := a[k]; ok {
			continue
		}
		merged[k] = valueB
	}
	return merged, nil
}
