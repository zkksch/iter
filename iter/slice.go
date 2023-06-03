// Helpers to work with Go slices
package iter

type sliceIterator[T any] struct {
	slice   []T
	started bool
	cursor  int
}

func (it *sliceIterator[T]) Next() bool {
	if it.started {
		it.cursor++
	} else {
		it.started = true
	}

	return it.cursor < len(it.slice)
}

func (it *sliceIterator[T]) Get() (T, error) {
	var value T
	if it.cursor >= len(it.slice) {
		return value, ErrStopIt
	}

	value = it.slice[it.cursor]
	return value, nil
}

// Iter creates an iterator based on slice
func Iter[T any](slice []T) Iterator[T] {
	return &sliceIterator[T]{
		slice: slice,
	}
}

// Finalize finalizer creates a slice from iterator values
func Finalize[T any](it Iterator[T]) ([]T, error) {
	slice := make([]T, 0)
	for it.Next() {
		v, err := it.Get()
		if err != nil {
			return nil, err
		}
		slice = append(slice, v)
	}

	return slice, nil
}
