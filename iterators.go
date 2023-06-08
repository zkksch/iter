// Contains constructors for iterators
package iter

import (
	"context"
	"sync/atomic"
)

// Slice iterator implementation
type sliceIterator[T any] struct {
	source []T
	len    int
	cursor int
}

func (it *sliceIterator[T]) Next() (T, error) {
	var value T
	if it.cursor >= it.len || it.cursor < 0 {
		return value, ErrStopIt
	}

	value = it.source[it.cursor]
	it.cursor++
	return value, nil
}

// Slice iterator implementation (thread safe)
type safeSliceIterator[T any] struct {
	source []T
	len    int64
	cursor *atomic.Int64
}

func (it *safeSliceIterator[T]) Next() (T, error) {
	cursor := it.cursor.Add(1) - 1
	if cursor >= it.len || cursor < 0 {
		var empty T
		it.cursor.Store(-1)
		return empty, ErrStopIt
	}

	return it.source[cursor], nil
}

// Function FromSlice creates a slice iterator
// Iterator will get all elements of the slice in order
func FromSlice[T any](slice []T) Iterator[T] {
	return &sliceIterator[T]{
		source: slice,
		len:    len(slice),
		cursor: 0,
	}
}

// Function FromSliceSafe creates a thread safe slice iterator
// Iterator will get all elements of the slice in order
func FromSliceSafe[T any](slice []T) Iterator[T] {
	return &safeSliceIterator[T]{
		source: slice,
		len:    int64(len(slice)),
		cursor: &atomic.Int64{},
	}
}

// Chan iterator implementation (thread safe)
type chanIterator[T any] struct {
	source chan T
	ctx    context.Context
}

func (it *chanIterator[T]) Next() (T, error) {
	select {
	case value, more := <-it.source:
		if !more {
			return value, ErrStopIt
		}
		return value, nil
	case <-it.ctx.Done():
		var empty T
		return empty, ErrStopIt
	}
}

// Function FromChan creates a thread safe iterator from channel
// Iterator will return values recieved from the channel
// and will stop when channel is closed
func FromChan[T any](ctx context.Context, source chan T) Iterator[T] {
	return &chanIterator[T]{
		source: source,
		ctx:    ctx,
	}
}
