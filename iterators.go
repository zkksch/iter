// Contains constructors for iterators
package iter

import "sync/atomic"

// Slice iterator implementation
type sliceIterator[T any] struct {
	slice  []T
	len    int
	cursor int
}

func (it *sliceIterator[T]) Next() (T, error) {
	var value T
	if it.cursor >= it.len || it.cursor < 0 {
		return value, ErrStopIt
	}

	value = it.slice[it.cursor]
	it.cursor++
	return value, nil
}

// Slice iterator implementation (thread safe)
type safeSliceIterator[T any] struct {
	slice   []T
	len     int64
	cursor  *atomic.Int64
	stopped *atomic.Bool
}

func (it *safeSliceIterator[T]) Next() (T, error) {
	var empty T
	if it.stopped.Load() {
		return empty, ErrStopIt
	}

	cursor := it.cursor.Add(1) - 1
	if cursor >= it.len || cursor < 0 {
		it.stopped.Store(true)
		return empty, ErrStopIt
	}

	return it.slice[cursor], nil
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

// Function FromSliceSafe creates a thread safe slice iterator
// Iterator will get all elements of the slice in order
func FromSliceSafe[T any](slice []T) Iterator[T] {
	return &safeSliceIterator[T]{
		slice:   slice,
		len:     int64(len(slice)),
		cursor:  &atomic.Int64{},
		stopped: &atomic.Bool{},
	}
}
