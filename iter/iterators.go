// Contains constructors for iterators
package iter

// Slice iterator implementation
type sliceIterator[T any] struct {
	slice  []T
	len    int
	cursor int
}

func (it *sliceIterator[T]) Next() bool {
	it.cursor++
	return it.cursor < len(it.slice)
}

func (it *sliceIterator[T]) Get() (T, error) {
	var value T
	if it.cursor >= it.len {
		return value, ErrStopIt
	}

	value = it.slice[it.cursor]
	return value, nil
}

// Function FromSlice creates a slice iterator
// Iterator will get all elements of the slice in order
func FromSlice[T any](slice []T) Iterator[T] {
	return &sliceIterator[T]{
		slice:  slice,
		len:    len(slice),
		cursor: -1,
	}
}
