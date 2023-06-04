// Contains constructors for iterators
package iter

// Slice iterator implementation
type sliceIterator[T any] struct {
	slice  []T
	len    int
	cursor int
}

func (it *sliceIterator[T]) Next() (T, error) {
	var value T
	if it.cursor >= it.len {
		return value, ErrStopIt
	}

	value = it.slice[it.cursor]
	it.cursor++
	return value, nil
}

// Function FromSlice creates a slice iterator
// Iterator will get all elements of the slice in order
func FromSlice[T any](slice []T) Iterator[T] {
	return &sliceIterator[T]{
		slice:  slice,
		len:    len(slice),
		cursor: 0,
	}
}
