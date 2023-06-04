// Contains constructors for iterators that generate values
package iter

// Sequence iterator implementation
type seqIterator struct {
	value, step int
}

func (it *seqIterator) Next() (int, error) {
	it.value += it.step
	return it.value, nil
}

// Function Sequence returns sequence iterator
// Iterator will generate ints from start with a given step
func Sequence(start, step int) Iterator[int] {
	return &seqIterator{
		value: start - step,
		step:  step,
	}
}

// Generating iterator implementation
type genIterator[T any] struct {
	generator func(T) (T, error)
	prev      T
	stopped   bool
}

func (it *genIterator[T]) Next() (T, error) {
	var empty T
	if it.stopped {
		return empty, ErrStopIt
	}

	value, err := it.generator(it.prev)
	if err != nil {
		it.stopped = true
		return empty, err
	}
	it.prev = value
	return value, nil
}

// Function Generate returns a generating iterator
// Iterator will generate values by using function from a previous generated value
// (will be equal to a type T default value for first call)
func Generate[T any](fn func(T) (T, error)) Iterator[T] {
	return &genIterator[T]{
		generator: fn,
	}
}

// Repeating iterator implementation
type repeatIterator[T any] struct {
	value T
}

func (it *repeatIterator[T]) Next() (T, error) {
	return it.value, nil
}

// Function Repeat returns a repeating iterator
// Iterator will repeat passed value indefinitely
func Repeat[T any](value T) Iterator[T] {
	return &repeatIterator[T]{
		value: value,
	}
}
