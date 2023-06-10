// Contains constructors for iterators
package iter

import (
	"context"
	"sync/atomic"
)

// Function FromSlice creates a slice iterator
// Iterator returns all elements of the slice
func FromSlice[T any](source []T) Iterator[T] {
	cursor := 0
	return func() (T, error) {
		if cursor >= len(source) {
			var empty T
			return empty, ErrStopIt
		}
		value := source[cursor]
		cursor++
		return value, nil
	}
}

// Function FromSliceSafe creates a thread safe slice iterator
// Iterator returns all elements of the slice
func FromSliceSafe[T any](source []T) Iterator[T] {
	cursor := &atomic.Int64{}
	l := int64(len(source))
	return func() (T, error) {
		c := cursor.Add(1) - 1
		if c >= l {
			var empty T
			cursor.Store(l)
			return empty, ErrStopIt
		}

		return source[c], nil
	}
}

// Function FromChan creates a thread safe iterator from a channel
// Iterator returns values recieved from the channel
// and stops when channel is closed
func FromChan[T any](ctx context.Context, source chan T) Iterator[T] {
	return func() (T, error) {
		select {
		case value, more := <-source:
			if !more {
				return value, ErrStopIt
			}
			return value, nil
		case <-ctx.Done():
			var empty T
			return empty, ErrStopIt
		}
	}
}
