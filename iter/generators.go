// Contains constructors for iterators that generate values
package iter

// Sequence iterator implementation
type seqIterator struct {
	value, step int
	started     bool
}

func (it *seqIterator) Next() bool {
	if it.started {
		it.value += it.step
	} else {
		it.started = true
	}
	return true
}

func (it *seqIterator) Get() (int, error) {
	return it.value, nil
}

// Function Sequence returns sequence iterator
// Iterator will generate ints from start with a given step
func Sequence(start, step int) Iterator[int] {
	return &seqIterator{
		value: start,
		step:  step,
	}
}

// Generating iterator implementation
type genIterator[T any] struct {
	generator func(T) (T, error)
	current   T
	err       error
}

func (it *genIterator[T]) Next() bool {
	if it.err != nil {
		return false
	}
	value, err := it.generator(it.current)
	if err != nil {
		var empty T
		it.current = empty
		it.err = err
		return false
	}
	it.current = value
	return true
}

func (it *genIterator[T]) Get() (T, error) {
	return it.current, it.err
}

// Function Generate returns a generating iterator
// Iterator will generate values by using function from a previous generated value
// (will be equal to a type T default value for first call)
func Generate[T any](fn func(T) (T, error)) Iterator[T] {
	return &genIterator[T]{
		generator: fn,
	}
}
