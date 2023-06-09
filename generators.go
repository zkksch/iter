// Contains constructors for iterators that generate values
package iter

import "sync/atomic"

// Function Sequence returns a sequence iterator
// Iterator will generate ints from start with a given step
func Sequence(start, step int) Iterator[int] {
	value := start - step
	return func() (int, error) {
		value += step
		return value, nil
	}
}

// Function SequenceSafe returns a thread safe sequence iterator
// Iterator will generate ints from start with a given step
func SequenceSafe(start, step int) Iterator[int] {
	value := &atomic.Int64{}
	value.Add(int64(start - step))
	step64 := int64(step)
	return func() (int, error) {
		v := value.Add(step64)
		return int(v), nil
	}
}

// Function Generator returns a thread safe generating iterator
// Iterator will generate values by using passed function
func Generator[T any](generator func() T) Iterator[T] {
	return func() (T, error) {
		return generator(), nil
	}
}

// Function Repeat returns a thread safe repeating iterator
// Iterator will repeat passed value indefinitely
func Repeat[T any](value T) Iterator[T] {
	return func() (T, error) {
		return value, nil
	}
}
