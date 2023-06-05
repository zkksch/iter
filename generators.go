// Contains constructors for iterators that generate values
package iter

import "sync/atomic"

// Sequence iterator implementation
type seqIterator struct {
	value, step int
}

func (it *seqIterator) Next() (int, error) {
	it.value += it.step
	return it.value, nil
}

// Sequence iterator implementation (thread safe)
type safeSeqIterator struct {
	value *atomic.Int64
	step  int64
}

func (it *safeSeqIterator) Next() (int, error) {
	value := it.value.Add(it.step)
	return int(value), nil
}

// Function Sequence returns a sequence iterator
// Iterator will generate ints from start with a given step
func Sequence(start, step int) Iterator[int] {
	return &seqIterator{
		value: start - step,
		step:  step,
	}
}

// Function SequenceSafe returns a thread safe sequence iterator
// Iterator will generate ints from start with a given step
func SequenceSafe(start, step int) Iterator[int] {
	v := &atomic.Int64{}
	v.Add(-int64(step))
	return &safeSeqIterator{
		value: v,
		step:  int64(step),
	}
}

// Generating iterator implementation (thread safe)
type genIterator[T any] struct {
	generator func() T
}

func (it *genIterator[T]) Next() (T, error) {
	return it.generator(), nil
}

// Function Generate returns a thread safe generating iterator
// Iterator will generate values by using passed function
func Generate[T any](fn func() T) Iterator[T] {
	return &genIterator[T]{
		generator: fn,
	}
}

// Repeating iterator implementation (thread safe)
type repeatIterator[T any] struct {
	value T
}

func (it *repeatIterator[T]) Next() (T, error) {
	return it.value, nil
}

// Function Repeat returns a thread safe repeating iterator
// Iterator will repeat passed value indefinitely
func Repeat[T any](value T) Iterator[T] {
	return &repeatIterator[T]{
		value: value,
	}
}
