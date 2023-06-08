// Contains finalizers for iterators
package iter

import (
	"context"
	"errors"
)

// Function Reduce takes an iterator, initial value and a function,
// after that it initializes accumulating value with init value,
// and for each element in an iterator executes function from element value and
// accumulating value that should return a new accumulating value
func Reduce[T, K any](it Iterator[T], init K, fn func(T, K) K) (K, error) {
	acc := init
	var (
		v   T
		err error
	)
	for v, err = it.Next(); err == nil; v, err = it.Next() {
		acc = fn(v, acc)
	}
	if !errors.Is(err, ErrStopIt) {
		var empty K
		return empty, err
	}
	return acc, nil
}

// Function ToSlice makes a slice from elements of an iterator
func ToSlice[T any](it Iterator[T]) ([]T, error) {
	slice := make([]T, 0)
	var (
		v   T
		err error
	)
	for v, err = it.Next(); err == nil; v, err = it.Next() {
		slice = append(slice, v)
	}
	if !errors.Is(err, ErrStopIt) {
		return nil, err
	}
	return slice, nil
}

// Function ToChanSimple makes a channel that will send values from the iterator
// Ignores errors returned from the iterator
func ToChan[T any](ctx context.Context, it Iterator[T]) <-chan T {
	c := make(chan T)
	go func() {
		defer close(c)
		for v, err := it.Next(); err == nil; v, err = it.Next() {
			select {
			case c <- v:
			case <-ctx.Done():
				return
			}
		}
	}()
	return c
}

// Final iterator, does not implement Iterator interface
// Only needed to make using iterators in for loop easier
type finalIterator[T any] struct {
	base Iterator[T]
	last T
	err  error
}

// Method Next switches iterator to the next element
// returns false if iterator stopped and true otherwise
func (it *finalIterator[T]) Next() bool {
	if it.err != nil {
		return false
	}
	it.last, it.err = it.base.Next()
	return it.err == nil
}

// Method Get returns current value of iterator
func (it *finalIterator[T]) Get() T {
	return it.last
}

// Method Stop returns error that caused iterator to stop
// nil if iterator still active
func (it *finalIterator[T]) Stop() error {
	return it.err
}

// Method Err returns error that caused iterator to stop unexpectedly
// nil if iterator stopped because stop iteration error or if it's still active
func (it *finalIterator[T]) Err() error {
	err := it.Stop()
	if errors.Is(err, ErrStopIt) {
		return nil
	}
	return err
}

// Function Final returns final iterator to use it in for loop
func Final[T any](it Iterator[T]) *finalIterator[T] {
	return &finalIterator[T]{
		base: it,
	}
}
